package scraper

// Source defines a web source to scrape for bonus information.
type Source struct {
	URL      string
	Name     string
	Type     string // "inps", "ade", "mef", "editorial", "caf"
	Priority int    // 1=primary, 2=secondary, 3=backup
	Parser   string // "inps", "ade", "editorial", "generic"
}

// GetSources returns the list of all sources to scrape.
func GetSources() []Source {
	return []Source{
		{URL: "https://www.inps.it/it/it/sostegni-sussidi-indennita/per-genitori.html", Name: "INPS Genitori", Type: "inps", Priority: 1, Parser: "inps"},
		{URL: "https://www.inps.it/it/it/sostegni-sussidi-indennita/per-famiglie.html", Name: "INPS Famiglie", Type: "inps", Priority: 1, Parser: "inps"},
		{URL: "https://www.agenziaentrate.gov.it/portale/web/guest/aree-tematiche/casa/agevolazioni", Name: "AdE Casa", Type: "ade", Priority: 1, Parser: "ade"},
		{URL: "https://www.agenziaentrate.gov.it/portale/web/guest/agevolazioni", Name: "AdE Agevolazioni", Type: "ade", Priority: 1, Parser: "ade"},
		{URL: "https://www.mef.gov.it", Name: "MEF", Type: "mef", Priority: 1, Parser: "generic"},
		{URL: "https://www.ticonsiglio.com/bonus-2025/", Name: "Ti Consiglio", Type: "editorial", Priority: 2, Parser: "editorial"},
		{URL: "https://www.fiscoetasse.com/new-rassegna-stampa/1542-legge-di-bilancio-2025-le-misure-per-le-famiglie.html", Name: "Fisco e Tasse", Type: "editorial", Priority: 2, Parser: "editorial"},
	}
}
