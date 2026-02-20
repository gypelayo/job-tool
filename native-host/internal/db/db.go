package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"native-host/internal/models"
	"strings"
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

	summary := job.RoleDetails.Summary
	keyResp := strings.Join(job.RoleDetails.KeyResponsibilities, "\n• ")
	teamStructure := job.RoleDetails.TeamStructure
	benefits := strings.Join(job.Compensation.Benefits, ", ")
	softSkills := strings.Join(job.Requirements.SoftSkills, ", ")
	niceToHave := strings.Join(job.Requirements.NiceToHave, "; ")

	query := `
        INSERT INTO jobs (
            source_url, extracted_at,
            job_title, company_name, company_size, industry,
            location_full, location_city, location_country,
            seniority_level, department, job_function,
            workplace_type, job_type, is_remote_friendly, timezone_requirements,
            years_experience_min, years_experience_max, education_level, requires_specific_degree,
            salary_min, salary_max, salary_currency, has_equity, has_remote_stipend,
            offers_visa_sponsorship, offers_health_insurance, offers_pto,
            offers_professional_development, offers_401k,
            urgency_level, interview_rounds, has_take_home, has_pair_programming,
            summary, key_responsibilities, team_structure, benefits, soft_skills, nice_to_have,
            status, raw_json
        ) VALUES (
            ?, ?,                             -- 1-2
            ?, ?, ?, ?,                       -- 3-6
            ?, ?, ?,                          -- 7-9
            ?, ?, ?,                          -- 10-12
            ?, ?, ?, ?,                       -- 13-16
            ?, ?, ?, ?,                       -- 17-20
            ?, ?, ?, ?, ?,                    -- 21-25
            ?, ?, ?, ?, ?,                    -- 26-30
            ?, ?, ?, ?,                       -- 31-34
            ?, ?, ?, ?, ?, ?,                 -- 35-40
            'saved', ?                        -- status literal, raw_json last
        )
        ON CONFLICT(source_url) DO UPDATE SET
            updated_at = CURRENT_TIMESTAMP,
            job_title = excluded.job_title,
            company_name = excluded.company_name,
            seniority_level = excluded.seniority_level,
            job_function = excluded.job_function,
            salary_min = excluded.salary_min,
            salary_max = excluded.salary_max,
            is_remote_friendly = excluded.is_remote_friendly,
            raw_json = excluded.raw_json
    `

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(query,
		// 1-2
		job.SourceURL, job.ExtractedAt,

		// 3-6
		job.Metadata.JobTitle,
		job.CompanyInfo.CompanyName,
		job.CompanyInfo.CompanySize,
		job.CompanyInfo.Industry,

		// 7-9
		job.CompanyInfo.LocationFull,
		job.CompanyInfo.LocationCity,
		job.CompanyInfo.LocationCountry,

		// 10-12
		job.Metadata.SeniorityLevel,
		job.Metadata.Department,
		job.Metadata.JobFunction,

		// 13-16
		job.WorkArrangement.WorkplaceType,
		job.WorkArrangement.JobType,
		job.WorkArrangement.IsRemoteFriendly,
		job.WorkArrangement.TimezoneRequirements,

		// 17-20
		job.Requirements.YearsExperienceMin,
		job.Requirements.YearsExperienceMax,
		job.Requirements.EducationLevel,
		job.Requirements.RequiresSpecificDegree,

		// 21-25
		job.Compensation.SalaryMin,
		job.Compensation.SalaryMax,
		job.Compensation.SalaryCurrency,
		job.Compensation.HasEquity,
		job.Compensation.HasRemoteStipend,

		// 26-30 (all the boolean benefit flags)
		job.Compensation.OffersVisa,            // offers_visa_sponsorship
		job.Compensation.OffersHealthInsurance, // offers_health_insurance
		job.Compensation.OffersPTO,             // offers_pto
		job.Compensation.OffersProfDev,         // offers_professional_development
		job.Compensation.Offers401k,            // offers_401k

		// 31-34
		job.MarketSignals.UrgencyLevel,
		job.MarketSignals.InterviewRounds,
		job.MarketSignals.HasTakeHome,
		job.MarketSignals.HasPairProgramming,

		// 35-40
		summary,
		keyResp,
		teamStructure,
		benefits,
		softSkills,
		niceToHave,

		// raw_json (last)
		string(rawJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("insert job: %w", err)
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	if err := db.saveSkills(tx, jobID, job.Requirements.TechnicalSkills); err != nil {
		return 0, fmt.Errorf("save skills: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return jobID, nil
}

func (db *DB) saveSkills(tx *sql.Tx, jobID int64, skills models.TechnicalSkills) error {
	_, err := tx.Exec("DELETE FROM job_skills WHERE job_id = ?", jobID)
	if err != nil {
		return err
	}

	insertSkill := func(category string, skillNames []string) error {
		for _, name := range skillNames {
			if name == "" {
				continue
			}
			if _, err := tx.Exec(
				"INSERT INTO job_skills (job_id, skill_name, skill_category, is_required) VALUES (?, ?, ?, ?)",
				jobID, name, category, true,
			); err != nil {
				return err
			}
		}
		return nil
	}

	if err := insertSkill("programming_language", skills.ProgrammingLanguages); err != nil {
		return err
	}
	if err := insertSkill("framework", skills.Frameworks); err != nil {
		return err
	}
	if err := insertSkill("database", skills.Databases); err != nil {
		return err
	}
	if err := insertSkill("cloud", skills.CloudPlatforms); err != nil {
		return err
	}
	if err := insertSkill("devops", skills.DevOpsTools); err != nil {
		return err
	}
	if err := insertSkill("other", skills.Other); err != nil {
		return err
	}

	return nil
}
