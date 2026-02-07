package scraper

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"log"
	"sync"
	"time"
)

var (
	cachedBonus []models.Bonus
	cacheMu     sync.RWMutex
)

// StartScheduler runs an initial scrape and then re-scrapes every 24 hours.
func StartScheduler() {
	go func() {
		refresh()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refresh()
		}
	}()
}

func refresh() {
	log.Println("[scraper] Avvio aggiornamento bonus...")
	bonuses := ScrapeBonus()
	cacheMu.Lock()
	cachedBonus = bonuses
	cacheMu.Unlock()
	log.Printf("[scraper] Cache aggiornata con %d bonus", len(bonuses))
}

// GetCachedBonus returns the cached list of bonuses.
// Falls back to hardcoded if cache is empty.
func GetCachedBonus() []models.Bonus {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if len(cachedBonus) == 0 {
		return matcher.GetAllBonus()
	}
	result := make([]models.Bonus, len(cachedBonus))
	copy(result, cachedBonus)
	return result
}
