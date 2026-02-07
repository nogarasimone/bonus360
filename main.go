package main

import (
	"bonus360/internal/handlers"
	"bonus360/internal/scraper"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize persistent counter
	handlers.InitCounter()

	// Start scraper scheduler (refreshes bonus data every 24h)
	scraper.StartScheduler()

	// API routes
	http.HandleFunc("/api/match", handlers.MatchHandler)
	http.HandleFunc("/api/stats", handlers.StatsHandler)
	http.HandleFunc("/api/health", handlers.HealthHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Printf("🎯 Bonus360 running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
