package scraper

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"log"
	"sync"
	"time"
)

// SourceStatus tracks the status of each scraping source.
type SourceStatus struct {
	LastFetch  time.Time `json:"last_fetch"`
	Success    bool      `json:"success"`
	BonusFound int       `json:"bonus_found"`
	Error      string    `json:"error,omitempty"`
}

// BonusCache holds the cached bonus data and source status information.
type BonusCache struct {
	mu            sync.RWMutex
	bonus         []models.Bonus
	lastUpdate    time.Time
	updateCount   int
	sourcesStatus map[string]SourceStatus
}

var cache = &BonusCache{
	sourcesStatus: make(map[string]SourceStatus),
}

// StartScheduler runs an initial scrape and then re-scrapes every 24 hours.
func StartScheduler() {
	go func() {
		RunScrape()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			RunScrape()
		}
	}()
}

// RunScrape performs a full scrape cycle across all sources.
func RunScrape() {
	log.Println("[scraper] Starting scrape cycle...")
	sources := GetSources()
	var allScraped []models.Bonus

	for i, src := range sources {
		if i > 0 {
			time.Sleep(2 * time.Second) // polite delay between sources
		}

		log.Printf("[scraper] Fetching %s (%s)...", src.Name, src.URL)
		bonuses := ParseSource(src)

		status := SourceStatus{
			LastFetch:  time.Now(),
			Success:    true, // success if no fetch error (even if 0 results)
			BonusFound: len(bonuses),
		}

		if len(bonuses) == 0 {
			status.Error = "no bonuses found"
		}

		cache.mu.Lock()
		cache.sourcesStatus[src.Name] = status
		cache.mu.Unlock()

		allScraped = append(allScraped, bonuses...)
		log.Printf("[scraper] %s: found %d items", src.Name, len(bonuses))
	}

	hardcoded := matcher.GetAllBonus()
	enriched := EnrichBonusData(allScraped, hardcoded)

	cache.mu.Lock()
	cache.bonus = enriched
	cache.lastUpdate = time.Now()
	cache.updateCount++
	cache.mu.Unlock()

	log.Printf("[scraper] Cache updated: %d total bonuses (cycle #%d)", len(enriched), cache.updateCount)
}

// GetCachedBonus returns the cached list of bonuses.
// Falls back to hardcoded if cache is empty.
func GetCachedBonus() []models.Bonus {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	if len(cache.bonus) == 0 {
		return matcher.GetAllBonus()
	}
	result := make([]models.Bonus, len(cache.bonus))
	copy(result, cache.bonus)
	return result
}

// GetScraperStatus returns a status map with scraper state and per-source info.
func GetScraperStatus() map[string]interface{} {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	sources := make(map[string]SourceStatus)
	for k, v := range cache.sourcesStatus {
		sources[k] = v
	}

	return map[string]interface{}{
		"last_run":     cache.lastUpdate,
		"next_run":     cache.lastUpdate.Add(24 * time.Hour),
		"bonus_count":  len(cache.bonus),
		"update_count": cache.updateCount,
		"sources":      sources,
	}
}
