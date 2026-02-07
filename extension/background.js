let port = null;

chrome.action.onClicked.addListener(async (tab) => {
  console.log("=== Extracting from:", tab.url);
  
  const jobInfo = extractGreenhouseInfo(tab.url);
  
  if (jobInfo) {
    console.log("Greenhouse job detected:", jobInfo);
    
    if (jobInfo.boardToken && jobInfo.jobId) {
      // Direct URL - we have everything
      const jobData = await fetchGreenhouseJob(jobInfo.boardToken, jobInfo.jobId);
      if (jobData) {
        sendToHost(jobData);
      } else {
        console.error("Failed to fetch from API");
      }
    } else if (jobInfo.jobId) {
      // Has job ID but no board token - need to find it from iframes
      chrome.webNavigation.getAllFrames({ tabId: tab.id }, async (frames) => {
        let boardToken = null;
        
        for (const frame of frames) {
          const match = frame.url.match(/greenhouse\.io.*[?&]for=([^&]+)/);
          if (match) {
            boardToken = match[1];
            console.log("Found board token from iframe:", boardToken);
            break;
          }
        }
        
        if (boardToken) {
          const jobData = await fetchGreenhouseJob(boardToken, jobInfo.jobId);
          if (jobData) {
            sendToHost(jobData);
          }
        } else {
          console.error("Could not find board token");
        }
      });
    }
  } else {
    console.log("Not a Greenhouse URL, use normal extraction");
    // Add your fallback scraping logic here if needed
  }
});

function extractGreenhouseInfo(url) {
  try {
    const urlObj = new URL(url);
    
    // Pattern 1: Direct Greenhouse URL
    // https://job-boards.greenhouse.io/{board_token}/jobs/{job_id}
    const directMatch = url.match(/greenhouse\.io\/([^\/]+)\/jobs\/(\d+)/);
    if (directMatch) {
      return {
        boardToken: directMatch[1],
        jobId: directMatch[2],
        type: 'direct'
      };
    }
    
    // Pattern 2: Custom domain with gh_jid parameter
    // https://careers.company.com/?gh_jid=123456
    const jobId = urlObj.searchParams.get('gh_jid');
    if (jobId) {
      return {
        boardToken: null, // Will find from iframe
        jobId: jobId,
        type: 'embedded'
      };
    }
    
    return null;
  } catch (e) {
    console.error("URL parse error:", e);
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
    
    // Extract text from HTML content
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
`;
    
    console.log("Extracted", formatted.length, "chars");
    return formatted;
    
  } catch (error) {
    console.error("API error:", error);
    return null;
  }
}

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
