package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"native-host/internal/config"
	"native-host/internal/db"
	"native-host/internal/extractor"
	"native-host/internal/messaging"
	"native-host/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// Logging
	logFile, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}
	log.Println("Native host started")

	if err := cfg.EnsureDirectories(); err != nil {
		log.Printf("Error creating directories: %v", err)
		return
	}

	log.Printf("Initializing database at: %s", cfg.DBPath)
	database, err := db.Init(cfg.DBPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		// we still allow extract to write files even if DB fails
	} else {
		defer database.Close()
		log.Printf("Database initialized successfully")
	}

	// First, read the raw JSON to decide which kind of message it is
	// We cannot re-use ReadMessage here because it already unmarshals into models.Message.
	// We'll read the length + bytes manually, then try to decode into APIRequest first.
	var length uint32
	if err := binary.Read(os.Stdin, binary.LittleEndian, &length); err != nil {
		log.Printf("Error reading length: %v", err)
		return
	}
	msgBytes := make([]byte, length)
	if _, err := io.ReadFull(os.Stdin, msgBytes); err != nil {
		log.Printf("Error reading message bytes: %v", err)
		return
	}

	// Try APIRequest first
	var apiReq messaging.APIRequest
	if err := json.Unmarshal(msgBytes, &apiReq); err == nil && apiReq.Action != "" {
		handleAPIRequest(apiReq, database)
		return
	}

	// Fallback: treat as old extraction Message
	var legacyMsg models.Message
	if err := json.Unmarshal(msgBytes, &legacyMsg); err != nil {
		log.Printf("Error unmarshaling legacy message: %v", err)
		return
	}
	handleExtractMessage(legacyMsg, cfg, database)
}

