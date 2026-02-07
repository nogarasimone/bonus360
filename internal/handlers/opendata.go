package handlers

import (
	"bonus360/internal/matcher"
	"bonus360/internal/scraper"
	"encoding/json"
	"net/http"
	"strings"
)

// BonusListHandler returns the full list of all bonuses (national + regional) as JSON.
// GET /api/bonus
func BonusListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try cached (enriched) bonuses first, fallback to hardcoded
	allBonus := scraper.GetCachedBonus()
	if len(allBonus) == 0 {
		allBonus = matcher.GetAllBonusWithRegional()
	} else {
		// Append regional bonuses if not already present
		regionals := matcher.GetRegionalBonus()
		allBonus = append(allBonus, regionals...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(allBonus)
}

// BonusDetailHandler returns a single bonus by ID.
// GET /api/bonus/{id}
func BonusDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/bonus/xxx
	path := strings.TrimPrefix(r.URL.Path, "/api/bonus/")
	bonusID := strings.TrimSuffix(path, "/")
	if bonusID == "" {
		http.Error(w, "Bonus ID richiesto", http.StatusBadRequest)
		return
	}

	allBonus := matcher.GetAllBonusWithRegional()
	for _, b := range allBonus {
		if b.ID == bonusID {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(b)
			return
		}
	}

	http.Error(w, "Bonus non trovato", http.StatusNotFound)
}
