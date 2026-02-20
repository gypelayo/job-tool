package db

import (
	"database/sql"
	"encoding/json"
	"native-host/internal/models"
)

type JobSummary struct {
	ID            int64
	JobTitle      string
	CompanyName   string
	Location      string
	JobType       string
	WorkplaceType string
	Level         string
	Department    string
	SalaryRange   string
	Status        string
	ExtractedAt   string
	SourceURL     string
}

// ListJobs uses existing columns: location_full, job_type, workplace_type, etc.
func (db *DB) ListJobs(limit, offset int, status string) ([]JobSummary, error) {
	query := `
        SELECT 
            id, 
            job_title, 
            company_name, 
            location_full, 
            job_type,
            workplace_type, 
            seniority_level,
            department,
            salary_min || '-' || salary_max || ' ' || IFNULL(salary_currency, '') as salary_range,
            status, 
            extracted_at, 
            source_url
        FROM jobs
        WHERE (? = '' OR status = ?)
        ORDER BY extracted_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.Query(query, status, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobSummary
	for rows.Next() {
		var job JobSummary
		var salaryRange sql.NullString

		if err := rows.Scan(
			&job.ID,
			&job.JobTitle,
			&job.CompanyName,
			&job.Location,
			&job.JobType,
			&job.WorkplaceType,
			&job.Level,
			&job.Department,
			&salaryRange,
			&job.Status,
			&job.ExtractedAt,
			&job.SourceURL,
		); err != nil {
			return nil, err
		}

		job.SalaryRange = salaryRange.String
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (db *DB) GetJobByID(id int64) (*models.JobPosting, string, string, int, error) {
	query := `SELECT raw_json, status, notes, rating FROM jobs WHERE id = ?`

	var rawJSON string
	var status sql.NullString
	var notes sql.NullString
	var rating sql.NullInt64

	if err := db.QueryRow(query, id).Scan(&rawJSON, &status, &notes, &rating); err != nil {
		return nil, "", "", 0, err
	}

	var job models.JobPosting
	if err := json.Unmarshal([]byte(rawJSON), &job); err != nil {
		return nil, "", "", 0, err
	}

	return &job, status.String, notes.String, int(rating.Int64), nil
}

func (db *DB) UpdateJobStatus(id int64, status string) error {
	query := `UPDATE jobs SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

func (db *DB) UpdateJobNotes(id int64, notes string) error {
	query := `UPDATE jobs SET notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, notes, id)
	return err
}

func (db *DB) UpdateJobRating(id int64, rating int) error {
	query := `UPDATE jobs SET rating = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, rating, id)
	return err
}

func (db *DB) SearchJobs(search string) ([]JobSummary, error) {
	query := `
		SELECT id, job_title, company_name, location_full, workplace_type, status, extracted_at, source_url
		FROM jobs
		WHERE job_title LIKE ? OR company_name LIKE ? OR location_full LIKE ?
		ORDER BY extracted_at DESC
		LIMIT 50
	`

	like := "%" + search + "%"
	rows, err := db.Query(query, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobSummary
	for rows.Next() {
		var job JobSummary
		if err := rows.Scan(
			&job.ID,
			&job.JobTitle,
			&job.CompanyName,
			&job.Location,
			&job.WorkplaceType,
			&job.Status,
			&job.ExtractedAt,
			&job.SourceURL,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (db *DB) GetJobStats() (map[string]int, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'saved' THEN 1 ELSE 0 END) as saved,
			SUM(CASE WHEN status = 'applied' THEN 1 ELSE 0 END) as applied,
			SUM(CASE WHEN status = 'interview' THEN 1 ELSE 0 END) as interview,
			SUM(CASE WHEN status = 'offer' THEN 1 ELSE 0 END) as offer,
			SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected
		FROM jobs
	`

	var total, saved, applied, interview, offer, rejected int
	if err := db.QueryRow(query).Scan(&total, &saved, &applied, &interview, &offer, &rejected); err != nil {
		return nil, err
	}

	return map[string]int{
		"total":     total,
		"saved":     saved,
		"applied":   applied,
		"interview": interview,
		"offer":     offer,
		"rejected":  rejected,
	}, nil
}
