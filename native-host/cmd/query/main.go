package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // ← ADD THIS LINE

	"native-host/internal/config"
	"native-host/internal/db"
	"native-host/internal/extractor"
	"native-host/internal/messaging"
	"native-host/internal/models"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	// Set up logging
	logFile, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("Native host started")

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		log.Printf("Error creating directories: %v", err)
		return
	}

	// Initialize database
	log.Printf("Initializing database at: %s", cfg.DBPath)
	database, err := db.Init(cfg.DBPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		// Don't return - continue with file save
	} else {
		defer database.Close()
		log.Printf("Database initialized successfully")
	}

	// Read message from extension
	message, err := messaging.ReadMessage(os.Stdin)
	if err != nil {
		log.Printf("Error reading message: %v", err)
		return
	}

	log.Printf("Received %d bytes of text", len(message.Text))
	log.Printf("Provider: %s", message.Settings.Provider)

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Save raw text
	rawFilename := fmt.Sprintf("job_%s_raw.txt", timestamp)
	rawPath := filepath.Join(cfg.OutputDir, rawFilename)
	if err := os.WriteFile(rawPath, []byte(message.Text), 0644); err != nil {
		log.Printf("Error writing raw file: %v", err)
		messaging.SendResponse(models.Response{Status: "error", Filename: ""})
		return
	}
	log.Printf("Saved raw text to %s", rawPath)

	// Extract structured data using configured AI
	log.Printf("Calling %s for structured extraction...", message.Settings.Provider)

	var structuredData *models.JobPosting
	if message.Settings.Provider == "perplexity" {
		structuredData, err = extractor.ExtractWithPerplexity(message.Text, message.Settings)
	} else {
		structuredData, err = extractor.ExtractWithOllama(message.Text, message.Settings)
	}

	if err != nil {
		log.Printf("Error extracting with %s: %v", message.Settings.Provider, err)
		messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	// Save to database
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
		messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		log.Printf("Error writing JSON file: %v", err)
		messaging.SendResponse(models.Response{Status: "error", Filename: rawPath})
		return
	}

	log.Printf("Saved structured data to %s", jsonPath)

	messaging.SendResponse(models.Response{
		Status:   "success",
		Filename: rawPath,
		JsonFile: jsonPath,
	})
}
