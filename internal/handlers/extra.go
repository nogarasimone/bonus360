package handlers

import (
	"bonus360/internal/matcher"
	"bonus360/internal/models"
	"bonus360/internal/scraper"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ---------- helpers ----------

var italianMonths = map[string]int{
	"gennaio":   1,
	"febbraio":  2,
	"marzo":     3,
	"aprile":    4,
	"maggio":    5,
	"giugno":    6,
	"luglio":    7,
	"agosto":    8,
	"settembre": 9,
	"ottobre":   10,
	"novembre":  11,
	"dicembre":  12,
}

var italianDateRe = regexp.MustCompile(
	`(\d{1,2})\s+(gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)\s+(\d{4})`,
)

var slashDateRe = regexp.MustCompile(`(\d{2})/(\d{2})/(\d{4})`)

func parseItalianDate(s string) time.Time {
	lower := strings.ToLower(s)

	if m := italianDateRe.FindStringSubmatch(lower); len(m) == 4 {
		day, _ := strconv.Atoi(m[1])
		month := italianMonths[m[2]]
		year, _ := strconv.Atoi(m[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	if m := slashDateRe.FindStringSubmatch(s); len(m) == 4 {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	// fallback: December 31 of current year
	return time.Date(time.Now().Year(), 12, 31, 0, 0, 0, 0, time.UTC)
}

func slugifyName(s string) string {
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
	return strings.Trim(s, "-")
}

func transliterate(s string) string {
	replacer := strings.NewReplacer(
		"à", "a", "è", "e", "é", "e", "ì", "i", "ò", "o", "ù", "u",
		"À", "A", "È", "E", "É", "E", "Ì", "I", "Ò", "O", "Ù", "U",
		"€", "EUR ", "\u2264", "<=", "\u2265", ">=",
	)
	return replacer.Replace(s)
}

func parseEuroAmount(s string) float64 {
	re := regexp.MustCompile(`[0-9][0-9.,]*`)
	m := re.FindString(s)
	if m == "" {
		return 0
	}
	m = strings.ReplaceAll(m, ".", "")
	m = strings.Replace(m, ",", ".", 1)
	v, _ := strconv.ParseFloat(m, 64)
	return v
}

// ---------- validateProfile ----------

func validateProfile(p models.UserProfile) (string, bool) {
	if p.Eta < 0 || p.Eta > 120 {
		return "Eta non valida (0-120)", false
	}
	if p.ISEE < 0 || p.ISEE > 500000 {
		return "ISEE non valido (0-500000)", false
	}
	if p.RedditoAnnuo < 0 || p.RedditoAnnuo > 1000000 {
		return "Reddito annuo non valido (0-1000000)", false
	}
	if p.NumeroFigli < 0 || p.NumeroFigli > 20 {
		return "Numero figli non valido (0-20)", false
	}
	if p.FigliMinorenni < 0 || p.FigliMinorenni > 20 {
		return "Figli minorenni non valido (0-20)", false
	}
	if p.FigliUnder3 < 0 || p.FigliUnder3 > 20 {
		return "Figli under 3 non valido (0-20)", false
	}
	if p.Over65 < 0 || p.Over65 > 10 {
		return "Over 65 non valido (0-10)", false
	}
	return "", true
}

// ---------- 1. CalendarHandler ----------

func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw := r.URL.Query().Get("bonuses")
	if raw == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var items []struct {
		Nome     string `json:"nome"`
		Scadenza string `json:"scadenza"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		http.Error(w, "Invalid bonuses JSON", http.StatusBadRequest)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	now := time.Now().UTC().Format("20060102T150405Z")

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Bonus360//IT\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")

	validCount := 0
	for _, item := range items {
		if item.Nome == "" {
			continue
		}
		validCount++
		dt := parseItalianDate(item.Scadenza)
		dtStr := dt.Format("20060102")
		uid := slugifyName(item.Nome) + "@bonus360.it"

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:" + uid + "\r\n")
		sb.WriteString("DTSTAMP:" + now + "\r\n")
		sb.WriteString("DTSTART;VALUE=DATE:" + dtStr + "\r\n")
		sb.WriteString("DTEND;VALUE=DATE:" + dtStr + "\r\n")
		sb.WriteString("SUMMARY:Scadenza: " + item.Nome + "\r\n")
		sb.WriteString("DESCRIPTION:Ricorda di presentare domanda per " + item.Nome + " prima della scadenza. Verifica requisiti su Bonus360.\r\n")
		sb.WriteString("BEGIN:VALARM\r\n")
		sb.WriteString("TRIGGER:-P7D\r\n")
		sb.WriteString("ACTION:DISPLAY\r\n")
		sb.WriteString("DESCRIPTION:Scadenza tra 7 giorni: " + item.Nome + "\r\n")
		sb.WriteString("END:VALARM\r\n")
		sb.WriteString("BEGIN:VALARM\r\n")
		sb.WriteString("TRIGGER:-P1D\r\n")
		sb.WriteString("ACTION:DISPLAY\r\n")
		sb.WriteString("DESCRIPTION:Scadenza domani: " + item.Nome + "\r\n")
		sb.WriteString("END:VALARM\r\n")
		sb.WriteString("END:VEVENT\r\n")
	}

	sb.WriteString("END:VCALENDAR\r\n")

	if validCount == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="bonus360-scadenze.ics"`)
	w.Write([]byte(sb.String()))
}

// ---------- 2. SimulateHandler ----------

func SimulateHandler(w http.ResponseWriter, r *http.Request) {
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

	if msg, ok := validateProfile(profile); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cachedBonus := scraper.GetCachedBonus()

	reale := matcher.MatchBonus(profile, cachedBonus)

	simProfile := profile
	simProfile.ISEE = profile.ISEESimulato
	simulato := matcher.MatchBonus(simProfile, cachedBonus)

	bonusExtra := simulato.BonusTrovati - reale.BonusTrovati
	if bonusExtra < 0 {
		bonusExtra = 0
	}

	risparmioReale := parseEuroAmount(reale.RisparmioStimato)
	risparmioSim := parseEuroAmount(simulato.RisparmioStimato)
	extraVal := risparmioSim - risparmioReale
	if extraVal < 0 {
		extraVal = 0
	}

	result := models.SimulateResult{
		Reale:          reale,
		Simulato:       simulato,
		BonusExtra:     bonusExtra,
		RisparmioExtra: fmt.Sprintf("EUR %.0f", extraVal),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(result)
}

// ---------- 3. ReportHandler ----------

func ReportHandler(w http.ResponseWriter, r *http.Request) {
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

	if msg, ok := validateProfile(profile); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cachedBonus := scraper.GetCachedBonus()
	result := matcher.MatchBonus(profile, cachedBonus)

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	dateTimeStr := now.Format("02/01/2006 15:04")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 25)

	// Footer on every page
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 10, transliterate("Generato da Bonus360.it -- Servizio gratuito e indipendente -- "+dateStr), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 12, transliterate("Bonus360 -- Report Personalizzato"), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, transliterate("Generato il "+dateTimeStr), "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Subtitle
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 5, transliterate("Documento a scopo orientativo -- Verifica sui portali ufficiali"), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Summary
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(0, 0, 0)
	summary := fmt.Sprintf("%d bonus trovati -- Risparmio stimato: %s", result.BonusTrovati, result.RisparmioStimato)
	pdf.CellFormat(0, 8, transliterate(summary), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Separator
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// Each bonus
	for i, b := range result.Bonus {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		// Nome
		pdf.SetFont("Helvetica", "B", 14)
		pdf.SetTextColor(0, 51, 102)
		pdf.CellFormat(0, 8, transliterate(b.Nome), "", 1, "L", false, 0, "")

		// Ente
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 5, transliterate("Ente erogatore: "+b.Ente), "", 1, "L", false, 0, "")

		// Importo
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 6, transliterate("Importo: "+b.Importo), "", 1, "L", false, 0, "")
		if b.ImportoReale != "" && b.ImportoReale != b.Importo {
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(0, 128, 0)
			pdf.CellFormat(0, 5, transliterate("Importo stimato per te: "+b.ImportoReale), "", 1, "L", false, 0, "")
		}

		// Descrizione
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.MultiCell(0, 5, transliterate(b.Descrizione), "", "L", false)
		pdf.Ln(2)

		// Requisiti
		if len(b.Requisiti) > 0 {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.CellFormat(0, 5, "Requisiti:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			for _, req := range b.Requisiti {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				pdf.CellFormat(5, 5, "", "", 0, "", false, 0, "")
				pdf.CellFormat(0, 5, transliterate("- "+req), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Come fare domanda
		if len(b.ComeRichiederlo) > 0 {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.CellFormat(0, 5, "Come fare domanda:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			for _, step := range b.ComeRichiederlo {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				pdf.CellFormat(5, 5, "", "", 0, "", false, 0, "")
				pdf.CellFormat(0, 5, transliterate("- "+step), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Documenti necessari
		if len(b.Documenti) > 0 {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.CellFormat(0, 5, "Documenti necessari:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			for _, doc := range b.Documenti {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				pdf.CellFormat(5, 5, "", "", 0, "", false, 0, "")
				pdf.CellFormat(0, 5, transliterate("[ ] "+doc), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Link ufficiale
		if b.LinkUfficiale != "" {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(0, 0, 200)
			pdf.CellFormat(0, 5, "Link: "+b.LinkUfficiale, "", 1, "L", false, 0, "")
		}

		// Scadenza
		if b.Scadenza != "" {
			pdf.SetFont("Helvetica", "I", 9)
			pdf.SetTextColor(180, 0, 0)
			pdf.CellFormat(0, 5, transliterate("Scadenza: "+b.Scadenza), "", 1, "L", false, 0, "")
		}

		// Separator between bonuses
		if i < len(result.Bonus)-1 {
			pdf.Ln(4)
			pdf.SetDrawColor(220, 220, 220)
			pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
			pdf.Ln(6)
		}
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="bonus360-report-%s.pdf"`, dateStr))

	if err := pdf.Output(w); err != nil {
		http.Error(w, "Errore generazione PDF", http.StatusInternalServerError)
		return
	}
}

// ---------- 4. NotifySignupHandler ----------

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NotifySignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	body.Email = strings.TrimSpace(body.Email)

	if !emailRe.MatchString(body.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email non valida"})
		return
	}

	f, err := os.OpenFile("emails.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Errore interno", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	line := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), body.Email)
	if _, err := f.WriteString(line); err != nil {
		http.Error(w, "Errore interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
