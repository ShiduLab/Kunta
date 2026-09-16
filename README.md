# Kunta

**ShiduLab · 2026** EXE · APK · PWA-> https://shidulab.github.io/Kunta/

Kunta è un utensile portatile per interrogare, contare, osservare e giocare con i testi.
Non genera frasi al posto di chi scrive: mette in evidenza strutture, ricorrenze, trasformazioni e peculiarità linguistiche già presenti nelle parole e nei testi.

## Perché “Kunta”?

In sardo **Kunta** richiama insieme due azioni: **contare** e **raccontare**.

> “Ma ixisi itt'est suzzediu?”  
> “Nou, kunta is noas.”  
> — Ma lo sai cos'è successo?  
> — No, dimmi / raccontami le novità.

> “Kunta kantu funti.”  
> — Conta quanti/e sono.

Kunta fa entrambe le cose: **conta il testo e, attraverso ciò che trova, lo rac-conta**.

## Un piccolo Giostoliere del testo

Kunta è anche una specie di **Giostoliere**: un pallottoliere, un letteriere.
Permette di curiosare dentro un testo a livello quasi chirurgico, lettera per lettera, parola per parola, ricorrenza per ricorrenza.

Dentro una pagina convivono anche:

- **matematica** — quantità, frequenze, proporzioni, ricorrenze;
- **geometria** — forme, simmetrie, disposizioni, ritorni;
- **musica** — ritmo, ripetizione, variazione, cadenza;
- **silenzio** — spazi, pause, assenze, ciò che non viene scritto.

Kunta non pretende di interpretare tutto questo al posto di chi scrive. Mette gli elementi in vista, li conta, li ordina e permette di osservarli da altre angolazioni.

---

# Stato del progetto

Il repository contiene tre destinazioni:

- `desktop/` — **Kunta Desktop nativo Windows**, scritto in Go e compilato come singolo `Kunta.exe`;
- `pwa/` — Progressive Web App autonoma;
- `android/` — applicazione Android;
- `.github/workflows/forge-kunta.yml` — build automatica degli artefatti Windows, PWA e Android.

Il **Desktop Windows** è attualmente il ramo più avanzato del motore e contiene le funzionalità descritte qui sotto.

## Desktop Windows

- applicazione nativa Win32;
- **nessun browser** e nessun server `localhost`;
- singolo eseguibile `Kunta.exe`;
- elaborazione completamente **locale**;
- nessun invio del testo verso servizi esterni;
- interfaccia DPI-aware e apertura massimizzata sul monitor corrente;
- drag & drop dei file di testo;
- apertura diretta dei file tramite percorso passato all'eseguibile;
- integrazione Esplora file: **tasto destro → “Apri in Kunta”** per i file testuali;
- copia del risultato negli appunti;
- logo **Botolo + ShiduLab** con collegamento al repository.

---

# Ingresso del testo

Puoi inserire il testo in più modi:

- incollando direttamente nell'area principale (`Ctrl+V` o pulsante **Incolla**);
- usando **Apri testo…**;
- trascinando un file sulla finestra;
- facendo **tasto destro su un file di testo → Apri in Kunta**.

Sono adatti TXT, MD, LOG, CSV, TSV e in generale file testuali leggibili.

Il testo originale nell'area superiore non viene modificato dalle analisi.

---

# 58 operazioni

## Conteggio e frequenze

1. Lettere indicate
2. Caratteri
3. Parole
4. Righe
5. Spazi
6. Cifre
7. Punteggiatura
8. Occorrenze parola/stringa
9. Parole uniche
10. Parole ripetute
11. Frequenza lettere
12. Frequenza parole
13. Parola più corta / più lunga
14. Lunghezza media parole
15. Distribuzione lunghezze

### Caratteri

Oltre al totale, Kunta mostra la frequenza di ogni carattere:

```text
Caratteri: 345

21: A
15: C
12: [spazio]
...
```

### Parole

Oltre al totale e al numero di parole distinte, Kunta mostra la frequenza lessicale:

```text
Parole: 128
Parole distinte: 47

12: che
9: di
7: il
...
```

## Enigmistica e trasformazioni

16. Palindromo puro
17. Bifronti
18. Anagrammi
19. Palindromo inverso
20. Palindromo contrario
21. Inversi
22. Antipodi
23. Sciarade / salti di dominio
24. LetterTransport
25. Isogrammi
26. Doppie / triple
27. Sequenze ripetute

Le quattro voci seguenti restano **distinte**, perché descrivono fenomeni differenti:

- **Palindromo inverso**
- **Palindromo contrario**
- **Inversi**
- **Antipodi**

Kunta non genera pseudo-parole per riempire i risultati: per le trasformazioni che richiedono una controparte, vengono mostrate solo corrispondenze valide secondo la specifica funzione.

## Anagrammi con dizionario italiano

La funzione **Anagrammi** non si limita più alle parole già presenti nel testo.

Il Desktop incorpora un **indice anagrammatico compresso** generato da un dizionario italiano di oltre **4,2 milioni di voci**. L'indice è incluso nell'eseguibile tramite `go:embed`: a runtime non servono file esterni né connessione Internet.

