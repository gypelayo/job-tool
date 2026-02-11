package webui

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"native-host/internal/db"
)

type Server struct {
	DB        *db.DB
	Templates *template.Template
}

func NewServer(database *db.DB) *Server {
	templates := loadTemplates()

	// List all loaded templates for debugging
	log.Println("Loaded templates:")
	for _, t := range templates.Templates() {
		log.Printf("  - %s", t.Name())
	}

	return &Server{
		DB:        database,
		Templates: templates,
	}
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	jobs, err := s.DB.ListJobs(100, 0, status)
	if err != nil {
		log.Printf("Error listing jobs: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d jobs", len(jobs))

	// DEBUG: Print first job details
	if len(jobs) > 0 {
		log.Printf("First job: Title=%s, Company=%s, Location=%s",
			jobs[0].JobTitle, jobs[0].CompanyName, jobs[0].Location)
	}

	stats, err := s.DB.GetJobStats()
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		stats = map[string]int{}
	}

	data := map[string]interface{}{
		"Jobs":          jobs,
		"Stats":         stats,
		"CurrentStatus": status,
	}

	log.Println("Executing template...")

	// Execute the named template explicitly
	err = s.Templates.Lookup("index.html").Execute(w, data)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Template rendered successfully")
}

func (s *Server) HandleJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := s.DB.GetJobByID(id)
	if err != nil {
		log.Printf("Error getting job: %v", err)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	err = s.Templates.Lookup("job.html").Execute(w, job)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	status := r.FormValue("status")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.UpdateJobStatus(id, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated row (for HTMX swap)
	jobs, _ := s.DB.ListJobs(1, 0, "")
	if len(jobs) > 0 {
		s.Templates.ExecuteTemplate(w, "job-row.html", jobs[0])
	}
}

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")

	jobs, err := s.DB.SearchJobs(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, job := range jobs {
		s.Templates.ExecuteTemplate(w, "job-row.html", job)
	}
}

func loadTemplates() *template.Template {
	tmpl := template.New("")

	// Parse all template files including subdirectories
	templates := []string{
		"web/templates/layout.html",
		"web/templates/index.html",
		"web/templates/job.html",
		"web/templates/components/job-row.html",
	}

	for _, t := range templates {
		_, err := tmpl.ParseFiles(t)
		if err != nil {
			log.Fatalf("Error parsing template %s: %v", t, err)
		}
		log.Printf("Loaded template: %s", t)
	}

	return tmpl
}
