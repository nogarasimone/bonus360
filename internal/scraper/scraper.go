package scraper

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ScrapeBonus attempts to scrape bonus information from official Italian sources.
// Falls back to hardcoded bonuses if scraping fails.
func ScrapeBonus() []models.Bonus {
	sources := []struct {
		name string
		url  string
	}{
		{"INPS", "https://www.inps.it/it/it/home.html"},
		{"Agenzia delle Entrate", "https://www.agenziaentrate.gov.it/portale/web/guest/home"},
		{"MEF", "https://www.mef.gov.it/"},
	}

	var scraped []models.Bonus

	client := &http.Client{Timeout: 15 * time.Second}

	for _, src := range sources {
		bonuses, err := scrapeSource(client, src.url, src.name)
		if err != nil {
			log.Printf("[scraper] Errore scraping %s (%s): %v", src.name, src.url, err)
			continue
		}
		scraped = append(scraped, bonuses...)
	}

	if len(scraped) == 0 {
		log.Printf("[scraper] Nessun bonus trovato dallo scraping, uso fallback hardcoded")
		return matcher.GetAllBonus()
	}

	return scraped
}

func scrapeSource(client *http.Client, url, ente string) ([]models.Bonus, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return extractBonusInfo(doc, ente), nil
}

func extractBonusInfo(doc *html.Node, ente string) []models.Bonus {
	var bonuses []models.Bonus
	var keywords = []string{"bonus", "agevolazione", "detrazione", "contributo", "assegno", "carta"}

	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "a" || n.Data == "h2" || n.Data == "h3" || n.Data == "h4") {
			text := getTextContent(n)
			textLower := strings.ToLower(text)

			for _, kw := range keywords {
				if strings.Contains(textLower, kw) && len(text) > 10 && len(text) < 200 {
					link := ""
					if n.Data == "a" {
						for _, attr := range n.Attr {
							if attr.Key == "href" {
								link = attr.Val
								break
							}
						}
					}

					bonus := models.Bonus{
						ID:              slugify(text),
						Nome:            strings.TrimSpace(text),
						Categoria:       categorize(textLower),
						Descrizione:     fmt.Sprintf("Informazione trovata su %s. Verificare sul sito ufficiale per dettagli aggiornati.", ente),
						Importo:         "Vedi sito ufficiale",
						Scadenza:        "Verificare sul sito ufficiale",
						Requisiti:       []string{"Consultare il sito ufficiale per i requisiti aggiornati"},
						ComeRichiederlo: []string{"Visitare il sito ufficiale dell'ente erogatore"},
						LinkUfficiale:   link,
						Ente:            ente,
					}
					bonuses = append(bonuses, bonus)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)

	return bonuses
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func categorize(text string) string {
	switch {
	case strings.Contains(text, "famiglia") || strings.Contains(text, "figlio") || strings.Contains(text, "nido") || strings.Contains(text, "nascita") || strings.Contains(text, "mamma"):
		return "famiglia"
	case strings.Contains(text, "casa") || strings.Contains(text, "ristruttur") || strings.Contains(text, "affitto") || strings.Contains(text, "abitazione"):
		return "casa"
	case strings.Contains(text, "salute") || strings.Contains(text, "psicolog"):
		return "salute"
	case strings.Contains(text, "studio") || strings.Contains(text, "cultura") || strings.Contains(text, "istruzione"):
		return "istruzione"
	case strings.Contains(text, "spesa") || strings.Contains(text, "alimentar"):
		return "spesa"
	case strings.Contains(text, "lavoro") || strings.Contains(text, "formazione"):
		return "lavoro"
	default:
		return "altro"
	}
}