// HandleExtractMessage runs the full legacy extraction flow:
// save raw text, call Perplexity/Ollama, save to DB, save JSON, send Response.
func handleExtractMessage(message models.Message, cfg *config.Config, database *db.DB) {
	log.Printf("Received %d bytes of text", len(message.Text))
	log.Printf("Provider: %s", message.Settings.Provider)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Save raw text
	rawFilename := fmt.Sprintf("job_%s_raw.txt", timestamp)
	rawPath := filepath.Join(cfg.OutputDir, rawFilename)
	if err := os.WriteFile(rawPath, []byte(message.Text), 0644); err != nil {
		log.Printf("Error writing raw file: %v", err)
		_ = messaging.SendResponse(models.Response{Status: "error", Filename: ""})
		return
	}
	log.Printf("Saved raw text to %s", rawPath)

	// Extract structured data
	log.Printf("Calling %s for structured extraction...", message.Settings.Provider)

	var (
		structuredData *models.JobPosting
		err            error
	)

	if message.Settings.Provider == "perplexity" {
		structuredData, err = extractor.ExtractWithPerplexity(message.Text, message.Settings)
	} else {
		structuredData, err = extractor.ExtractWithOllama(message.Text, message.Settings)
	}
	if message.Settings.SourceURL != "" {
		structuredData.SourceURL = message.Settings.SourceURL
	}

	if err != nil {
		log.Printf("Error extracting with %s: %v", message.Settings.Provider, err)
		_ = messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	// Save to database (if available)
	if database != nil {
		log.Printf("Attempting to save job to database...")
		jobID, err := database.SaveJob(structuredData)
		if err != nil {
			log.Printf("Error saving to database: %v", err)
		} else {
			log.Printf("Saved to database with ID: %d", jobID)
		}
	} else {
		log.Printf("Database not initialized, skipping save")
	}

	// Save structured JSON
	jsonFilename := fmt.Sprintf("job_%s_structured.json", timestamp)
	jsonPath := filepath.Join(cfg.OutputDir, jsonFilename)

	jsonData, err := json.MarshalIndent(structuredData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		_ = messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		log.Printf("Error writing JSON file: %v", err)
		_ = messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	log.Printf("Saved structured data to %s", jsonPath)

	_ = messaging.SendResponse(models.Response{
		Status:   "success",
		Filename: rawPath,
		JsonFile: jsonPath,
	})
}

func handleAPIRequest(req messaging.APIRequest, database *db.DB) {
	if database == nil {
		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:    false,
			Error: "database not initialized",
		})
		return
	}

	switch req.Action {
	case "ping":
		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:      true,
			Payload: map[string]any{"ok": true},
		})

	case "deleteJob":
		idF, ok := req.Data["id"].(float64)
		if !ok {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: "missing id"})
			return
		}
		id := int64(idF)

		if err := database.DeleteJob(id); err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
			return
		}

		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:      true,
			Payload: map[string]any{"deleted": true},
		})

	case "listJobs":
		jobs, err := database.ListJobs(100, 0, "")
		if err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
			return
		}

		summaries := make([]map[string]any, 0, len(jobs))
		for _, j := range jobs {
			summaries = append(summaries, map[string]any{
				"id":            j.ID,
				"title":         j.JobTitle,
				"company":       j.CompanyName,
				"location":      j.Location, // this is location_full from company_info
				"job_type":      j.JobType,
				"workplaceType": j.WorkplaceType,
				"level":         j.Level,
				"department":    j.Department,
				"salaryRange":   j.SalaryRange,
				"status":        j.Status,
				"extractedAt":   j.ExtractedAt,
				"url":           j.SourceURL, // original link available in list
			})
		}

		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:      true,
			Payload: map[string]any{"jobs": summaries},
		})

	case "getJob":
		// id comes from JSON -> float64
		idF, ok := req.Data["id"].(float64)
		if !ok {
			_ = messaging.SendAPIResponse(messaging.APIResponse{
				OK:    false,
				Error: "missing or invalid id",
			})
			return
		}
		id := int64(idF)

		// Get full JobPosting + status/notes/rating from DB
		job, status, notes, rating, err := database.GetJobByID(id)
		if err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		// Flatten technical skills into a single slice
		var skills []string
		ts := job.Requirements.TechnicalSkills
		skills = append(skills, ts.ProgrammingLanguages...)
		skills = append(skills, ts.Frameworks...)
		skills = append(skills, ts.Databases...)
		skills = append(skills, ts.CloudPlatforms...)
		skills = append(skills, ts.DevOpsTools...)
		skills = append(skills, ts.Other...)

		respJob := map[string]any{
			"id":       id,
			"title":    job.Metadata.JobTitle,
			"company":  job.CompanyInfo.CompanyName,
			"location": job.CompanyInfo.LocationFull,
			"url":      job.SourceURL, // original link from extracted data

			"status": status,
			"notes":  notes,
			"rating": rating,
			"skills": skills,

			// full extracted JSON structure
			"extracted": job,
		}

		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:      true,
			Payload: map[string]any{"job": respJob},
		})

	case "updateJob":
		idF, ok := req.Data["id"].(float64)
		if !ok {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: "missing id"})
			return
		}
		id := int64(idF)

		if status, ok := req.Data["status"].(string); ok && status != "" {
			if err := database.UpdateJobStatus(id, status); err != nil {
				_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
				return
			}
		}
		if notes, ok := req.Data["notes"].(string); ok {
			if err := database.UpdateJobNotes(id, notes); err != nil {
				_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
				return
			}
		}

		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:      true,
			Payload: map[string]any{"updated": true},
		})

	case "getAnalytics":
		statusStats, err := database.GetJobStats()
		if err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
			return
		}

		topSkills, err := database.GetTopSkills(15)
		if err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
			return
		}

		skillsByStatus, err := database.GetSkillsByStatus(5)
		if err != nil {
			_ = messaging.SendAPIResponse(messaging.APIResponse{OK: false, Error: err.Error()})
			return
		}

		// Optionally pick one “focus” skill: the most frequent one
		focusSkill := ""
		if len(topSkills) > 0 {
			focusSkill = topSkills[0].SkillName
		}

		var focusSkillLocations []map[string]any
		if focusSkill != "" {
			locs, err := database.GetSkillLocations(focusSkill, 10)
			if err == nil {
				for _, l := range locs {
					focusSkillLocations = append(focusSkillLocations, map[string]any{
						"location": l.Location,
						"count":    l.Count,
					})
				}
			}
		}

		// Convert SkillSummary slices to plain maps for JSON
		topSkillsPayload := make([]map[string]any, 0, len(topSkills))
		for _, s := range topSkills {
			topSkillsPayload = append(topSkillsPayload, map[string]any{
				"skill":    s.SkillName,
				"category": s.SkillCategory,
				"count":    s.Count,
			})
		}

		skillsByStatusPayload := make(map[string][]map[string]any)
		for status, list := range skillsByStatus {
			arr := make([]map[string]any, 0, len(list))
			for _, s := range list {
				arr = append(arr, map[string]any{
					"skill":    s.SkillName,
					"category": s.SkillCategory,
					"count":    s.Count,
				})
			}
			skillsByStatusPayload[status] = arr
		}

		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK: true,
			Payload: map[string]any{
				"statusStats":         statusStats,
				"topSkills":           topSkillsPayload,
				"skillsByStatus":      skillsByStatusPayload,
				"focusSkill":          focusSkill,
				"focusSkillLocations": focusSkillLocations,
			},
		})

	default:
		_ = messaging.SendAPIResponse(messaging.APIResponse{
			OK:    false,
			Error: "unknown action: " + req.Action,
		})
	}
}
