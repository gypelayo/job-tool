let port = null;
let frameData = [];
let extractionTimer = null;

chrome.action.onClicked.addListener(async (tab) => {
  console.log("=== Extracting from:", tab.url);
  
  const jobInfo = extractGreenhouseInfo(tab.url);
  
  if (jobInfo) {
    console.log("Greenhouse job detected:", jobInfo);
    await handleGreenhouseJob(tab, jobInfo);
  } else {
    console.log("Non-Greenhouse site - using scraping");
    await handleGenericScraping(tab);
  }
});

// ========== GREENHOUSE API HANDLING ==========

async function handleGreenhouseJob(tab, jobInfo) {
  if (jobInfo.boardToken && jobInfo.jobId) {
    // Direct URL - fetch immediately
    const jobData = await fetchGreenhouseJob(jobInfo.boardToken, jobInfo.jobId);
    if (jobData) {
      sendToHost(jobData);
    } else {
      console.error("API failed, falling back to scraping");
      await handleGenericScraping(tab);
    }
  } else if (jobInfo.jobId) {
    // Need to find board token from iframes
    chrome.webNavigation.getAllFrames({ tabId: tab.id }, async (frames) => {
      let boardToken = null;
      
      for (const frame of frames) {
        const match = frame.url.match(/greenhouse\.io.*[?&]for=([^&]+)/);
        if (match) {
          boardToken = match[1];
          console.log("Found board token:", boardToken);
          break;
        }
      }
      
      if (boardToken) {
        const jobData = await fetchGreenhouseJob(boardToken, jobInfo.jobId);
        if (jobData) {
          sendToHost(jobData);
        } else {
          console.error("API failed, falling back to scraping");
          await handleGenericScraping(tab);
        }
      } else {
        console.error("No board token found, falling back to scraping");
        await handleGenericScraping(tab);
      }
    });
  }
}

function extractGreenhouseInfo(url) {
  try {
    const urlObj = new URL(url);
    
    // Pattern 1: Direct Greenhouse URL
    const directMatch = url.match(/greenhouse\.io\/([^\/]+)\/jobs\/(\d+)/);
    if (directMatch) {
      return {
        boardToken: directMatch[1],
        jobId: directMatch[2],
        type: 'direct'
      };
    }
    
    // Pattern 2: Custom domain with gh_jid parameter
    const jobId = urlObj.searchParams.get('gh_jid');
    if (jobId) {
      return {
        boardToken: null,
        jobId: jobId,
        type: 'embedded'
      };
    }
    
    return null;
  } catch (e) {
    return null;
  }
}

async function fetchGreenhouseJob(boardToken, jobId) {
  const url = `https://boards-api.greenhouse.io/v1/boards/${boardToken}/jobs/${jobId}`;
  console.log("Fetching:", url);
  
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`API returned ${response.status}`);
    }
    
    const job = await response.json();
    console.log("✓ Got job:", job.title);
    
    const parser = new DOMParser();
    const doc = parser.parseFromString(job.content, 'text/html');
    const textContent = doc.body.textContent;
    
    const formatted = `
JOB TITLE: ${job.title}

LOCATION: ${job.location?.name || 'Not specified'}

DEPARTMENT: ${job.departments?.map(d => d.name).join(', ') || 'Not specified'}

DESCRIPTION:
${textContent}

URL: ${job.absolute_url}
UPDATED: ${job.updated_at}
SOURCE: Greenhouse API
`;
    
    console.log("Extracted", formatted.length, "chars");
    return formatted;
    
  } catch (error) {
    console.error("API error:", error);
    return null;
  }
}

// ========== GENERIC SCRAPING ==========

async function handleGenericScraping(tab) {
  frameData = [];
  clearTimeout(extractionTimer);
  
  // Inject content script and extract
  try {
    await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      files: ['content.js']
    });
    
    console.log("Content script injected, waiting for extraction...");
    
    // Wait a moment for injection
    setTimeout(() => {
      chrome.tabs.sendMessage(tab.id, { action: "extract" }).catch(err => {
        console.error("Send message failed:", err);
      });
    }, 500);
    
    // Timeout after 20 seconds
    extractionTimer = setTimeout(() => {
      console.log("Timeout - collected", frameData.length, "responses");
      combineAndSend();
    }, 20000);
    
  } catch (err) {
    console.error("Injection failed:", err);
  }
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extractText") {
    const data = request.data;
    
    console.log("✓ Received scraped content:", data.contentLength, "chars");
    
    frameData.push(data);
    
    // If we got substantial content, send it
    if (data.contentLength > 500) {
      clearTimeout(extractionTimer);
      setTimeout(() => combineAndSend(), 2000);
    }
    
    sendResponse({ received: true });
  }
});

function combineAndSend() {
  if (frameData.length === 0) {
    console.error("No data collected");
    return;
  }
  
  frameData.sort((a, b) => b.contentLength - a.contentLength);
  
  const combined = frameData.map((frame, index) => {
    return `
========== SCRAPED CONTENT ==========
URL: ${frame.url}
TITLE: ${frame.title}
LENGTH: ${frame.contentLength} chars
SOURCE: Web Scraping

${frame.text}
`;
  }).join("\n\n");
  
  console.log("Sending", combined.length, "chars");
  sendToHost(combined);
  
  frameData = [];
}

// ========== NATIVE HOST ==========

function sendToHost(text) {
  try {
    if (!port) {
      port = chrome.runtime.connectNative('com.textextractor.host');
      
      port.onMessage.addListener((msg) => {
        console.log("✓ Host response:", msg);
      });
      
      port.onDisconnect.addListener(() => {
        if (chrome.runtime.lastError) {
          console.error("✗", chrome.runtime.lastError.message);
        }
        port = null;
      });
    }
    
    port.postMessage({ text: text });
    console.log("✓ Sent to host");
  } catch (err) {
    console.error("✗ Host error:", err);
  }
}
