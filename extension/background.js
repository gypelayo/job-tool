let port = null;
let frameData = [];
let extractionTimer = null;

chrome.action.onClicked.addListener((tab) => {
  console.log("Extension clicked on tab:", tab.id);
  
  // Reset collection
  frameData = [];
  clearTimeout(extractionTimer);
  
  // Trigger extraction in main frame
  chrome.tabs.sendMessage(tab.id, { action: "extract" }).catch((err) => {
    console.error("Error sending to main frame:", err);
  });
  
  // Trigger extraction in all subframes
  chrome.webNavigation.getAllFrames({ tabId: tab.id }, (frames) => {
    if (frames) {
      console.log("Found", frames.length, "frames");
      frames.forEach(frame => {
        chrome.tabs.sendMessage(tab.id, { action: "extract" }, { frameId: frame.frameId }).catch(() => {});
      });
    }
  });
  
  // Wait 7 seconds for all frames to respond, then combine and send
  extractionTimer = setTimeout(() => {
    console.log("Timer finished, collected", frameData.length, "frames");
    if (frameData.length > 0) {
      const combined = frameData.join("\n\n========== NEXT FRAME ==========\n\n");
      console.log("Combining", frameData.length, "frames, total length:", combined.length);
      sendToHost(combined);
    } else {
      console.error("No frame data collected!");
    }
  }, 7000);  // Increased to 7 seconds
});

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extractText") {
    console.log("Received from frame:", sender.frameId, "length:", request.text.length);
    frameData.push(request.text);
  }
});

function sendToHost(text) {
  console.log("Attempting to connect to native host...");
  
  try {
    if (!port) {
      port = chrome.runtime.connectNative('com.textextractor.host');
      console.log("Connected to native host");
      
      port.onMessage.addListener((msg) => {
        console.log("Host response:", msg);
      });
      
      port.onDisconnect.addListener(() => {
        console.log("Disconnected from host");
        if (chrome.runtime.lastError) {
          console.error("Disconnect error:", chrome.runtime.lastError.message);
        }
        port = null;
      });
    }
    
    console.log("Posting message to host, text length:", text.length);
    port.postMessage({ text: text });
  } catch (err) {
    console.error("Error connecting to native host:", err);
  }
}
