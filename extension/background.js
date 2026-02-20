let port = null;
let frameData = [];
let extractionTimer = null;

// In MV2 + Firefox, there is no chrome.action.
// If you want click-to-extract on icon click, use browserAction:
chrome.browserAction.onClicked.addListener(async (tab) => {
  console.log('=== Extracting from:', tab.url);

  const jobInfo = extractGreenhouseInfo(tab.url);

  if (jobInfo) {
    console.log('Greenhouse job detected:', jobInfo);
    await handleGreenhouseJob(tab, jobInfo);
  } else if (isWellfound(tab.url)) {
    console.log('Wellfound job detected');
    await handleWellfoundJob(tab);
  } else if (isRemoteRocketship(tab.url)) {
    console.log('Remote Rocketship job detected');
    await handleRemoteRocketshipJob(tab);
  } else if (isLinkedIn(tab.url)) {
    console.log('LinkedIn job detected');
    await handleLinkedInJob(tab);
  } else {
    console.log('Generic site - using basic scraping');
    await handleGenericScraping(tab);
  }
});

// ========== GREENHOUSE API HANDLING ==========

async function handleGreenhouseJob(tab, jobInfo) {
  if (jobInfo.boardToken && jobInfo.jobId) {
    const jobData = await fetchGreenhouseJob(jobInfo.boardToken, jobInfo.jobId);
    if (jobData) {
      sendToHost(jobData);
    } else {
      console.error('API failed');
    }
  } else if (jobInfo.jobId) {
    chrome.webNavigation.getAllFrames({ tabId: tab.id }, async (frames) => {
      let boardToken = null;

      for (const frame of frames) {
        const match = frame.url.match(/greenhouse\.io.*[?&]for=([^&]+)/);
        if (match) {
          boardToken = match[1];
          console.log('Found board token:', boardToken);
          break;
        }
      }

      if (boardToken) {
        const jobData = await fetchGreenhouseJob(boardToken, jobInfo.jobId);
        if (jobData) {
          sendToHost(jobData);
        }
      } else {
        console.error('No board token found');
      }
    });
  }
}

function extractGreenhouseInfo(url) {
  try {
    const urlObj = new URL(url);

    const directMatch = url.match(/greenhouse\.io\/([^/]+)\/jobs\/(\d+)/);
    if (directMatch) {
      return {
        boardToken: directMatch[1],
        jobId: directMatch[2],
        type: 'direct',
      };
    }

    const jobId = urlObj.searchParams.get('gh_jid');
    if (jobId) {
      return {
        boardToken: null,
        jobId: jobId,
        type: 'embedded',
      };
    }

    return null;
  } catch (e) {
    return null;
  }
}

async function fetchGreenhouseJob(boardToken, jobId) {
  const url = `https://boards-api.greenhouse.io/v1/boards/${boardToken}/jobs/${jobId}`;
  console.log('Fetching:', url);

  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`API returned ${response.status}`);
    }

    const job = await response.json();
    console.log('✓ Got job:', job.title);

    const parser = new DOMParser();
    const doc = parser.parseFromString(job.content, 'text/html');
    const textContent = doc.body.textContent;

    const formatted = `
JOB TITLE: ${job.title}

LOCATION: ${job.location?.name || 'Not specified'}

DEPARTMENT: ${job.departments?.map((d) => d.name).join(', ') || 'Not specified'}

DESCRIPTION:
${textContent}

URL: ${job.absolute_url}
UPDATED: ${job.updated_at}
SOURCE: Greenhouse API
`;

    console.log('Extracted', formatted.length, 'chars');
    return formatted;
  } catch (error) {
    console.error('API error:', error);
    return null;
  }
}

// ========== WELLFOUND HANDLING ==========

function isWellfound(url) {
  return url.includes('wellfound.com/jobs/') || url.includes('angel.co/jobs/');
}

async function handleWellfoundJob(tab) {
  frameData = [];
  clearTimeout(extractionTimer);

  try {
    // MV2 / Firefox: use tabs.executeScript to inject content.js
    chrome.tabs.executeScript(tab.id, { file: 'content.js' }, () => {
      if (chrome.runtime.lastError) {
        console.error('Injection failed:', chrome.runtime.lastError.message);
        return;
      }
      console.log('Content script injected');

      setTimeout(() => {
        chrome.tabs.sendMessage(tab.id, { action: 'extractWellfound' }).catch((err) => {
          console.error('Message failed:', err);
        });
      }, 500);

      extractionTimer = setTimeout(() => {
        console.log('Timeout - collected', frameData.length, 'responses');
        combineAndSend();
      }, 15000);
    });
  } catch (err) {
    console.error('Injection failed:', err);
  }
}

