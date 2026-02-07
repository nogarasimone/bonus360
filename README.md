# BonusPerMe

**Scopri in 2 minuti quali bonus e agevolazioni ti spettano — gratuito, anonimo, open source.**

![Build](https://img.shields.io/github/actions/workflow/status/bonusperme/bonusperme/ci.yml?branch=main&style=flat-square&label=build)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)
![Updated](https://img.shields.io/badge/bonus_aggiornati-Febbraio_2025-green?style=flat-square)
![Languages](https://img.shields.io/badge/lingue-IT_EN_FR_ES_RO-orange?style=flat-square)
![PRs](https://img.shields.io/badge/PRs-benvenute-brightgreen?style=flat-square)
![Privacy](https://img.shields.io/badge/tracking-zero-red?style=flat-square)

---

> **[Prova BonusPerMe →](https://bonusperme.it)** (nessuna registrazione richiesta)

<!-- ![Screenshot](docs/screenshot.png) -->
<!-- Aggiungi screenshot quando il sito è live -->

---

## Perché BonusPerMe

Ogni anno in Italia miliardi di euro in bonus e agevolazioni restano non richiesti. Le informazioni sono frammentate tra INPS, Agenzia delle Entrate, MEF, Regioni e Comuni — e capire a cosa si ha diritto richiede ore di ricerca. BonusPerMe centralizza tutto in un'unica verifica anonima di 2 minuti: rispondi a poche domande sulla tua situazione e ricevi la lista completa dei bonus compatibili, con importi, requisiti e istruzioni. Nessun dato viene salvato — mai.

## Funzionalità principali

- Analisi di 20+ bonus e agevolazioni attive nel 2025
- Calcolo personalizzato dell'importo basato su ISEE
- Report stampabile per CAF e commercialisti
- Report PDF scaricabile con fonti e riferimenti normativi
- Calendario scadenze .ics importabile in Google Calendar / Apple Calendar
- Checklist documenti spuntabile per ogni bonus
- Simulatore "cosa cambia con un ISEE diverso"
- FAQ specifiche per ogni bonus
- Mappa CAF più vicino
- Scraper automatico che aggiorna i dati da fonti ufficiali ogni 24h
- 5 lingue: Italiano, English, Français, Español, Română
- Accessibilità WCAG AA
- Zero cookie, zero tracking, zero database

## Come funziona

1. **Inserisci i tuoi dati** — età, famiglia, ISEE e situazione abitativa
2. **Analisi istantanea** — il sistema confronta con tutti i bonus attivi
3. **Risultati personalizzati** — ricevi la lista con importi reali calcolati per te
4. **Porta tutto al CAF** — stampa il report o scarica il PDF con fonti e riferimenti

## Stack tecnico

| Componente | Tecnologia |
|------------|-----------|
| Backend | Go 1.21+ (standard library, `net/http`) |
| Frontend | Vanilla HTML/CSS/JS (singolo file, zero framework) |
| PDF | [gofpdf](https://github.com/jung-kurt/gofpdf) |
| Scraping | [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html) |
| Parsing ISEE | [ledongthuc/pdf](https://github.com/ledongthuc/pdf) |
| Database | Nessuno (tutto in memoria) |
| JS dependencies | Zero |

## Quick Start

```bash
# Clona il repository
git clone https://github.com/bonusperme/bonusperme.git
cd bonusperme

# Build e avvia
go build -o bonusperme .
./bonusperme

# Oppure con go run
go run main.go

# Il server parte su http://localhost:8080
```

**Requisiti:**
- Go 1.21 o superiore
- Nessun'altra dipendenza di sistema

## Struttura del progetto

```
bonusperme/
├── main.go                       # Entry point, server HTTP, routing
├── internal/
│   ├── handlers/
│   │   ├── handlers.go           # Handler API principali (match, stats, ISEE)
│   │   ├── extra.go              # Calendar, simulate, report PDF, notify
│   │   ├── infra.go              # SEO pages, sitemap, analytics, health
│   │   ├── counter.go            # Contatore persistente con debounce
│   │   └── ratelimit.go          # Rate limiter per IP (token bucket)
│   ├── matcher/
│   │   └── matcher.go            # Engine di matching + database 20 bonus
│   ├── models/
│   │   └── models.go             # Struct dati (UserProfile, Bonus, FAQ)
│   ├── scraper/
│   │   ├── sources.go            # Lista fonti istituzionali (7 sorgenti)
│   │   ├── parsers.go            # Parser per INPS, AdE, editoriali, generico
│   │   ├── enricher.go           # Deduplicazione e merge con dati hardcoded
│   │   └── scheduler.go          # Scheduler 24h + cache thread-safe
│   ├── i18n/
│   │   └── translations.go       # Traduzioni 5 lingue (~100 chiavi ciascuna)
│   └── telegram/
│       └── bot.go                # Bot Telegram (coming soon)
├── static/
│   └── index.html                # Frontend completo (single file)
├── counter.json                  # Contatore persistente verifiche
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .github/
│   └── workflows/
│       └── ci.yml                # CI/CD GitHub Actions
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
└── README.md
```

## API

| Metodo | Path | Descrizione |
|--------|------|-------------|
| POST | `/api/match` | Calcola bonus compatibili |
| POST | `/api/simulate` | Simula con ISEE diverso |
| POST | `/api/parse-isee` | Estrai ISEE da PDF |
| POST | `/api/report` | Genera report PDF |
| GET | `/api/calendar?bonuses=...` | Scarica calendario .ics |
| GET | `/api/translations?lang=XX` | Dizionario traduzioni |
| GET | `/api/stats` | Contatore verifiche |
| GET | `/api/health` | Stato del server e scraper |
| GET | `/api/scraper-status` | Dettaglio fonti scraper |
| GET | `/bonus/{id}` | Pagina SEO singolo bonus |
| GET | `/sitemap.xml` | Sitemap per motori di ricerca |
| POST | `/api/notify-signup` | Iscrizione lista d'attesa notifiche |
| POST | `/api/analytics` | Evento analytics (anonimo) |

## Fonti dati

BonusPerMe raccoglie informazioni da fonti istituzionali e giornalistiche verificate:

- **INPS** — [inps.it](https://www.inps.it) (bonus famiglia, assegno unico, nido, ADI)
- **Agenzia delle Entrate** — [agenziaentrate.gov.it](https://www.agenziaentrate.gov.it) (detrazioni casa, bonus edilizi)
- **Ministero dell'Economia** — [mef.gov.it](https://www.mef.gov.it) (carta dedicata a te, misure fiscali)
- **Ti Consiglio un Lavoro** — [ticonsiglio.com](https://www.ticonsiglio.com) (guide aggiornate)
- **Fisco e Tasse** — [fiscoetasse.com](https://www.fiscoetasse.com) (approfondimenti fiscali)

I dati vengono aggiornati automaticamente ogni 24 ore. Se una fonte non è raggiungibile, il sistema usa i dati verificati più recenti.

## Privacy

BonusPerMe non raccoglie, salva o trasmette **nessun dato personale**. Non esistono database, cookie, sistemi di tracking o profilazione. I dati inseriti dall'utente esistono solo nella sessione corrente e vengono cancellati al refresh della pagina. Il codice è open source e verificabile da chiunque. I server sono in Unione Europea. Il progetto è conforme al GDPR.

## Deploy

### Docker

```bash
docker build -t bonusperme .
docker run -p 8080:8080 bonusperme
```

### Docker Compose

```bash
docker-compose up -d
```

### Manuale

```bash
go build -o bonusperme .
./bonusperme

# Variabili d'ambiente opzionali:
# PORT=8080 (default)
```

## Contribuire

Le contribuzioni sono benvenute! Leggi [CONTRIBUTING.md](CONTRIBUTING.md) per le linee guida.

**Aree dove serve aiuto:**

- Traduzione bonus in nuove lingue
- Aggiunta bonus regionali e comunali
- Miglioramento parser scraper per nuove fonti
- Test di accessibilità
- Segnalazione bonus mancanti o importi errati

## Roadmap

| Feature | Stato |
|---------|-------|
| Scraper fonti istituzionali | Attivo |
| 5 lingue (IT EN FR ES RO) | Attivo |
| Report PDF | Attivo |
| Calendario scadenze | Attivo |
| Pagine SEO per bonus | Attivo |
| Bot Telegram | In sviluppo |
| Notifiche nuovi bonus | In sviluppo |
| Bonus regionali | Pianificato |
| App mobile (PWA) | Pianificato |
| API pubblica documentata | Pianificato |
| Widget embeddabile per CAF | Pianificato |

## Licenza

Questo progetto è rilasciato sotto licenza [AGPL-3.0](LICENSE). Puoi usarlo, modificarlo e ridistribuirlo liberamente, a patto che il codice derivato rimanga open source.

**Perché AGPL e non MIT:** abbiamo scelto AGPL-3.0 per garantire che qualsiasi versione modificata di BonusPerMe resti accessibile come software libero. Se usi BonusPerMe come servizio web, devi rendere disponibile il codice sorgente.

---

## For International Contributors

BonusPerMe is a free, open-source tool that helps families in Italy discover government benefits they're entitled to. The interface supports 5 languages (Italian, English, French, Spanish, Romanian). We welcome contributions from developers worldwide — see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. The codebase is in Go with a single-file vanilla HTML/CSS/JS frontend.

---

Fatto con cura per le famiglie italiane.

[Segnala un problema](https://github.com/bonusperme/bonusperme/issues/new?template=bug_report.md) · [Richiedi una feature](https://github.com/bonusperme/bonusperme/issues/new?template=feature_request.md) · [Segnala un bonus errato](https://github.com/bonusperme/bonusperme/issues/new?template=bonus_errato.md)
