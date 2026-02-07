package main

import (
	"bonusperme/internal/handlers"
	"bonusperme/internal/i18n"
	"bonusperme/internal/linkcheck"
	"bonusperme/internal/logger"
	"bonusperme/internal/matcher"
	"bonusperme/internal/middleware"
	"bonusperme/internal/models"
	"bonusperme/internal/scraper"
	sentryutil "bonusperme/internal/sentry"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Initialize Sentry (non-blocking if SENTRY_DSN is empty)
	sentryutil.Init()
	defer sentryutil.Flush()

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

	// New API routes
	mux.HandleFunc("/api/encode-profile", handlers.EncodeProfileHandler)
	mux.HandleFunc("/api/decode-profile", handlers.DecodeProfileHandler)
	mux.HandleFunc("/api/bonus", handlers.BonusListHandler)
	mux.HandleFunc("/api/bonus/", handlers.BonusDetailHandler)

	// Pages
	mux.HandleFunc("/per-caf", handlers.PerCAFHandler)

	// SEO routes
	mux.HandleFunc("/bonus/", handlers.BonusPageHandler)
	mux.HandleFunc("/sitemap.xml", handlers.SitemapHandler)
	mux.HandleFunc("/robots.txt", handlers.RobotsTxtHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// Wrap with middleware: Recovery → Gzip → Rate Limiter
	handler := middleware.Recovery(middleware.Gzip(limiter.Middleware(mux)))

	// Link check at boot (background, delayed 5s to not slow startup)
	go func() {
		time.Sleep(5 * time.Second)
		allBonus := matcher.GetAllBonusWithRegional()
		ptrs := make([]*models.Bonus, len(allBonus))
		for i := range allBonus {
			ptrs[i] = &allBonus[i]
		}
		broken := linkcheck.CheckAllLinks(ptrs)
		if broken > 0 {
			logger.Warn("link check: broken links found at boot", map[string]interface{}{"broken": broken})
		}
	}()

	// Periodic link check every 24h
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			allBonus := matcher.GetAllBonusWithRegional()
			ptrs := make([]*models.Bonus, len(allBonus))
			for i := range allBonus {
				ptrs[i] = &allBonus[i]
			}
			linkcheck.CheckAllLinks(ptrs)
		}
	}()

	logger.Info("server starting", map[string]interface{}{"port": port})
	fmt.Printf("BonusPerMe running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
