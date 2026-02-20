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

func (db *DB) DeleteJob(id int64) error {
	// Also delete from job_skills to keep it clean
	_, err := db.Exec(`DELETE FROM job_skills WHERE job_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM jobs WHERE id = ?`, id)
	return err
}

// SkillSummary is used for analytics responses.
type SkillSummary struct {
	SkillName     string
	SkillCategory string
	Count         int
}

func (db *DB) GetTopSkills(limit int) ([]SkillSummary, error) {
	query := `
        SELECT 
            skill_name,
            skill_category,
            COUNT(*) AS cnt
        FROM job_skills
        GROUP BY skill_name, skill_category
        ORDER BY cnt DESC, skill_name ASC
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []SkillSummary
	for rows.Next() {
		var s SkillSummary
		if err := rows.Scan(&s.SkillName, &s.SkillCategory, &s.Count); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

func (db *DB) GetSkillLocations(skill string, limit int) ([]struct {
	Location string
	Count    int
}, error) {
	query := `
        SELECT 
            COALESCE(j.location_city, j.location_full, 'Unknown') AS loc,
            COUNT(*) AS cnt
        FROM job_skills s
        JOIN jobs j ON j.id = s.job_id
        WHERE s.skill_name = ?
        GROUP BY COALESCE(j.location_city, j.location_full, 'Unknown')
        ORDER BY cnt DESC
        LIMIT ?
    `
	rows, err := db.Query(query, skill, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []struct {
		Location string
		Count    int
	}

	for rows.Next() {
		var loc string
		var cnt int
		if err := rows.Scan(&loc, &cnt); err != nil {
			return nil, err
		}
		res = append(res, struct {
			Location string
			Count    int
		}{Location: loc, Count: cnt})
	}
	return res, nil
}

// GetTopSkillsByCategory returns top skills per category.
func (db *DB) GetTopSkillsByCategory(category string, limit int) ([]SkillSummary, error) {
	query := `
        SELECT 
            skill_name,
            skill_category,
            COUNT(*) AS cnt
        FROM job_skills
        WHERE skill_category = ?
        GROUP BY skill_name, skill_category
        ORDER BY cnt DESC, skill_name ASC
        LIMIT ?
    `
	rows, err := db.Query(query, category, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []SkillSummary
	for rows.Next() {
		var s SkillSummary
		if err := rows.Scan(&s.SkillName, &s.SkillCategory, &s.Count); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

// GetSkillsByStatus returns top skills per pipeline stage.
func (db *DB) GetSkillsByStatus(limitPerStatus int) (map[string][]SkillSummary, error) {
	query := `
        SELECT 
            j.status,
            s.skill_name,
            s.skill_category,
            COUNT(*) AS cnt
        FROM job_skills s
        JOIN jobs j ON j.id = s.job_id
        GROUP BY j.status, s.skill_name, s.skill_category
        ORDER BY j.status, cnt DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]SkillSummary)

	for rows.Next() {
		var status, name, cat string
		var cnt int
		if err := rows.Scan(&status, &name, &cat, &cnt); err != nil {
			return nil, err
		}

		list := result[status]
		if len(list) < limitPerStatus {
			list = append(list, SkillSummary{
				SkillName:     name,
				SkillCategory: cat,
				Count:         cnt,
			})
			result[status] = list
		}
	}

	return result, nil
}

// JobTitleSummary is used for job title frequency.
type JobTitleSummary struct {
	Title string
	Count int
}

// GetTopJobTitles returns most frequent job titles.
func (db *DB) GetTopJobTitles(limit int) ([]JobTitleSummary, error) {
	query := `
        SELECT 
            job_title,
            COUNT(*) AS cnt
        FROM jobs
        WHERE job_title IS NOT NULL AND job_title != ''
        GROUP BY job_title
        ORDER BY cnt DESC, job_title ASC
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []JobTitleSummary
	for rows.Next() {
		var jt JobTitleSummary
		if err := rows.Scan(&jt.Title, &jt.Count); err != nil {
			return nil, err
		}
		res = append(res, jt)
	}
	return res, nil
}
