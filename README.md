# Kunta

**Piccolo laboratorio testuale portatile per Windows e mobile**  
**ShiduLab 2026**

Kunta serve per **contare, osservare, ordinare, smontare e curiosare dentro i testi**.
È un utensile da scrivania: nessuna installazione, nessun account, nessun servizio esterno. Si apre `Kunta.exe`, si incolla o si carica un testo e si sceglie cosa Kuntare.

## Perché “Kunta”?

In sardo **Kuntai** attraversa due azioni: **contare** e **rac-contare**.

> «Ma ixisi itt'est suzzediu?»  
> «Nou, kunta is noas.»  
> — Ma lo sai cos'è successo?  
> — No, dimmi / raccontami le novità.

> «Kunta kantu funti.»  
> — Conta quanti/e sono.

Kunta fa entrambe le cose: **conta il testo e, attraverso ciò che trova, lo racconta.**

## Un piccolo Giostoliere del testo

Kunta è anche una specie di **Giostoliere**: un pallottoliere, un **letteriere**.
Permette di operare e curiosare dentro la scrittura quasi chirurgicamente: **lettera per lettera, parola per parola, ricorrenza per ricorrenza**.

La scrittura non appartiene a un solo dominio. Dentro una pagina convivono anche:

- **matematica** — quantità, frequenze, proporzioni, ricorrenze;
- **geometria** — forme, simmetrie, disposizioni, ritorni;
- **musica** — ritmo, ripetizione, variazione, cadenza;
- **silenzio** — spazi, pause, assenze, ciò che non viene scritto.

Kunta non pretende di interpretare al posto di chi scrive: **mette gli elementi in vista**.

## Uso

Puoi:

- incollare direttamente il testo nell'area superiore (`Ctrl+V` o **Incolla**);
- usare **Apri testo...** per caricare file testuali;
- trascinare un file di testo sulla finestra;
- scegliere una funzione dal menu;
- inserire un eventuale parametro;
- premere **KUNTA!**;
- copiare il risultato senza modificare il testo originale.

## Funzioni

### Conteggi e frequenze

- lettere indicate, ad esempio `aeiou`;
- caratteri Unicode e byte UTF-8;
- parole e parole distinte;
- righe e righe non vuote;
- spazi e caratteri bianchi;
- cifre;
- punteggiatura;
- occorrenze di una parola o stringa;
- parole uniche;
- parole ripetute;
- frequenza delle lettere;
- frequenza delle parole;
- parola più corta / più lunga.

### Osservazione editoriale

- **Ordina parole alfabeticamente** — rende immediatamente visibili ripetizioni, ricorrenze, famiglie e tic lessicali;
- inverti il testo;
- inverti l'ordine delle parole;
- estrai solo numeri.

### Enigmistica e strutture del testo

- palindromi;
- bifronti presenti nel testo;
- coppie e gruppi di anagrammi;
- acrostico delle righe;
- telestico delle righe;
- sciarade interne;
- candidati a salti di dominio;
- LetterTransport;
- inclusioni progressive;
- schema delle rime / desinenze.

## Schema rime / desinenze

Nel lessico operativo di Kunta, **rima = desinenza grafica**.
Kunta prende l'ultima parola di ogni verso e risale dalla coda, una lettera alla volta, mostrando dove le famiglie coincidono e dove si separano.

Esempio:

```text
condottiero → -ero → x
nero        → -ero → x
vero        → -ero → x
duro        → -uro → z

Schema: x-x-x-z
```

La profondità della desinenza è regolabile tramite il parametro.

## Rima baciata / inclusione progressiva

Nel lessico sperimentale di Kunta, una **inclusione progressiva** rileva catene in cui una parola rimane intera nella coda della successiva:

```text
oro → onoro → sonoro
```

Kunta mostra il nucleo trasportato e la profondità della catena.

## LetterTransport

**LetterTransport** cerca materia alfabetica che viaggia tra parole consecutive: nuclei di lettere conservati, trasportati o rilanciati da una parola alla successiva.

È un modo per chiedere a Kunta:

> **Mostrami cosa viaggia dentro il testo.**

## Sciarade e salti tra domini

Kunta può cercare segmentazioni interne costruite con parole realmente presenti nel testo.

Esempio:

```text
contesto
con + testo
con + te + sto
```

Se la stessa forma intera ricorre insieme a segmentazioni differenti, Kunta la segnala come **possibile salto di dominio da verificare nel contesto**. La macchina rileva la struttura; il senso resta a chi legge e scrive.

## Kunta il testo

La funzione **Kunta il testo (fenomeni)** esegue una ricognizione combinata e riunisce in un solo risultato:

- ripetizioni;
- schema rime / desinenze;
- inclusioni progressive;
- LetterTransport;
- sciarade e possibili salti di dominio.

## Parametro

Il campo **Parametro** cambia funzione a seconda dell'operazione scelta. Per esempio:

- **Lettere indicate** → `aeiou`;
- **Occorrenze parola/stringa** → parola o frase da cercare;
- **Palindromi / Bifronti / Anagrammi** → lunghezza minima;
- **Schema rime / desinenze** → profondità della coda;
- **Rima baciata / LetterTransport** → nucleo minimo;
- **Sciarade** → lunghezza minima dei segmenti.

La casella **Maiuscole/minuscole distinte** decide se, per l'analisi scelta, `A` e `a` devono essere considerate differenti.

## Portabilità e privacy

Kunta è un'applicazione Windows portatile: l'eseguibile funziona senza installazione. Le analisi vengono eseguite localmente sul testo caricato nell'applicazione.

L'icona del programma è l'avatar di Kunta; **Botolo + ShiduLab** restano la firma grafica nell'interfaccia.

## Contatti

**ShiduLab**  
Email: `ShiduLab@gmail.com`

---

**Kunta — ShiduLab 2026**

## Versioni

### Kunta Desktop — Windows

La versione Windows è portatile e non richiede installazione.
Il codice sorgente si trova in `desktop/`; l'eseguibile viene distribuito tramite le **Releases** del repository.

### Kunta Mobile — PWA

La versione mobile si trova in `docs/` ed è una Progressive Web App (PWA): può essere aperta dal browser, aggiunta alla schermata Home e, dopo il primo caricamento, usata offline quando il browser lo consente.

Quando GitHub Pages è attivo per questo repository, Kunta Mobile è raggiungibile da:

`https://shidulab.github.io/Kunta/`

Su sistemi/browser compatibili, la PWA è predisposta anche per ricevere testo tramite il menu **Condividi**.

## Struttura del repository

```text
Kunta/
├─ README.md
├─ .gitignore
├─ desktop/
│  ├─ main.go
│  ├─ embed_icon.py
│  ├─ BUILD.md
│  └─ assets/
└─ docs/
   ├─ index.html
   ├─ app.js
   ├─ styles.css
   ├─ manifest.webmanifest
   ├─ sw.js
   ├─ share.html
   └─ assets/
```
