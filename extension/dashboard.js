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
let currentJobId = null;

// Charts
let chartProgrammingLanguages = null;
let chartFrameworks = null;
let chartDatabases = null;
let chartCloudPlatforms = null;
let chartDevopsTools = null;
let chartOtherSkills = null;
let chartJobTitles = null;
let chartSkillsByStatus = null;

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
    if (job.id === currentJobId) {
      card.classList.add('selected');
    }

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
    viewBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      openJobDetail(job.id);
    });

    const deleteBtn = document.createElement('button');
    deleteBtn.textContent = '❌';
    deleteBtn.className = 'job-delete-btn';
    deleteBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      deleteJob(job.id, deleteBtn);
    });

    actions.appendChild(statusSelect);
    actions.appendChild(viewBtn);
    actions.appendChild(deleteBtn);

    card.appendChild(main);
    card.appendChild(actions);

    card.addEventListener('click', (e) => {
      if (e.target.closest('.job-actions')) return;
      openJobDetail(job.id);
    });

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

async function deleteJob(jobId, buttonEl) {
  if (!confirm('Remove this job from your list?')) return;
  try {
    if (buttonEl) buttonEl.disabled = true;
    await sendNativeMessage({ action: 'deleteJob', data: { id: jobId } });
    allJobs = allJobs.filter((j) => j.id !== jobId);
    if (currentJobId === jobId) {
      currentJobId = null;
      jobDetailEl.classList.add('hidden');
      jobDetailEl.innerHTML = '';
    }
    renderJobs();
  } catch (err) {
    console.error('Failed to delete job', err);
  } finally {
    if (buttonEl) buttonEl.disabled = false;
  }
}

