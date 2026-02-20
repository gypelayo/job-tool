// Utility: send message to native host (JobFlow Desktop)
function sendNativeMessage(payload) {
  return new Promise((resolve, reject) => {
    chrome.runtime.sendNativeMessage('com.textextractor.host', payload, (response) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message));
      } else if (!response || response.ok === false) {
        reject(new Error(response && response.error ? response.error : 'Unknown host error'));
      } else {
        // Host returns { ok, error?, payload }
        resolve(response.payload || {});
      }
    });
  });
}

// Tabs
const navButtons = document.querySelectorAll('.nav-btn');
const tabs = document.querySelectorAll('.tab');

navButtons.forEach((btn) => {
  btn.addEventListener('click', () => {
    const tab = btn.dataset.tab;
    navButtons.forEach((b) => b.classList.toggle('active', b === btn));
    tabs.forEach((t) => t.classList.toggle('active', t.id === `tab-${tab}`));

    // Lazy-load analytics when switching to that tab
    if (tab === 'analytics') {
      loadAnalytics();
    }
  });
});

// Jobs tab elements
const jobsListEl = document.getElementById('jobsList');
const jobDetailEl = document.getElementById('jobDetail');
const searchInputEl = document.getElementById('searchInput');
const statusFilterEl = document.getElementById('statusFilter');

let allJobs = [];

