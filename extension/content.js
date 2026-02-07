console.log("[Content] Loaded on:", window.location.href);

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "extractWellfound") {
    console.log("[Content] Wellfound extraction");
    
    waitForContent().then(() => {
      const jobContent = extractWellfound();
      sendJobData(jobContent, "Wellfound");
      sendResponse({ started: true });
    });
    
    return true;
  } else if (request.action === "extractRemoteRocketship") {
    console.log("[Content] Remote Rocketship extraction");
    
    waitForContent().then(() => {
      const jobContent = extractRemoteRocketship();
      sendJobData(jobContent, "Remote Rocketship");
      sendResponse({ started: true });
    });
    
    return true;
  } else if (request.action === "extractLinkedIn") {
    console.log("[Content] LinkedIn extraction");
    
    waitForContent().then(() => {
      const jobContent = extractLinkedIn();
      sendJobData(jobContent, "LinkedIn");
      sendResponse({ started: true });
    });
    
    return true;
  } else if (request.action === "extract") {
    console.log("[Content] Generic extraction");
    
    waitForContent().then(() => {
      const jobContent = document.body.innerText;
      sendJobData(jobContent, "Generic");
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
    const maxWait = 10000;
    
    const checkInterval = setInterval(() => {
      elapsed += 1000;
      const currentLength = document.body.innerText.length;
      
      if (currentLength === lastLength) {
        stableCount++;
      } else {
        stableCount = 0;
      }
      
      lastLength = currentLength;
      
      if ((stableCount >= 2 && currentLength > 1000) || elapsed >= maxWait) {
        clearInterval(checkInterval);
        console.log("[Content] Content ready");
        resolve();
      }
    }, 1000);
  });
}

// ========== WELLFOUND EXTRACTOR ==========

function extractWellfound() {
  console.log("[Wellfound] Starting extraction");
  
  let jobData = {
    title: '',
    salary: '',
    location: '',
    company: '',
    description: ''
  };
  
  // Extract job title
  const titleSelectors = [
    'h1',
    '[class*="JobTitle"]',
    '[class*="job-title"]'
  ];
  
  for (const selector of titleSelectors) {
    const el = document.querySelector(selector);
    if (el && el.innerText.trim().length > 0 && el.innerText.trim().length < 150) {
      jobData.title = el.innerText.trim();
      break;
    }
  }
  
  // Extract salary/equity
  const salarySelectors = [
    '[class*="salary"]',
    '[class*="compensation"]',
    '[class*="Salary"]'
  ];
  
  for (const selector of salarySelectors) {
    const el = document.querySelector(selector);
    if (el) {
      const text = el.innerText.trim();
      if (text.includes('$') || text.includes('€') || text.includes('%')) {
        jobData.salary = text;
        break;
      }
    }
  }
  
  // Extract location
  const locationSelectors = [
    '[class*="location"]',
    '[class*="Location"]'
  ];
  
  for (const selector of locationSelectors) {
    const el = document.querySelector(selector);
    if (el) {
      const text = el.innerText.trim();
      if (text.length > 0 && text.length < 200) {
        jobData.location = text;
        break;
      }
    }
  }
  
  // Extract job description
  const descriptionSelectors = [
    '[class*="JobDescription"]',
    '[class*="job-description"]',
    '[class*="description"]',
    'main section',
    'main'
  ];
  
  for (const selector of descriptionSelectors) {
    const el = document.querySelector(selector);
    if (el) {
      const clone = el.cloneNode(true);
      
      const removeSelectors = [
        'script',
        'style',
        'nav',
        'header',
        'footer',
        '[class*="similar"]',
        '[class*="Similar"]',
        '[class*="navigation"]',
        '[class*="cookie"]',
        '[class*="consent"]',
        '[id*="consent"]'
      ];
      
      removeSelectors.forEach(sel => {
        clone.querySelectorAll(sel).forEach(elem => elem.remove());
      });
      
      const text = clone.innerText.trim();
      
      if (text.length > 500) {
        jobData.description = text;
        console.log("[Wellfound] Found description:", text.length, "chars");
        break;
      }
    }
  }
  
  if (!jobData.description || jobData.description.length < 500) {
    const main = document.querySelector('main');
    if (main) {
      const clone = main.cloneNode(true);
      clone.querySelectorAll('script, style, nav, header, footer, [class*="similar"], [class*="Similar"], [class*="cookie"], [class*="consent"], [id*="consent"]').forEach(el => el.remove());
      jobData.description = clone.innerText.trim();
    }
  }
  
  // Format output
  let formatted = '';
  
  if (jobData.title) {
    formatted += `JOB TITLE: ${jobData.title}\n\n`;
  }
  
  if (jobData.salary) {
    formatted += `COMPENSATION: ${jobData.salary}\n\n`;
  }
  
  if (jobData.location) {
    formatted += `LOCATION: ${jobData.location}\n\n`;
  }
  
  formatted += `DESCRIPTION:\n${jobData.description}`;
  
  formatted = cleanText(formatted);
  
  console.log("[Wellfound] Final output:", formatted.length, "chars");
  
  return formatted;
}

