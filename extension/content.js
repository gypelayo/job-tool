console.log("Content script loaded on:", window.location.href);

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extract") {
    console.log("Extracting from:", window.location.href);
    
    // Wait 5 seconds for dynamic content to load
    setTimeout(() => {
      const completeHTML = document.documentElement.outerHTML;
      console.log("Extracted length:", completeHTML.length);
      
      chrome.runtime.sendMessage({
        action: "extractText",
        text: completeHTML
      }).catch(err => console.error("Error:", err));
    }, 5000);  // Changed from 3000 to 5000
    
    return true;
  }
});
