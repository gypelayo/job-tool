// Load saved settings
document.addEventListener('DOMContentLoaded', async () => {
  // Use browser.storage instead of chrome.storage for Firefox
  const storage = typeof browser !== 'undefined' ? browser.storage : chrome.storage;
  
  const settings = await storage.sync.get({
    provider: 'ollama',
    ollamaModel: 'qwen2.5:7b',
    perplexityKey: '',
    perplexityModel: 'sonar-pro'
  });
  
  document.getElementById('provider').value = settings.provider;
  document.getElementById('ollama-model').value = settings.ollamaModel;
  document.getElementById('perplexity-key').value = settings.perplexityKey;
  document.getElementById('perplexity-model').value = settings.perplexityModel;
  
  toggleProviderConfig(settings.provider);
});

// Toggle between Ollama and Perplexity config
document.getElementById('provider').addEventListener('change', (e) => {
  toggleProviderConfig(e.target.value);
});

function toggleProviderConfig(provider) {
  const ollamaConfig = document.getElementById('ollama-config');
  const perplexityConfig = document.getElementById('perplexity-config');
  
  if (provider === 'ollama') {
    ollamaConfig.style.display = 'block';
    perplexityConfig.style.display = 'none';
  } else {
    ollamaConfig.style.display = 'none';
    perplexityConfig.style.display = 'block';
  }
}

// Save settings
document.getElementById('save').addEventListener('click', async () => {
  const settings = {
    provider: document.getElementById('provider').value,
    ollamaModel: document.getElementById('ollama-model').value,
    perplexityKey: document.getElementById('perplexity-key').value,
    perplexityModel: document.getElementById('perplexity-model').value
  };
  
  // Validate
  if (settings.provider === 'perplexity' && !settings.perplexityKey) {
    showStatus('Please enter your Perplexity API key', 'error');
    return;
  }
  
  try {
    // Use browser.storage for Firefox compatibility
    const storage = typeof browser !== 'undefined' ? browser.storage : chrome.storage;
    await storage.sync.set(settings);
    showStatus('Settings saved successfully!', 'success');
  } catch (err) {
    showStatus('Error saving settings: ' + err.message, 'error');
  }
});

function showStatus(message, type) {
  const status = document.getElementById('status');
  status.textContent = message;
  status.className = 'status ' + type;
  status.style.display = 'block';
  
  setTimeout(() => {
    status.style.display = 'none';
  }, 3000);
}