// ========== REMOTE ROCKETSHIP EXTRACTOR ==========

function extractRemoteRocketship() {
  console.log("[RemoteRocketship] Starting extraction");
  
  const clone = document.body.cloneNode(true);
  
  const removeSelectors = [
    'script',
    'style',
    'nav',
    'header',
    'footer',
    '[class*="similar"]',
    '[class*="Similar"]'
  ];
  
  removeSelectors.forEach(selector => {
    clone.querySelectorAll(selector).forEach(el => el.remove());
  });
  
  let fullText = clone.innerText;
  
  const similarJobsIndex = fullText.indexOf('Similar Jobs');
  if (similarJobsIndex > 0) {
    fullText = fullText.substring(0, similarJobsIndex);
  }
  
  const discoverIndex = fullText.indexOf('Discover 100,000+ Remote Jobs');
  if (discoverIndex > 0) {
    fullText = fullText.substring(0, discoverIndex);
  }
  
  fullText = cleanText(fullText);
  
  console.log("[RemoteRocketship] Extracted", fullText.length, "chars");
  
  return fullText;
}

// ========== LINKEDIN EXTRACTOR ==========

function extractLinkedIn() {
  console.log("[LinkedIn] Starting extraction");
  
  let jobData = {
    title: '',
    company: '',
    location: '',
    description: ''
  };
  
  // Extract job title
  const titleSelectors = [
    '.top-card-layout__title',
    'h1.topcard__title',
    'h1'
  ];
  
  for (const selector of titleSelectors) {
    const el = document.querySelector(selector);
    if (el && el.innerText.trim().length > 0 && el.innerText.trim().length < 150) {
      jobData.title = el.innerText.trim();
      break;
    }
  }
  
  // Extract company name
  const companySelectors = [
    '.topcard__org-name-link',
    '.topcard__flavor',
    'a.topcard__org-name-link'
  ];
  
  for (const selector of companySelectors) {
    const el = document.querySelector(selector);
    if (el && el.innerText.trim().length > 0) {
      jobData.company = el.innerText.trim();
      break;
    }
  }
  
  // Extract location
  const locationSelectors = [
    '.topcard__flavor--bullet',
    '.top-card-layout__second-subline'
  ];
  
  for (const selector of locationSelectors) {
    const el = document.querySelector(selector);
    if (el) {
      const text = el.innerText.trim();
      if (text.length > 0 && text.length < 200 && !text.includes('applicants')) {
        jobData.location = text;
        break;
      }
    }
  }
  
  // Extract job description
  const descriptionSelectors = [
    '.show-more-less-html__markup',
    '.jobs-description__content',
    '.description__text',
    '[class*="job-description"]'
  ];
  
  for (const selector of descriptionSelectors) {
    const el = document.querySelector(selector);
    if (el) {
      const clone = el.cloneNode(true);
      
      clone.querySelectorAll('button, [class*="show-more"]').forEach(elem => elem.remove());
      
      const text = clone.innerText.trim();
      
      if (text.length > 200) {
        jobData.description = text;
        console.log("[LinkedIn] Found description:", text.length, "chars");
        break;
      }
    }
  }
  
  if (!jobData.description || jobData.description.length < 200) {
    const mainContent = document.querySelector('[class*="jobs-description"]') || 
                       document.querySelector('article') ||
                       document.querySelector('main');
    
    if (mainContent) {
      const clone = mainContent.cloneNode(true);
      
      const removeSelectors = [
        'nav',
        'header',
        'aside',
        '[class*="messaging"]',
        '[class*="global-nav"]',
        '[class*="artdeco-modal"]',
        'button',
        'form'
      ];
      
      removeSelectors.forEach(sel => {
        clone.querySelectorAll(sel).forEach(elem => elem.remove());
      });
      
      jobData.description = clone.innerText.trim();
    }
  }
  
  // Format output
  let formatted = '';
  
  if (jobData.title) {
    formatted += `JOB TITLE: ${jobData.title}\n\n`;
  }
  
  if (jobData.company) {
    formatted += `COMPANY: ${jobData.company}\n\n`;
  }
  
  if (jobData.location) {
    formatted += `LOCATION: ${jobData.location}\n\n`;
  }
  
  formatted += `DESCRIPTION:\n${jobData.description}`;
  
  formatted = cleanLinkedInText(formatted);
  
  console.log("[LinkedIn] Final output:", formatted.length, "chars");
  
  return formatted;
}

