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
	WorkplaceType string
	Status        string
	ExtractedAt   string
	SourceURL     string
}

func (db *DB) ListJobs(limit, offset int, status string) ([]JobSummary, error) {
	query := `
		SELECT id, job_title, company_name, location, workplace_type, status, extracted_at, source_url
		FROM jobs
		WHERE ($1 = '' OR status = $1)
		ORDER BY extracted_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobSummary
	for rows.Next() {
		var job JobSummary
		err := rows.Scan(
			&job.ID, &job.JobTitle, &job.CompanyName, &job.Location,
			&job.WorkplaceType, &job.Status, &job.ExtractedAt, &job.SourceURL,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (db *DB) GetJobByID(id int64) (*models.JobPosting, error) {
	query := `SELECT raw_json, status, notes, rating FROM jobs WHERE id = $1`

	var rawJSON, status, notes string
	var rating sql.NullInt64

	err := db.QueryRow(query, id).Scan(&rawJSON, &status, &notes, &rating)
	if err != nil {
		return nil, err
	}

	var job models.JobPosting
	if err := json.Unmarshal([]byte(rawJSON), &job); err != nil {
		return nil, err
	}

	// Add database-only fields that aren't in raw_json
	// These would need to be added to models.JobPosting if you want them

	return &job, nil
}

func (db *DB) UpdateJobStatus(id int64, status string) error {
	query := `UPDATE jobs SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := db.Exec(query, status, id)
	return err
}

func (db *DB) UpdateJobNotes(id int64, notes string) error {
	query := `UPDATE jobs SET notes = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := db.Exec(query, notes, id)
	return err
}

func (db *DB) UpdateJobRating(id int64, rating int) error {
	query := `UPDATE jobs SET rating = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := db.Exec(query, rating, id)
	return err
}

func (db *DB) SearchJobs(search string) ([]JobSummary, error) {
	query := `
		SELECT id, job_title, company_name, location, workplace_type, status, extracted_at, source_url
		FROM jobs
		WHERE job_title LIKE $1 OR company_name LIKE $1 OR location LIKE $1
		ORDER BY extracted_at DESC
		LIMIT 50
	`

	rows, err := db.Query(query, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobSummary
	for rows.Next() {
		var job JobSummary
		err := rows.Scan(
			&job.ID, &job.JobTitle, &job.CompanyName, &job.Location,
			&job.WorkplaceType, &job.Status, &job.ExtractedAt, &job.SourceURL,
		)
		if err != nil {
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
	err := db.QueryRow(query).Scan(&total, &saved, &applied, &interview, &offer, &rejected)
	if err != nil {
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
