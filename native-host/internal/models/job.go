package models

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
