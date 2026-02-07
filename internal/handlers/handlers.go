package handlers

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"bonus360/internal/scraper"
	sentryutil "bonus360/internal/sentry"
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
		sentryutil.CaptureError(err, map[string]string{"handler": "match", "phase": "decode"})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if msg, ok := validateProfile(profile); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

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

type iseeResponse struct {
	ISEE  float64 `json:"isee"`
	Found bool    `json:"found"`
}

func ParseISEEHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10MB max

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File non trovato", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		sentryutil.CaptureError(err, map[string]string{"handler": "parse-isee", "phase": "read"})
		http.Error(w, "Errore lettura file", http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		sentryutil.CaptureError(err, map[string]string{"handler": "parse-isee", "phase": "pdf-parse"})
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		json.NewEncoder(w).Encode(iseeResponse{ISEE: 0, Found: false})
		return
	}

	var textBuilder strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		p := pdfReader.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		textBuilder.WriteString(text)
		textBuilder.WriteString(" ")
	}

	fullText := textBuilder.String()
	iseeValue := extractISEE(fullText)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	if iseeValue > 0 {
		json.NewEncoder(w).Encode(iseeResponse{ISEE: iseeValue, Found: true})
	} else {
		json.NewEncoder(w).Encode(iseeResponse{ISEE: 0, Found: false})
	}
}

func extractISEE(text string) float64 {
	re := regexp.MustCompile(`(?i)(?:isee|indicatore)[:\s€]*([0-9][0-9.,]*)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return 0
	}
	numStr := matches[1]
	// Handle Italian number format: 18.432,00 -> 18432.00
	if strings.Contains(numStr, ",") {
		// Remove thousand separators (dots), replace comma with dot
		numStr = strings.ReplaceAll(numStr, ".", "")
		numStr = strings.Replace(numStr, ",", ".", 1)
	}
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	return val
}
