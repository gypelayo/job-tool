CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Core identification
    source_url TEXT UNIQUE NOT NULL,
    extracted_at TIMESTAMP NOT NULL,
    
    -- Job metadata
    job_title TEXT,
    company_name TEXT,
    department TEXT,
    level TEXT, -- JSON array as text
    job_type TEXT,
    workplace_type TEXT,
    
    -- Company info
    industry TEXT, -- JSON array as text
    company_size TEXT,
    location TEXT,
    remote_policy TEXT,
    
    -- Role details
    summary TEXT,
    key_responsibilities TEXT, -- JSON array as text
    team_structure TEXT,
    
    -- Requirements
    years_of_experience TEXT,
    programming_languages TEXT, -- JSON array as text
    frameworks TEXT, -- JSON array as text
    databases TEXT, -- JSON array as text
    cloud_platforms TEXT, -- JSON array as text
    devops_tools TEXT, -- JSON array as text
    other_skills TEXT, -- JSON array as text
    soft_skills TEXT, -- JSON array as text
    education TEXT, -- JSON array as text
    certifications TEXT, -- JSON array as text
    nice_to_have TEXT, -- JSON array as text
    
    -- Compensation
    salary_range TEXT,
    equity TEXT,
    benefits TEXT, -- JSON array as text
    bonus_structure TEXT,
    
    -- Application info
    posted_date TEXT,
    application_deadline TEXT,
    interview_process TEXT, -- JSON array as text
    time_to_hire TEXT,
    contact_info TEXT,
    
    -- User tracking fields
    status TEXT DEFAULT 'saved', -- saved, applied, interview, offer, rejected, archived
    applied_date TIMESTAMP,
    notes TEXT,
    rating INTEGER, -- 1-5 stars
    
    -- Full JSON backup
    raw_json TEXT NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_company ON jobs(company_name);
CREATE INDEX IF NOT EXISTS idx_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_location ON jobs(location);
CREATE INDEX IF NOT EXISTS idx_workplace_type ON jobs(workplace_type);
CREATE INDEX IF NOT EXISTS idx_extracted_at ON jobs(extracted_at DESC);

-- Full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS jobs_fts USING fts5(
    job_title,
    company_name,
    summary,
    content=jobs,
    content_rowid=id
);

-- Trigger to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS jobs_fts_insert AFTER INSERT ON jobs BEGIN
    INSERT INTO jobs_fts(rowid, job_title, company_name, summary)
    VALUES (new.id, new.job_title, new.company_name, new.summary);
END;

CREATE TRIGGER IF NOT EXISTS jobs_fts_update AFTER UPDATE ON jobs BEGIN
    UPDATE jobs_fts SET 
        job_title = new.job_title,
        company_name = new.company_name,
        summary = new.summary
    WHERE rowid = new.id;
END;

CREATE TRIGGER IF NOT EXISTS jobs_fts_delete AFTER DELETE ON jobs BEGIN
    DELETE FROM jobs_fts WHERE rowid = old.id;
END;