// ========== HELPER FUNCTIONS ==========

function cleanText(text) {
  // Remove script/style content
  text = text.replace(/@font-face[\s\S]*?\}/g, '');
  text = text.replace(/\/\*[\s\S]*?\*\//g, '');
  text = text.replace(/\{[\s\S]*?\}/g, '');
  
  // Remove URLs and technical stuff
  text = text.replace(/https?:\/\/[^\s]+/g, '');
  text = text.replace(/[a-z-]+\.(woff2?|ttf|eot)/gi, '');
  
  // Clean whitespace
  text = text.replace(/\n{4,}/g, '\n\n');
  text = text.replace(/\s{3,}/g, ' ');
  
  // Remove common footer/nav text
  text = text.replace(/DiscoverFind JobsFor Recruiters.*?Sign Up/gi, '');
  text = text.replace(/Copyright.*$/gi, '');
  text = text.replace(/©.*\d{4}.*$/gi, '');
  text = text.replace(/Cookie Preferences.*/gi, '');
  text = text.replace(/Browse by:.*/gi, '');
  text = text.replace(/Similar Jobs.*/gi, '');
  text = text.replace(/Find Your Dream Remote Job.*/gi, '');
  text = text.replace(/Loved by \d+.*remote workers/gi, '');
  text = text.replace(/Wall of Love.*/gi, '');
  text = text.replace(/Frequently asked questions.*/gi, '');
  
  return text.trim();
}

function cleanLinkedInText(text) {
  text = cleanText(text);
  
  // LinkedIn-specific cleanup
  text = text.replace(/Skip to search.*?Skip to main content/gi, '');
  text = text.replace(/Keyboard shortcuts/gi, '');
  text = text.replace(/\d+ notifications total/gi, '');
  text = text.replace(/Home My Network Jobs Messaging Notifications/gi, '');
  text = text.replace(/Show more options/gi, '');
  text = text.replace(/Apply.*?Reposted.*?ago/gi, '');
  text = text.replace(/Over \d+ people clicked apply/gi, '');
  text = text.replace(/Responses managed off LinkedIn/gi, '');
  text = text.replace(/Matches your job preferences.*/gi, '');
  text = text.replace(/No longer accepting applications/gi, '');
  text = text.replace(/Job activity.*/gi, '');
  text = text.replace(/See more jobs like this/gi, '');
  text = text.replace(/Learn skills to get a new job.*/gi, '');
  text = text.replace(/Looking for talent.*Post a job/gi, '');
  text = text.replace(/About Accessibility.*LinkedIn Corporation/gi, '');
  text = text.replace(/Status is online Messaging.*/gi, '');
  text = text.replace(/You are on the messaging overlay.*/gi, '');
  text = text.replace(/Premium insights/gi, '');
  text = text.replace(/More jobs.*/gi, '');
  text = text.replace(/Get hired faster with the help of AI tools.*/gi, '');
  text = text.replace(/See how you compare to others.*/gi, '');
  text = text.replace(/Exclusive Job Seeker Insights.*/gi, '');
  text = text.replace(/Powered by Bing.*/gi, '');
  
  const messageStart = text.indexOf('Close your conversation with');
  if (messageStart > 0) {
    text = text.substring(0, messageStart);
  }
  
  return text.trim();
}

function sendJobData(text, source) {
  const formatted = `
URL: ${window.location.href}
SOURCE: ${source} (Scraped)

${text}
`;
  
  chrome.runtime.sendMessage({
    action: "extractText",
    data: {
      text: formatted,
      url: window.location.href,
      title: document.title,
      contentLength: text.length
    }
  }).catch(err => console.error("[Content] Send failed:", err));
}