// ========== REMOTE ROCKETSHIP HANDLING ==========

function isRemoteRocketship(url) {
  return url.includes('remoterocketship.com/company/');
}

async function handleRemoteRocketshipJob(tab) {
  frameData = [];
  clearTimeout(extractionTimer);

  try {
    chrome.tabs.executeScript(tab.id, { file: 'content.js' }, () => {
      if (chrome.runtime.lastError) {
        console.error('Injection failed:', chrome.runtime.lastError.message);
        return;
      }
      console.log('Content script injected');

      setTimeout(() => {
        chrome.tabs.sendMessage(tab.id, { action: 'extractRemoteRocketship' }).catch((err) => {
          console.error('Message failed:', err);
        });
      }, 500);

      extractionTimer = setTimeout(() => {
        console.log('Timeout - collected', frameData.length, 'responses');
        combineAndSend();
      }, 15000);
    });
  } catch (err) {
    console.error('Injection failed:', err);
  }
}

// ========== LINKEDIN HANDLING ==========

function isLinkedIn(url) {
  return url.includes('linkedin.com/jobs/view/');
}

async function handleLinkedInJob(tab) {
  frameData = [];
  clearTimeout(extractionTimer);

  try {
    chrome.tabs.executeScript(tab.id, { file: 'content.js' }, () => {
      if (chrome.runtime.lastError) {
        console.error('Injection failed:', chrome.runtime.lastError.message);
        return;
      }
      console.log('Content script injected');

      setTimeout(() => {
        chrome.tabs.sendMessage(tab.id, { action: 'extractLinkedIn' }).catch((err) => {
          console.error('Message failed:', err);
        });
      }, 500);

      extractionTimer = setTimeout(() => {
        console.log('Timeout - collected', frameData.length, 'responses');
        combineAndSend();
      }, 15000);
    });
  } catch (err) {
    console.error('Injection failed:', err);
  }
}

// ========== GENERIC SCRAPING ==========

async function handleGenericScraping(tab) {
  frameData = [];
  clearTimeout(extractionTimer);

  try {
    chrome.tabs.executeScript(tab.id, { file: 'content.js' }, () => {
      if (chrome.runtime.lastError) {
        console.error('Injection failed:', chrome.runtime.lastError.message);
        return;
      }

      setTimeout(() => {
        chrome.tabs.sendMessage(tab.id, { action: 'extract' }).catch((err) => {
          console.error('Message failed:', err);
        });
      }, 500);

      extractionTimer = setTimeout(() => {
        console.log('Timeout');
        combineAndSend();
      }, 20000);
    });
  } catch (err) {
    console.error('Injection failed:', err);
  }
}

// ========== MESSAGE HANDLING ==========

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'extractText') {
    const data = request.data;

    console.log('✓ Received:', data.contentLength, 'chars');

    frameData.push(data);

    if (data.contentLength > 500) {
      clearTimeout(extractionTimer);
      setTimeout(() => combineAndSend(), 2000);
    }

    sendResponse({ received: true });
  }
});

function combineAndSend() {
  if (frameData.length === 0) {
    console.error('No data collected');
    return;
  }

  const data = frameData[0];
  console.log('Sending', data.contentLength, 'chars');
  sendToHost(data.text);

  frameData = [];
}

// ========== NATIVE HOST ==========

async function sendToHost(text) {
  const storage = chrome.storage; // Firefox supports chrome.* alias

  const defaults = {
    provider: 'ollama',
    ollamaModel: 'qwen2.5:7b',
    perplexityKey: '',
    perplexityModel: 'sonar-pro',
  };

  storage.sync.get(defaults, (settings) => {
    try {
      if (!port) {
        port = chrome.runtime.connectNative('com.textextractor.host');

        port.onMessage.addListener((msg) => {
          console.log('✓ Host response:', msg);
        });

        port.onDisconnect.addListener(() => {
          if (chrome.runtime.lastError) {
            console.error('✗', chrome.runtime.lastError.message);
          }
          port = null;
        });
      }

      port.postMessage({
        text: text,
        settings: settings,
      });
      console.log('✓ Sent to host with settings:', settings.provider);
    } catch (err) {
      console.error('✗ Host error:', err);
    }
  });
}
