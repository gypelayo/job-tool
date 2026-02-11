package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Message struct {
	Text     string   `json:"text"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	Provider        string `json:"provider"`
	OllamaModel     string `json:"ollamaModel"`
	PerplexityKey   string `json:"perplexityKey"`
	PerplexityModel string `json:"perplexityModel"`
}

type Response struct {
	Status   string `json:"status"`
	Filename string `json:"filename"`
	JsonFile string `json:"json_file,omitempty"`
}

// Ollama API types
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Perplexity API types
type PerplexityRequest struct {
	Model    string              `json:"model"`
	Messages []PerplexityMessage `json:"messages"`
}

type PerplexityMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PerplexityResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Structured job data
type JobPosting struct {
	Metadata        JobMetadata     `json:"metadata"`
	CompanyInfo     CompanyInfo     `json:"company_info"`
	RoleDetails     RoleDetails     `json:"role_details"`
	Requirements    Requirements    `json:"requirements"`
	Compensation    Compensation    `json:"compensation"`
	ApplicationInfo ApplicationInfo `json:"application_info"`
	ExtractedAt     string          `json:"extracted_at"`
	SourceURL       string          `json:"source_url"`
}

type JobMetadata struct {
	JobTitle      string   `json:"job_title"`
	Department    string   `json:"department"`
	Level         []string `json:"level"`
	JobType       string   `json:"job_type"`
	WorkplaceType string   `json:"workplace_type"`
}

type CompanyInfo struct {
	CompanyName  string   `json:"company_name"`
	Industry     []string `json:"industry"`
	CompanySize  string   `json:"company_size"`
	Location     string   `json:"location"`
	RemotePolicy string   `json:"remote_policy"`
}

type RoleDetails struct {
	Summary             string   `json:"summary"`
	KeyResponsibilities []string `json:"key_responsibilities"`
	TeamStructure       string   `json:"team_structure"`
}

type Requirements struct {
	YearsOfExperience string          `json:"years_of_experience"`
	TechnicalSkills   TechnicalSkills `json:"technical_skills"`
	SoftSkills        []string        `json:"soft_skills"`
	Education         []string        `json:"education"`
	Certifications    []string        `json:"certifications"`
	NiceToHave        []string        `json:"nice_to_have"`
}

type TechnicalSkills struct {
	ProgrammingLanguages []SkillDetail `json:"programming_languages"`
	Frameworks           []SkillDetail `json:"frameworks"`
	Databases            []SkillDetail `json:"databases"`
	CloudPlatforms       []SkillDetail `json:"cloud_platforms"`
	DevOpsTools          []SkillDetail `json:"devops_tools"`
	Other                []SkillDetail `json:"other"`
}

type SkillDetail struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description,omitempty"`
}

type Compensation struct {
	SalaryRange    string   `json:"salary_range"`
	Equity         string   `json:"equity"`
	Benefits       []string `json:"benefits"`
	BonusStructure string   `json:"bonus_structure"`
}

