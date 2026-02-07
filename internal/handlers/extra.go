package handlers

import (
	"bonusperme/internal/linkcheck"
	"bonusperme/internal/matcher"
	"bonusperme/internal/models"
	"bonusperme/internal/scraper"
	sentryutil "bonusperme/internal/sentry"
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
	if p.Eta < 18 || p.Eta > 120 {
		return "Eta non valida (18-120)", false
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
	sb.WriteString("PRODID:-//BonusPerMe//IT\r\n")
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
		uid := slugifyName(item.Nome) + "@bonusperme.it"

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:" + uid + "\r\n")
		sb.WriteString("DTSTAMP:" + now + "\r\n")
		sb.WriteString("DTSTART;VALUE=DATE:" + dtStr + "\r\n")
		sb.WriteString("DTEND;VALUE=DATE:" + dtStr + "\r\n")
		sb.WriteString("SUMMARY:Scadenza: " + item.Nome + "\r\n")
		sb.WriteString("DESCRIPTION:Ricorda di presentare domanda per " + item.Nome + " prima della scadenza. Verifica requisiti su BonusPerMe.\r\n")
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
	w.Header().Set("Content-Disposition", `attachment; filename="bonusperme-scadenze.ics"`)
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
	linkcheck.ApplyStatus(reale.Bonus)

	simProfile := profile
	simProfile.ISEE = profile.ISEESimulato
	simulato := matcher.MatchBonus(simProfile, cachedBonus)
	linkcheck.ApplyStatus(simulato.Bonus)

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
	linkcheck.ApplyStatus(result.Bonus)

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	dateTimeStr := now.Format("02/01/2006 15:04")

	// Category color map
	categoryColors := map[string][3]int{
		"famiglia":   {232, 115, 90},
		"casa":       {229, 165, 73},
		"salute":     {184, 169, 212},
		"istruzione": {108, 155, 207},
		"spesa":      {43, 138, 126},
		"lavoro":     {92, 184, 92},
		"sostegno":   {240, 158, 140},
	}
	defaultColor := [3]int{142, 155, 160}

	pageW := 210.0
	marginL := 15.0
	marginR := 15.0
	contentW := pageW - marginL - marginR

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginL, 15, marginR)
	pdf.SetAutoPageBreak(true, 25)

	// Footer on every page
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(contentW/2, 10, transliterate("BonusPerMe.it -- Servizio gratuito e indipendente"), "", 0, "L", false, 0, "")
		pdf.CellFormat(contentW/2, 10, fmt.Sprintf("Pagina %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// ── 1. Header bar ──
	pdf.SetFillColor(26, 58, 92)
	pdf.Rect(0, 0, pageW, 20, "F")
	pdf.SetY(3)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(contentW, 8, transliterate("BonusPerMe"), "", 1, "L", false, 0, "")
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(contentW, 7, transliterate("Report Personalizzato"), "", 1, "L", false, 0, "")

	// ── 2. Date and subtitle ──
	pdf.SetY(24)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(contentW, 5, transliterate("Generato il "+dateTimeStr), "", 1, "L", false, 0, "")
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(contentW, 5, transliterate("Documento a scopo orientativo"), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// ── 3. Summary box ──
	summaryY := pdf.GetY()
	boxH := 22.0
	pdf.SetFillColor(232, 240, 248)
	pdf.SetDrawColor(180, 210, 235)
	pdf.RoundedRect(marginL, summaryY, contentW, boxH, 3, "1234", "FD")
	pdf.SetY(summaryY + 3)
	pdf.SetX(marginL + 5)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(26, 58, 92)
	pdf.CellFormat(contentW-10, 6, transliterate("RIEPILOGO"), "", 1, "L", false, 0, "")
	pdf.SetX(marginL + 5)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(contentW/2-5, 6, transliterate(fmt.Sprintf("Bonus compatibili: %d", result.BonusTrovati)), "", 0, "L", false, 0, "")
	pdf.CellFormat(contentW/2-5, 6, transliterate(fmt.Sprintf("Risparmio stimato: %s", result.RisparmioStimato)), "", 1, "L", false, 0, "")
	pdf.SetY(summaryY + boxH + 6)

	// ── 4. Each bonus ──
	for i, b := range result.Bonus {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		startY := pdf.GetY()

		// Determine category color
		cat := strings.ToLower(b.Categoria)
		col, ok := categoryColors[cat]
		if !ok {
			col = defaultColor
		}

		// Colored left border (drawn after content to know height; save startY)
		contentX := marginL + 5

		// Nome
		pdf.SetX(contentX)
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetTextColor(26, 58, 92)
		pdf.CellFormat(contentW-5, 7, transliterate(b.Nome), "", 1, "L", false, 0, "")

		// Ente
		pdf.SetX(contentX)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(contentW-5, 5, transliterate(b.Ente), "", 1, "L", false, 0, "")

		// Compatibilita
		pdf.SetX(contentX)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(contentW-5, 6, transliterate(fmt.Sprintf("Compatibilita: %d%%", b.Compatibilita)), "", 1, "L", false, 0, "")

		// Importo
		pdf.SetX(contentX)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(contentW-5, 6, transliterate("Importo: "+b.Importo), "", 1, "L", false, 0, "")

		if b.ImportoReale != "" && b.ImportoReale != b.Importo {
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(0, 128, 0)
			pdf.CellFormat(contentW-5, 5, transliterate("Importo stimato per te: "+b.ImportoReale), "", 1, "L", false, 0, "")
		}

		// Descrizione
		pdf.SetX(contentX)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		pdf.MultiCell(contentW-5, 4.5, transliterate(b.Descrizione), "", "L", false)
		pdf.Ln(2)

		// Requisiti
		if len(b.Requisiti) > 0 {
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetTextColor(26, 58, 92)
			pdf.CellFormat(contentW-5, 6, transliterate("Requisiti:"), "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(50, 50, 50)
			for _, req := range b.Requisiti {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				pdf.SetX(contentX + 3)
				pdf.CellFormat(contentW-8, 5, transliterate("  "+req), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Come fare domanda
		if len(b.ComeRichiederlo) > 0 {
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetTextColor(26, 58, 92)
			pdf.CellFormat(contentW-5, 6, transliterate("Come fare domanda:"), "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(50, 50, 50)
			for stepIdx, step := range b.ComeRichiederlo {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				pdf.SetX(contentX + 3)
				pdf.CellFormat(contentW-8, 5, transliterate(fmt.Sprintf("%d. %s", stepIdx+1, step)), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Documenti necessari
		if len(b.Documenti) > 0 {
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetTextColor(26, 58, 92)
			pdf.CellFormat(contentW-5, 6, transliterate("Documenti necessari:"), "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(50, 50, 50)
			for _, doc := range b.Documenti {
				if pdf.GetY() > 270 {
					pdf.AddPage()
				}
				cbX := contentX + 3
				cbY := pdf.GetY() + 1
				pdf.SetDrawColor(100, 100, 100)
				pdf.Rect(cbX, cbY, 3, 3, "D")
				pdf.SetX(cbX + 5)
				pdf.CellFormat(contentW-13, 5, transliterate(doc), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}

		// Link ufficiale — smart rendering based on verification
		if b.LinkUfficiale != "" {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "", 9)
			if b.LinkVerificato {
				// Verified: blue link
				pdf.SetTextColor(0, 51, 153)
				pdf.Write(5, transliterate("Link ufficiale: "))
				pdf.WriteLinkString(5, b.LinkUfficiale, b.LinkUfficiale)
			} else if b.LinkRicerca != "" {
				// Broken: warning color with search fallback
				pdf.SetTextColor(184, 134, 11)
				pdf.Write(5, transliterate("Cerca su sito ufficiale: "))
				pdf.WriteLinkString(5, b.LinkRicerca, b.LinkRicerca)
			} else {
				// Not verified, no search: show as-is
				pdf.SetTextColor(0, 51, 153)
				pdf.Write(5, transliterate("Link ufficiale: "))
				pdf.WriteLinkString(5, b.LinkUfficiale, b.LinkUfficiale)
			}
			pdf.Ln(6)
		}

		// Verifica manuale disclaimer for regional bonuses
		if b.VerificaManualeNecessaria {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFillColor(255, 251, 235)
			pdf.SetDrawColor(184, 134, 11)
			noteW := contentW - 5
			noteY := pdf.GetY()
			pdf.Rect(contentX, noteY, noteW, 12, "FD")
			pdf.SetX(contentX + 3)
			pdf.SetFont("Helvetica", "I", 8)
			pdf.SetTextColor(130, 100, 0)
			pdf.CellFormat(noteW-6, 12, transliterate("Dato inserito manualmente -- potrebbe non essere aggiornato. Verificare sul sito della regione."), "", 1, "L", false, 0, "")
		}

		// Scadenza
		if b.Scadenza != "" {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}
			pdf.SetX(contentX)
			pdf.SetFont("Helvetica", "I", 9)
			pdf.SetTextColor(200, 0, 0)
			pdf.CellFormat(contentW-5, 5, transliterate("Scadenza: "+b.Scadenza), "", 1, "L", false, 0, "")
		}

		// Draw colored left border now that we know the height
		endY := pdf.GetY()
		borderH := endY - startY
		if borderH < 10 {
			borderH = 10
		}
		pdf.SetFillColor(col[0], col[1], col[2])
		pdf.Rect(marginL, startY, 3, borderH, "F")

		// Separator between bonuses
		if i < len(result.Bonus)-1 {
			pdf.Ln(3)
			pdf.SetDrawColor(200, 200, 200)
			pdf.Line(marginL, pdf.GetY(), pageW-marginR, pdf.GetY())
			pdf.Ln(5)
		}
	}

	// ── 5. "Come procedere" box ──
	if pdf.GetY() > 250 {
		pdf.AddPage()
	}
	pdf.Ln(6)
	boxY := pdf.GetY()
	procBoxH := 38.0
	pdf.SetFillColor(232, 245, 232)
	pdf.SetDrawColor(180, 220, 180)
	pdf.RoundedRect(marginL, boxY, contentW, procBoxH, 3, "1234", "FD")
	pdf.SetY(boxY + 3)
	pdf.SetX(marginL + 5)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(26, 58, 92)
	pdf.CellFormat(contentW-10, 7, transliterate("PROSSIMI PASSI"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(50, 50, 50)
	steps := []string{
		"1. Verifica i requisiti specifici per ogni bonus sui siti ufficiali indicati",
		"2. Prepara la documentazione necessaria (ISEE, SPID, documenti)",
		"3. Presenta le domande online o presso un CAF/patronato",
	}
	for _, step := range steps {
		pdf.SetX(marginL + 5)
		pdf.CellFormat(contentW-10, 7, transliterate(step), "", 1, "L", false, 0, "")
	}

	// ── Output ──
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="bonusperme-report-%s.pdf"`, dateStr))

	if err := pdf.Output(w); err != nil {
		sentryutil.CaptureError(err, map[string]string{"handler": "report", "phase": "pdf-output"})
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
		Tipo  string `json:"tipo,omitempty"`
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

	tipo := body.Tipo
	if tipo == "" {
		tipo = "user"
	}
	line := fmt.Sprintf("[%s] [%s] %s\n", time.Now().Format(time.RFC3339), tipo, body.Email)
	if _, err := f.WriteString(line); err != nil {
		http.Error(w, "Errore interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
