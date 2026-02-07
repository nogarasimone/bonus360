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
}

type Bonus struct {
	ID              string   `json:"id"`
	Nome            string   `json:"nome"`
	Categoria       string   `json:"categoria"`
	Descrizione     string   `json:"descrizione"`
	Importo         string   `json:"importo"`
	Scadenza        string   `json:"scadenza"`
	Requisiti       []string `json:"requisiti"`
	ComeRichiederlo []string `json:"come_richiederlo"`
	LinkUfficiale   string   `json:"link_ufficiale"`
	Ente            string   `json:"ente"`
	Compatibilita   int      `json:"compatibilita"`
}

type MatchResult struct {
	BonusTrovati     int     `json:"bonus_trovati"`
	RisparmioStimato string  `json:"risparmio_stimato"`
	Bonus            []Bonus `json:"bonus"`
}
