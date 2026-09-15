# Kunta — ShiduLab

Kunta è un utensile portatile per interrogare, contare, osservare e giocare con i testi.

Il nome richiama in sardo il doppio gesto del **contare** e del **raccontare**: Kunta conta ciò che trova nel testo e lo rende osservabile.

## Windows

La cartella `Portable/` contiene `Kunta.exe` per Windows x64. Nessuna installazione.

Funzioni principali:

- riepilogo del testo: caratteri, lettere, parole, righe, spazi/bianchi, cifre, punteggiatura, parole uniche/ripetute;
- lettere indicate;
- occorrenze di parola/stringa;
- frequenza lettere e parole;
- parole più corte / più lunghe;
- palindromi;
- bifronti presenti nel testo;
- gruppi di anagrammi;
- acrostico e telestico delle righe;
- isovocaliche e isoconsonantiche;
- omovocaliche e omoconsonantiche;
- omovocaliche iniziali / finali;
- omoconsonantiche iniziali / finali;
- inversione del testo e dell'ordine delle parole;
- ordinamento alfabetico delle parole;
- estrazione dei numeri.

## Nuove analisi linguistiche

- **Isovocaliche**: parole con la stessa sequenza vocalica, nello stesso ordine e con la stessa molteplicità. Esempio: `CASA / RAMA / FATA` → `A-A`.
- **Isoconsonantiche**: stessa sequenza consonantica.
- **Omovocaliche**: stesso patrimonio vocalico, indipendentemente dall'ordine.
- **Omoconsonantiche**: stesso patrimonio consonantico, indipendentemente dall'ordine.
- **Iniziali / finali**: il parametro numerico indica quante vocali o consonanti iniziali/finali confrontare; default consigliato `2`.
- **Acrostico**: prende il primo carattere significativo di ogni riga non vuota.
- **Telestico**: prende l'ultimo carattere significativo di ogni riga non vuota.

## Uso

1. Incolla un testo oppure usa **APRI TESTO...**.
2. Scegli l'operazione.
3. Se serve, imposta `Parametro`.
4. Premi **KUNTA**.

Per palindromi, bifronti, anagrammi e gruppi lessicali il parametro indica la lunghezza minima della parola. Per le analisi `iniziali/finali` indica invece il numero di vocali/consonanti da confrontare.

La casella **Maiuscole/minuscole distinte** riguarda le operazioni che usano il confronto testuale diretto.

Il testo originale non viene modificato dalle analisi.

## Struttura repository

- `analysis/` — motore linguistico e test automatici;
- `desktop/` — applicativo nativo Windows in Go/Win32;
- `desktop/assets/` — icona Kunta e logo ShiduLab;
- `docs/` — versione web/PWA;
- `Portable/` — build Windows pronta;
- `.github/workflows/` — build/test automatizzati.

## Build

Vedi `desktop/BUILD.md`.

ShiduLab 2026
