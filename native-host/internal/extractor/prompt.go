package extractor

import (
	"fmt"
	"time"
)

func BuildPrompt(jobText, sourceURL string) string {
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
