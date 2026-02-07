console.log("[Content] Loaded on:", window.location.href);

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extract") {
    console.log("[Content] Starting extraction");
    
    waitForContent().then(() => {
      extractAndSend();
      sendResponse({ started: true });
    });
    
    return true;
  }
});

function waitForContent() {
  return new Promise((resolve) => {
    let lastLength = document.body.innerText.length;
    let stableCount = 0;
    let elapsed = 0;
    const maxWait = 15000;
    
    console.log("[Content] Initial content:", lastLength, "chars");
    
    const checkInterval = setInterval(() => {
      elapsed += 1000;
      const currentLength = document.body.innerText.length;
      
      if (currentLength === lastLength) {
        stableCount++;
      } else {
        stableCount = 0;
        console.log("[Content] Content changed:", currentLength, "chars");
      }
      
      lastLength = currentLength;
      
      // Stable for 2 seconds OR timeout
      if ((stableCount >= 2 && currentLength > 500) || elapsed >= maxWait) {
        clearInterval(checkInterval);
        console.log("[Content] Content ready");
        resolve();
      }
    }, 1000);
  });
}

function extractAndSend() {
  const fullText = document.body.innerText;
  
  // Try to find main content area
  const contentSelectors = [
    'main',
    'article',
    '[role="main"]',
    '#content',
    '.content',
    '[class*="job"]',
    '[class*="posting"]',
    '[class*="description"]'
  ];
  
  let bestContent = fullText;
  let bestLength = 0;
  
  for (const selector of contentSelectors) {
    const element = document.querySelector(selector);
    if (element) {
      const text = element.innerText;
      if (text.length > bestLength && text.length > 500) {
        bestContent = text;
        bestLength = text.length;
      }
    }
  }
  
  console.log("[Content] Extracted", bestContent.length, "chars");
  
  chrome.runtime.sendMessage({
    action: "extractText",
    data: {
      text: bestContent,
      url: window.location.href,
      title: document.title,
      contentLength: bestContent.length
    }
  }).catch(err => console.error("[Content] Send failed:", err));
}
