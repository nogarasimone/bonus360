package linkcheck

import (
	"bonus360/internal/logger"
	"bonus360/internal/models"
	sentryutil "bonus360/internal/sentry"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

// CheckLink verifies if a URL responds with a 2xx/3xx status using HEAD.
func CheckLink(url string) (ok bool, statusCode int) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Bonus360/1.0; +https://bonus360.it)")
	req.Header.Set("Accept-Language", "it-IT,it;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400, resp.StatusCode
}

// CheckAllLinks checks all bonus links and updates verification fields.
// Returns the number of broken links found.
func CheckAllLinks(bonusList []*models.Bonus) int {
	broken := 0
	today := time.Now().Format("2006-01-02")

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5) // max 5 concurrent checks

	for _, b := range bonusList {
		if b.LinkUfficiale == "" {
			continue
		}
		wg.Add(1)
		go func(bonus *models.Bonus) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ok, status := CheckLink(bonus.LinkUfficiale)
			bonus.LinkVerificatoAl = today

			if ok {
				bonus.LinkVerificato = true
				logger.Info("linkcheck: OK", map[string]interface{}{
					"bonus_id": bonus.ID, "url": bonus.LinkUfficiale,
				})
			} else {
				bonus.LinkVerificato = false
				mu.Lock()
				broken++
				mu.Unlock()
				logger.Warn("linkcheck: broken", map[string]interface{}{
					"bonus_id": bonus.ID, "url": bonus.LinkUfficiale,
					"status": status, "fallback": bonus.LinkRicerca,
				})
				sentryutil.CaptureMessage(
					"Broken link: "+bonus.ID,
					sentryutil.LevelWarning(),
					map[string]string{
						"component": "linkcheck",
						"bonus_id":  bonus.ID,
						"url":       bonus.LinkUfficiale,
						"status":    fmt.Sprintf("%d", status),
					},
				)
			}
		}(b)
	}
	wg.Wait()

	logger.Info("linkcheck: completed", map[string]interface{}{"broken": broken, "total": len(bonusList)})
	return broken
}