// Render jobs list
function renderJobs() {
  const search = (searchInputEl.value || '').toLowerCase();
  const statusFilter = statusFilterEl.value;

  jobsListEl.innerHTML = '';

  const filtered = allJobs.filter((job) => {
    const matchesSearch =
      job.title.toLowerCase().includes(search) ||
      job.company.toLowerCase().includes(search);
    const matchesStatus = !statusFilter || job.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  if (filtered.length === 0) {
    jobsListEl.textContent =
      'No jobs found yet. Extract a job from a posting to get started.';
    jobDetailEl.classList.add('hidden');
    return;
  }

  filtered.forEach((job) => {
    const card = document.createElement('div');
    card.className = 'job-card';

    const main = document.createElement('div');
    main.className = 'job-main';

    const title = document.createElement('div');
    title.className = 'job-title';
    title.textContent = job.title;

    const meta = document.createElement('div');
    meta.className = 'job-meta';
    meta.textContent = `${job.company} · ${job.location || 'Location not set'}`;

    main.appendChild(title);
    main.appendChild(meta);

    const actions = document.createElement('div');
    actions.className = 'job-actions';

    const statusSelect = document.createElement('select');
    ['saved', 'applied', 'interview', 'offer', 'rejected'].forEach((s) => {
      const opt = document.createElement('option');
      opt.value = s;
      opt.textContent = s.charAt(0).toUpperCase() + s.slice(1);
      if (job.status === s) opt.selected = true;
      statusSelect.appendChild(opt);
    });

    statusSelect.addEventListener('change', () => {
      updateJobStatus(job.id, statusSelect.value);
    });

    const viewBtn = document.createElement('button');
    viewBtn.textContent = 'View';
    viewBtn.addEventListener('click', () => openJobDetail(job.id));

    actions.appendChild(statusSelect);
    actions.appendChild(viewBtn);

    card.appendChild(main);
    card.appendChild(actions);
    jobsListEl.appendChild(card);
  });
}

async function loadJobs() {
  try {
    const resp = await sendNativeMessage({ action: 'listJobs' });
    allJobs = resp.jobs || [];
    renderJobs();
  } catch (err) {
    console.error('Failed to load jobs', err);
    jobsListEl.textContent =
      'Could not load jobs. Check native helper installation.';
  }
}

async function updateJobStatus(jobId, status) {
  try {
    await sendNativeMessage({ action: 'updateJob', data: { id: jobId, status } });
    const job = allJobs.find((j) => j.id === jobId);
    if (job) job.status = status;
  } catch (err) {
    console.error('Failed to update job status', err);
  }
}

async function openJobDetail(jobId) {
  try {
    const resp = await sendNativeMessage({ action: 'getJob', data: { id: jobId } });
    const job = resp.job;
    if (!job) return;

    const extracted = job.extracted || {};
    const meta = extracted.metadata || {};
    const company = extracted.company_info || {};
    const role = extracted.role_details || {};
    const reqs = extracted.requirements || {};
    const comp = extracted.compensation || {};
    const work = extracted.work_arrangement || {};
    const market = extracted.market_signals || {};

    jobDetailEl.innerHTML = '';
    jobDetailEl.classList.remove('hidden');

    // Header: title, company, original link
    const title = document.createElement('h2');
    title.textContent = meta.job_title || job.title || 'Untitled role';

    const subtitle = document.createElement('div');
    subtitle.className = 'subtitle';
    subtitle.textContent = `${company.company_name || job.company || 'Unknown company'} · ${
      company.location_full || job.location || 'Location not set'
    }`;
    
    console.log('job from host', job);
    console.log('extracted.source_url', extracted.source_url);

    const link = document.createElement('a');
    const url = job.url || extracted.source_url || '#';
    link.href = url;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.textContent = 'Open original posting';
    link.style.display = url && url !== '#' ? 'inline-block' : 'none';
    link.style.marginBottom = '8px';

    const skills = document.createElement('div');
    skills.innerHTML = `<strong>Skills:</strong> ${
      (job.skills || []).join(', ') || 'None extracted'
    }`;

    // Role (from extracted metadata/role_details)
    const metaSection = document.createElement('div');
    metaSection.innerHTML = `
      <h3>Role</h3>
      <p><strong>Title:</strong> ${meta.job_title || ''}</p>
      <p><strong>Department:</strong> ${meta.department || ''}</p>
      <p><strong>Seniority:</strong> ${meta.seniority_level || ''}</p>
      <p><strong>Function:</strong> ${meta.job_function || ''}</p>
    `;

    // Company
    const companySection = document.createElement('div');
    companySection.innerHTML = `
      <h3>Company</h3>
      <p><strong>Name:</strong> ${company.company_name || ''}</p>
      <p><strong>Industry:</strong> ${company.industry || ''}</p>
      <p><strong>Size:</strong> ${company.company_size || ''}</p>
      <p><strong>Location:</strong> ${company.location_full || ''}</p>
    `;

    // Role details
    const roleSection = document.createElement('div');
    const summaryHtml = (role.summary || '').replace(/\n/g, '<br>');
    const responsibilitiesHtml =
      (role.key_responsibilities || []).map((r) => `• ${r}`).join('<br>') || '—';
    roleSection.innerHTML = `
      <h3>Role details</h3>
      <p><strong>Summary:</strong><br>${summaryHtml}</p>
      <p><strong>Key responsibilities:</strong><br>${responsibilitiesHtml}</p>
      <p><strong>Team structure:</strong> ${role.team_structure || '—'}</p>
    `;

    // Requirements
    const reqSection = document.createElement('div');
    const ts = reqs.technical_skills || {};
    const techLines = [
      ['Programming languages', ts.programming_languages],
      ['Frameworks', ts.frameworks],
      ['Databases', ts.databases],
      ['Cloud platforms', ts.cloud_platforms],
      ['DevOps tools', ts.devops_tools],
      ['Other', ts.other],
    ]
      .map(([label, arr]) =>
        arr && arr.length ? `<p><strong>${label}:</strong> ${arr.join(', ')}</p>` : ''
      )
      .join('');

    reqSection.innerHTML = `
      <h3>Requirements</h3>
      <p><strong>Experience:</strong> ${reqs.years_experience_min || 0}–${
        reqs.years_experience_max || 0
      } years</p>
      <p><strong>Education:</strong> ${reqs.education_level || 'Not specified'}</p>
      <p><strong>Specific degree required:</strong> ${
        reqs.requires_specific_degree ? 'Yes' : 'No'
      }</p>
      ${techLines || ''}
      <p><strong>Soft skills:</strong> ${(reqs.soft_skills || []).join(', ') || '—'}</p>
      <p><strong>Nice to have:</strong> ${(reqs.nice_to_have || []).join(', ') || '—'}</p>
    `;

    // Compensation
    const compSection = document.createElement('div');
    const salaryMin = comp.salary_min || 0;
    const salaryMax = comp.salary_max || 0;
    const currency = comp.salary_currency || '';
    compSection.innerHTML = `
      <h3>Compensation & benefits</h3>
      <p><strong>Salary:</strong> ${
        salaryMin || salaryMax
          ? `${salaryMin}–${salaryMax} ${currency}`.trim()
          : 'Not specified'
      }</p>
      <p><strong>Equity:</strong> ${comp.has_equity ? 'Yes' : 'No'}</p>
      <p><strong>Remote stipend:</strong> ${comp.has_remote_stipend ? 'Yes' : 'No'}</p>
      <p><strong>Visa sponsorship:</strong> ${comp.offers_visa_sponsorship ? 'Yes' : 'No'}</p>
      <p><strong>Health insurance:</strong> ${comp.offers_health_insurance ? 'Yes' : 'No'}</p>
      <p><strong>PTO:</strong> ${comp.offers_pto ? 'Yes' : 'No'}</p>
      <p><strong>Professional development:</strong> ${
        comp.offers_professional_development ? 'Yes' : 'No'
      }</p>
      <p><strong>401k:</strong> ${comp.offers_401k ? 'Yes' : 'No'}</p>
      <p><strong>Benefits:</strong> ${(comp.benefits || []).join(', ') || '—'}</p>
    `;

    // Work arrangement
    const workSection = document.createElement('div');
    const remoteFriendly =
      typeof work.is_remote_friendly === 'boolean'
        ? work.is_remote_friendly
          ? 'Yes'
          : 'No'
        : 'Not specified';
    workSection.innerHTML = `
      <h3>Work arrangement</h3>
      <p><strong>Workplace type:</strong> ${work.workplace_type || 'Not specified'}</p>
      <p><strong>Job type:</strong> ${work.job_type || 'Not specified'}</p>
      <p><strong>Remote friendly:</strong> ${remoteFriendly}</p>
      <p><strong>Timezone requirements:</strong> ${work.timezone_requirements || '—'}</p>
    `;

    // Market signals
    const marketSection = document.createElement('div');
    marketSection.innerHTML = `
      <h3>Market signals</h3>
      <p><strong>Urgency:</strong> ${market.urgency_level || 'Standard'}</p>
      <p><strong>Interview rounds:</strong> ${
        market.interview_rounds !== undefined && market.interview_rounds !== null
          ? market.interview_rounds
          : 'Not specified'
      }</p>
      <p><strong>Take home:</strong> ${market.has_take_home ? 'Yes' : 'No'}</p>
      <p><strong>Pair programming:</strong> ${market.has_pair_programming ? 'Yes' : 'No'}</p>
      <p><strong>Extracted at:</strong> ${extracted.extracted_at || ''}</p>
    `;

    // Notes
    const notesLabel = document.createElement('label');
    notesLabel.textContent = 'Notes';
    const notesArea = document.createElement('textarea');
    notesArea.value = job.notes || '';
    notesArea.rows = 4;
    notesArea.style.width = '100%';
    notesArea.style.marginTop = '4px';
    notesArea.addEventListener('change', () => {
      saveJobNotes(job.id, notesArea.value);
    });
    notesLabel.appendChild(notesArea);

    jobDetailEl.appendChild(title);
    jobDetailEl.appendChild(subtitle);
    jobDetailEl.appendChild(link);
    jobDetailEl.appendChild(skills);
    jobDetailEl.appendChild(document.createElement('hr'));
    jobDetailEl.appendChild(metaSection);
    jobDetailEl.appendChild(companySection);
    jobDetailEl.appendChild(roleSection);
    jobDetailEl.appendChild(reqSection);
    jobDetailEl.appendChild(compSection);
    jobDetailEl.appendChild(workSection);
    jobDetailEl.appendChild(marketSection);
    jobDetailEl.appendChild(document.createElement('hr'));
    jobDetailEl.appendChild(notesLabel);
  } catch (err) {
    console.error('Failed to load job detail', err);
  }
}


async function saveJobNotes(jobId, notes) {
  try {
    await sendNativeMessage({ action: 'updateJob', data: { id: jobId, notes } });
    const job = allJobs.find((j) => j.id === jobId);
    if (job) job.notes = notes;
  } catch (err) {
    console.error('Failed to save notes', err);
  }
}

// Analytics tab (placeholder: just fetch and show JSON)
async function loadAnalytics() {
  const el = document.getElementById('analyticsContent');
  el.textContent = 'Loading analytics...';
  try {
    const resp = await sendNativeMessage({ action: 'getAnalytics' });
    el.textContent = JSON.stringify(resp, null, 2);
  } catch (err) {
    console.error('Failed to load analytics', err);
    el.textContent =
      'Could not load analytics. Check native helper installation.';
  }
}

// Settings tab: Perplexity key & host test
const apiKeyInput = document.getElementById('perplexityApiKey');
const saveApiKeyBtn = document.getElementById('saveApiKeyBtn');
const apiKeyStatus = document.getElementById('apiKeyStatus');
const testHostBtn = document.getElementById('testHostBtn');
const hostStatus = document.getElementById('hostStatus');

function loadApiKey() {
  chrome.storage.sync.get(['perplexityApiKey'], (res) => {
    if (res.perplexityApiKey) {
      apiKeyInput.value = res.perplexityApiKey;
      apiKeyStatus.textContent = 'API key is set.';
    }
  });
}

saveApiKeyBtn.addEventListener('click', () => {
  const key = apiKeyInput.value.trim();
  if (!key) {
    apiKeyStatus.textContent = 'Please enter an API key.';
    return;
  }
  chrome.storage.sync.set({ perplexityApiKey: key }, () => {
    apiKeyStatus.textContent = 'Saved.';
  });
});

testHostBtn.addEventListener('click', async () => {
  hostStatus.textContent = 'Testing...';
  try {
    const resp = await sendNativeMessage({ action: 'ping' });
    hostStatus.textContent = resp && resp.ok
      ? 'Native helper is connected.'
      : 'Unexpected response from helper.';
  } catch (err) {
    hostStatus.textContent =
      'Native helper not reachable. Make sure JobFlow Desktop is installed.';
  }
});

// Event bindings
searchInputEl.addEventListener('input', renderJobs);
statusFilterEl.addEventListener('change', renderJobs);

// Initial load
loadJobs();
loadApiKey();
// Analytics will load when tab is opened
