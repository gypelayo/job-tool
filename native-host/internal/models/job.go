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
	SourceURL       string `json:"sourceUrl"` // NEW
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
	WorkArrangement WorkArrangement `json:"work_arrangement"`
	MarketSignals   MarketSignals   `json:"market_signals"`
	ExtractedAt     string          `json:"extracted_at"`
	SourceURL       string          `json:"source_url"`
}

type JobMetadata struct {
	JobTitle       string `json:"job_title"`
	Department     string `json:"department"`
	SeniorityLevel string `json:"seniority_level"` // Junior, Mid, Senior, Staff, Principal, Lead
	JobFunction    string `json:"job_function"`    // Backend, Frontend, FullStack, DevOps, Data
}

type CompanyInfo struct {
	CompanyName     string `json:"company_name"`
	Industry        string `json:"industry"` // Single value now
	CompanySize     string `json:"company_size"`
	LocationFull    string `json:"location_full"`
	LocationCity    string `json:"location_city"`
	LocationCountry string `json:"location_country"`
}

type RoleDetails struct {
	Summary             string   `json:"summary"`
	KeyResponsibilities []string `json:"key_responsibilities"`
	TeamStructure       string   `json:"team_structure"`
}

type Requirements struct {
	YearsExperienceMin     int             `json:"years_experience_min"`
	YearsExperienceMax     int             `json:"years_experience_max"`
	EducationLevel         string          `json:"education_level"`
	RequiresSpecificDegree bool            `json:"requires_specific_degree"`
	TechnicalSkills        TechnicalSkills `json:"technical_skills"`
	SoftSkills             []string        `json:"soft_skills"`
	NiceToHave             []string        `json:"nice_to_have"`
}

type TechnicalSkills struct {
	ProgrammingLanguages []string `json:"programming_languages"`
	Frameworks           []string `json:"frameworks"`
	Databases            []string `json:"databases"`
	CloudPlatforms       []string `json:"cloud_platforms"`
	DevOpsTools          []string `json:"devops_tools"`
	Other                []string `json:"other"`
}

type Compensation struct {
	SalaryMin             int      `json:"salary_min"`
	SalaryMax             int      `json:"salary_max"`
	SalaryCurrency        string   `json:"salary_currency"`
	HasEquity             bool     `json:"has_equity"`
	HasRemoteStipend      bool     `json:"has_remote_stipend"`
	Benefits              []string `json:"benefits"`
	OffersVisa            bool     `json:"offers_visa_sponsorship"`
	OffersHealthInsurance bool     `json:"offers_health_insurance"`
	OffersPTO             bool     `json:"offers_pto"`
	OffersProfDev         bool     `json:"offers_professional_development"`
	Offers401k            bool     `json:"offers_401k"`
}

type WorkArrangement struct {
	WorkplaceType        string `json:"workplace_type"` // Remote, Hybrid, On-site
	JobType              string `json:"job_type"`       // Full-time, Part-time, Contract
	IsRemoteFriendly     bool   `json:"is_remote_friendly"`
	TimezoneRequirements string `json:"timezone_requirements"`
}

type MarketSignals struct {
	UrgencyLevel       string `json:"urgency_level"` // Standard, Urgent, Immediate
	InterviewRounds    int    `json:"interview_rounds"`
	HasTakeHome        bool   `json:"has_take_home"`
	HasPairProgramming bool   `json:"has_pair_programming"`
}
