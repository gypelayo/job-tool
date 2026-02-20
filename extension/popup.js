const extractBtn = document.getElementById('extractBtn');
const dashboardBtn = document.getElementById('dashboardBtn');
const statusEl = document.getElementById('status');

// Open full dashboard
dashboardBtn.addEventListener('click', () => {
  chrome.tabs.create({ url: chrome.runtime.getURL('dashboard.html') });
});

// Helper: get page text in MV2 / Firefox compatible way
function getPageText(tabId) {
  return new Promise((resolve, reject) => {
    chrome.tabs.executeScript(
      tabId,
      {
        code: 'document.body ? document.body.innerText : ""'
      },
      (results) => {
        if (chrome.runtime.lastError) {
          reject(new Error(chrome.runtime.lastError.message));
        } else if (!results || !results.length) {
          resolve('');
        } else {
          resolve(results[0]);
        }
      }
    );
  });
}

// Extract current tab's page text and send to native host (legacy flow)
extractBtn.addEventListener('click', () => {
  statusEl.textContent = 'Extracting...';

  chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
    if (!tabs || !tabs.length) {
      statusEl.textContent = 'No active tab.';
      return;
    }
    const tab = tabs[0];

    getPageText(tab.id)
      .then((pageText) => {
        chrome.storage.sync.get(['perplexityApiKey'], (res) => {
          const key = res.perplexityApiKey || '';
          if (!key) {
            statusEl.textContent = 'Set your Perplexity API key in Settings first.';
            return;
          }

          const message = {
            text: pageText,
            settings: {
              provider: 'perplexity',
              perplexityKey: key,
              perplexityModel: 'sonar-pro',
              sourceUrl: tab.url,           // NEW: flows into Settings.SourceURL
            }
          };

          chrome.runtime.sendNativeMessage(
            'com.textextractor.host',
            message,
            (response) => {
              if (chrome.runtime.lastError) {
                statusEl.textContent = 'Error: ' + chrome.runtime.lastError.message;
              } else if (!response || response.status !== 'success') {
                statusEl.textContent = 'Extraction failed.';
              } else {
                statusEl.textContent = 'Job extracted and saved.';
              }
            }
          );
        });
      })
      .catch((err) => {
        console.error(err);
        statusEl.textContent = 'Error extracting page text.';
      });
  });
});
