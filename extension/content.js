console.log(`[Content Script] Loaded in: ${window.location.href}`);
console.log(`[Content Script] Frame type: ${window === window.top ? "MAIN" : "IFRAME"}`);
console.log(`[Content Script] Hostname: ${window.location.hostname}`);

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extract") {
    console.log(`[Extract] Starting extraction`);
    
    if (window.location.hostname.includes('greenhouse.io')) {
      console.log(`[Extract] Greenhouse iframe detected`);
      waitForGreenhouseContent();
    } else {
      console.log(`[Extract] Regular page`);
      waitForMainPageContent();
    }
    
    sendResponse({received: true});
    return true;
  }
});

function waitForGreenhouseContent() {
  let attempts = 0;
  const maxAttempts = 30;
  
  const checkInterval = setInterval(() => {
    attempts++;
    const bodyText = document.body ? document.body.innerText : '';
    
    console.log(`[Greenhouse ${attempts}s] Content: ${bodyText.length} chars`);
    
    if (bodyText.length > 1000 || attempts >= maxAttempts) {
      clearInterval(checkInterval);
      
      console.log(`[Greenhouse] Extracting ${bodyText.length} chars`);
      console.log(`[Greenhouse] Preview: ${bodyText.substring(0, 200)}`);
      
      chrome.runtime.sendMessage({
        action: "extractText",
        data: {
          text: bodyText,
          url: window.location.href,
          title: document.title,
          isMainFrame: false,
          contentLength: bodyText.length,
          source: 'greenhouse-iframe'
        }
      }).catch(err => console.error("[Greenhouse] Send error:", err));
    }
  }, 1000);
}

function waitForMainPageContent() {
  setTimeout(() => {
    const bodyText = document.body ? document.body.innerText : '';
    console.log(`[Main] Content: ${bodyText.length} chars`);
    
    chrome.runtime.sendMessage({
      action: "extractText",
      data: {
        text: bodyText,
        url: window.location.href,
        title: document.title,
        isMainFrame: true,
        contentLength: bodyText.length,
        source: 'main-page'
      }
    }).catch(err => console.error("[Main] Send error:", err));
  }, 2000);
}
