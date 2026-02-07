package handlers

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"bonus360/internal/scraper"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

func MatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var profile models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	IncrementCounter()

	cachedBonus := scraper.GetCachedBonus()
	result := matcher.MatchBonus(profile, cachedBonus)

	w.Header().Set("Content-Type", "application/json")
	// No caching - data is ephemeral
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	json.NewEncoder(w).Encode(result)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"scansioni": GetCounter(),
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
