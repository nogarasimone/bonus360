package handlers

import (
	"net/http"
	"strings"
)

// PerCAFHandler serves the /per-caf landing page.
func PerCAFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>BonusPerMe per i CAF — Centro di Assistenza Fiscale</title>
<meta name="description" content="BonusPerMe per i Centri di Assistenza Fiscale. I tuoi clienti arrivano con il report gia pronto.">
<meta property="og:title" content="BonusPerMe per i CAF">
<meta property="og:description" content="I tuoi clienti arrivano con il report gia pronto.">
<link rel="canonical" href="/per-caf">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap" rel="stylesheet">
<style>
:root{--ink:#1B1B1F;--surface:#FFFFFF;--surface-alt:#F5F4F0;--primary:#1A3A5C;--accent:#C45A2C;--teal:#2B8A7E;--text-soft:#555E60;--text-muted:#8E9BA0;--border:rgba(0,0,0,0.08);--radius:8px}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'IBM Plex Sans',-apple-system,sans-serif;background:var(--surface);color:var(--ink);min-height:100vh;-webkit-font-smoothing:antialiased;font-size:15px;line-height:1.6}
.topbar{background:var(--primary);color:#fff;font-size:.72rem;font-weight:600;letter-spacing:.3px;text-align:center;padding:8px 16px}
header{background:rgba(255,255,255,.95);border-bottom:1px solid var(--border);height:56px;display:flex;align-items:center;justify-content:center}
.logo{font-family:'Instrument Serif',serif;font-size:1.4rem;color:var(--primary);display:flex;align-items:center;gap:8px;text-decoration:none}
.logo-mark{width:32px;height:32px;background:var(--primary);border-radius:var(--radius);display:flex;align-items:center;justify-content:center;color:#fff;font-family:'Instrument Serif',serif;font-size:1rem}
.container{max-width:800px;margin:0 auto;padding:0 24px}
.hero-caf{padding:64px 0 48px;text-align:center}
.hero-caf .badge{display:inline-block;padding:6px 16px;background:rgba(43,138,126,0.1);color:var(--teal);font-size:.78rem;font-weight:700;border-radius:var(--radius);margin-bottom:16px;text-transform:uppercase;letter-spacing:1px}
.hero-caf h1{font-family:'Instrument Serif',serif;font-size:clamp(1.6rem,4vw,2.4rem);font-weight:400;margin-bottom:16px}
.hero-caf p{color:var(--text-soft);max-width:560px;margin:0 auto 32px;font-size:1.05rem}
.steps{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:24px;margin:48px 0}
.step{background:var(--surface-alt);border:1px solid var(--border);border-radius:var(--radius);padding:28px 24px;text-align:center}
.step .num{width:36px;height:36px;background:var(--primary);color:#fff;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-weight:700;margin-bottom:12px}
.step h3{font-size:1rem;margin-bottom:6px}
.step p{font-size:.88rem;color:var(--text-soft)}
.advantages{margin:48px 0}
.advantages h2{font-family:'Instrument Serif',serif;font-size:1.5rem;text-align:center;margin-bottom:24px}
.adv-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px}
.adv{padding:20px;border:1px solid var(--border);border-radius:var(--radius)}
.adv h4{font-size:.92rem;margin-bottom:4px}
.adv p{font-size:.82rem;color:var(--text-soft)}
.coming-box{background:var(--surface-alt);border:1px solid var(--border);border-radius:var(--radius);padding:32px;text-align:center;margin:48px 0}
.coming-box h3{font-size:1.1rem;margin-bottom:8px}
.coming-box p{color:var(--text-soft);margin-bottom:20px}
.email-form{display:flex;gap:8px;max-width:400px;margin:0 auto}
.email-form input{flex:1;padding:10px 14px;border:1px solid var(--border);border-radius:var(--radius);font-family:inherit;font-size:.92rem}
.email-form button{padding:10px 20px;background:var(--accent);color:#fff;border:none;border-radius:var(--radius);font-family:inherit;font-weight:600;cursor:pointer}
.email-form button:hover{background:#A84A22}
.api-box{border:1px solid var(--border);border-radius:var(--radius);padding:24px;margin:32px 0}
.api-box h3{font-size:1rem;margin-bottom:8px}
.api-box code{display:block;background:var(--surface-alt);padding:12px;border-radius:var(--radius);font-size:.85rem;margin:8px 0;overflow-x:auto}
footer{border-top:1px solid var(--border);padding:24px 0;text-align:center;color:var(--text-muted);font-size:.82rem;margin-top:48px}
footer a{color:var(--primary);text-decoration:none}
#thanks{display:none;color:var(--teal);font-weight:600;margin-top:12px}
</style>
</head>
<body>
<div class="topbar">BonusPerMe per i Centri di Assistenza Fiscale</div>
<header><a href="/" class="logo"><div class="logo-mark">B</div> BonusPerMe</a></header>

<div class="container">
<section class="hero-caf">
<div class="badge">Per i CAF</div>
<h1>I tuoi clienti arrivano con il report pronto</h1>
<p>BonusPerMe aiuta le famiglie a scoprire i bonus a cui hanno diritto. Il risultato? Clienti che arrivano al CAF sapendo esattamente cosa chiedere, con tutta la documentazione.</p>
</section>

<section class="steps">
<div class="step"><div class="num">1</div><h3>Il cliente usa BonusPerMe</h3><p>Risponde a 4 domande — eta, famiglia, ISEE, situazione abitativa — e scopre i bonus compatibili.</p></div>
<div class="step"><div class="num">2</div><h3>Scarica il report PDF</h3><p>Un report professionale con elenco bonus, requisiti, documenti necessari e link ufficiali.</p></div>
<div class="step"><div class="num">3</div><h3>Viene al CAF preparato</h3><p>Il cliente arriva con documenti pronti. Meno tempo per la consulenza, piu pratiche evase.</p></div>
</section>

<section class="advantages">
<h2>Vantaggi per il tuo CAF</h2>
<div class="adv-grid">
<div class="adv"><h4>Clienti informati</h4><p>Arrivano sapendo cosa chiedere, con report e documenti.</p></div>
<div class="adv"><h4>Meno consulenza base</h4><p>Le domande generiche ("a cosa ho diritto?") vengono filtrate prima.</p></div>
<div class="adv"><h4>Piu pratiche/giorno</h4><p>Clienti preparati = consulenze piu rapide = piu appuntamenti.</p></div>
<div class="adv"><h4>Zero costi</h4><p>BonusPerMe e gratuito, open source e senza pubblicita.</p></div>
</div>
</section>

<section class="coming-box">
<h3>Prossimamente</h3>
<p>Widget embeddabile per il sito del tuo CAF e API per integrazione con i vostri gestionali.</p>
<p>Lascia la tua email per essere avvisato al lancio:</p>
<div class="email-form">
<input type="email" id="caf-email" placeholder="La tua email...">
<button onclick="submitCAF()">Avvisami</button>
</div>
<div id="thanks">Grazie! Ti avviseremo al lancio.</div>
</section>

<section class="api-box">
<h3>Open Data API (disponibile ora)</h3>
<p>Accedi ai dati di tutti i bonus italiani tramite API REST:</p>
<code>GET /api/bonus — Lista completa bonus (nazionali + regionali)</code>
<code>GET /api/bonus/{id} — Dettaglio singolo bonus</code>
<p style="font-size:.85rem;color:var(--text-soft);margin-top:8px">Formato JSON, CORS abilitato, cache 1 ora. Rate limit: 60 richieste/minuto.</p>
</section>
</div>

<footer>
<div class="container">
<p>BonusPerMe — Progetto gratuito e open source. Non siamo un CAF o un patronato.</p>
<p><a href="/">Torna a BonusPerMe</a></p>
</div>
</footer>

<script>
function submitCAF(){
var email=document.getElementById('caf-email').value.trim();
if(!email||!email.includes('@'))return;
fetch('/api/notify-signup',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email:email,tipo:'caf'})})
.then(function(){document.getElementById('thanks').style.display='block';document.querySelector('.email-form').style.display='none';});
}
</script>
</body>
</html>`)

	w.Write([]byte(sb.String()))
}
