package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	baseURL    = "http://metabase:3000"
	setupToken = "my-setup-token-12345"
	adminEmail = "admin@example.com"  // ← Valid email format
	adminPass  = "SecurePassword123!" // ← Stronger password
)

type MetabaseClient struct {
	baseURL string
	session string
}

func main() {
	log.Println("🚀 Starting Metabase setup...")

	// Wait for Metabase to be ready
	waitForMetabase()

	client := &MetabaseClient{baseURL: baseURL}

	// Setup admin user
	if err := client.setupAdmin(); err != nil {
		log.Printf("ℹ️  Admin setup skipped (may already exist): %v", err)
	}

	// Login
	if err := client.login(); err != nil {
		log.Fatalf("❌ Login failed: %v", err)
	}
	log.Println("✅ Logged in successfully")

	// Create database connection
	dbID, err := client.createDatabase()
	if err != nil {
		log.Fatalf("❌ Database creation failed: %v", err)
	}
	log.Printf("✅ Database created (ID: %d)", dbID)

	// Sync database
	if err := client.syncDatabase(dbID); err != nil {
		log.Printf("⚠️  Database sync failed: %v", err)
	}
	time.Sleep(5 * time.Second) // Wait for sync

	// Create queries and dashboard
	dashboardID, err := client.createSkillsDashboard(dbID)
	if err != nil {
		log.Fatalf("❌ Dashboard creation failed: %v", err)
	}
	log.Printf("✅ Skills Dashboard created (ID: %d)", dashboardID)

	log.Println("🎉 Metabase setup complete!")
	log.Printf("📊 Access your dashboard at: %s/dashboard/%d", baseURL, dashboardID)
}

func waitForMetabase() {
	log.Println("⏳ Waiting for Metabase to be ready...")
	for i := 0; i < 60; i++ {
		resp, err := http.Get(baseURL + "/api/health")
		if err == nil && resp.StatusCode == 200 {
			log.Println("✅ Metabase is ready")
			return
		}
		time.Sleep(2 * time.Second)
	}
	log.Fatal("❌ Metabase failed to start")
}

func (c *MetabaseClient) setupAdmin() error {
	data := map[string]interface{}{
		"token": setupToken,
		"user": map[string]string{
			"first_name": "Admin",
			"last_name":  "User",
			"email":      adminEmail,
			"password":   adminPass,
		},
		"prefs": map[string]string{
			"site_name": "Job Tracker Analytics",
		},
	}

	_, err := c.post("/api/setup", data, "")
	return err
}

func (c *MetabaseClient) login() error {
	data := map[string]string{
		"username": adminEmail,
		"password": adminPass,
	}

	body, err := c.post("/api/session", data, "")
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	c.session = result["id"].(string)
	return nil
}

func (c *MetabaseClient) createDatabase() (int, error) {
	data := map[string]interface{}{
		"engine": "sqlite",
		"name":   "Job Tracker",
		"details": map[string]string{
			"db": "/data/jobs.db",
		},
	}

	body, err := c.post("/api/database", data, c.session)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	return int(result["id"].(float64)), nil
}

func (c *MetabaseClient) syncDatabase(dbID int) error {
	_, err := c.post(fmt.Sprintf("/api/database/%d/sync_schema", dbID), nil, c.session)
	return err
}

