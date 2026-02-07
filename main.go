package main

import (
	"bonus360/internal/handlers"
	"bonus360/internal/i18n"
	"bonus360/internal/scraper"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

	// Connect i18n translations to handler
	handlers.SetTranslationLoader(i18n.GetAll)

	// Rate limiter: 30 requests per second per IP, burst of 60
	limiter := handlers.NewRateLimiter(30, 60, time.Second)

	// Create mux
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/match", handlers.MatchHandler)
	mux.HandleFunc("/api/stats", handlers.StatsHandler)
	mux.HandleFunc("/api/health", handlers.HealthDetailedHandler)
	mux.HandleFunc("/api/parse-isee", handlers.ParseISEEHandler)
	mux.HandleFunc("/api/calendar", handlers.CalendarHandler)
	mux.HandleFunc("/api/simulate", handlers.SimulateHandler)
	mux.HandleFunc("/api/report", handlers.ReportHandler)
	mux.HandleFunc("/api/notify-signup", handlers.NotifySignupHandler)
	mux.HandleFunc("/api/analytics", handlers.AnalyticsHandler)
	mux.HandleFunc("/api/analytics-summary", handlers.AnalyticsSummaryHandler)
	mux.HandleFunc("/api/scraper-status", handlers.ScraperStatusHandler)
	mux.HandleFunc("/api/translations", handlers.TranslationsHandler)

	// SEO routes
	mux.HandleFunc("/bonus/", handlers.BonusPageHandler)
	mux.HandleFunc("/sitemap.xml", handlers.SitemapHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// Wrap with rate limiter
	handler := limiter.Middleware(mux)

	fmt.Printf("Bonus360 running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
