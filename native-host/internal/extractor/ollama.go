package extractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"native-host/internal/models"
	"native-host/pkg/utils"
	"net/http"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func ExtractWithOllama(jobText string, settings models.Settings) (*models.JobPosting, error) {
	sourceURL := utils.ExtractURL(jobText)
	prompt := BuildPrompt(jobText, sourceURL)

	model := settings.OllamaModel
	if model == "" {
		model = "qwen2.5:7b"
	}

	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("parse ollama response: %w", err)
	}

	log.Printf("Ollama response length: %d bytes", len(ollamaResp.Response))

	jsonStr := utils.CleanJSONResponse(ollamaResp.Response)

	var jobPosting models.JobPosting
	if err := json.Unmarshal([]byte(jsonStr), &jobPosting); err != nil {
		log.Printf("Failed to parse JSON. Response was: %s", jsonStr)
		return nil, fmt.Errorf("parse job data: %w", err)
	}

	return &jobPosting, nil
}
