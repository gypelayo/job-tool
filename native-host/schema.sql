CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_url TEXT UNIQUE NOT NULL,
    extracted_at TIMESTAMP NOT NULL,
    
    -- Basic Info
    job_title TEXT,
    company_name TEXT,
    company_size TEXT,
    industry TEXT,                    -- Single value, not array
    location TEXT,
    location_country TEXT,            -- NEW: Extract country
    location_city TEXT,               -- NEW: Extract city
    
    -- Role Classification
    seniority_level TEXT,             -- Junior, Mid, Senior, Staff, Principal, Lead
    department TEXT,
    job_function TEXT,                -- NEW: Backend, Frontend, FullStack, DevOps, Data, etc.
    
    -- Work Arrangement
    workplace_type TEXT,              -- Remote, Hybrid, On-site
    job_type TEXT,                    -- Full-time, Part-time, Contract
    is_remote_friendly BOOLEAN,       -- NEW: True if Remote or Hybrid
    timezone_requirements TEXT,       -- NEW: "EMEA", "US hours", "Flexible"
    
    -- Experience & Skills (NORMALIZED)
    years_experience_min INTEGER,     -- NEW: Extract minimum (e.g., "3-5" → 3)
    years_experience_max INTEGER,     -- NEW: Extract maximum (e.g., "3-5" → 5)
    education_level TEXT,             -- "Bachelor's", "Master's", "PhD", "None specified"
    
    -- Compensation (STRUCTURED)
    salary_min INTEGER,               -- NEW: Extract from range
    salary_max INTEGER,               -- NEW: Extract from range
    salary_currency TEXT,             -- NEW: USD, EUR, GBP
    has_equity BOOLEAN,               -- NEW: True if equity mentioned
    has_remote_stipend BOOLEAN,       -- NEW: True if remote equipment/stipend
    
    -- Market Signals
    urgency_level TEXT,               -- "Standard", "Urgent", "Immediate"
    num_required_skills INTEGER,      -- NEW: Count of must-have technical skills
    num_nice_to_have INTEGER,         -- NEW: Count of nice-to-have skills
    requires_specific_degree BOOLEAN, -- NEW: True if degree is required
    
    -- Interview Process
    interview_rounds INTEGER,         -- NEW: Extract number of rounds if mentioned
    has_take_home BOOLEAN,           -- NEW: True if take-home assignment mentioned
    has_pair_programming BOOLEAN,    -- NEW: True if pair programming mentioned
    
    -- Benefits & Perks (BOOLEAN FLAGS for analysis)
    offers_visa_sponsorship BOOLEAN,
    offers_health_insurance BOOLEAN,
    offers_pto BOOLEAN,
    offers_professional_development BOOLEAN,
    offers_401k BOOLEAN,
    
    -- Tracking (you add manually)
    status TEXT DEFAULT 'saved',
    applied_date TIMESTAMP,
    notes TEXT,
    rating INTEGER,
    
    -- Raw data
    raw_json TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- NEW: Normalized skills table for trend analysis
CREATE TABLE IF NOT EXISTS job_skills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    skill_name TEXT NOT NULL,
    skill_category TEXT,  -- programming_language, framework, database, cloud, devops, other
    is_required BOOLEAN DEFAULT 1,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_skills_name ON job_skills(skill_name);
CREATE INDEX IF NOT EXISTS idx_job_skills_category ON job_skills(skill_category);
