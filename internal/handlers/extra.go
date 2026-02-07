package handlers

import (
	"bonusperme/internal/linkcheck"
	"bonusperme/internal/matcher"
	"bonusperme/internal/models"
	"bonusperme/internal/scraper"
	sentryutil "bonusperme/internal/sentry"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
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
		"\u2264", "<=", "\u2265", ">=",
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

// =====================================================================
// 3. ReportHandler — Professional PDF Report
// =====================================================================

// PDF design system colors
var (
	cBlue    = [3]int{27, 58, 84}
	cBlueMid = [3]int{44, 95, 124}
	cTerra   = [3]int{192, 82, 46}
	cGreen   = [3]int{42, 107, 69}
	cGreenBg = [3]int{233, 245, 237}
	cAmber   = [3]int{154, 123, 46}
	cAmberBg = [3]int{250, 244, 230}
	cCream   = [3]int{244, 243, 238}
	cInk90   = [3]int{38, 38, 38}
	cInk75   = [3]int{64, 64, 64}
	cInk50   = [3]int{107, 107, 107}
	cInk30   = [3]int{160, 160, 160}
	cInk15   = [3]int{217, 217, 217}
	cRed     = [3]int{220, 38, 38}
	cRedBg   = [3]int{254, 226, 226}
	cWhite   = [3]int{255, 255, 255}
)

const (
	pageW   = 210.0
	marginL = 18.0
	marginR = 18.0
	contentW = pageW - marginL - marginR // 174mm
)

func setFill(pdf *gofpdf.Fpdf, c [3]int)  { pdf.SetFillColor(c[0], c[1], c[2]) }
func setText(pdf *gofpdf.Fpdf, c [3]int)   { pdf.SetTextColor(c[0], c[1], c[2]) }
func setDraw(pdf *gofpdf.Fpdf, c [3]int)   { pdf.SetDrawColor(c[0], c[1], c[2]) }

func fmtEuro(amount float64) string {
	if amount == 0 {
		return "0"
	}
	neg := amount < 0
	if neg {
		amount = -amount
	}
	whole := int(amount)
	frac := int(math.Round((amount - float64(whole)) * 100))
	s := addDotSep(fmt.Sprintf("%d", whole))
	prefix := ""
	if neg {
		prefix = "-"
	}
	if frac > 0 {
		return fmt.Sprintf("%s%s,%02d", prefix, s, frac)
	}
	return prefix + s
}

func addDotSep(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	return addDotSep(s[:n-3]) + "." + s[n-3:]
}

func compatColor(pct int) ([3]int, [3]int) {
	if pct >= 80 {
		return cGreen, cGreenBg
	}
	if pct >= 50 {
		return cAmber, cAmberBg
	}
	return cInk30, [3]int{240, 240, 240}
}

func truncURL(url string, max int) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	if len(url) > max {
		return url[:max-3] + "..."
	}
	return url
}

func estimateBonusH(b models.Bonus) float64 {
	if b.Scaduto {
		return 40
	}
	h := 22.0 // header
	h += 18   // importo box
	h += 16   // desc
	h += 3    // separator
	if len(b.Requisiti) > 0 {
		h += float64(len(b.Requisiti))*5.5 + 10
	}
	if len(b.ComeRichiederlo) > 0 {
		h += float64(len(b.ComeRichiederlo))*5.5 + 10
	}
	if len(b.Documenti) > 0 {
		h += float64(len(b.Documenti))*5.5 + 10
	}
	h += 14 // footer
	return h
}

func ensureSpace(pdf *gofpdf.Fpdf, needed float64) float64 {
	y := pdf.GetY()
	if y+needed > 277 {
		pdf.AddPage()
		return 18
	}
	return y
}

func drawPill(pdf *gofpdf.Fpdf, x, y float64, text string, bg, fg [3]int) float64 {
	pdf.SetFont("Helvetica", "B", 7.5)
	w := pdf.GetStringWidth(transliterate(text)) + 8
	setFill(pdf, bg)
	pdf.RoundedRect(x, y, w, 5.5, 2, "1234", "F")
	setText(pdf, fg)
	pdf.SetXY(x, y+0.5)
	pdf.CellFormat(w, 5, transliterate(text), "", 0, "C", false, 0, "")
	return w
}

