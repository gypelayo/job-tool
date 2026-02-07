package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type Message struct {
	Text string `json:"text"`
}

type Response struct {
	Status   string `json:"status"`
	Filename string `json:"filename"`
}

func main() {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home dir: %v", err)
		return
	}

	// Set up logging to Downloads folder
	logPath := filepath.Join(homeDir, "Downloads", "extractor.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("Native host started")

	// Read message from extension
	msg, err := readMessage(os.Stdin)
	if err != nil {
		log.Printf("Error reading message: %v", err)
		return
	}

	// Parse message
	var message Message
	if err := json.Unmarshal(msg, &message); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	log.Printf("Received %d bytes of text", len(message.Text))

	// Create output directory in Downloads
	outputDir := filepath.Join(homeDir, "Downloads", "extracted_jobs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("job_%s.html", timestamp)
	fullPath := filepath.Join(outputDir, filename)

	// Save to file
	if err := os.WriteFile(fullPath, []byte(message.Text), 0644); err != nil {
		log.Printf("Error writing file: %v", err)
		sendMessage(Response{Status: "error", Filename: ""})
		return
	}

	log.Printf("Saved to %s", fullPath)

	// Send success response
	sendMessage(Response{Status: "success", Filename: fullPath})
}

// readMessage reads a message from stdin using Chrome native messaging format
func readMessage(reader io.Reader) ([]byte, error) {
	// Read 4-byte length header
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	// Read message body
	message := make([]byte, length)
	if _, err := io.ReadFull(reader, message); err != nil {
		return nil, err
	}

	return message, nil
}

// sendMessage sends a message to stdout using Chrome native messaging format
func sendMessage(response Response) {
	// Marshal response to JSON
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	// Write 4-byte length header
	length := uint32(len(data))
	if err := binary.Write(os.Stdout, binary.LittleEndian, length); err != nil {
		log.Printf("Error writing length: %v", err)
		return
	}

	// Write message body
	if _, err := os.Stdout.Write(data); err != nil {
		log.Printf("Error writing message: %v", err)
		return
	}
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	return reg.ReplaceAllString(name, "_")
}
