package models

type UserProfile struct {
	Eta              int     `json:"eta"`
	Residenza        string  `json:"residenza"`
	Comune           string  `json:"comune"`
	StatoCivile      string  `json:"stato_civile"`
	Occupazione      string  `json:"occupazione"`
	NumeroFigli      int     `json:"numero_figli"`
	FigliMinorenni   int     `json:"figli_minorenni"`
	FigliUnder3      int     `json:"figli_under3"`
	Disabilita       bool    `json:"disabilita"`
	Over65           int     `json:"over65"`
	ISEE             float64 `json:"isee"`
	RedditoAnnuo     float64 `json:"reddito_annuo"`
	Affittuario      bool    `json:"affittuario"`
	PrimaAbitazione  bool    `json:"prima_abitazione"`
	RistrutturazCasa bool    `json:"ristrutturaz_casa"`
	Studente         bool    `json:"studente"`
	NuovoNato2025    bool    `json:"nuovo_nato_2025"`
	ISEESimulato     float64 `json:"isee_simulato,omitempty"`
}

type FAQ struct {
	Domanda  string `json:"domanda"`
	Risposta string `json:"risposta"`
}

type BonusTrad struct {
	Descrizione     string   `json:"descrizione"`
	Requisiti       []string `json:"requisiti,omitempty"`
	ComeRichiederlo []string `json:"come_richiederlo,omitempty"`
	FAQ             []FAQ    `json:"faq,omitempty"`
}

type Bonus struct {
	ID                   string               `json:"id"`
	Nome                 string               `json:"nome"`
	Categoria            string               `json:"categoria"`
	Descrizione          string               `json:"descrizione"`
	Importo              string               `json:"importo"`
	ImportoReale         string               `json:"importo_reale,omitempty"`
	Scadenza             string               `json:"scadenza"`
	Requisiti            []string             `json:"requisiti"`
	ComeRichiederlo      []string             `json:"come_richiederlo"`
	Documenti            []string             `json:"documenti,omitempty"`
	FAQ                  []FAQ                `json:"faq,omitempty"`
	LinkUfficiale        string               `json:"link_ufficiale"`
	Ente                 string               `json:"ente"`
	Compatibilita        int                  `json:"compatibilita"`
	UltimoAggiornamento  string               `json:"ultimo_aggiornamento,omitempty"`
	Fonte                string               `json:"fonte,omitempty"`
	Stato                string               `json:"stato,omitempty"`
	FonteURL             string               `json:"fonte_url,omitempty"`
	FonteNome            string               `json:"fonte_nome,omitempty"`
	RiferimentiNormativi []string             `json:"riferimenti_normativi,omitempty"`
	RegioniApplicabili   []string             `json:"regioni,omitempty"`
	SogliaISEE           float64              `json:"soglia_isee,omitempty"`
	Traduzioni           map[string]BonusTrad `json:"traduzioni,omitempty"`
}

type MatchResult struct {
	BonusTrovati     int     `json:"bonus_trovati"`
	RisparmioStimato string  `json:"risparmio_stimato"`
	PersoFinora      string  `json:"perso_finora,omitempty"`
	Bonus            []Bonus `json:"bonus"`
}

type SimulateResult struct {
	Reale          MatchResult `json:"reale"`
	Simulato       MatchResult `json:"simulato"`
	BonusExtra     int         `json:"bonus_extra"`
	RisparmioExtra string      `json:"risparmio_extra"`
}