func drawStepCircle(pdf *gofpdf.Fpdf, x, y float64, num int) {
	setFill(pdf, cBlue)
	pdf.Circle(x+1.5, y+1.8, 2.5, "F")
	pdf.SetFont("Courier", "B", 6)
	setText(pdf, cWhite)
	pdf.SetXY(x-1, y-0.2)
	pdf.CellFormat(5, 4.5, fmt.Sprintf("%d", num), "", 0, "C", false, 0, "")
}

func drawCheckGreen(pdf *gofpdf.Fpdf, x, y float64) {
	setFill(pdf, cGreenBg)
	pdf.Circle(x+1.5, y+1.5, 2, "F")
	setDraw(pdf, cGreen)
	pdf.SetLineWidth(0.35)
	pdf.Line(x+0.5, y+1.5, x+1.2, y+2.2)
	pdf.Line(x+1.2, y+2.2, x+2.5, y+0.8)
}

func drawCheckboxEmpty(pdf *gofpdf.Fpdf, x, y float64) {
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(x, y, 3, 3, 0.5, "1234", "D")
}

// ReportHandler generates a professional PDF report.
// Accepts JSON body or form field "data" with JSON.
// Query param ?mode=inline opens in browser instead of downloading.
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var profile models.UserProfile

	// Support both JSON body and form-encoded "data" field
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		r.ParseForm()
		dataStr := r.FormValue("data")
		if dataStr == "" {
			http.Error(w, "Missing data field", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal([]byte(dataStr), &profile); err != nil {
			http.Error(w, "Invalid profile data", http.StatusBadRequest)
			return
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
	}

	if msg, ok := validateProfile(profile); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cachedBonus := scraper.GetCachedBonus()
	result := matcher.MatchBonus(profile, cachedBonus)
	linkcheck.ApplyStatus(result.Bonus)

	// Generate profile code for footer
	profileCode := "BPM-..."
	compact := toCompact(profile)
	if data, err := json.Marshal(compact); err == nil {
		b64 := base64.RawURLEncoding.EncodeToString(data)
		code := codePrefix + b64
		if len(code) > 64 {
			code = code[:64]
		}
		profileCode = code
	}

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	dateDisplay := now.Format("02/01/2006")

	// Separate active and expired bonuses
	var activeBonuses, expiredBonuses []models.Bonus
	for _, b := range result.Bonus {
		if b.Scaduto {
			expiredBonuses = append(expiredBonuses, b)
		} else {
			activeBonuses = append(activeBonuses, b)
		}
	}

	risparmioVal := parseEuroAmount(result.RisparmioStimato)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginL, 15, marginR)
	pdf.SetAutoPageBreak(false, 20)

	isFirstPage := true

	// Footer on every page
	pdf.SetFooterFunc(func() {
		pdf.SetY(-18)
		setDraw(pdf, cInk15)
		pdf.SetLineWidth(0.2)
		pdf.Line(marginL, pdf.GetY(), pageW-marginR, pdf.GetY())
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 7)
		setText(pdf, cInk30)
		pdf.SetX(marginL)
		pdf.CellFormat(contentW/2, 10, transliterate("bonusperme.it -- Servizio gratuito"), "", 0, "L", false, 0, "")
		pdf.CellFormat(contentW/2, 10, fmt.Sprintf("Pagina %d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	// Header on pages 2+ (set via SetHeaderFunc)
	pdf.SetHeaderFunc(func() {
		if isFirstPage {
			return
		}
		pdf.SetY(10)
		pdf.SetX(marginL)
		pdf.SetFont("Helvetica", "B", 7)
		setText(pdf, cBlue)
		pdf.CellFormat(contentW/2, 4, "BonusPerMe", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 7)
		setText(pdf, cInk30)
		pdf.CellFormat(contentW/2, 4, "Report Personalizzato", "", 0, "R", false, 0, "")
		setDraw(pdf, cBlue)
		pdf.SetLineWidth(0.4)
		pdf.Line(marginL, 15, pageW-marginR, 15)
	})

	// ═══════════════════════════════════════════════════
	// PAGE 1 — COVER
	// ═══════════════════════════════════════════════════
	pdf.AddPage()

	// 1. HEADER BLU (full width, 55mm)
	setFill(pdf, cBlue)
	pdf.Rect(0, 0, pageW, 55, "F")

	pdf.SetXY(marginL, 15)
	pdf.SetFont("Helvetica", "B", 26)
	setText(pdf, cWhite)
	pdf.CellFormat(contentW, 10, "BonusPerMe", "", 1, "L", false, 0, "")

	// Decorative line
	pdf.SetXY(marginL, 27)
	pdf.SetDrawColor(255, 255, 255)
	pdf.SetLineWidth(0.3)
	pdf.Line(marginL, 27, marginL+40, 27)

	pdf.SetXY(marginL, 30)
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(255, 255, 255) // white 70%
	pdf.CellFormat(contentW, 6, "Report Personalizzato", "", 1, "L", false, 0, "")

	pdf.SetXY(marginL, 37)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(200, 210, 220) // white 50%
	pdf.CellFormat(contentW, 5, transliterate("Generato il "+dateDisplay), "", 1, "L", false, 0, "")

	// 2. SUMMARY CARD (overlapping the blue header)
	cardW := 140.0
	cardX := (pageW - cardW) / 2
	cardY := 44.0
	cardH := 22.0

	// Shadow
	setFill(pdf, [3]int{200, 200, 200})
	pdf.RoundedRect(cardX+1, cardY+1, cardW, cardH, 4, "1234", "F")
	// Card bg
	setFill(pdf, cWhite)
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(cardX, cardY, cardW, cardH, 4, "1234", "FD")

	// Left: bonus count
	pdf.SetXY(cardX+8, cardY+3)
	pdf.SetFont("Courier", "B", 30)
	setText(pdf, cBlue)
	pdf.CellFormat(cardW/2-8, 10, fmt.Sprintf("%d", result.BonusAttivi), "", 0, "L", false, 0, "")

	pdf.SetXY(cardX+8, cardY+14)
	pdf.SetFont("Helvetica", "", 8)
	setText(pdf, cInk50)
	label := "bonus attivi"
	if result.BonusScaduti > 0 {
		label = fmt.Sprintf("bonus attivi + %d scadut", result.BonusScaduti)
		if result.BonusScaduti == 1 {
			label += "o"
		} else {
			label += "i"
		}
	}
	pdf.CellFormat(cardW/2-8, 4, transliterate(label), "", 0, "L", false, 0, "")

	// Right: risparmio
	pdf.SetXY(cardX+cardW/2, cardY+3)
	pdf.SetFont("Courier", "B", 30)
	setText(pdf, cGreen)
	euroStr := fmtEuro(risparmioVal)
	// Use smaller font if amount is very large
	if len(euroStr) > 8 {
		pdf.SetFont("Courier", "B", 22)
	}
	pdf.CellFormat(cardW/2-8, 10, transliterate("EUR "+euroStr), "", 0, "R", false, 0, "")

	pdf.SetXY(cardX+cardW/2, cardY+14)
	pdf.SetFont("Helvetica", "", 8)
	setText(pdf, cInk50)
	pdf.CellFormat(cardW/2-8, 4, "risparmio stimato", "", 0, "R", false, 0, "")

	// 3. PROFILE SECTION
	pdf.SetY(75)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", 8)
	setText(pdf, cInk30)
	pdf.CellFormat(contentW, 5, "IL TUO PROFILO", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	profY := pdf.GetY()
	profH := 24.0
	setFill(pdf, cCream)
	pdf.RoundedRect(marginL, profY, contentW, profH, 3, "1234", "F")

	colW := contentW / 3
	row1Y := profY + 4
	row2Y := profY + 14

	// Row 1
	profileCell(pdf, marginL+5, row1Y, colW, "Eta", fmt.Sprintf("%d anni", profile.Eta))
	iseeStr := fmtEuro(profile.ISEE)
	profileCell(pdf, marginL+5+colW, row1Y, colW, "ISEE", transliterate("EUR "+iseeStr))
	regioneVal := profile.Residenza
	if regioneVal == "" {
		regioneVal = "-"
	}
	profileCell(pdf, marginL+5+colW*2, row1Y, colW, "Regione", transliterate(regioneVal))

	// Row 2
	figliStr := fmt.Sprintf("%d", profile.NumeroFigli)
	if profile.FigliMinorenni > 0 {
		figliStr += fmt.Sprintf(" (%d min.)", profile.FigliMinorenni)
	}
	profileCell(pdf, marginL+5, row2Y, colW, "Figli", figliStr)
	occVal := profile.Occupazione
	if occVal == "" {
		occVal = "-"
	}
	profileCell(pdf, marginL+5+colW, row2Y, colW, "Occupazione", transliterate(occVal))
	civVal := profile.StatoCivile
	if civVal == "" {
		civVal = "-"
	}
	profileCell(pdf, marginL+5+colW*2, row2Y, colW, "Stato civile", transliterate(civVal))

	// 4. PANORAMICA BONUS
	pdf.SetY(profY + profH + 8)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", 8)
	setText(pdf, cInk30)
	pdf.CellFormat(contentW, 5, "PANORAMICA", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Active bonuses list
	for _, b := range activeBonuses {
		y := pdf.GetY()
		fg, _ := compatColor(b.Compatibilita)

		// Colored dot
		setFill(pdf, fg)
		pdf.Circle(marginL+3, y+2, 1.5, "F")

		// Name
		pdf.SetXY(marginL+8, y)
		pdf.SetFont("Helvetica", "", 9)
		setText(pdf, cInk75)
		pdf.CellFormat(contentW-50, 4.5, transliterate(b.Nome), "", 0, "L", false, 0, "")

		// Importo aligned right
		pdf.SetFont("Courier", "B", 9)
		setText(pdf, cBlue)
		importoDisplay := transliterate(b.Importo)
		if importoDisplay == "" {
			importoDisplay = "-"
		}
		pdf.CellFormat(42, 4.5, importoDisplay, "", 1, "R", false, 0, "")
		pdf.Ln(1)
	}

	// Separator
	if len(expiredBonuses) > 0 {
		sepY := pdf.GetY() + 1
		setDraw(pdf, cInk15)
		pdf.SetLineWidth(0.2)
		pdf.Line(marginL, sepY, pageW-marginR, sepY)
		pdf.SetY(sepY + 3)

		for _, b := range expiredBonuses {
			y := pdf.GetY()
			// Red X dot
			setFill(pdf, cRed)
			pdf.Circle(marginL+3, y+2, 1.5, "F")
			setText(pdf, cWhite)
			pdf.SetFont("Helvetica", "B", 5)
			pdf.SetXY(marginL+1.5, y+0.2)
			pdf.CellFormat(3, 3.5, "x", "", 0, "C", false, 0, "")

			// Name in grey
			pdf.SetXY(marginL+8, y)
			pdf.SetFont("Helvetica", "", 9)
			setText(pdf, cInk30)
			pdf.CellFormat(contentW-50, 4.5, transliterate(b.Nome), "", 0, "L", false, 0, "")

			// SCADUTO label
			pdf.SetFont("Helvetica", "B", 7)
			setText(pdf, cRed)
			pdf.CellFormat(42, 4.5, "SCADUTO", "", 1, "R", false, 0, "")
			pdf.Ln(1)
		}
	}

	// Legend
	pdf.Ln(2)
	legendY := pdf.GetY()
	pdf.SetFont("Helvetica", "", 7)
	setText(pdf, cInk30)
	// Green dot
	setFill(pdf, cGreen)
	pdf.Circle(marginL+3, legendY+1.5, 1, "F")
	pdf.SetXY(marginL+6, legendY)
	pdf.CellFormat(15, 3, "alta", "", 0, "L", false, 0, "")
	// Amber dot
	setFill(pdf, cAmber)
	pdf.Circle(marginL+25, legendY+1.5, 1, "F")
	pdf.SetXY(marginL+28, legendY)
	pdf.CellFormat(15, 3, "media", "", 0, "L", false, 0, "")
	// Grey dot
	setFill(pdf, cInk30)
	pdf.Circle(marginL+47, legendY+1.5, 1, "F")
	pdf.SetXY(marginL+50, legendY)
	pdf.CellFormat(15, 3, "bassa", "", 0, "L", false, 0, "")
	// Red X
	if len(expiredBonuses) > 0 {
		setFill(pdf, cRed)
		pdf.Circle(marginL+69, legendY+1.5, 1, "F")
		pdf.SetXY(marginL+72, legendY)
		pdf.CellFormat(15, 3, "scaduto", "", 0, "L", false, 0, "")
	}

	// 5. FOOTER COPERTINA
	pdf.SetY(275)
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.2)
	pdf.Line(marginL, 275, pageW-marginR, 275)
	pdf.SetY(277)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "", 7)
	setText(pdf, cInk30)
	pdf.CellFormat(contentW/2, 4, transliterate("bonusperme.it -- Documento a scopo orientativo"), "", 0, "L", false, 0, "")
	pdf.CellFormat(contentW/2, 4, transliterate("Codice profilo: "+profileCode), "", 0, "R", false, 0, "")

	isFirstPage = false

	// ═══════════════════════════════════════════════════
	// PAGES 2+ — BONUS CARDS
	// ═══════════════════════════════════════════════════

	// Active bonus cards
	for _, b := range activeBonuses {
		needed := estimateBonusH(b)
		y := ensureSpace(pdf, needed)
		if y < 18 {
			y = 18
		}
		pdf.SetY(y)
		drawBonusCardActive(pdf, b, profile)
		pdf.Ln(6)
	}

	// Expired bonus cards (compact)
	for _, b := range expiredBonuses {
		y := ensureSpace(pdf, 40)
		if y < 18 {
			y = 18
		}
		pdf.SetY(y)
		drawBonusCardExpired(pdf, b)
		pdf.Ln(6)
	}

	// ═══════════════════════════════════════════════════
	// LAST PAGE — PROSSIMI PASSI
	// ═══════════════════════════════════════════════════
	ensureSpace(pdf, 160)
	pdf.SetY(ensureSpace(pdf, 160))

	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", 16)
	setText(pdf, cBlue)
	pdf.CellFormat(contentW, 10, "Prossimi passi", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	nextSteps := []struct {
		Num   string
		Title string
		Desc  string
	}{
		{"1", "Verifica i requisiti", "Controlla ogni bonus sui siti ufficiali indicati."},
		{"2", "Prepara i documenti", "ISEE aggiornato, SPID o CIE, documenti d'identita."},
		{"3", "Presenta le domande", "Online sui portali ufficiali o presso un CAF/patronato."},
	}

	for _, step := range nextSteps {
		stepY := pdf.GetY()
		stepH := 22.0
		setFill(pdf, cCream)
		pdf.RoundedRect(marginL, stepY, contentW, stepH, 3, "1234", "F")

		// Number
		pdf.SetXY(marginL+6, stepY+3)
		pdf.SetFont("Courier", "B", 18)
		setText(pdf, cTerra)
		pdf.CellFormat(12, 8, step.Num, "", 0, "L", false, 0, "")

		// Title
		pdf.SetXY(marginL+20, stepY+3)
		pdf.SetFont("Helvetica", "B", 10)
		setText(pdf, cBlue)
		pdf.CellFormat(contentW-26, 6, transliterate(step.Title), "", 1, "L", false, 0, "")

		// Description
		pdf.SetXY(marginL+20, stepY+11)
		pdf.SetFont("Helvetica", "", 8)
		setText(pdf, cInk75)
		pdf.CellFormat(contentW-26, 5, transliterate(step.Desc), "", 1, "L", false, 0, "")

		pdf.SetY(stepY + stepH + 4)
	}

	// Double line separator
	pdf.Ln(4)
	sepLineY := pdf.GetY()
	setDraw(pdf, cInk30)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginL, sepLineY, pageW-marginR, sepLineY)
	pdf.SetLineWidth(0.5)
	pdf.Line(marginL, sepLineY+1.5, pageW-marginR, sepLineY+1.5)

	// Legal footer
	pdf.SetY(sepLineY + 8)
	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", 11)
	setText(pdf, cBlue)
	pdf.CellFormat(contentW, 6, "BonusPerMe", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	setText(pdf, cInk50)
	pdf.CellFormat(contentW, 5, "bonusperme.it", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "", 7.5)
	setText(pdf, cInk30)
	pdf.CellFormat(contentW, 4, "Simone Nogara", "", 1, "C", false, 0, "")
	pdf.CellFormat(contentW, 4, "P.IVA 03817020138 -- C.F. NGRSMN91P14C933V", "", 1, "C", false, 0, "")
	pdf.CellFormat(contentW, 4, "Via Morazzone 4, 22100 Como (CO), Italia", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "I", 7)
	setText(pdf, cInk30)
	pdf.CellFormat(contentW, 4, "Questo documento e a scopo orientativo.", "", 1, "C", false, 0, "")
	pdf.CellFormat(contentW, 4, "Non sostituisce la consulenza di un professionista, CAF o patronato.", "", 1, "C", false, 0, "")

	// ═══════════════════════════════════════════════════
	// OUTPUT
	// ═══════════════════════════════════════════════════
	disposition := "attachment"
	if r.URL.Query().Get("mode") == "inline" {
		disposition = "inline"
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`%s; filename="bonusperme-report-%s.pdf"`, disposition, dateStr))

	if err := pdf.Output(w); err != nil {
		sentryutil.CaptureError(err, map[string]string{"handler": "report", "phase": "pdf-output"})
		http.Error(w, "Errore generazione PDF", http.StatusInternalServerError)
	}
}

// profileCell draws a label+value pair in the profile grid.
func profileCell(pdf *gofpdf.Fpdf, x, y, w float64, label, value string) {
	pdf.SetXY(x, y)
	pdf.SetFont("Helvetica", "", 7)
	setText(pdf, cInk50)
	pdf.CellFormat(w-5, 3.5, label, "", 1, "L", false, 0, "")
	pdf.SetXY(x, y+4)
	pdf.SetFont("Helvetica", "B", 9)
	setText(pdf, cInk75)
	pdf.CellFormat(w-5, 4, value, "", 0, "L", false, 0, "")
}

// drawBonusCardActive draws a full bonus card with all details.
func drawBonusCardActive(pdf *gofpdf.Fpdf, b models.Bonus, profile models.UserProfile) {
	cardX := marginL
	cardInner := marginL + 6
	innerW := contentW - 12

	startY := pdf.GetY()

	// We'll draw the card border at the end once we know the height

	// A) HEADER
	y := startY + 6
	fg, _ := compatColor(b.Compatibilita)

	// Dot
	setFill(pdf, fg)
	pdf.Circle(cardInner+2, y+3, 2.5, "F")

	// Name
	pdf.SetXY(cardInner+7, y)
	pdf.SetFont("Helvetica", "B", 12)
	setText(pdf, cBlue)
	pdf.CellFormat(innerW-50, 6, transliterate(b.Nome), "", 0, "L", false, 0, "")

	// Pill
	pillText := fmt.Sprintf("%d%%", b.Compatibilita)
	pillFg, pillBg := compatColor(b.Compatibilita)
	pillX := pageW - marginR - 6 - pdf.GetStringWidth(pillText) - 8
	drawPill(pdf, pillX, y, pillText, pillBg, pillFg)

	// Ente + Scadenza line
	y += 8
	pdf.SetXY(cardInner+7, y)
	pdf.SetFont("Helvetica", "", 8)
	setText(pdf, cInk50)
	enteScad := transliterate(b.Ente)
	if b.Scadenza != "" {
		enteScad += " -- Scad: "
	}
	pdf.CellFormat(0, 4.5, enteScad, "", 0, "L", false, 0, "")
	if b.Scadenza != "" {
		x := cardInner + 7 + pdf.GetStringWidth(enteScad)
		pdf.SetXY(x, y)
		setText(pdf, cTerra)
		pdf.CellFormat(0, 4.5, transliterate(b.Scadenza), "", 0, "L", false, 0, "")
	}
	y += 7

	// B) IMPORTO BOX
	boxY := y
	boxH := 16.0
	setFill(pdf, cCream)
	pdf.RoundedRect(cardInner, boxY, innerW, boxH, 2, "1234", "F")

	pdf.SetXY(cardInner+4, boxY+3)
	pdf.SetFont("Courier", "B", 11)
	setText(pdf, cGreen)
	importoText := transliterate(b.Importo)
	if importoText == "" {
		importoText = "Vedi sito ufficiale"
		pdf.SetFont("Helvetica", "", 9)
		setText(pdf, cInk50)
	}
	pdf.CellFormat(innerW-8, 5, importoText, "", 1, "L", false, 0, "")

	if b.ImportoReale != "" && b.ImportoReale != b.Importo {
		pdf.SetXY(cardInner+4, boxY+9)
		pdf.SetFont("Helvetica", "", 7.5)
		setText(pdf, cInk50)
		pdf.CellFormat(innerW-8, 4, transliterate("Stimato per te: "+b.ImportoReale), "", 0, "L", false, 0, "")

		// "STIMATO PER TE" label top right
		pdf.SetFont("Helvetica", "B", 6.5)
		setText(pdf, cGreen)
		labelW := pdf.GetStringWidth("STIMATO PER TE") + 4
		pdf.SetXY(cardInner+innerW-labelW-4, boxY+2)
		pdf.CellFormat(labelW, 3.5, "STIMATO PER TE", "", 0, "R", false, 0, "")
	}
	y = boxY + boxH + 3

	// C) DESCRIZIONE
	pdf.SetXY(cardInner, y)
	pdf.SetFont("Helvetica", "", 8.5)
	setText(pdf, cInk75)
	desc := b.Descrizione
	if len(desc) > 300 {
		desc = desc[:297] + "..."
	}
	pdf.MultiCell(innerW, 4.5, transliterate(desc), "", "L", false)
	y = pdf.GetY() + 2

	// D) SEPARATOR
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.2)
	pdf.Line(cardInner, y, cardInner+innerW, y)
	y += 3

	// E) REQUISITI
	if len(b.Requisiti) > 0 {
		y = ensureSpace(pdf, float64(len(b.Requisiti))*5.5+10)
		pdf.SetXY(cardInner, y)
		pdf.SetFont("Helvetica", "B", 7.5)
		setText(pdf, cInk30)
		pdf.CellFormat(innerW, 5, "REQUISITI", "", 1, "L", false, 0, "")
		y += 5

		for _, req := range b.Requisiti {
			drawCheckGreen(pdf, cardInner+1, y)
			pdf.SetXY(cardInner+6, y-0.5)
			pdf.SetFont("Helvetica", "", 8)
			setText(pdf, cInk75)
			pdf.CellFormat(innerW-6, 4.5, transliterate(req), "", 1, "L", false, 0, "")
			y += 5.5
		}
		y += 2
	}

	// F) COME FARE DOMANDA
	if len(b.ComeRichiederlo) > 0 {
		y = ensureSpace(pdf, float64(len(b.ComeRichiederlo))*5.5+10)
		pdf.SetXY(cardInner, y)
		pdf.SetFont("Helvetica", "B", 7.5)
		setText(pdf, cInk30)
		pdf.CellFormat(innerW, 5, "COME FARE DOMANDA", "", 1, "L", false, 0, "")
		y += 5

		for stepIdx, step := range b.ComeRichiederlo {
			drawStepCircle(pdf, cardInner+1, y, stepIdx+1)
			pdf.SetXY(cardInner+6, y-0.5)
			pdf.SetFont("Helvetica", "", 8)
			setText(pdf, cInk75)
			pdf.CellFormat(innerW-6, 4.5, transliterate(step), "", 1, "L", false, 0, "")
			y += 5.5
		}
		y += 2
	}

	// G) DOCUMENTI
	if len(b.Documenti) > 0 {
		y = ensureSpace(pdf, float64(len(b.Documenti))*5.5+10)
		pdf.SetXY(cardInner, y)
		pdf.SetFont("Helvetica", "B", 7.5)
		setText(pdf, cInk30)
		pdf.CellFormat(innerW, 5, "DOCUMENTI", "", 1, "L", false, 0, "")
		y += 5

		for _, doc := range b.Documenti {
			drawCheckboxEmpty(pdf, cardInner+1, y)
			pdf.SetXY(cardInner+6, y-0.5)
			pdf.SetFont("Helvetica", "", 8)
			setText(pdf, cInk75)
			pdf.CellFormat(innerW-6, 4.5, transliterate(doc), "", 1, "L", false, 0, "")
			y += 5.5
		}
		y += 2
	}

	// H) FOOTER CARD
	y = pdf.GetY()
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.2)
	pdf.Line(cardInner, y, cardInner+innerW, y)
	y += 3

	if b.LinkUfficiale != "" {
		pdf.SetXY(cardInner, y)
		pdf.SetFont("Helvetica", "", 7.5)
		setText(pdf, cBlueMid)
		linkText := truncURL(b.LinkUfficiale, 55)
		pdf.WriteLinkString(4, transliterate(linkText), b.LinkUfficiale)
	}

	if b.Scadenza != "" {
		pdf.SetFont("Helvetica", "", 7.5)
		setText(pdf, cTerra)
		scadW := pdf.GetStringWidth(transliterate(b.Scadenza)) + 2
		pdf.SetXY(cardInner+innerW-scadW, y)
		pdf.CellFormat(scadW, 4, transliterate(b.Scadenza), "", 0, "R", false, 0, "")
	}
	y += 6

	// Draw card border around everything
	endY := y
	cardH := endY - startY
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(cardX, startY, contentW, cardH, 3, "1234", "D")

	pdf.SetY(endY)
}

// drawBonusCardExpired draws a compact card for an expired bonus.
func drawBonusCardExpired(pdf *gofpdf.Fpdf, b models.Bonus) {
	cardX := marginL
	cardInner := marginL + 6
	innerW := contentW - 12

	startY := pdf.GetY()
	y := startY + 6

	// Red X dot
	setFill(pdf, cRed)
	pdf.Circle(cardInner+2, y+3, 2.5, "F")
	setText(pdf, cWhite)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(cardInner, y+0.5)
	pdf.CellFormat(4, 5, "x", "", 0, "C", false, 0, "")

	// Name in grey
	pdf.SetXY(cardInner+7, y)
	pdf.SetFont("Helvetica", "B", 12)
	setText(pdf, cInk30)
	pdf.CellFormat(innerW-50, 6, transliterate(b.Nome), "", 0, "L", false, 0, "")

	// [SCADUTO] pill
	drawPill(pdf, pageW-marginR-6-pdf.GetStringWidth("SCADUTO")-8, y, "SCADUTO", cRedBg, cRed)
	y += 10

	// Importo in grey (barred)
	pdf.SetXY(cardInner+7, y)
	pdf.SetFont("Courier", "B", 10)
	setText(pdf, cInk30)
	importoText := transliterate(b.Importo)
	if importoText != "" {
		pdf.CellFormat(0, 5, importoText, "", 0, "L", false, 0, "")
		// Strikethrough line
		strW := pdf.GetStringWidth(importoText)
		setDraw(pdf, cInk30)
		pdf.SetLineWidth(0.3)
		pdf.Line(cardInner+7, y+2.5, cardInner+7+strW, y+2.5)
	}
	y += 8

	// Note
	pdf.SetXY(cardInner+7, y)
	pdf.SetFont("Helvetica", "I", 8)
	setText(pdf, cInk50)
	nota := "Questo bonus non e piu disponibile."
	if b.Scadenza != "" {
		nota += " Scaduto il " + b.Scadenza + "."
	}
	pdf.CellFormat(innerW-7, 4.5, transliterate(nota), "", 1, "L", false, 0, "")
	y += 8

	// Card border
	cardH := y - startY
	setDraw(pdf, cInk15)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(cardX, startY, contentW, cardH, 3, "1234", "D")

	pdf.SetY(y)
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