async function openJobDetail(jobId) {
  try {
    const resp = await sendNativeMessage({ action: 'getJob', data: { id: jobId } });
    const job = resp.job;
    if (!job) return;

    currentJobId = jobId;
    renderJobs();

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

    // Header
    const title = document.createElement('h2');
    title.textContent = meta.job_title || job.title || 'Untitled role';

    const subtitle = document.createElement('div');
    subtitle.className = 'subtitle';
    subtitle.textContent = `${company.company_name || job.company || 'Unknown company'} · ${company.location_full || job.location || 'Location not set'
      }`;

    const closeBtn = document.createElement('button');
    closeBtn.textContent = 'Close';
    closeBtn.className = 'job-detail-close-btn';
    closeBtn.addEventListener('click', () => {
      currentJobId = null;
      jobDetailEl.classList.add('hidden');
      jobDetailEl.innerHTML = '';
      renderJobs();
    });

    const leftHeader = document.createElement('div');
    leftHeader.appendChild(title);
    leftHeader.appendChild(subtitle);

    const headerRow = document.createElement('div');
    headerRow.style.display = 'flex';
    headerRow.style.justifyContent = 'space-between';
    headerRow.style.alignItems = 'center';
    headerRow.style.gap = '8px';
    headerRow.appendChild(leftHeader);
    headerRow.appendChild(closeBtn);

    const link = document.createElement('a');
    const url = job.url || extracted.source_url || '#';
    link.href = url;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.textContent = 'Open original posting';
    link.style.display = url && url !== '#' ? 'inline-block' : 'none';
    link.style.marginBottom = '8px';

    const skills = document.createElement('div');
    skills.innerHTML = `<strong>Skills:</strong> ${(job.skills || []).join(', ') || 'None extracted'
      }`;

    // Sections (role, company, etc.) – same as before
    const metaSection = document.createElement('div');
    metaSection.innerHTML = `
      <h3>Role</h3>
      <p><strong>Title:</strong> ${meta.job_title || ''}</p>
      <p><strong>Department:</strong> ${meta.department || ''}</p>
      <p><strong>Seniority:</strong> ${meta.seniority_level || ''}</p>
      <p><strong>Function:</strong> ${meta.job_function || ''}</p>
    `;

    const companySection = document.createElement('div');
    companySection.innerHTML = `
      <h3>Company</h3>
      <p><strong>Name:</strong> ${company.company_name || ''}</p>
      <p><strong>Industry:</strong> ${company.industry || ''}</p>
      <p><strong>Size:</strong> ${company.company_size || ''}</p>
      <p><strong>Location:</strong> ${company.location_full || ''}</p>
    `;

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
      <p><strong>Experience:</strong> ${reqs.years_experience_min || 0}–${reqs.years_experience_max || 0
      } years</p>
      <p><strong>Education:</strong> ${reqs.education_level || 'Not specified'}</p>
      <p><strong>Specific degree required:</strong> ${reqs.requires_specific_degree ? 'Yes' : 'No'
      }</p>
      ${techLines || ''}
      <p><strong>Soft skills:</strong> ${(reqs.soft_skills || []).join(', ') || '—'}</p>
      <p><strong>Nice to have:</strong> ${(reqs.nice_to_have || []).join(', ') || '—'}</p>
    `;

    const compSection = document.createElement('div');
    const salaryMin = comp.salary_min || 0;
    const salaryMax = comp.salary_max || 0;
    const currency = comp.salary_currency || '';
    compSection.innerHTML = `
      <h3>Compensation & benefits</h3>
      <p><strong>Salary:</strong> ${salaryMin || salaryMax
        ? `${salaryMin}–${salaryMax} ${currency}`.trim()
        : 'Not specified'
      }</p>
      <p><strong>Equity:</strong> ${comp.has_equity ? 'Yes' : 'No'}</p>
      <p><strong>Remote stipend:</strong> ${comp.has_remote_stipend ? 'Yes' : 'No'}</p>
      <p><strong>Visa sponsorship:</strong> ${comp.offers_visa_sponsorship ? 'Yes' : 'No'}</p>
      <p><strong>Health insurance:</strong> ${comp.offers_health_insurance ? 'Yes' : 'No'}</p>
      <p><strong>PTO:</strong> ${comp.offers_pto ? 'Yes' : 'No'}</p>
      <p><strong>Professional development:</strong> ${comp.offers_professional_development ? 'Yes' : 'No'
      }</p>
      <p><strong>401k:</strong> ${comp.offers_401k ? 'Yes' : 'No'}</p>
      <p><strong>Benefits:</strong> ${(comp.benefits || []).join(', ') || '—'}</p>
    `;

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

    const marketSection = document.createElement('div');
    marketSection.innerHTML = `
      <h3>Market signals</h3>
      <p><strong>Urgency:</strong> ${market.urgency_level || 'Standard'}</p>
      <p><strong>Interview rounds:</strong> ${market.interview_rounds !== undefined && market.interview_rounds !== null
        ? market.interview_rounds
        : 'Not specified'
      }</p>
      <p><strong>Take home:</strong> ${market.has_take_home ? 'Yes' : 'No'}</p>
      <p><strong>Pair programming:</strong> ${market.has_pair_programming ? 'Yes' : 'No'}</p>
      <p><strong>Extracted at:</strong> ${extracted.extracted_at || ''}</p>
    `;

    const notesWrapper = document.createElement('div');
    const notesLabel = document.createElement('label');
    notesLabel.textContent = 'Notes';

    const notesArea = document.createElement('textarea');
    notesArea.value = job.notes || '';
    notesArea.rows = 4;
    notesArea.style.width = '100%';
    notesArea.style.marginTop = '4px';

    const notesSaveBtn = document.createElement('button');
    notesSaveBtn.textContent = 'Save notes';
    notesSaveBtn.className = 'notes-save-btn';
    notesSaveBtn.addEventListener('click', () => {
      saveJobNotes(job.id, notesArea.value);
    });

    notesWrapper.appendChild(notesLabel);
    notesWrapper.appendChild(notesArea);
    notesWrapper.appendChild(notesSaveBtn);

    jobDetailEl.appendChild(headerRow);
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
    jobDetailEl.appendChild(notesWrapper);
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

// Analytics tab – one chart per skill type, one for titles, one for skills per stage
async function loadAnalytics() {
  const summaryEl = document.getElementById('analyticsSummary');
  summaryEl.textContent = 'Loading analytics...';

  try {
    const resp = await sendNativeMessage({ action: 'getAnalytics' });

    const statusStats = resp.statusStats || {};
    const skillsByCategory = resp.skillsByCategory || {};
    const skillsByStatus = resp.skillsByStatus || {};
    const topJobTitles = resp.topJobTitles || [];

    const total = statusStats.total || 0;
    summaryEl.textContent =
      `You have ${total} jobs tracked. ` +
      `Applied: ${statusStats.applied || 0}, ` +
      `Interview: ${statusStats.interview || 0}, ` +
      `Offer: ${statusStats.offer || 0}.`;

    // Helper to build a simple bar chart
    function buildBarChart(chartRef, canvasId, labels, data, label, color) {
      const ctx = document.getElementById(canvasId);
      if (!ctx) return null;
      if (chartRef) chartRef.destroy();
      return new Chart(ctx.getContext('2d'), {
        type: 'bar',
        data: {
          labels,
          datasets: [
            {
              label,
              data,
              backgroundColor: color,
            },
          ],
        },
        options: {
          responsive: true,
          plugins: {
            legend: { display: false },
          },
          scales: {
            x: {
              ticks: { maxRotation: 60, minRotation: 0, autoSkip: true },
            },
            y: {
              beginAtZero: true,
            },
          },
        },
      });
    }

    const prog = skillsByCategory['programming_language'] || [];
    chartProgrammingLanguages = buildBarChart(
      chartProgrammingLanguages,
      'chartProgrammingLanguages',
      prog.map((s) => s.skill),
      prog.map((s) => s.count),
      'Jobs',
      '#2563eb'
    );

    // Frameworks chart currently has no category in DB, so either:
    // - Hide it for now, or
    // - Reuse "other" or "devops" until you add a real frameworks category

    // Databases
    const dbs = skillsByCategory['database'] || [];
    chartDatabases = buildBarChart(
      chartDatabases,
      'chartDatabases',
      dbs.map((s) => s.skill),
      dbs.map((s) => s.count),
      'Jobs',
      '#f97316'
    );

    // Cloud platforms
    const clouds = skillsByCategory['cloud'] || [];
    chartCloudPlatforms = buildBarChart(
      chartCloudPlatforms,
      'chartCloudPlatforms',
      clouds.map((s) => s.skill),
      clouds.map((s) => s.count),
      'Jobs',
      '#6366f1'
    );

    // DevOps tools
    const devops = skillsByCategory['devops'] || [];
    chartDevopsTools = buildBarChart(
      chartDevopsTools,
      'chartDevopsTools',
      devops.map((s) => s.skill),
      devops.map((s) => s.count),
      'Jobs',
      '#22c55e'
    );

    // Other skills
    const other = skillsByCategory['other'] || [];
    chartOtherSkills = buildBarChart(
      chartOtherSkills,
      'chartOtherSkills',
      other.map((s) => s.skill),
      other.map((s) => s.count),
      'Jobs',
      '#a855f7'
    );

    // Job titles
    chartJobTitles = buildBarChart(
      chartJobTitles,
      'chartJobTitles',
      topJobTitles.map((t) => t.title),
      topJobTitles.map((t) => t.count),
      'Jobs',
      '#0ea5e9'
    );

    // Skills by pipeline stage – grouped bar chart
    const statuses = ['saved', 'applied', 'interview', 'offer', 'rejected'];
    const skillNamesSet = new Set();
    statuses.forEach((status) => {
      (skillsByStatus[status] || []).forEach((s) => skillNamesSet.add(s.skill));
    });
    const skillNames = Array.from(skillNamesSet);

    const datasetsByStatus = statuses.map((status, idx) => {
      const palette = ['#3b82f6', '#10b981', '#f59e0b', '#6366f1', '#ef4444'];
      const data = skillNames.map((skill) => {
        const list = skillsByStatus[status] || [];
        const match = list.find((s) => s.skill === skill);
        return match ? match.count : 0;
      });
      return {
        label: status.charAt(0).toUpperCase() + status.slice(1),
        data,
        backgroundColor: palette[idx % palette.length],
      };
    });

    const ctxStatus = document.getElementById('chartSkillsByStatus');
    if (ctxStatus) {
      if (chartSkillsByStatus) chartSkillsByStatus.destroy();
      chartSkillsByStatus = new Chart(ctxStatus.getContext('2d'), {
        type: 'bar',
        data: {
          labels: skillNames,
          datasets: datasetsByStatus,
        },
        options: {
          responsive: true,
          interaction: { mode: 'index', intersect: false },
          plugins: {
            legend: { position: 'top' },
            title: { display: false },
          },
          scales: {
            x: {
              ticks: { maxRotation: 60, minRotation: 0, autoSkip: true },
            },
            y: {
              beginAtZero: true,
            },
          },
        },
      });
    }
  } catch (err) {
    console.error('Failed to load analytics', err);
    summaryEl.textContent =
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
