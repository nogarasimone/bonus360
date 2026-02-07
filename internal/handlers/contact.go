package handlers

import (
	"bonusperme/internal/config"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// ContattiHandler serves the GET /contatti page.
func ContattiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	siteKey := config.Cfg.TurnstileSiteKey
	turnstileScript := ""
	turnstileWidget := ""
	if siteKey != "" {
		turnstileScript = `<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>`
		turnstileWidget = `<div class="cf-turnstile" data-sitekey="` + siteKey + `" data-callback="onContactTurnstile" data-theme="light" style="margin-bottom:16px"></div>`
	}

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="it">
<head>
` + SharedMetaTags("Contatti — BonusPerMe", "Contattaci per informazioni, segnalazioni o partnership. BonusPerMe — servizio gratuito per le famiglie italiane.", "/contatti") + `
` + turnstileScript + `
<style>` + SharedCSS() + `
.page-hero{padding:48px 0 32px;text-align:center}
.page-hero h1{font-size:clamp(1.5rem,3.5vw,2rem);margin-bottom:10px}
.page-hero p{color:var(--ink-75);font-size:1rem}
.contact-grid{display:grid;grid-template-columns:1fr 1fr;gap:32px;margin:32px 0 48px}
.contact-form-wrap{background:#fff;border:1px solid var(--ink-15);border-radius:var(--radius-lg);padding:28px 24px;box-shadow:var(--shadow-card)}
.contact-info{display:flex;flex-direction:column;gap:16px}
.info-card{background:#fff;border:1px solid var(--ink-15);border-radius:var(--radius-lg);padding:20px;box-shadow:var(--shadow-card)}
.info-card h3{font-size:1rem;margin-bottom:8px}
.info-card p{font-size:.88rem;color:var(--ink-75)}
.info-card a{word-break:break-all}
.field{margin-bottom:14px}
.field label{display:block;font-size:.82rem;font-weight:600;margin-bottom:4px;color:var(--ink-75)}
.field input,.field select,.field textarea{width:100%;padding:10px 14px;border:1px solid var(--ink-15);border-radius:var(--radius);font-family:inherit;font-size:.92rem;background:#fff}
.field textarea{min-height:120px;resize:vertical}
.field input:focus,.field select:focus,.field textarea:focus{outline:none;border-color:var(--blue-mid)}
.privacy-check{display:flex;align-items:flex-start;gap:8px;margin:16px 0;font-size:.85rem;color:var(--ink-75)}
.privacy-check input{margin-top:3px}
.btn-contact{display:block;width:100%;padding:12px 20px;background:var(--terra);color:#fff;border:none;border-radius:var(--radius);font-family:inherit;font-size:.95rem;font-weight:600;cursor:pointer}
.btn-contact:hover{background:var(--terra-dark)}
#contactResult{display:none;margin-top:16px;padding:12px 16px;border-radius:var(--radius);font-size:.9rem;font-weight:600;text-align:center}
@media(max-width:640px){.contact-grid{grid-template-columns:1fr}}
</style>
</head>
<body>
` + SharedTopbar() + `
` + SharedHeader("/contatti") + `

<div class="container">
<section class="page-hero">
<h1>Contattaci</h1>
<p>Hai domande, suggerimenti o vuoi segnalare un errore? Scrivici.</p>
</section>

<div class="contact-grid">
<div class="contact-form-wrap">
<div class="field"><label>Nome *</label><input type="text" id="ct-nome" placeholder="Il tuo nome"></div>
<div class="field"><label>Email *</label><input type="email" id="ct-email" placeholder="La tua email"></div>
<div class="field">
<label>Oggetto</label>
<select id="ct-oggetto">
<option value="info">Informazioni generali</option>
<option value="bug">Segnalazione errore</option>
<option value="partner">Partnership / CAF</option>
<option value="altro">Altro</option>
</select>
</div>
<div class="field"><label>Messaggio *</label><textarea id="ct-messaggio" placeholder="Scrivi il tuo messaggio..."></textarea></div>
<label class="privacy-check"><input type="checkbox" id="ct-privacy"> Ho letto e accetto la <a href="/privacy" target="_blank">Privacy Policy</a></label>
` + turnstileWidget + `
<button class="btn-contact" onclick="submitContact()">Invia messaggio</button>
<div id="contactResult"></div>
</div>

<div class="contact-info">
<div class="info-card">
<h3>Email</h3>
<p><a href="mailto:info@bonusperme.it">info@bonusperme.it</a></p>
</div>
<div class="info-card">
<h3>Sede</h3>
<p>Via Morazzone 4<br>22100 Como (CO)</p>
</div>
<div class="info-card">
<h3>Titolare</h3>
<p>Simone Nogara<br>P.IVA 03817020138</p>
</div>
<div class="info-card">
<h3>Rispondiamo entro</h3>
<p>24-48 ore lavorative</p>
</div>
</div>
</div>
</div>

` + SharedFooter() + `
` + SharedCookieBanner() + `

<script>
var contactTurnstileToken='';
function onContactTurnstile(t){contactTurnstileToken=t;}

function submitContact(){
  var nome=document.getElementById('ct-nome').value.trim();
  var email=document.getElementById('ct-email').value.trim();
  var oggetto=document.getElementById('ct-oggetto').value;
  var messaggio=document.getElementById('ct-messaggio').value.trim();
  var privacy=document.getElementById('ct-privacy').checked;
  var result=document.getElementById('contactResult');

  if(!nome||!email||!messaggio){
    result.style.display='block';result.style.background='var(--terra-light)';result.style.color='var(--terra)';
    result.textContent='Compila tutti i campi obbligatori (*).';return;
  }
  if(!email.includes('@')){
    result.style.display='block';result.style.background='var(--terra-light)';result.style.color='var(--terra)';
    result.textContent='Inserisci un indirizzo email valido.';return;
  }
  if(!privacy){
    result.style.display='block';result.style.background='var(--terra-light)';result.style.color='var(--terra)';
    result.textContent='Devi accettare la Privacy Policy.';return;
  }

  var headers={'Content-Type':'application/json'};
  if(contactTurnstileToken)headers['X-Turnstile-Token']=contactTurnstileToken;

  fetch('/api/contact',{method:'POST',headers:headers,body:JSON.stringify({nome:nome,email:email,oggetto:oggetto,messaggio:messaggio})})
  .then(function(r){return r.json();})
  .then(function(data){
    result.style.display='block';
    if(data.ok){
      result.style.background='var(--green-light)';result.style.color='var(--green)';
      result.textContent='Grazie! Il tuo messaggio è stato inviato.';
      document.getElementById('ct-nome').value='';
      document.getElementById('ct-email').value='';
      document.getElementById('ct-messaggio').value='';
      document.getElementById('ct-privacy').checked=false;
    } else {
      result.style.background='var(--terra-light)';result.style.color='var(--terra)';
      result.textContent=data.error||'Errore. Riprova.';
    }
  })
  .catch(function(){
    result.style.display='block';result.style.background='var(--terra-light)';result.style.color='var(--terra)';
    result.textContent='Errore di rete. Riprova.';
  });
}
</script>
` + SharedScripts() + `
</body>
</html>`)

	w.Write([]byte(sb.String()))
}

type contactRequest struct {
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Oggetto  string `json:"oggetto"`
	Messaggio string `json:"messaggio"`
}

// ContactHandler handles POST /api/contact
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !verifyTurnstile(getTurnstileToken(r)) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Verifica di sicurezza non superata"})
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Dati non validi"})
		return
	}
	defer r.Body.Close()

	req.Nome = strings.TrimSpace(req.Nome)
	req.Email = strings.TrimSpace(req.Email)
	req.Messaggio = strings.TrimSpace(req.Messaggio)

	if req.Nome == "" || req.Email == "" || req.Messaggio == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Compila tutti i campi obbligatori"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Email non valida"})
		return
	}

	log.Printf("[contact] nome=%s email=%s oggetto=%s", req.Nome, req.Email, req.Oggetto)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}
