package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	Text string `json:"text"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// Setup logging to file
	logFile, _ := os.Create(filepath.Join(os.TempDir(), "textextractor.log"))
	defer logFile.Close()
	log.SetOutput(logFile)

	for {
		msg, err := readMessage(os.Stdin)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Error reading: %v", err)
			continue
		}

		log.Printf("Received text length: %d", len(msg.Text))

		// Get user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting home dir: %v", err)
			homeDir = os.TempDir()
		}

		// Save to Downloads folder - RAW HTML, NO PROCESSING
		downloadsDir := filepath.Join(homeDir, "Downloads")
		filename := filepath.Join(downloadsDir, "extracted_text_"+time.Now().Format("20060102_150405")+".txt")

		// Save raw HTML directly
		err = os.WriteFile(filename, []byte(msg.Text), 0644)

		var response Response
		if err != nil {
			log.Printf("Error writing file: %v", err)
			response = Response{Status: "error", Message: err.Error()}
		} else {
			log.Printf("Successfully saved to: %s", filename)
			response = Response{Status: "success", Message: "Saved to " + filename}
		}

		if err := sendMessage(os.Stdout, response); err != nil {
			log.Printf("Error sending: %v", err)
		}
	}
}

func readMessage(r io.Reader) (*Message, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func sendMessage(w io.Writer, data interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	length := uint32(len(buf))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return err
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}
