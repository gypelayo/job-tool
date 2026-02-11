package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"native-host/internal/models"
)

type DB struct {
	*sql.DB
}

func Init(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := sqlDB.Exec(Schema); err != nil {
		return nil, fmt.Errorf("execute schema: %w", err)
	}

	return &DB{sqlDB}, nil
}

func (db *DB) SaveJob(job *models.JobPosting) (int64, error) {
	rawJSON, err := json.Marshal(job)
	if err != nil {
		return 0, fmt.Errorf("marshal job: %w", err)
	}

	toJSON := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	}

	query := `
		INSERT INTO jobs (
			source_url, extracted_at, job_title, company_name, department, level,
			job_type, workplace_type, industry, company_size, location, remote_policy,
			summary, key_responsibilities, team_structure, years_of_experience,
			programming_languages, frameworks, databases, cloud_platforms, devops_tools,
			other_skills, soft_skills, education, certifications, nice_to_have,
			salary_range, equity, benefits, bonus_structure, posted_date,
			application_deadline, interview_process, time_to_hire, contact_info,
			raw_json, status
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, 'saved'
		)
		ON CONFLICT(source_url) DO UPDATE SET
			updated_at = CURRENT_TIMESTAMP,
			job_title = excluded.job_title,
			company_name = excluded.company_name,
			raw_json = excluded.raw_json
	`

	result, err := db.Exec(query,
		job.SourceURL, job.ExtractedAt, job.Metadata.JobTitle, job.CompanyInfo.CompanyName,
		job.Metadata.Department, toJSON(job.Metadata.Level), job.Metadata.JobType,
		job.Metadata.WorkplaceType, toJSON(job.CompanyInfo.Industry), job.CompanyInfo.CompanySize,
		job.CompanyInfo.Location, job.CompanyInfo.RemotePolicy, job.RoleDetails.Summary,
		toJSON(job.RoleDetails.KeyResponsibilities), job.RoleDetails.TeamStructure,
		job.Requirements.YearsOfExperience, toJSON(job.Requirements.TechnicalSkills.ProgrammingLanguages),
		toJSON(job.Requirements.TechnicalSkills.Frameworks), toJSON(job.Requirements.TechnicalSkills.Databases),
		toJSON(job.Requirements.TechnicalSkills.CloudPlatforms), toJSON(job.Requirements.TechnicalSkills.DevOpsTools),
		toJSON(job.Requirements.TechnicalSkills.Other), toJSON(job.Requirements.SoftSkills),
		toJSON(job.Requirements.Education), toJSON(job.Requirements.Certifications),
		toJSON(job.Requirements.NiceToHave), job.Compensation.SalaryRange, job.Compensation.Equity,
		toJSON(job.Compensation.Benefits), job.Compensation.BonusStructure, job.ApplicationInfo.PostedDate,
		job.ApplicationInfo.ApplicationDeadline, toJSON(job.ApplicationInfo.InterviewProcess),
		job.ApplicationInfo.TimeToHire, job.ApplicationInfo.ContactInfo, string(rawJSON),
	)

	if err != nil {
		return 0, fmt.Errorf("insert job: %w", err)
	}

	return result.LastInsertId()
}
