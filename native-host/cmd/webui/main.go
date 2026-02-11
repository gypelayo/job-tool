package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"native-host/internal/config"
	"native-host/internal/db"
	"native-host/internal/webui"

	_ "github.com/mattn/go-sqlite3" // ← ADD THIS LINE
)

func main() {
	log.Println("Starting web server...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	log.Printf("Config loaded. DB path: %s", cfg.DBPath)

	database, err := db.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected successfully")

	server := webui.NewServer(database)
	log.Println("Server initialized")

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", server.HandleIndex)
	http.HandleFunc("/job", server.HandleJob)
	http.HandleFunc("/update-status", server.HandleUpdateStatus)
	http.HandleFunc("/search", server.HandleSearch)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	fmt.Printf("\n🚀 Server running at http://localhost:%s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
