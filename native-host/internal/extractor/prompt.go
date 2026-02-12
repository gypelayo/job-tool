package extractor

import (
	"fmt"
	"time"
)

func BuildPrompt(jobText, sourceURL string) string {
	return fmt.Sprintf(`Extract job posting information into structured JSON for analytics. Extract ONLY what is explicitly stated.

Job Posting:
%s

Return this JSON structure:
{
  "metadata": {
    "job_title": "exact title from posting",
    "department": "Engineering, Product, Sales, etc.",
    "seniority_level": "Junior|Mid|Senior|Staff|Principal|Lead",
    "job_function": "Backend|Frontend|FullStack|DevOps|Data|Mobile|Security|Embedded"
  },
  "company_info": {
    "company_name": "exact company name",
    "industry": "single primary industry: SaaS, E-commerce, Finance, Healthcare, etc.",
    "company_size": "10-50, 50-200, 200-1000, 1000+, or empty",
    "location_full": "full location as stated",
    "location_city": "extract city name",
    "location_country": "extract country name or region (e.g., USA, UK, EMEA, Remote)"
  },
  "role_details": {
    "summary": "1-2 sentence role summary",
    "key_responsibilities": ["extract exact bullet points"],
    "team_structure": "team info if mentioned"
  },
  "requirements": {
    "years_experience_min": 0,
    "years_experience_max": 0,
    "education_level": "None|Bachelor's|Master's|PhD",
    "requires_specific_degree": false,
    "technical_skills": {
      "programming_languages": ["Go", "Python"],
      "frameworks": ["React", "Django"],
      "databases": ["PostgreSQL", "Redis"],
      "cloud_platforms": ["AWS", "GCP", "Azure"],
      "devops_tools": ["Docker", "Kubernetes", "Terraform"],
      "other": ["Git", "Linux"]
    },
    "soft_skills": ["Communication", "Problem-solving"],
    "nice_to_have": ["skill or experience that's nice to have"]
  },
  "compensation": {
    "salary_min": 0,
    "salary_max": 0,
    "salary_currency": "USD|EUR|GBP|empty",
    "has_equity": false,
    "has_remote_stipend": false,
    "benefits": ["401k", "health insurance"],
    "offers_visa_sponsorship": false,
    "offers_health_insurance": false,
    "offers_pto": false,
    "offers_professional_development": false,
    "offers_401k": false
  },
  "work_arrangement": {
    "workplace_type": "Remote|Hybrid|On-site",
    "job_type": "Full-time|Part-time|Contract|Internship",
    "is_remote_friendly": true,
    "timezone_requirements": "EMEA|US|APAC|Flexible|empty"
  },
  "market_signals": {
    "urgency_level": "Standard|Urgent|Immediate",
    "interview_rounds": 0,
    "has_take_home": false,
    "has_pair_programming": false
  },
  "extracted_at": "%s",
  "source_url": "%s"
}

CRITICAL EXTRACTION RULES:
1. years_experience_min/max: Extract numbers from "3-5 years" → min:3, max:5. If "5+ years" → min:5, max:0
2. seniority_level: Infer from title (Junior/Mid/Senior/Staff/Principal/Lead)
3. job_function: Categorize the role type (Backend/Frontend/etc)
4. salary_min/max: Extract numbers only. "€80k-100k" → min:80000, max:100000
5. technical_skills: Use simple names only ["Go", "Python"], not full sentences
6. Boolean fields: Set to true ONLY if explicitly mentioned
7. urgency_level: "Urgent" if mentions "immediate", "ASAP", "urgent". Otherwise "Standard"

Return ONLY valid JSON.`, jobText, time.Now().Format("2006-01-02T15:04:05Z07:00"), sourceURL)
}
