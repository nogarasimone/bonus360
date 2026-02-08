package matcher

import (
	"bonusperme/internal/models"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var italianMonthsMap = map[string]time.Month{
	"gennaio": time.January, "febbraio": time.February, "marzo": time.March,
	"aprile": time.April, "maggio": time.May, "giugno": time.June,
	"luglio": time.July, "agosto": time.August, "settembre": time.September,
	"ottobre": time.October, "novembre": time.November, "dicembre": time.December,
}

var itDateRe = regexp.MustCompile(`(\d{1,2})\s+(gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)\s+(\d{4})`)
var yearOnlyRe = regexp.MustCompile(`\b(20\d{2})\b`)

// isScaduto determines whether a bonus deadline has passed.
func isScaduto(scadenza string) bool {
	if scadenza == "" {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(scadenza))

	// Never-expiring patterns
	neverExpired := []string{"in vigore", "permanente", "annuale", "esaurimento fondi", "erogazione automatica", "entro 60 giorni"}
	for _, pat := range neverExpired {
		if strings.Contains(lower, pat) {
			return false
		}
	}
	// "Bando regionale" / "Bando annuale" patterns
	if strings.Contains(lower, "bando") {
		return false
	}
	// "Domanda entro il 28 febbraio per arretrati" — recurring deadline, not expired
	if strings.Contains(lower, "per arretrati") {
		return false
	}

	now := time.Now()

	// Try Italian date pattern: "31 dicembre 2025"
	// Also handles "Entro il 31 dicembre 2025" since regex finds within string
	if m := itDateRe.FindStringSubmatch(lower); len(m) == 4 {
		day, _ := strconv.Atoi(m[1])
		month := italianMonthsMap[m[2]]
		year, _ := strconv.Atoi(m[3])
		deadline := time.Date(year, month, day, 23, 59, 59, 0, time.UTC)
		return now.After(deadline)
	}

	// Try dd/mm/yyyy pattern
	slashRe := regexp.MustCompile(`(\d{2})/(\d{2})/(\d{4})`)
	if m := slashRe.FindStringSubmatch(scadenza); len(m) == 4 {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		deadline := time.Date(year, time.Month(month), day, 23, 59, 59, 0, time.UTC)
		return now.After(deadline)
	}

	// If just a past year is mentioned (e.g. "2024"), consider expired
	if m := yearOnlyRe.FindStringSubmatch(scadenza); len(m) == 2 {
		year, _ := strconv.Atoi(m[1])
		if year < now.Year() {
			return true
		}
	}

	return false
}

func GetAllBonus() []models.Bonus {
	bonuses := []models.Bonus{
		{
			ID: "assegno-unico", Nome: "Assegno Unico Universale", Categoria: "famiglia",
			Descrizione: "Assegno mensile per ogni figlio a carico fino a 21 anni. Importo da €57 a €199,4/mese per figlio in base all'ISEE, con maggiorazioni per famiglie numerose e figli piccoli.",
			Importo: "da €57 a €199,4/mese per figlio", Scadenza: "Domanda entro il 28 febbraio per arretrati",
			Requisiti:       []string{"Figli a carico sotto i 21 anni", "Residenza in Italia", "ISEE valido (facoltativo)"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Assegno Unico'", "Compilare domanda online"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Codici fiscali di tutti i figli", "Coordinate bancarie/postali (IBAN)"},
			FAQ: []models.FAQ{
				{Domanda: "Posso richiederlo se sono separato/a?", Risposta: "Sì, l'assegno spetta al genitore che ha i figli a carico. In caso di affido condiviso, può essere diviso al 50%."},
				{Domanda: "Serve il commercialista?", Risposta: "No, la domanda si fa online sul portale INPS con SPID o CIE. In alternativa puoi rivolgerti a un patronato gratuitamente."},
				{Domanda: "Quanto tempo ci vuole per ricevere i soldi?", Risposta: "Generalmente 30-60 giorni dalla domanda. Il pagamento avviene mensilmente tramite bonifico."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.assegno-unico-e-universale-per-i-figli-a-carico-55984.assegno-unico-e-universale-per-i-figli-a-carico.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=assegno+unico+universale+figli",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.assegno-unico-e-universale-per-i-figli-a-carico-55984.assegno-unico-e-universale-per-i-figli-a-carico.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"D.Lgs. 29 dicembre 2021, n. 230", "Circolare INPS n. 33 del 4 febbraio 2025 — Aggiornamento importi"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-nido", Nome: "Bonus Asilo Nido", Categoria: "famiglia",
			Descrizione: "Contributo per rette asilo nido pubblico/privato o supporto domiciliare per bimbi sotto 3 anni con patologie croniche.",
			Importo: "fino a €3.600/anno (ISEE ≤ €25.000)", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Figli sotto i 3 anni", "Iscrizione asilo nido", "ISEE in corso di validità"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Bonus Nido'", "Allegare ricevute rette + ISEE"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Ricevute di pagamento rette asilo", "Iscrizione/frequenza del minore"},
			FAQ: []models.FAQ{
				{Domanda: "Vale anche per asili nido privati?", Risposta: "Sì, il bonus copre sia asili nido pubblici che privati autorizzati, con importi diversi in base all'ISEE."},
				{Domanda: "Posso cumularlo con l'Assegno Unico?", Risposta: "Sì, bonus nido e Assegno Unico sono pienamente cumulabili."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.bonus-asilo-nido-e-forme-di-supporto-presso-la-propria-abitazione-51105.bonus-asilo-nido-e-forme-di-supporto-presso-la-propria-abitazione.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=bonus+asilo+nido",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.bonus-asilo-nido-e-forme-di-supporto-presso-la-propria-abitazione-51105.bonus-asilo-nido-e-forme-di-supporto-presso-la-propria-abitazione.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"Legge di Bilancio 2025, art. 1 comma 177", "Circolare INPS n. 27/2025"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-nascita", Nome: "Carta per i Nuovi Nati", Categoria: "famiglia",
			Descrizione: "Contributo una tantum di €1.000 per ogni figlio nato o adottato dal 2025 per nuclei con ISEE fino a €40.000.",
			Importo: "€1.000 una tantum", Scadenza: "Entro 60 giorni dalla nascita",
			Requisiti:       []string{"Figlio nato/adottato dal 2025", "ISEE fino a €40.000", "Residenza in Italia"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Carta nuovi nati'", "Domanda online entro 60 giorni"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Certificato di nascita o adozione", "Coordinate bancarie (IBAN)"},
			FAQ: []models.FAQ{
				{Domanda: "Vale per adozioni internazionali?", Risposta: "Sì, il bonus spetta anche per adozioni nazionali e internazionali perfezionate dal 2025."},
				{Domanda: "Entro quando devo fare domanda?", Risposta: "La domanda va presentata entro 60 giorni dalla nascita o dall'ingresso in famiglia del minore adottato."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.carta-per-i-nuovi-nati.carta-per-i-nuovi-nati.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=carta+nuovi+nati+bonus+nascita",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.carta-per-i-nuovi-nati.carta-per-i-nuovi-nati.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"Legge di Bilancio 2025, art. 1 commi 206-208"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-mamma", Nome: "Bonus Mamme Lavoratrici", Categoria: "famiglia",
			Descrizione: "Esonero totale contributi previdenziali (fino a €3.000/anno) per madri lavoratrici dipendenti con almeno 2 figli.",
			Importo: "fino a €3.000/anno", Scadenza: "In vigore",
			Requisiti:       []string{"Madre lavoratrice dipendente", "Almeno 2 figli", "Figlio più piccolo sotto i 10 anni"},
			ComeRichiederlo: []string{"Comunicare al datore di lavoro i CF dei figli", "Esonero automatico in busta paga"},
			Documenti:       []string{"Codici fiscali dei figli", "Comunicazione al datore di lavoro"},
			FAQ: []models.FAQ{
				{Domanda: "Vale per le lavoratrici part-time?", Risposta: "Sì, l'esonero si applica anche alle lavoratrici part-time, proporzionalmente all'orario."},
				{Domanda: "Devo fare domanda io o il datore di lavoro?", Risposta: "Basta comunicare al datore di lavoro i codici fiscali dei figli. L'esonero viene applicato automaticamente in busta paga."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/schede/prestazioni-e-servizi/esonero-contributivo-per-le-lavoratrici-madri.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=esonero+contributivo+lavoratrici+madri",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/schede/prestazioni-e-servizi/esonero-contributivo-per-le-lavoratrici-madri.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"Legge di Bilancio 2024, art. 1 commi 180-182", "Circolare INPS n. 7/2024"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-ristrutturazione", Nome: "Bonus Ristrutturazione", Categoria: "casa",
			Descrizione: "Detrazione IRPEF 50% sulle spese di ristrutturazione edilizia fino a €96.000 per unità immobiliare (prima casa). 36% per seconde case dal 2025.",
			Importo: "detrazione 50% fino a €96.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Proprietario/titolare diritto reale", "Lavori manutenzione straordinaria", "Pagamento con bonifico parlante"},
			ComeRichiederlo: []string{"Pagare con bonifico parlante", "Conservare fatture", "Indicare in dichiarazione dei redditi"},
			Documenti:       []string{"Fatture e ricevute dei lavori", "Bonifici parlanti", "Titoli abilitativi (CILA/SCIA)", "Dati catastali dell'immobile"},
			FAQ: []models.FAQ{
				{Domanda: "Posso cedere il credito?", Risposta: "Dal 2025 la cessione del credito e lo sconto in fattura non sono più disponibili per le nuove pratiche, salvo eccezioni per il Superbonus."},
				{Domanda: "Devo fare la pratica prima di iniziare i lavori?", Risposta: "Per la manutenzione straordinaria serve la CILA prima dell'inizio lavori. Per la manutenzione ordinaria su parti condominiali basta la delibera assembleare."},
				{Domanda: "In quanti anni si recupera?", Risposta: "La detrazione si recupera in 10 rate annuali di pari importo nella dichiarazione dei redditi."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni/detrazione-per-le-ristrutturazioni-edilizie",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=ristrutturazione+edilizia+detrazione",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni/detrazione-per-le-ristrutturazioni-edilizie",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Art. 16-bis DPR 917/1986 (TUIR)", "Legge di Bilancio 2025 — Aliquote 50% prima casa, 36% altre"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-mobili", Nome: "Bonus Mobili ed Elettrodomestici", Categoria: "casa",
			Descrizione: "Detrazione 50% su acquisto mobili e grandi elettrodomestici per immobile in ristrutturazione, fino a €5.000.",
			Importo: "detrazione 50% fino a €5.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Lavori di ristrutturazione avviati", "Elettrodomestici classe A+ (A per forni)", "Pagamento tracciabile"},
			ComeRichiederlo: []string{"Pagamenti tracciabili", "Conservare ricevute", "Indicare in dichiarazione dei redditi"},
			Documenti:       []string{"Fatture di acquisto mobili/elettrodomestici", "Ricevute bonifico o carta", "Documentazione ristrutturazione in corso"},
			FAQ: []models.FAQ{
				{Domanda: "Posso comprare mobili anche prima della fine dei lavori?", Risposta: "Sì, basta che la ristrutturazione sia iniziata. I mobili possono essere acquistati anche prima della conclusione dei lavori."},
				{Domanda: "Quali elettrodomestici sono inclusi?", Risposta: "Grandi elettrodomestici di classe energetica A+ (A per forni): frigoriferi, lavatrici, lavastoviglie, forni, condizionatori, ecc."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/bonus-mobili",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=bonus+mobili+elettrodomestici",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni/bonus-mobili",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Art. 16, comma 2, DL 63/2013", "Legge di Bilancio 2025"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-affitto-giovani", Nome: "Bonus Affitto Giovani Under 31", Categoria: "casa",
			Descrizione: "Detrazione fino a €2.000/anno per 4 anni per giovani tra 20 e 31 anni che prendono in affitto un'abitazione principale.",
			Importo: "fino a €2.000/anno per 4 anni", Scadenza: "In vigore",
			Requisiti:       []string{"Età 20-31 anni", "Reddito ≤ €15.493,71", "Contratto di locazione registrato"},
			ComeRichiederlo: []string{"Indicare in dichiarazione dei redditi", "Conservare contratto registrato"},
			Documenti:       []string{"Contratto di locazione registrato", "Dichiarazione dei redditi", "Documento d'identità"},
			FAQ: []models.FAQ{
				{Domanda: "Vale se convivo con il mio partner?", Risposta: "Sì, purché il contratto sia intestato a te e l'immobile sia la tua abitazione principale."},
				{Domanda: "Posso usarlo se sono studente fuori sede?", Risposta: "Sì, anche gli studenti fuori sede possono accedere al bonus se rispettano i requisiti di età e reddito."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=bonus+affitto+giovani",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Art. 16, comma 1-ter, TUIR", "Decreto Sostegni-bis, art. 31"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "prima-casa-under36", Nome: "Agevolazioni Prima Casa Under 36", Categoria: "casa",
			Descrizione: "Esenzione imposte registro, ipotecaria e catastale per acquisto prima casa under 36 con ISEE fino a €40.000.",
			Importo: "esenzione imposte (risparmio €2.000-€8.000)", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Età sotto 36 anni", "ISEE fino a €40.000", "Acquisto prima abitazione"},
			ComeRichiederlo: []string{"Dichiarare requisiti nell'atto notarile", "Presentare ISEE valido"},
			Documenti:       []string{"ISEE in corso di validità", "Atto notarile di acquisto", "Documento d'identità", "Autocertificazione requisiti"},
			FAQ: []models.FAQ{
				{Domanda: "Vale anche per mutui già in corso?", Risposta: "No, le agevolazioni si applicano solo ai nuovi acquisti con atto stipulato entro la scadenza prevista."},
				{Domanda: "Posso comprare con il mio partner?", Risposta: "Sì, ma entrambi gli acquirenti devono avere meno di 36 anni e rispettare il limite ISEE."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=prima+casa+under+36+agevolazioni",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"DL 73/2021, art. 64, commi 6-10", "Legge di Bilancio 2025 — Proroga"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "ecobonus", Nome: "Ecobonus", Categoria: "casa",
			Descrizione: "Detrazione dal 50% al 65% per interventi di efficientamento energetico: caldaie, infissi, cappotto termico, pannelli solari.",
			Importo: "detrazione 50-65% fino a €100.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Immobile esistente", "Interventi di efficientamento energetico", "Asseverazione tecnica"},
			ComeRichiederlo: []string{"Comunicazione ENEA entro 90 giorni da fine lavori", "Bonifico parlante", "Dichiarazione dei redditi"},
			Documenti:       []string{"Asseverazione tecnica", "APE pre e post intervento", "Fatture e bonifici parlanti", "Comunicazione ENEA"},
			FAQ: []models.FAQ{
				{Domanda: "Serve un tecnico per la pratica ENEA?", Risposta: "Sì, per la maggior parte degli interventi serve un tecnico abilitato per l'asseverazione e la comunicazione ENEA."},
				{Domanda: "Posso combinare ecobonus e bonus ristrutturazione?", Risposta: "No, per lo stesso intervento non puoi cumulare le due detrazioni. Devi scegliere quella più conveniente."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=ecobonus+efficientamento+energetico",
			Ente:                "Agenzia delle Entrate / ENEA",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			FonteNome:           "Agenzia delle Entrate / ENEA",
			RiferimentiNormativi: []string{"Art. 14, DL 63/2013", "Legge di Bilancio 2025"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-verde", Nome: "Bonus Verde", Categoria: "casa",
			Descrizione: "Detrazione 36% su spese per sistemazione a verde di giardini, terrazze, coperture, impianti di irrigazione.",
			Importo: "detrazione 36% fino a €5.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Proprietario o nudo proprietario", "Interventi di sistemazione a verde", "Pagamento tracciabile"},
			ComeRichiederlo: []string{"Pagamento tracciabile", "Conservare fatture", "Dichiarazione dei redditi"},
			Documenti:       []string{"Fatture dei lavori", "Ricevute pagamento tracciabile", "Autocertificazione proprietà"},
			FAQ: []models.FAQ{
				{Domanda: "Il taglio dell'erba è incluso?", Risposta: "No, la manutenzione ordinaria come il taglio dell'erba non rientra. Sono inclusi solo interventi di sistemazione a verde straordinari."},
				{Domanda: "Vale per i balconi?", Risposta: "Sì, rientrano anche la realizzazione di giardini pensili e coperture a verde su balconi e terrazzi."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=bonus+verde+giardini",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Legge 205/2017, art. 1 commi 12-15"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-psicologo", Nome: "Bonus Psicologo", Categoria: "salute",
			Descrizione: "Contributo per sessioni di psicoterapia con professionisti iscritti all'albo. Importo variabile in base all'ISEE: fino a €1.500 (ISEE ≤ €15.000), €1.000 (ISEE ≤ €30.000), €500 (ISEE ≤ €50.000).",
			Importo: "da €500 a €1.500 in base all'ISEE", Scadenza: "Bando annuale",
			Requisiti:       []string{"ISEE valido", "Residenza in Italia", "Psicoterapeuta iscritto all'albo"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Bonus Psicologo'", "Domanda nel periodo di apertura"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Dati dello psicoterapeuta (nome, cognome, codice albo)"},
			FAQ: []models.FAQ{
				{Domanda: "Quanto ricevo per seduta?", Risposta: "Il bonus copre fino a €50 per seduta, fino al raggiungimento dell'importo totale assegnato in base al tuo ISEE."},
				{Domanda: "Posso scegliere qualsiasi psicologo?", Risposta: "Deve essere uno psicoterapeuta iscritto nell'elenco degli aderenti al bonus psicologo sul portale INPS."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.contributo-per-sostenere-le-spese-relative-a-sessioni-di-psicoterapia-bonus-psicologo.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=bonus+psicologo+psicoterapia",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.contributo-per-sostenere-le-spese-relative-a-sessioni-di-psicoterapia-bonus-psicologo.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"DL 228/2021, art. 1-quater", "DM 24 novembre 2023"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "carta-dedicata", Nome: "Carta Dedicata a Te", Categoria: "spesa",
			Descrizione: "Carta prepagata €500 per acquisto beni alimentari di prima necessità e carburante per nuclei con ISEE fino a €15.000.",
			Importo: "€500 su carta prepagata", Scadenza: "Erogazione automatica",
			Requisiti:       []string{"ISEE fino a €15.000", "Nessun altro sostegno al reddito", "Nucleo ≥ 3 persone"},
			ComeRichiederlo: []string{"Erogazione automatica dal Comune", "Ritiro presso uffici postali", "Nessuna domanda necessaria"},
			Documenti:       []string{"Documento d'identità", "Codice fiscale", "ISEE in corso di validità"},
			FAQ: []models.FAQ{
				{Domanda: "Come faccio a sapere se mi spetta?", Risposta: "L'erogazione è automatica: il Comune identifica i beneficiari in base all'ISEE. Riceverai una comunicazione per il ritiro."},
				{Domanda: "Dove posso usare la carta?", Risposta: "Nei supermercati e negozi alimentari convenzionati, e per l'acquisto di carburante."},
			},
			LinkUfficiale:       "https://www.mef.gov.it/focus/Carta-dedicata-a-te/",
			LinkRicerca:         "https://www.mef.gov.it/cerca/?q=carta+dedicata+a+te",
			Ente:                "Comune / MEF",
			FonteURL:            "https://www.mef.gov.it/focus/Carta-dedicata-a-te/",
			FonteNome:           "Ministero dell'Economia e delle Finanze",
			RiferimentiNormativi: []string{"DL 48/2023, art. 1 comma 450", "Decreto interministeriale 18 luglio 2024"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "carta-cultura", Nome: "Carta della Cultura / Merito", Categoria: "istruzione",
			Descrizione: "€500 Carta Cultura per neodiciottenni (ISEE ≤ €35.000) + €500 Carta Merito (diploma con 100). Cumulabili fino a €1.000.",
			Importo: "€500 (fino a €1.000 cumulate)", Scadenza: "Entro 30 giugno dell'anno successivo ai 18 anni",
			Requisiti:       []string{"18 anni compiuti nell'anno precedente", "ISEE ≤ €35.000 (Carta Cultura)", "Diploma con 100 (Carta Merito)"},
			ComeRichiederlo: []string{"Registrarsi su cartacultura.gov.it", "Accesso con SPID", "Generare buoni per acquisti culturali"},
			Documenti:       []string{"SPID", "Diploma di maturità (per Carta Merito)", "ISEE in corso di validità"},
			FAQ: []models.FAQ{
				{Domanda: "Cosa posso comprare?", Risposta: "Libri, musica, biglietti cinema/teatro/concerti/musei, corsi di formazione, abbonamenti a quotidiani digitali."},
				{Domanda: "Posso averle entrambe?", Risposta: "Sì, se hai sia ISEE ≤ €35.000 sia diploma con 100, puoi cumulare Carta Cultura e Carta Merito per un totale di €1.000."},
			},
			LinkUfficiale:       "https://www.cartacultura.gov.it",
			LinkRicerca:         "https://www.google.com/search?q=site:cartacultura.gov.it+carta+cultura+merito",
			Ente:                "Ministero della Cultura",
			FonteURL:            "https://www.cartacultura.gov.it",
			FonteNome:           "Ministero della Cultura",
			RiferimentiNormativi: []string{"DL 230/2023, art. 1", "DPCM 20 luglio 2023"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "borsa-studio", Nome: "Borse di Studio Universitarie", Categoria: "istruzione",
			Descrizione: "Borsa di studio regionale per studenti universitari meritevoli e con basso ISEE. Copre tasse, vitto e alloggio.",
			Importo: "da €2.000 a €6.000/anno + esenzione tasse", Scadenza: "Bando regionale (luglio-settembre)",
			Requisiti:       []string{"Iscrizione università/AFAM", "ISEE universitario ≤ €23.000-€26.000", "Requisiti di merito (CFU minimi)"},
			ComeRichiederlo: []string{"Portale ente regionale diritto allo studio", "Domanda online nel periodo del bando", "Allegare ISEE universitario"},
			Documenti:       []string{"ISEE universitario", "Iscrizione universitaria", "Piano di studi", "Certificato esami sostenuti"},
			FAQ: []models.FAQ{
				{Domanda: "Devo ripresentare domanda ogni anno?", Risposta: "Sì, la domanda va rinnovata ogni anno accademico, verificando il possesso dei requisiti di reddito e merito."},
				{Domanda: "Se perdo i requisiti di merito devo restituire i soldi?", Risposta: "Non devi restituire quanto già ricevuto, ma perdi il diritto alla borsa per l'anno successivo."},
			},
			LinkUfficiale:       "https://www.miur.gov.it",
			LinkRicerca:         "https://www.google.com/search?q=site:miur.gov.it+borse+di+studio+universitarie",
			Ente:                "Regione / Ente DSU",
			FonteURL:            "https://www.miur.gov.it/borse-di-studio",
			FonteNome:           "Ministero dell'Istruzione e del Merito",
			RiferimentiNormativi: []string{"D.Lgs. 68/2012", "DPCM annuale soglie ISEE"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-decoder-tv", Nome: "Bonus TV / Decoder", Categoria: "altro",
			Descrizione: "Contributo per acquisto TV e decoder compatibili con il nuovo digitale terrestre DVB-T2 per famiglie con ISEE fino a €20.000. Fondi esauriti nel 2024.",
			Importo: "fino a €50 (decoder) / €100 (TV)", Scadenza: "Fondi esauriti (2024)",
			Requisiti:       []string{"ISEE ≤ €20.000", "Residenza in Italia", "Rottamazione vecchio apparecchio (per bonus TV)"},
			ComeRichiederlo: []string{"Acquistare presso rivenditori aderenti", "Presentare autocertificazione ISEE", "Sconto diretto in negozio"},
			Documenti:       []string{"Documento d'identità", "Autocertificazione ISEE", "Vecchio apparecchio da rottamare"},
			FAQ: []models.FAQ{
				{Domanda: "Posso usarlo online?", Risposta: "No, lo sconto si applica solo in negozi fisici aderenti all'iniziativa."},
				{Domanda: "Serve rottamare la vecchia TV?", Risposta: "Per il bonus TV sì, serve la rottamazione. Per il solo decoder non è necessario."},
			},
			LinkUfficiale:       "https://www.mise.gov.it",
			LinkRicerca:         "https://www.google.com/search?q=site:mise.gov.it+bonus+tv+decoder",
			Ente:                "MISE",
			FonteURL:            "https://www.mise.gov.it/index.php/it/incentivi/bonus-tv",
			FonteNome:           "Ministero delle Imprese e del Made in Italy",
			RiferimentiNormativi: []string{"DM 18 ottobre 2021"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "scaduto",
			Scaduto:             true,
		},
		{
			ID: "adi", Nome: "Assegno di Inclusione (ADI)", Categoria: "sostegno",
			Descrizione: "Sostegno economico per nuclei con minori, disabili, over 60 o in condizione di svantaggio. Sostituisce il Reddito di Cittadinanza.",
			Importo: "fino a €6.000/anno (+ integrazione affitto fino a €3.360)", Scadenza: "In vigore",
			Requisiti:       []string{"ISEE ≤ €9.360", "Nucleo con minori, disabili, over 60", "Residenza in Italia da almeno 5 anni", "Patrimonio mobiliare ≤ €6.000"},
			ComeRichiederlo: []string{"Portale INPS o patronato", "Iscrizione al SIISL", "Colloquio presso servizi sociali"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Documento d'identità", "Attestazione disabilità (se applicabile)"},
			FAQ: []models.FAQ{
				{Domanda: "È compatibile con un lavoro part-time?", Risposta: "Sì, fino a un certo reddito da lavoro. L'importo dell'ADI viene ricalcolato in base al reddito percepito."},
				{Domanda: "Quanto dura?", Risposta: "L'ADI dura 18 mesi, rinnovabili per periodi di 12 mesi previo aggiornamento dei requisiti."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.assegno-di-inclusione.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=assegno+di+inclusione+ADI",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/dettaglio-scheda.it.schede-servizio-strumento.schede-servizi.assegno-di-inclusione.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"DL 48/2023, convertito in L. 85/2023", "Circolare INPS n. 105/2023"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "sfl", Nome: "Supporto Formazione e Lavoro", Categoria: "lavoro",
			Descrizione: "Indennità di €350/mese per 12 mesi per persone tra 18 e 59 anni occupabili che partecipano a percorsi di formazione o lavoro.",
			Importo: "€350/mese per 12 mesi", Scadenza: "In vigore",
			Requisiti:       []string{"Età 18-59 anni", "ISEE ≤ €6.000", "Non beneficiario ADI", "Partecipazione a percorsi formativi"},
			ComeRichiederlo: []string{"Portale INPS o patronato", "Iscrizione al SIISL", "Adesione a percorso formativo/lavorativo"},
			Documenti:       []string{"SPID o CIE", "ISEE in corso di validità", "Curriculum vitae", "Iscrizione centro per l'impiego"},
			FAQ: []models.FAQ{
				{Domanda: "Devo frequentare un corso di formazione?", Risposta: "Sì, l'indennità è condizionata alla partecipazione attiva a percorsi formativi o di riqualificazione professionale."},
				{Domanda: "Posso rifiutare offerte di lavoro?", Risposta: "Il rifiuto di un'offerta di lavoro congrua comporta la decadenza dal beneficio."},
			},
			LinkUfficiale:       "https://www.inps.it/it/it/schede/prestazioni-e-servizi/supporto-formazione-e-lavoro.html",
			LinkRicerca:         "https://www.inps.it/it/it/ricerca.html?q=supporto+formazione+e+lavoro+SFL",
			Ente:                "INPS",
			FonteURL:            "https://www.inps.it/it/it/schede/prestazioni-e-servizi/supporto-formazione-e-lavoro.html",
			FonteNome:           "INPS",
			RiferimentiNormativi: []string{"DL 48/2023, art. 12", "Circolare INPS n. 77/2023"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-animali", Nome: "Bonus Animali Domestici", Categoria: "altro",
			Descrizione: "Detrazione del 19% sulle spese veterinarie per animali domestici legalmente detenuti, fino a €550.",
			Importo: "detrazione 19% fino a €550", Scadenza: "In vigore (annuale)",
			Requisiti:       []string{"Possesso legale di animale domestico", "Spese veterinarie documentate", "Franchigia di €129,11"},
			ComeRichiederlo: []string{"Conservare fatture/scontrini veterinario", "Indicare in dichiarazione dei redditi"},
			Documenti:       []string{"Fatture/scontrini veterinario", "Documentazione possesso animale"},
			FAQ: []models.FAQ{
				{Domanda: "Vale per tutti gli animali?", Risposta: "Solo per animali domestici legalmente detenuti (cani, gatti, ecc.). Non si applica ad animali da reddito o allevamento."},
				{Domanda: "Come funziona la franchigia?", Risposta: "La detrazione si applica sulle spese che superano €129,11, fino a un massimo di €550."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/spese-sanitarie-per-animali-da-compagnia",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=spese+veterinarie+animali+detrazione",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/spese-sanitarie-per-animali-da-compagnia",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Art. 15, comma 1, lett. c-bis, TUIR"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-colonnine", Nome: "Bonus Colonnine Ricarica Elettrica", Categoria: "casa",
			Descrizione: "Contributo fino all'80% (max €1.500 per privati) per installazione di infrastrutture di ricarica per veicoli elettrici in ambito domestico.",
			Importo: "fino a €1.500 (80% delle spese)", Scadenza: "Fino ad esaurimento fondi",
			Requisiti:       []string{"Persona fisica residente in Italia", "Installazione in ambito domestico", "Installatore qualificato"},
			ComeRichiederlo: []string{"Portale del Ministero dell'Ambiente", "Domanda online con documentazione", "Erogazione post-installazione"},
			Documenti:       []string{"Fattura installazione", "Certificato installatore qualificato", "Documentazione immobile"},
			FAQ: []models.FAQ{
				{Domanda: "Serve un contatore dedicato?", Risposta: "Non è obbligatorio, ma l'installatore potrebbe consigliarlo per ottimizzare i consumi."},
				{Domanda: "Vale per le colonnine condominiali?", Risposta: "Sì, anche le installazioni in parti comuni condominiali sono ammesse, con limiti di spesa diversi."},
			},
			LinkUfficiale:       "https://www.mase.gov.it",
			LinkRicerca:         "https://www.google.com/search?q=site:mase.gov.it+bonus+colonnine+ricarica+elettrica",
			Ente:                "MASE",
			FonteURL:            "https://www.mase.gov.it/pagina/infrastrutture-di-ricarica-per-veicoli-elettrici",
			FonteNome:           "Ministero dell'Ambiente e della Sicurezza Energetica",
			RiferimentiNormativi: []string{"DM 25 agosto 2021, n. 358"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
		{
			ID: "bonus-acqua-potabile", Nome: "Bonus Acqua Potabile", Categoria: "casa",
			Descrizione: "Credito d'imposta del 50% sulle spese per sistemi di filtraggio e mineralizzazione dell'acqua potabile, fino a €1.000.",
			Importo: "credito d'imposta 50% fino a €1.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Acquisto sistemi filtraggio/mineralizzazione", "Comunicazione spese all'Agenzia delle Entrate"},
			ComeRichiederlo: []string{"Comunicazione spese su sito Agenzia Entrate entro febbraio anno successivo", "Indicare in dichiarazione dei redditi"},
			Documenti:       []string{"Fattura acquisto sistema filtraggio", "Comunicazione all'Agenzia delle Entrate"},
			FAQ: []models.FAQ{
				{Domanda: "Quali sistemi sono ammessi?", Risposta: "Sistemi di filtraggio, mineralizzazione, raffreddamento e addizione di anidride carbonica alimentare."},
				{Domanda: "Devo comunicare l'acquisto?", Risposta: "Sì, devi comunicare le spese sostenute tramite il sito dell'Agenzia delle Entrate entro febbraio dell'anno successivo."},
			},
			LinkUfficiale:       "https://www.agenziaentrate.gov.it/portale/web/guest/bonus-acqua-potabile",
			LinkRicerca:         "https://www.agenziaentrate.gov.it/portale/web/guest/ricerca/-/search/q=bonus+acqua+potabile",
			Ente:                "Agenzia delle Entrate",
			FonteURL:            "https://www.agenziaentrate.gov.it/portale/web/guest/bonus-acqua-potabile",
			FonteNome:           "Agenzia delle Entrate",
			RiferimentiNormativi: []string{"Legge di Bilancio 2021, art. 1 commi 1087-1089"},
			UltimoAggiornamento: "15 gennaio 2025",
			Stato:               "attivo",
		},
	}
	populateValidity(bonuses)
	return bonuses
}

// GetAllBonusWithRegional returns national + regional bonuses combined.
func GetAllBonusWithRegional() []models.Bonus {
	all := GetAllBonus()
	all = append(all, GetRegionalBonus()...)
	return all
}

// MatchBonus matches user profile against available bonuses.
// If bonusList is provided, uses that; otherwise falls back to GetAllBonusWithRegional().
func MatchBonus(profile models.UserProfile, bonusList ...[]models.Bonus) models.MatchResult {
	var allBonus []models.Bonus
	if len(bonusList) > 0 && len(bonusList[0]) > 0 {
		allBonus = bonusList[0]
	} else {
		allBonus = GetAllBonusWithRegional()
	}
	var matched []models.Bonus
	totalSaving := 0.0

	userRegion := strings.ToLower(strings.TrimSpace(profile.Residenza))

	for _, b := range allBonus {
		// Regional filter: if bonus has RegioniApplicabili, user must match
		if len(b.RegioniApplicabili) > 0 {
			if userRegion == "" {
				continue
			}
			found := false
			for _, r := range b.RegioniApplicabili {
				if strings.ToLower(r) == userRegion {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		score := calcScore(b.ID, profile)
		// For regional bonuses not in calcScore, use generic ISEE-based scoring
		if score == 0 && len(b.RegioniApplicabili) > 0 {
			score = calcRegionalScore(b, profile)
		}
		if score > 0 {
			b.Compatibilita = score
			b.ImportoReale = calcImportoReale(b.ID, profile.ISEE, profile)
			matched = append(matched, b)
			saving := estimateSaving(b.ID, profile)
			if saving == 0 && len(b.RegioniApplicabili) > 0 {
				saving = estimateRegionalSaving(b)
			}
			totalSaving += saving
		}
	}

	// Mark expired bonuses and count
	attivi := 0
	scaduti := 0
	activeSaving := 0.0
	for i := range matched {
		matched[i].Scaduto = isScaduto(matched[i].Scadenza)
		if matched[i].Scaduto {
			scaduti++
		} else {
			attivi++
			// Re-estimate saving for active only
			s := estimateSaving(matched[i].ID, profile)
			if s == 0 && len(matched[i].RegioniApplicabili) > 0 {
				s = estimateRegionalSaving(matched[i])
			}
			activeSaving += s
		}
	}

	// Sort: active first (by compat desc), then expired (by compat desc)
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Scaduto != matched[j].Scaduto {
			return !matched[i].Scaduto
		}
		return matched[i].Compatibilita > matched[j].Compatibilita
	})

	// Calculate "perso finora" — monthly bonuses × months elapsed since January
	perso := calcPersoFinora(activeSaving)

	return models.MatchResult{
		BonusTrovati:     len(matched),
		BonusAttivi:      attivi,
		BonusScaduti:     scaduti,
		RisparmioStimato: fmt.Sprintf("€%.0f", activeSaving),
		PersoFinora:      perso,
		Bonus:            matched,
	}
}

// calcRegionalScore scores regional bonuses based on ISEE threshold and category match.
func calcRegionalScore(b models.Bonus, p models.UserProfile) int {
	if b.SogliaISEE > 0 && p.ISEE > 0 && p.ISEE > b.SogliaISEE {
		return 0
	}
	cat := strings.ToLower(b.Categoria)
	switch cat {
	case "famiglia":
		if p.NumeroFigli > 0 || p.FigliMinorenni > 0 || p.FigliUnder3 > 0 {
			if b.SogliaISEE == 0 || p.ISEE == 0 || p.ISEE <= b.SogliaISEE {
				return 75
			}
		}
		return 0
	case "istruzione":
		if p.Studente || p.FigliMinorenni > 0 || p.NumeroFigli > 0 {
			return 70
		}
		return 0
	case "casa":
		if p.Affittuario || p.PrimaAbitazione {
			return 70
		}
		return 0
	case "trasporti":
		// Show to students, young people, over 65
		if p.Studente || p.Eta < 26 || p.Over65 > 0 || (p.ISEE > 0 && p.ISEE <= 30000) {
			return 65
		}
		return 0
	}
	return 50
}

// estimateRegionalSaving estimates annual savings for regional bonuses.
func estimateRegionalSaving(b models.Bonus) float64 {
	cat := strings.ToLower(b.Categoria)
	switch cat {
	case "famiglia":
		return 800
	case "istruzione":
		return 200
	case "casa":
		return 2000
	case "trasporti":
		return 400
	}
	return 300
}

// calcPersoFinora calculates the estimated amount lost since January.
func calcPersoFinora(annualSaving float64) string {
	if annualSaving <= 0 {
		return ""
	}
	monthsElapsed := float64(time.Now().Month() - 1)
	if monthsElapsed <= 0 {
		return ""
	}
	monthlySaving := annualSaving / 12
	perso := monthlySaving * monthsElapsed
	if perso < 10 {
		return ""
	}
	return fmt.Sprintf("€%.0f", perso)
}

func calcImportoReale(bonusID string, isee float64, profile models.UserProfile) string {
	switch bonusID {
	case "assegno-unico":
		var perFiglio float64
		switch {
		case isee > 0 && isee <= 17090.61:
			perFiglio = 199.4
		case isee > 17090.61 && isee <= 45574.96:
			// Linear interpolation between 199.4 (at 17090.61) and 57 (at 45574.96)
			perFiglio = 199.4 - (isee-17090.61)/(45574.96-17090.61)*(199.4-57)
		default:
			// ISEE > 45574.96 or ISEE == 0
			perFiglio = 57
		}

		// Maggiorazione per figli under 3 (used as proxy for under 1 year)
		under3Bonus := 91.40 * float64(profile.FigliUnder3)

		// Maggiorazione from 3rd child onward
		var thirdChildBonus float64
		if profile.NumeroFigli >= 3 {
			thirdChildBonus = 17.10 * float64(profile.NumeroFigli-2)
		}

		monthly := math.Round(perFiglio*float64(profile.NumeroFigli)*100) / 100
		monthly += under3Bonus + thirdChildBonus
		monthly = math.Round(monthly*100) / 100
		yearly := math.Round(monthly*12*100) / 100

		return fmt.Sprintf("€%.2f/mese (€%.2f/anno)", monthly, yearly)

	case "bonus-nido":
		switch {
		case isee > 0 && isee <= 25000:
			return "€3.600/anno"
		case isee > 25000 && isee <= 40000:
			return "€2.500/anno"
		default:
			return "€1.500/anno"
		}

	case "bonus-psicologo":
		switch {
		case isee > 0 && isee <= 15000:
			return "fino a €1.500"
		case isee > 15000 && isee <= 30000:
			return "fino a €1.000"
		case isee > 30000 && isee <= 50000:
			return "fino a €500"
		}
	}

	return ""
}

func calcScore(id string, p models.UserProfile) int {
	switch id {
	case "assegno-unico":
		if p.NumeroFigli > 0 {
			if p.ISEE > 0 && p.ISEE <= 17000 {
				return 98
			}
			return 85
		}
	case "bonus-nido":
		if p.FigliUnder3 > 0 {
			if p.ISEE > 0 && p.ISEE <= 25000 {
				return 95
			}
			return 70
		}
	case "bonus-nascita":
		if p.NuovoNato2025 && (p.ISEE == 0 || p.ISEE <= 40000) {
			return 95
		}
	case "bonus-mamma":
		if p.NumeroFigli >= 2 && p.Occupazione == "dipendente" && p.StatoCivile != "single" {
			return 85
		}
	case "bonus-ristrutturazione":
		if p.RistrutturazCasa {
			return 90
		}
	case "bonus-mobili":
		if p.RistrutturazCasa {
			return 80
		}
	case "bonus-affitto-giovani":
		if p.Eta >= 20 && p.Eta <= 31 && p.Affittuario {
			if p.RedditoAnnuo > 0 && p.RedditoAnnuo <= 15493 {
				return 95
			}
			return 60
		}
	case "prima-casa-under36":
		if p.Eta > 0 && p.Eta < 36 && p.PrimaAbitazione {
			if p.ISEE > 0 && p.ISEE <= 40000 {
				return 95
			}
			return 70
		}
	case "ecobonus":
		if p.RistrutturazCasa {
			return 75
		}
	case "bonus-verde":
		if p.RistrutturazCasa || p.PrimaAbitazione {
			return 50
		}
	case "bonus-psicologo":
		if p.ISEE > 0 && p.ISEE <= 50000 {
			return 70
		}
		return 40 // always show, very popular
	case "carta-dedicata":
		if p.ISEE > 0 && p.ISEE <= 15000 && (p.NumeroFigli+1+p.Over65) >= 3 {
			return 90
		}
	case "carta-cultura":
		if p.Eta == 18 || p.Eta == 19 {
			if p.ISEE > 0 && p.ISEE <= 35000 {
				return 95
			}
			return 70
		}
	case "borsa-studio":
		if p.Studente && (p.ISEE == 0 || p.ISEE <= 26000) {
			return 90
		}
	case "bonus-decoder-tv":
		if p.ISEE > 0 && p.ISEE <= 20000 {
			return 40
		}
	case "adi":
		if p.ISEE > 0 && p.ISEE <= 9360 {
			if p.FigliMinorenni > 0 || p.Disabilita || p.Over65 > 0 {
				return 95
			}
		}
	case "sfl":
		if p.Eta >= 18 && p.Eta <= 59 && p.ISEE > 0 && p.ISEE <= 6000 {
			if p.Occupazione == "disoccupato" || p.Occupazione == "inoccupato" {
				return 90
			}
		}
	case "bonus-animali":
		return 30 // generic, always show low
	case "bonus-colonnine":
		if p.PrimaAbitazione || p.RistrutturazCasa {
			return 40
		}
	case "bonus-acqua-potabile":
		if p.PrimaAbitazione || p.RistrutturazCasa {
			return 35
		}
	}
	return 0
}

func estimateSaving(id string, p models.UserProfile) float64 {
	switch id {
	case "assegno-unico":
		base := 1500.0
		if p.ISEE > 0 && p.ISEE <= 17000 {
			base = 2400.0
		}
		return base * float64(p.NumeroFigli)
	case "bonus-nido":
		if p.ISEE <= 25000 {
			return 3600
		}
		return 1500
	case "bonus-nascita":
		return 1000
	case "bonus-mamma":
		return 3000
	case "bonus-ristrutturazione":
		return 5000
	case "bonus-mobili":
		return 2500
	case "bonus-affitto-giovani":
		return 2000
	case "prima-casa-under36":
		return 5000
	case "ecobonus":
		return 3000
	case "bonus-verde":
		return 900
	case "bonus-psicologo":
		switch {
		case p.ISEE > 0 && p.ISEE <= 15000:
			return 1500
		case p.ISEE > 15000 && p.ISEE <= 30000:
			return 1000
		case p.ISEE > 30000 && p.ISEE <= 50000:
			return 500
		default:
			return 600
		}
	case "carta-dedicata":
		return 500
	case "carta-cultura":
		return 500
	case "borsa-studio":
		return 4000
	case "bonus-decoder-tv":
		return 50
	case "adi":
		return 6000
	case "sfl":
		return 4200
	case "bonus-animali":
		return 100
	case "bonus-colonnine":
		return 1500
	case "bonus-acqua-potabile":
		return 500
	}
	return 0
}

// populateValidity auto-derives TipoScadenza, ScadenzaDomanda, AnnoConferma, UltimaVerifica
// from existing Scadenza text for each bonus.
func populateValidity(bonuses []models.Bonus) {
	now := time.Now()
	for i := range bonuses {
		b := &bonuses[i]
		lower := strings.ToLower(strings.TrimSpace(b.Scadenza))

		// Derive TipoScadenza from Scadenza text
		switch {
		case strings.Contains(lower, "in vigore"):
			b.TipoScadenza = "permanente"
		case strings.Contains(lower, "erogazione automatica"):
			b.TipoScadenza = "permanente"
		case strings.Contains(lower, "entro 60 giorni"):
			b.TipoScadenza = "permanente"
		case strings.Contains(lower, "per arretrati"):
			b.TipoScadenza = "permanente"
		case strings.Contains(lower, "entro 30 giugno"):
			b.TipoScadenza = "permanente"
		case strings.Contains(lower, "esaurimento fondi"):
			b.TipoScadenza = "esaurimento_fondi"
		case strings.Contains(lower, "bando"):
			b.TipoScadenza = "bando_annuale"
		default:
			// Try to parse Italian date
			if m := itDateRe.FindStringSubmatch(lower); len(m) == 4 {
				day, _ := strconv.Atoi(m[1])
				month := italianMonthsMap[m[2]]
				year, _ := strconv.Atoi(m[3])
				b.TipoScadenza = "data_fissa"
				b.ScadenzaDomanda = time.Date(year, month, day, 23, 59, 59, 0, time.UTC)
			} else {
				b.TipoScadenza = "permanente"
			}
		}

		// AnnoConferma: derive from UltimoAggiornamento text
		if b.UltimoAggiornamento != "" {
			if m := yearOnlyRe.FindStringSubmatch(b.UltimoAggiornamento); len(m) == 2 {
				y, _ := strconv.Atoi(m[1])
				b.AnnoConferma = y
			}
		}
		if b.AnnoConferma == 0 {
			b.AnnoConferma = 2025
		}

		// UltimaVerifica: set to now (data is loaded from code)
		b.UltimaVerifica = now
	}
}
