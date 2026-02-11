package db

const Schema = `
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_url TEXT UNIQUE NOT NULL,
    extracted_at TIMESTAMP NOT NULL,
    job_title TEXT,
    company_name TEXT,
    department TEXT,
    level TEXT,
    job_type TEXT,
    workplace_type TEXT,
    industry TEXT,
    company_size TEXT,
    location TEXT,
    remote_policy TEXT,
    summary TEXT,
    key_responsibilities TEXT,
    team_structure TEXT,
    years_of_experience TEXT,
    programming_languages TEXT,
    frameworks TEXT,
    databases TEXT,
    cloud_platforms TEXT,
    devops_tools TEXT,
    other_skills TEXT,
    soft_skills TEXT,
    education TEXT,
    certifications TEXT,
    nice_to_have TEXT,
    salary_range TEXT,
    equity TEXT,
    benefits TEXT,
    bonus_structure TEXT,
    posted_date TEXT,
    application_deadline TEXT,
    interview_process TEXT,
    time_to_hire TEXT,
    contact_info TEXT,
    status TEXT DEFAULT 'saved',
    applied_date TIMESTAMP,
    notes TEXT,
    rating INTEGER,
    raw_json TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_company ON jobs(company_name);
CREATE INDEX IF NOT EXISTS idx_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_workplace_type ON jobs(workplace_type);
CREATE INDEX IF NOT EXISTS idx_extracted_at ON jobs(extracted_at DESC);
`
