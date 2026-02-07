package matcher

import (
	"bonus360/internal/models"
	"fmt"
	"sort"
)

func GetAllBonus() []models.Bonus {
	return []models.Bonus{
		{
			ID: "assegno-unico", Nome: "Assegno Unico Universale", Categoria: "famiglia",
			Descrizione: "Assegno mensile per ogni figlio a carico fino a 21 anni. Importo da €57 a €199,4/mese per figlio in base all'ISEE, con maggiorazioni per famiglie numerose e figli piccoli.",
			Importo: "da €57 a €199,4/mese per figlio", Scadenza: "Domanda entro il 28 febbraio per arretrati",
			Requisiti:       []string{"Figli a carico sotto i 21 anni", "Residenza in Italia", "ISEE valido (facoltativo)"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Assegno Unico'", "Compilare domanda online"},
			LinkUfficiale:   "https://www.inps.it/it/it/schede/prestazioni-e-servizi/assegno-unico-e-universale-per-i-figli-a-carico.html",
			Ente:            "INPS",
		},
		{
			ID: "bonus-nido", Nome: "Bonus Asilo Nido", Categoria: "famiglia",
			Descrizione: "Contributo per rette asilo nido pubblico/privato o supporto domiciliare per bimbi sotto 3 anni con patologie croniche.",
			Importo: "fino a €3.600/anno (ISEE ≤ €25.000)", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Figli sotto i 3 anni", "Iscrizione asilo nido", "ISEE in corso di validità"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Bonus Nido'", "Allegare ricevute rette + ISEE"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "bonus-nascita", Nome: "Carta per i Nuovi Nati", Categoria: "famiglia",
			Descrizione: "Contributo una tantum di €1.000 per ogni figlio nato o adottato dal 2025 per nuclei con ISEE fino a €40.000.",
			Importo: "€1.000 una tantum", Scadenza: "Entro 60 giorni dalla nascita",
			Requisiti:       []string{"Figlio nato/adottato dal 2025", "ISEE fino a €40.000", "Residenza in Italia"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Carta nuovi nati'", "Domanda online entro 60 giorni"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "bonus-mamma", Nome: "Bonus Mamme Lavoratrici", Categoria: "famiglia",
			Descrizione: "Esonero totale contributi previdenziali (fino a €3.000/anno) per madri lavoratrici dipendenti con almeno 2 figli.",
			Importo: "fino a €3.000/anno", Scadenza: "In vigore",
			Requisiti:       []string{"Madre lavoratrice dipendente", "Almeno 2 figli", "Figlio più piccolo sotto i 10 anni"},
			ComeRichiederlo: []string{"Comunicare al datore di lavoro i CF dei figli", "Esonero automatico in busta paga"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "bonus-ristrutturazione", Nome: "Bonus Ristrutturazione", Categoria: "casa",
			Descrizione: "Detrazione IRPEF 50% sulle spese di ristrutturazione edilizia fino a €96.000 per unità immobiliare (prima casa). 36% per seconde case dal 2025.",
			Importo: "detrazione 50% fino a €96.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Proprietario/titolare diritto reale", "Lavori manutenzione straordinaria", "Pagamento con bonifico parlante"},
			ComeRichiederlo: []string{"Pagare con bonifico parlante", "Conservare fatture", "Indicare in dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it/portale/ristrutturazioni-edilizie",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "bonus-mobili", Nome: "Bonus Mobili ed Elettrodomestici", Categoria: "casa",
			Descrizione: "Detrazione 50% su acquisto mobili e grandi elettrodomestici per immobile in ristrutturazione, fino a €5.000.",
			Importo: "detrazione 50% fino a €5.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Lavori di ristrutturazione avviati", "Elettrodomestici classe A+ (A per forni)", "Pagamento tracciabile"},
			ComeRichiederlo: []string{"Pagamenti tracciabili", "Conservare ricevute", "Indicare in dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it/portale/bonus-mobili",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "bonus-affitto-giovani", Nome: "Bonus Affitto Giovani Under 31", Categoria: "casa",
			Descrizione: "Detrazione fino a €2.000/anno per 4 anni per giovani tra 20 e 31 anni che prendono in affitto un'abitazione principale.",
			Importo: "fino a €2.000/anno per 4 anni", Scadenza: "In vigore",
			Requisiti:       []string{"Età 20-31 anni", "Reddito ≤ €15.493,71", "Contratto di locazione registrato"},
			ComeRichiederlo: []string{"Indicare in dichiarazione dei redditi", "Conservare contratto registrato"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "prima-casa-under36", Nome: "Agevolazioni Prima Casa Under 36", Categoria: "casa",
			Descrizione: "Esenzione imposte registro, ipotecaria e catastale per acquisto prima casa under 36 con ISEE fino a €40.000.",
			Importo: "esenzione imposte (risparmio €2.000-€8.000)", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Età sotto 36 anni", "ISEE fino a €40.000", "Acquisto prima abitazione"},
			ComeRichiederlo: []string{"Dichiarare requisiti nell'atto notarile", "Presentare ISEE valido"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "ecobonus", Nome: "Ecobonus", Categoria: "casa",
			Descrizione: "Detrazione dal 50% al 65% per interventi di efficientamento energetico: caldaie, infissi, cappotto termico, pannelli solari.",
			Importo: "detrazione 50-65% fino a €100.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Immobile esistente", "Interventi di efficientamento energetico", "Asseverazione tecnica"},
			ComeRichiederlo: []string{"Comunicazione ENEA entro 90 giorni da fine lavori", "Bonifico parlante", "Dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate / ENEA",
		},
		{
			ID: "bonus-verde", Nome: "Bonus Verde", Categoria: "casa",
			Descrizione: "Detrazione 36% su spese per sistemazione a verde di giardini, terrazze, coperture, impianti di irrigazione.",
			Importo: "detrazione 36% fino a €5.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Proprietario o nudo proprietario", "Interventi di sistemazione a verde", "Pagamento tracciabile"},
			ComeRichiederlo: []string{"Pagamento tracciabile", "Conservare fatture", "Dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "bonus-psicologo", Nome: "Bonus Psicologo", Categoria: "salute",
			Descrizione: "Contributo fino a €1.500 per sessioni di psicoterapia con professionisti iscritti all'albo. Importo variabile in base all'ISEE.",
			Importo: "fino a €1.500 (ISEE ≤ €15.000)", Scadenza: "Bando annuale",
			Requisiti:       []string{"ISEE valido", "Residenza in Italia", "Psicoterapeuta iscritto all'albo"},
			ComeRichiederlo: []string{"Portale INPS con SPID/CIE", "Sezione 'Bonus Psicologo'", "Domanda nel periodo di apertura"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "carta-dedicata", Nome: "Carta Dedicata a Te", Categoria: "spesa",
			Descrizione: "Carta prepagata €500 per acquisto beni alimentari di prima necessità e carburante per nuclei con ISEE fino a €15.000.",
			Importo: "€500 su carta prepagata", Scadenza: "Erogazione automatica",
			Requisiti:       []string{"ISEE fino a €15.000", "Nessun altro sostegno al reddito", "Nucleo ≥ 3 persone"},
			ComeRichiederlo: []string{"Erogazione automatica dal Comune", "Ritiro presso uffici postali", "Nessuna domanda necessaria"},
			LinkUfficiale:   "https://www.mef.gov.it",
			Ente:            "Comune / MEF",
		},
		{
			ID: "carta-cultura", Nome: "Carta della Cultura / Merito", Categoria: "istruzione",
			Descrizione: "€500 Carta Cultura per neodiciottenni (ISEE ≤ €35.000) + €500 Carta Merito (diploma con 100). Cumulabili fino a €1.000.",
			Importo: "€500 (fino a €1.000 cumulate)", Scadenza: "Entro 30 giugno dell'anno successivo ai 18 anni",
			Requisiti:       []string{"18 anni compiuti nell'anno precedente", "ISEE ≤ €35.000 (Carta Cultura)", "Diploma con 100 (Carta Merito)"},
			ComeRichiederlo: []string{"Registrarsi su cartacultura.gov.it", "Accesso con SPID", "Generare buoni per acquisti culturali"},
			LinkUfficiale:   "https://www.cartacultura.gov.it",
			Ente:            "Ministero della Cultura",
		},
		{
			ID: "borsa-studio", Nome: "Borse di Studio Universitarie", Categoria: "istruzione",
			Descrizione: "Borsa di studio regionale per studenti universitari meritevoli e con basso ISEE. Copre tasse, vitto e alloggio.",
			Importo: "da €2.000 a €6.000/anno + esenzione tasse", Scadenza: "Bando regionale (luglio-settembre)",
			Requisiti:       []string{"Iscrizione università/AFAM", "ISEE universitario ≤ €23.000-€26.000", "Requisiti di merito (CFU minimi)"},
			ComeRichiederlo: []string{"Portale ente regionale diritto allo studio", "Domanda online nel periodo del bando", "Allegare ISEE universitario"},
			LinkUfficiale:   "https://www.miur.gov.it",
			Ente:            "Regione / Ente DSU",
		},
		{
			ID: "bonus-decoder-tv", Nome: "Bonus TV / Decoder", Categoria: "altro",
			Descrizione: "Contributo per acquisto TV e decoder compatibili con il nuovo digitale terrestre DVB-T2 per famiglie con ISEE fino a €20.000.",
			Importo: "fino a €50 (decoder) / €100 (TV)", Scadenza: "Fino ad esaurimento fondi",
			Requisiti:       []string{"ISEE ≤ €20.000", "Residenza in Italia", "Rottamazione vecchio apparecchio (per bonus TV)"},
			ComeRichiederlo: []string{"Acquistare presso rivenditori aderenti", "Presentare autocertificazione ISEE", "Sconto diretto in negozio"},
			LinkUfficiale:   "https://www.mise.gov.it",
			Ente:            "MISE",
		},
		{
			ID: "adi", Nome: "Assegno di Inclusione (ADI)", Categoria: "sostegno",
			Descrizione: "Sostegno economico per nuclei con minori, disabili, over 60 o in condizione di svantaggio. Sostituisce il Reddito di Cittadinanza.",
			Importo: "fino a €6.000/anno (+ integrazione affitto fino a €3.360)", Scadenza: "In vigore",
			Requisiti:       []string{"ISEE ≤ €9.360", "Nucleo con minori, disabili, over 60", "Residenza in Italia da almeno 5 anni", "Patrimonio mobiliare ≤ €6.000"},
			ComeRichiederlo: []string{"Portale INPS o patronato", "Iscrizione al SIISL", "Colloquio presso servizi sociali"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "sfl", Nome: "Supporto Formazione e Lavoro", Categoria: "lavoro",
			Descrizione: "Indennità di €350/mese per 12 mesi per persone tra 18 e 59 anni occupabili che partecipano a percorsi di formazione o lavoro.",
			Importo: "€350/mese per 12 mesi", Scadenza: "In vigore",
			Requisiti:       []string{"Età 18-59 anni", "ISEE ≤ €6.000", "Non beneficiario ADI", "Partecipazione a percorsi formativi"},
			ComeRichiederlo: []string{"Portale INPS o patronato", "Iscrizione al SIISL", "Adesione a percorso formativo/lavorativo"},
			LinkUfficiale:   "https://www.inps.it",
			Ente:            "INPS",
		},
		{
			ID: "bonus-animali", Nome: "Bonus Animali Domestici", Categoria: "altro",
			Descrizione: "Detrazione del 19% sulle spese veterinarie per animali domestici legalmente detenuti, fino a €550.",
			Importo: "detrazione 19% fino a €550", Scadenza: "In vigore (annuale)",
			Requisiti:       []string{"Possesso legale di animale domestico", "Spese veterinarie documentate", "Franchigia di €129,11"},
			ComeRichiederlo: []string{"Conservare fatture/scontrini veterinario", "Indicare in dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate",
		},
		{
			ID: "bonus-colonnine", Nome: "Bonus Colonnine Ricarica Elettrica", Categoria: "casa",
			Descrizione: "Contributo fino all'80% (max €1.500 per privati) per installazione di infrastrutture di ricarica per veicoli elettrici in ambito domestico.",
			Importo: "fino a €1.500 (80% delle spese)", Scadenza: "Fino ad esaurimento fondi",
			Requisiti:       []string{"Persona fisica residente in Italia", "Installazione in ambito domestico", "Installatore qualificato"},
			ComeRichiederlo: []string{"Portale del Ministero dell'Ambiente", "Domanda online con documentazione", "Erogazione post-installazione"},
			LinkUfficiale:   "https://www.mase.gov.it",
			Ente:            "MASE",
		},
		{
			ID: "bonus-acqua-potabile", Nome: "Bonus Acqua Potabile", Categoria: "casa",
			Descrizione: "Credito d'imposta del 50% sulle spese per sistemi di filtraggio e mineralizzazione dell'acqua potabile, fino a €1.000.",
			Importo: "credito d'imposta 50% fino a €1.000", Scadenza: "31 dicembre 2025",
			Requisiti:       []string{"Acquisto sistemi filtraggio/mineralizzazione", "Comunicazione spese all'Agenzia delle Entrate"},
			ComeRichiederlo: []string{"Comunicazione spese su sito Agenzia Entrate entro febbraio anno successivo", "Indicare in dichiarazione dei redditi"},
			LinkUfficiale:   "https://www.agenziaentrate.gov.it",
			Ente:            "Agenzia delle Entrate",
		},
	}
}

// MatchBonus matches user profile against available bonuses.
// If bonusList is provided, uses that; otherwise falls back to GetAllBonus().
func MatchBonus(profile models.UserProfile, bonusList ...[]models.Bonus) models.MatchResult {
	var allBonus []models.Bonus
	if len(bonusList) > 0 && len(bonusList[0]) > 0 {
		allBonus = bonusList[0]
	} else {
		allBonus = GetAllBonus()
	}
	var matched []models.Bonus
	totalSaving := 0.0

	for _, b := range allBonus {
		score := calcScore(b.ID, profile)
		if score > 0 {
			b.Compatibilita = score
			matched = append(matched, b)
			totalSaving += estimateSaving(b.ID, profile)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Compatibilita > matched[j].Compatibilita
	})

	return models.MatchResult{
		BonusTrovati:     len(matched),
		RisparmioStimato: fmt.Sprintf("€%.0f", totalSaving),
		Bonus:            matched,
	}
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
		return 600
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
