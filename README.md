# Kunta

Repository completo e pulito per tre destinazioni:

- `desktop/` — sorgente Windows EXE in Go, con interfaccia web incorporata.
- `pwa/` — Progressive Web App autonoma.
- `android/` — progetto Android nativo Java + WebView con gli asset Kunta incorporati.
- `.github/workflows/forge-kunta.yml` — build automatica di EXE, PWA ZIP e APK.
- `.github/workflows/pages.yml` — pubblicazione della PWA su GitHub Pages.
- `Kunta.exe` e `Kunta-PWA.zip` — build correnti già pronte anche nella root.
- `release/` — copie delle build correnti.

Il motore Kunta contiene 56 operazioni e la stessa `app.js` viene usata per Desktop, PWA e Android.

## Pubblicazione senza svuotare prima il repository

Estrarre questo pacchetto sul PC e fare doppio clic su `PUBBLICA_KUNTA.cmd`.

Lo script prepara un commit locale e sostituisce `main` con un unico push atomico. Il repository remoto rimane quello vecchio fino all'istante in cui il nuovo commit viene accettato.

Commit usato:

`Kunta: ricostruzione completa pulita Desktop PWA Android`
