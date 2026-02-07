package scraper

import (
	"bonusperme/internal/matcher"
	"bonusperme/internal/models"
	"log"
	"strings"
	"time"
)

// EnrichBonusData merges scraped bonus data with the hardcoded bonus list.
// Hardcoded bonuses serve as the authoritative base; scraped data supplements them.
func EnrichBonusData(scraped []models.Bonus, hardcoded []models.Bonus) []models.Bonus {
	// 1. Start with hardcoded as base (they have complete data + proper IDs for scoring)
	result := make([]models.Bonus, len(hardcoded))
	copy(result, hardcoded)

	// 2. Set UltimoAggiornamento on hardcoded if not set
	now := time.Now().Format("2 January 2006")
	for i := range result {
		if result[i].UltimoAggiornamento == "" {
			result[i].UltimoAggiornamento = now
		}
		if result[i].Stato == "" {
			result[i].Stato = "attivo"
		}
	}

	// 3. Build lookup by normalized name
	existing := make(map[string]int) // normalized name -> index in result
	for i, b := range result {
		existing[normalizeName(b.Nome)] = i
	}

	// 4. Merge scraped data
	for _, s := range scraped {
		norm := normalizeName(s.Nome)
		if idx, ok := existing[norm]; ok {
			// Update existing bonus with scraped data (only non-empty fields)
			mergeBonus(&result[idx], &s)
		} else {
			// New bonus from scraper
			s.UltimoAggiornamento = now
			if s.Stato == "" {
				s.Stato = "attivo"
			}
			if s.Categoria == "" {
				s.Categoria = "altro"
			}
			result = append(result, s)
			existing[norm] = len(result) - 1
		}
	}

	// 5. Validate
	var valid []models.Bonus
	for _, b := range result {
		if b.Nome != "" && b.ID != "" {
			valid = append(valid, b)
		}
	}

	log.Printf("[enricher] Result: %d hardcoded + %d scraped -> %d unique bonuses", len(hardcoded), len(scraped), len(valid))
	return valid
}

// ensureHardcodedBase is a compile-time check that matcher.GetAllBonus is accessible.
var _ = matcher.GetAllBonus

func normalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func mergeBonus(dst *models.Bonus, src *models.Bonus) {
	// Only update if scraped data has something new
	if src.Importo != "" && src.Importo != "Vedi sito ufficiale" && dst.Importo == "" {
		dst.Importo = src.Importo
	}
	if src.Scadenza != "" && src.Scadenza != "Verificare sul sito ufficiale" {
		dst.Scadenza = src.Scadenza
	}
	if src.FonteURL != "" && dst.FonteURL == "" {
		dst.FonteURL = src.FonteURL
	}
	if src.UltimoAggiornamento != "" {
		dst.UltimoAggiornamento = src.UltimoAggiornamento
	}
}
