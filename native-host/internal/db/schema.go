package db

const Schema = `
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_url TEXT UNIQUE NOT NULL,
    extracted_at TIMESTAMP NOT NULL,
    
    -- Basic Info
    job_title TEXT,
    company_name TEXT,
    company_size TEXT,
    industry TEXT,
    location_full TEXT,
    location_city TEXT,
    location_country TEXT,
    
    -- Role Classification
    seniority_level TEXT,
    department TEXT,
    job_function TEXT,
    
    -- Work Arrangement
    workplace_type TEXT,
    job_type TEXT,
    is_remote_friendly BOOLEAN,
    timezone_requirements TEXT,
    
    -- Experience & Skills
    years_experience_min INTEGER,
    years_experience_max INTEGER,
    education_level TEXT,
    requires_specific_degree BOOLEAN,
    
    -- Compensation
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency TEXT,
    has_equity BOOLEAN,
    has_remote_stipend BOOLEAN,
    offers_visa_sponsorship BOOLEAN,
    offers_health_insurance BOOLEAN,
    offers_pto BOOLEAN,
    offers_professional_development BOOLEAN,
    offers_401k BOOLEAN,
    
    -- Market Signals
    urgency_level TEXT,
    interview_rounds INTEGER,
    has_take_home BOOLEAN,
    has_pair_programming BOOLEAN,
    
    -- Aggregated data
    summary TEXT,
    key_responsibilities TEXT,
    team_structure TEXT,
    benefits TEXT,
    soft_skills TEXT,
    nice_to_have TEXT,
    
    -- Tracking
    status TEXT DEFAULT 'saved',
    applied_date TIMESTAMP,
    notes TEXT,
    rating INTEGER,
    
    -- Raw
    raw_json TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_skills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    skill_name TEXT NOT NULL,
    skill_category TEXT,
    is_required BOOLEAN DEFAULT 1,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_company ON jobs(company_name);
CREATE INDEX IF NOT EXISTS idx_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_workplace_type ON jobs(workplace_type);
CREATE INDEX IF NOT EXISTS idx_seniority ON jobs(seniority_level);
CREATE INDEX IF NOT EXISTS idx_is_remote ON jobs(is_remote_friendly);
CREATE INDEX IF NOT EXISTS idx_extracted_at ON jobs(extracted_at DESC);
CREATE INDEX IF NOT EXISTS idx_job_skills_name ON job_skills(skill_name);
CREATE INDEX IF NOT EXISTS idx_job_skills_category ON job_skills(skill_category);
`