Esempio:

```text
caos →
    caso
    cosa
```

Per ogni parola del testo Kunta cerca nel dizionario tutte le parole reali formabili con **esattamente le stesse lettere e le stesse molteplicità**.

Regole applicate:

- la parola di partenza non viene proposta come anagramma di sé stessa;
- varianti che differiscono solo per accento grafico non vengono considerate anagrammi differenti (`Empatia` / `empatìa`);
- articoli ed elisioni iniziali vengono esclusi dalla parola analizzata:
  - `L'Empatia` → `Empatia`
  - `nell'amare` → `amare`
  - `un'altra` → `altra`
- il parametro indica la **lunghezza minima**, default `3`.

L'indice è nel repository come:

```text
desktop/assets/anagram_index.tsv.gz
```

## Strutture vocaliche e consonantiche

28. Isovocaliche
29. Isoconsonantiche
30. Omovocaliche
31. Omoconsonantiche
32. Omovocaliche iniziali
33. Omovocaliche finali
34. Omoconsonantiche iniziali
35. Omoconsonantiche finali

Per **Omovocaliche** e **Omoconsonantiche** la sequenza mostrata conserva l'ordine reale delle lettere nella parola.

Esempio:

```text
Energia → E-E-I-A
```

non una sequenza riordinata alfabeticamente.

## Strutture posizionali e foniche

36. Acrostico
37. Telestico
38. Schema rime / desinenze
39. Rima baciata / inclusione progressiva
40. Allitterazioni
41. Assonanze
42. Consonanze
43. Ossimori — candidati

## Filtri e ricerca nelle parole

44. Parole con iniziale
45. Parole con finale
46. Parole contenenti sequenza
47. Parole di lunghezza N
48. Parole alfabetiche
49. Parole alfabetiche inverse

### Parole alfabetiche

Qui “alfabetiche” significa **sequenza consecutiva dell'alfabeto**, non semplice ordine crescente.

Valide:

```text
AB
ABC
BCD
XYZ
```

Non valide:

```text
AL
CI
DEL
```

Per **Parole alfabetiche inverse** vale lo stesso criterio in direzione opposta:

```text
BA
CBA
FED
ZYX
```

## Manipolazione ed estrazione

50. Inverti testo
51. Inverti ordine parole
52. Ordina parole alfabeticamente
53. Estrai numeri
54. Elimina duplicati
55. Estrai parole
56. Testo in MAIUSCOLO
57. Testo in minuscolo
58. Kunta il testo (fenomeni)

---

# Parametro

Alcune operazioni utilizzano il campo **Parametro**.

Esempi:

- **Lettere indicate** — lettere da contare, es. `aeiou`;
- **Occorrenze parola/stringa** — parola o frase da cercare;
- **Palindromi / Bifronti / Anagrammi** — lunghezza minima, default `3`;
- **Schema rime / desinenze** — profondità della coda, default `3`;
- **Omovocaliche/Omoconsonantiche iniziali/finali** — quantità di vocali o consonanti da confrontare;
- **Parole di lunghezza N** — lunghezza esatta;
- **Sequenze ripetute** — lunghezza minima della sequenza.

La casella **Maiuscole/minuscole distinte** decide se, dove previsto, Kunta deve considerare `A` e `a` differenti.

---

# Output

I risultati vengono mostrati in una finestra multilinea e vengono **sostituiti a ogni nuova elaborazione**, evitando la sovrapposizione con il risultato precedente.

Le liste e i gruppi vengono presentati verticalmente, una voce per riga, per mantenere leggibili anche analisi molto estese.

---

# Privacy e funzionamento locale

Kunta Desktop lavora localmente.

- nessun testo viene inviato a ShiduLab;
- nessun testo viene inviato a servizi cloud;
- la ricerca degli anagrammi usa l'indice incorporato nell'eseguibile;
- non è necessaria una connessione Internet per analizzare i testi.

---

# Struttura principale del repository

```text
Kunta/
├─ .github/
│  └─ workflows/
│     └─ forge-kunta.yml
├─ android/
├─ desktop/
│  ├─ main.go
│  ├─ ui_windows.go
│  ├─ engine_test.go
│  └─ assets/
│     ├─ Kunta.ico
│     ├─ Botolo_ShiduLab.*
│     └─ anagram_index.tsv.gz
├─ pwa/
├─ go.mod
└─ README.md
```

---

# Build

Il workflow **Forge Kunta** produce gli artefatti delle tre destinazioni del progetto.

Per il Desktop Windows il sorgente Go viene compilato come applicazione GUI x64 e l'icona Kunta viene incorporata nell'eseguibile.

I test del motore sono contenuti in:

```text
desktop/engine_test.go
```

---

# Obiettivo

**Uno strumento per gli scrittori così come la calcolatrice lo è per i matematici.**

Kunta nasce dalla pratica della parola: usata, ascoltata, piegata, rimata, invertita, osservata e riconosciuta nel suo accadere.

**ShiduLab · Ricerca, visione, narrazione.**