type ApplicationInfo struct {
	PostedDate          string   `json:"posted_date"`
	ApplicationDeadline string   `json:"application_deadline"`
	InterviewProcess    []string `json:"interview_process"`
	TimeToHire          string   `json:"time_to_hire"`
	ContactInfo         string   `json:"contact_info"`
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home dir: %v", err)
		return
	}

	logPath := filepath.Join(homeDir, "Downloads", "extractor.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("Native host started")

	msg, err := readMessage(os.Stdin)
	if err != nil {
		log.Printf("Error reading message: %v", err)
		return
	}

	var message Message
	if err := json.Unmarshal(msg, &message); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	log.Printf("Received %d bytes of text", len(message.Text))
	log.Printf("Provider: %s", message.Settings.Provider)

	outputDir := filepath.Join(homeDir, "Downloads", "extracted_jobs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Save raw text
	rawFilename := fmt.Sprintf("job_%s_raw.txt", timestamp)
	rawPath := filepath.Join(outputDir, rawFilename)
	if err := os.WriteFile(rawPath, []byte(message.Text), 0644); err != nil {
		log.Printf("Error writing raw file: %v", err)
		sendMessage(Response{Status: "error", Filename: ""})
		return
	}
	log.Printf("Saved raw text to %s", rawPath)

	// Extract structured data using configured AI
	log.Printf("Calling %s for structured extraction...", message.Settings.Provider)

	var structuredData *JobPosting
	if message.Settings.Provider == "perplexity" {
		structuredData, err = extractJobDataPerplexity(message.Text, message.Settings)
	} else {
		structuredData, err = extractJobDataOllama(message.Text, message.Settings)
	}

	if err != nil {
		log.Printf("Error extracting with %s: %v", message.Settings.Provider, err)
		sendMessage(Response{Status: "error", Filename: rawPath})
		return
	}

	// Save structured JSON
	jsonFilename := fmt.Sprintf("job_%s_structured.json", timestamp)
	jsonPath := filepath.Join(outputDir, jsonFilename)

	jsonData, err := json.MarshalIndent(structuredData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		sendMessage(Response{Status: "error", Filename: rawPath})
		return
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		log.Printf("Error writing JSON file: %v", err)
		sendMessage(Response{Status: "error", Filename: rawPath})
		return
	}

	log.Printf("Saved structured data to %s", jsonPath)

	sendMessage(Response{
		Status:   "success",
		Filename: rawPath,
		JsonFile: jsonPath,
	})
}

func extractJobDataOllama(jobText string, settings Settings) (*JobPosting, error) {
	sourceURL := extractURL(jobText)
	prompt := buildPrompt(jobText, sourceURL)

	model := settings.OllamaModel
	if model == "" {
		model = "qwen2.5:7b"
	}

	reqBody := OllamaRequest{
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

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("parse ollama response: %w", err)
	}

	log.Printf("Ollama response length: %d bytes", len(ollamaResp.Response))

	jsonStr := cleanJSONResponse(ollamaResp.Response)

	var jobPosting JobPosting
	if err := json.Unmarshal([]byte(jsonStr), &jobPosting); err != nil {
		log.Printf("Failed to parse JSON. Response was: %s", jsonStr)
		return nil, fmt.Errorf("parse job data: %w", err)
	}

	return &jobPosting, nil
}

func extractJobDataPerplexity(jobText string, settings Settings) (*JobPosting, error) {
	sourceURL := extractURL(jobText)
	prompt := buildPrompt(jobText, sourceURL)

	model := settings.PerplexityModel
	if model == "" {
		model = "sonar-pro"
	}

	reqBody := PerplexityRequest{
		Model: model,
		Messages: []PerplexityMessage{
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

	var perplexityResp PerplexityResponse
	if err := json.NewDecoder(resp.Body).Decode(&perplexityResp); err != nil {
		return nil, fmt.Errorf("parse perplexity response: %w", err)
	}

	if len(perplexityResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from perplexity")
	}

	log.Printf("Perplexity response length: %d bytes", len(perplexityResp.Choices[0].Message.Content))

	jsonStr := cleanJSONResponse(perplexityResp.Choices[0].Message.Content)

	var jobPosting JobPosting
	if err := json.Unmarshal([]byte(jsonStr), &jobPosting); err != nil {
		log.Printf("Failed to parse JSON. Response was: %s", jsonStr)
		return nil, fmt.Errorf("parse job data: %w", err)
	}

	return &jobPosting, nil
}

func buildPrompt(jobText string, sourceURL string) string {
	return fmt.Sprintf(`Extract job posting information into JSON. Be extremely precise - only extract what is EXPLICITLY stated.

Job Posting:
%s

Key instructions:
1. years_of_experience: ONLY if posting says "5+ years", "3-5 years", etc. Use empty string "" if not stated.
2. job_type: "Full-time", "Part-time", "Contract", or "Internship" (work schedule, not location)
3. workplace_type: "Remote", "Hybrid", or "On-site" (location flexibility)
4. soft_skills: Use SHORT keywords only. Examples: "Communication", "Problem-solving", "Leadership", "Adaptability"
   BAD: "Effective communicator who can explain technical concepts"
   GOOD: "Communication", "Technical writing"
5. For skills with notes, put notes in "description", keep "level" empty unless you can infer from context like "expert in", "5+ years of"

Return this JSON structure:
{
  "metadata": {
    "job_title": "exact title",
    "department": "only if stated",
    "level": ["Senior"],
    "job_type": "Full-time",
    "workplace_type": "Remote"
  },
  "company_info": {
    "company_name": "exact name",
    "industry": ["Marketplace", "E-commerce", "Classifieds"],
    "company_size": "1000+",
    "location": "Portugal",
    "remote_policy": "remote work details"
  },
  "role_details": {
    "summary": "1-2 sentence summary",
    "key_responsibilities": ["exact bullet points"],
    "team_structure": "team info"
  },
  "requirements": {
    "years_of_experience": "ONLY if explicitly stated like '5+ years'. Otherwise empty string",
    "technical_skills": {
      "programming_languages": [{"name": "Go", "level": "", "description": "context if mentioned"}],
      "frameworks": [],
      "databases": [{"name": "MySQL", "level": "", "description": ""}],
      "cloud_platforms": [{"name": "AWS", "level": "", "description": ""}],
      "devops_tools": [{"name": "Terraform", "level": "", "description": ""}],
      "other": []
    },
    "soft_skills": ["Communication", "Problem-solving", "Leadership"],
    "education": [],
    "certifications": [],
    "nice_to_have": []
  },
  "compensation": {
    "salary_range": "only exact range if stated",
    "equity": "",
    "benefits": ["exact benefits"],
    "bonus_structure": ""
  },
  "application_info": {
    "posted_date": "",
    "application_deadline": "",
    "interview_process": ["steps"],
    "time_to_hire": "",
    "contact_info": ""
  },
  "extracted_at": "%s",
  "source_url": "%s"
}

CRITICAL: Do NOT infer years of experience. Do NOT copy full requirement sentences into soft_skills. Return ONLY JSON.`, jobText, time.Now().Format("2006-01-02T15:04:05Z07:00"), sourceURL)
}

func cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response)
}

func extractURL(text string) string {
	// Look for URL: pattern in the text
	if idx := strings.Index(text, "URL:"); idx != -1 {
		urlLine := text[idx:]
		if endIdx := strings.Index(urlLine, "\n"); endIdx != -1 {
			urlLine = urlLine[:endIdx]
		}
		url := strings.TrimSpace(strings.TrimPrefix(urlLine, "URL:"))
		return url
	}
	return ""
}

func readMessage(reader io.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	message := make([]byte, length)
	if _, err := io.ReadFull(reader, message); err != nil {
		return nil, err
	}

	return message, nil
}

func sendMessage(response Response) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	length := uint32(len(data))
	if err := binary.Write(os.Stdout, binary.LittleEndian, length); err != nil {
		log.Printf("Error writing length: %v", err)
		return
	}

	if _, err := os.Stdout.Write(data); err != nil {
		log.Printf("Error writing message: %v", err)
		return
	}
}