func (c *MetabaseClient) createSkillsDashboard(dbID int) (int, error) {
	dashData := map[string]interface{}{
		"name":        "Skills Intelligence Dashboard",
		"description": "Analysis of skills demand from job postings",
	}

	body, err := c.post("/api/dashboard", dashData, c.session)
	if err != nil {
		return 0, err
	}

	var dashResult map[string]interface{}
	if err := json.Unmarshal(body, &dashResult); err != nil {
		return 0, err
	}
	dashboardID := int(dashResult["id"].(float64))

	cards := []struct {
		name   string
		sql    string
		viz    string
		row    int
		col    int
		width  int
		height int
	}{
		{
			name: "Top 20 Most In-Demand Skills",
			sql: `SELECT 
				skill_name,
				COUNT(*) as job_count,
				ROUND(COUNT(*) * 100.0 / (SELECT COUNT(DISTINCT job_id) FROM job_skills), 1) as percentage
			FROM job_skills
			GROUP BY skill_name
			ORDER BY job_count DESC
			LIMIT 20`,
			viz:    "bar",
			row:    0,
			col:    0,
			width:  8,
			height: 4,
		},
		{
			name: "Skills by Category",
			sql: `SELECT 
				skill_category,
				COUNT(DISTINCT skill_name) as unique_skills,
				COUNT(*) as total_mentions
			FROM job_skills
			GROUP BY skill_category
			ORDER BY total_mentions DESC`,
			viz:    "pie",
			row:    0,
			col:    8,
			width:  4,
			height: 4,
		},
		{
			name: "Top Programming Languages",
			sql: `SELECT 
				skill_name,
				COUNT(*) as mentions
			FROM job_skills
			WHERE skill_category = 'programming_language'
			GROUP BY skill_name
			ORDER BY mentions DESC
			LIMIT 10`,
			viz:    "bar",
			row:    4,
			col:    0,
			width:  6,
			height: 4,
		},
		{
			name: "Cloud Platforms Demand",
			sql: `SELECT 
				skill_name,
				COUNT(*) as mentions
			FROM job_skills
			WHERE skill_category = 'cloud'
			GROUP BY skill_name
			ORDER BY mentions DESC`,
			viz:    "bar",
			row:    4,
			col:    6,
			width:  6,
			height: 4,
		},
		{
			name: "Common Skill Pairs",
			sql: `SELECT 
				s1.skill_name as skill_1,
				s2.skill_name as skill_2,
				COUNT(*) as times_together
			FROM job_skills s1
			JOIN job_skills s2 ON s1.job_id = s2.job_id AND s1.skill_name < s2.skill_name
			GROUP BY s1.skill_name, s2.skill_name
			HAVING times_together > 1
			ORDER BY times_together DESC
			LIMIT 20`,
			viz:    "table",
			row:    8,
			col:    0,
			width:  12,
			height: 4,
		},
	}

	// Create all cards first
	dashcards := []map[string]interface{}{}
	for _, card := range cards {
		cardID, err := c.createCard(dbID, card.name, card.sql, card.viz)
		if err != nil {
			log.Printf("⚠️  Failed to create card '%s': %v", card.name, err)
			continue
		}
		log.Printf("✅ Created card: %s (ID: %d)", card.name, cardID)

		// Prepare dashcard data
		dashcards = append(dashcards, map[string]interface{}{
			"id":      -1, // -1 for new dashcards
			"card_id": cardID,
			"row":     card.row,
			"col":     card.col,
			"size_x":  card.width,
			"size_y":  card.height,
		})
	}

	// Update dashboard with all cards at once using PUT
	if len(dashcards) > 0 {
		updatePayload := map[string]interface{}{
			"cards": dashcards,
		}

		_, err = c.put(fmt.Sprintf("/api/dashboard/%d", dashboardID), updatePayload, c.session)
		if err != nil {
			log.Printf("⚠️  Failed to add cards to dashboard: %v", err)
		} else {
			log.Printf("✅ Added %d cards to dashboard", len(dashcards))
		}
	}

	return dashboardID, nil
}

func (c *MetabaseClient) createCard(dbID int, name, sql, vizType string) (int, error) {
	data := map[string]interface{}{
		"name": name,
		"dataset_query": map[string]interface{}{
			"type":     "native",
			"native":   map[string]string{"query": sql},
			"database": dbID,
		},
		"display":                vizType,
		"visualization_settings": map[string]interface{}{},
	}

	body, err := c.post("/api/card", data, c.session)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	return int(result["id"].(float64)), nil
}

func (c *MetabaseClient) addCardToDashboard(dashboardID, cardID, row, col, width, height int) error {
	// Updated API endpoint - use PUT instead of POST
	data := map[string]interface{}{
		"cards": []map[string]interface{}{
			{
				"id":      cardID,
				"card_id": cardID,
				"row":     row,
				"col":     col,
				"size_x":  width,
				"size_y":  height,
			},
		},
	}

	_, err := c.put(fmt.Sprintf("/api/dashboard/%d", dashboardID), data, c.session)
	return err
}

func (c *MetabaseClient) put(path string, data interface{}, session string) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("PUT", c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if session != "" {
		req.Header.Set("X-Metabase-Session", session)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *MetabaseClient) post(path string, data interface{}, session string) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if session != "" {
		req.Header.Set("X-Metabase-Session", session)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
