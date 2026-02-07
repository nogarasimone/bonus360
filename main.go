package main

import (
	"bonusperme/internal/config"
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
	"time"
)

func main() {
	// Load configuration from .env and environment variables
	config.Load()

	// Initialize Sentry (non-blocking if SENTRY_DSN is empty)
	sentryutil.Init()
	defer sentryutil.Flush()

	// Initialize persistent counter
	handlers.InitCounter()

	// Start scraper scheduler (respects SCRAPER_ENABLED config)
	scraper.StartScheduler()

	// Connect i18n translations to handler
	handlers.SetTranslationLoader(i18n.GetAll)

	// Rate limiter from config
	limiter := handlers.NewRateLimiter(
		config.Cfg.RateLimitRPS,
		config.Cfg.RateLimitBurst,
		time.Second,
	)

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
	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/privacy.html")
	})
	mux.HandleFunc("/cookie-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/cookie-policy.html")
	})

	// SEO routes
	mux.HandleFunc("/bonus/", handlers.BonusPageHandler)
	mux.HandleFunc("/sitemap.xml", handlers.SitemapHandler)
	mux.HandleFunc("/robots.txt", handlers.RobotsTxtHandler)

	// Serve static files (index.html served via template handler for GTM injection)
	mux.HandleFunc("/index.html", handlers.IndexHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	// Root handler: serve index.html via template for GTM, fallback to static for other files
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handlers.IndexHandler(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Wrap with middleware: Recovery → Gzip (if enabled) → Rate Limiter
	var handler http.Handler = limiter.Middleware(mux)
	if config.Cfg.GzipEnabled {
		handler = middleware.Gzip(handler)
	}
	handler = middleware.Recovery(handler)

	// Link check at boot (background, respects config)
	if config.Cfg.LinkCheckEnabled {
		go func() {
			time.Sleep(config.Cfg.LinkCheckDelay)
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

		// Periodic link check
		go func() {
			ticker := time.NewTicker(config.Cfg.LinkCheckInterval)
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
	}

	logger.Info("server starting", map[string]interface{}{"port": config.Cfg.Port})
	fmt.Printf("BonusPerMe running on http://localhost:%s\n", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(":"+config.Cfg.Port, handler))
}
