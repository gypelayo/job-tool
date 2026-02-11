package extractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"native-host/internal/models"
	"native-host/pkg/utils"
)

type perplexityRequest struct {
	Model    string              `json:"model"`
	Messages []perplexityMessage `json:"messages"`
}

type perplexityMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type perplexityResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func ExtractWithPerplexity(jobText string, settings models.Settings) (*models.JobPosting, error) {
	sourceURL := utils.ExtractURL(jobText)
	prompt := BuildPrompt(jobText, sourceURL)

	model := settings.PerplexityModel
	if model == "" {
		model = "sonar-pro"
	}

	reqBody := perplexityRequest{
		Model: model,
		Messages: []perplexityMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.PerplexityKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call perplexity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("perplexity returned status %d: %s", resp.StatusCode, string(body))
	}

	var perplexityResp perplexityResponse
	if err := json.NewDecoder(resp.Body).Decode(&perplexityResp); err != nil {
		return nil, fmt.Errorf("parse perplexity response: %w", err)
	}

	if len(perplexityResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from perplexity")
	}

	log.Printf("Perplexity response length: %d bytes", len(perplexityResp.Choices[0].Message.Content))

	jsonStr := utils.CleanJSONResponse(perplexityResp.Choices[0].Message.Content)

	var jobPosting models.JobPosting
	if err := json.Unmarshal([]byte(jsonStr), &jobPosting); err != nil {
		log.Printf("Failed to parse JSON. Response was: %s", jsonStr)
		return nil, fmt.Errorf("parse job data: %w", err)
	}

	return &jobPosting, nil
}
