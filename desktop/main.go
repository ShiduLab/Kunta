package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Kunta Desktop native: single Windows executable, no browser, no local HTTP server.

//go:embed assets/anagram_index.tsv.gz
var anagramIndexGzip []byte

var operations = []struct{ Name, Hint string }{
	{"Lettere indicate", "Scrivi le lettere da contare, es. aeiou"},
	{"Caratteri", "Conta caratteri Unicode e byte UTF-8"},
	{"Parole", "Conta parole e parole distinte"},
	{"Righe", "Conta righe e righe non vuote"},
	{"Spazi", "Conta spazi semplici e caratteri bianchi"},
	{"Cifre", "Conta le cifre"},
	{"Punteggiatura", "Conta i segni di punteggiatura"},
	{"Occorrenze parola/stringa", "Parametro: parola o frase da cercare"},
	{"Parole uniche", "Elenca le parole distinte"},
	{"Parole ripetute", "Elenca solo le parole che ricorrono"},
	{"Frequenza lettere", "Classifica le lettere per frequenza"},
	{"Frequenza parole", "Classifica le parole per frequenza"},
	{"Parola più corta / più lunga", "Trova gli estremi di lunghezza"},
	{"Palindromo puro", "Parametro: lunghezza minima, default 3"},
	{"Bifronti", "Parametro: lunghezza minima, default 3"},
	{"Anagrammi", "Parametro: lunghezza minima, default 3"},
	{"Acrostico", "Prima lettera/cifra utile di ogni riga non vuota"},
	{"Telestico", "Ultima lettera/cifra utile di ogni riga non vuota"},
	{"Inverti testo", "Inverte tutti i caratteri"},
	{"Inverti ordine parole", "Inverte la sequenza delle parole"},
	{"Ordina parole alfabeticamente", "Visualizza subito ripetizioni e famiglie lessicali"},
	{"Estrai numeri", "Estrae le sequenze numeriche"},
	{"Schema rime / desinenze", "Parametro: profondità della coda, default 3"},
	{"Rima baciata / inclusione progressiva", "Parametro: nucleo minimo, default 3"},
	{"LetterTransport", "Parametro: trasporto minimo, default 3"},
	{"Sciarade / salti di dominio", "Parametro: lunghezza minima segmento, default 2"},
	{"Kunta il testo (fenomeni)", "Ricognizione combinata di strutture e ricorrenze"},
	{"Isovocaliche", "Raggruppa parole con lo stesso scheletro vocalico"},
	{"Isoconsonantiche", "Raggruppa parole con lo stesso scheletro consonantico"},
	{"Omovocaliche", "Stesso materiale vocalico; la sequenza resta nell’ordine reale di ogni parola"},
	{"Omoconsonantiche", "Stesso materiale consonantico; la sequenza resta nell’ordine reale di ogni parola"},
	{"Omovocaliche iniziali", "Parametro: quante vocali iniziali confrontare, default 1"},
	{"Omovocaliche finali", "Parametro: quante vocali finali confrontare, default 1"},
	{"Omoconsonantiche iniziali", "Parametro: quante consonanti iniziali confrontare, default 1"},
	{"Omoconsonantiche finali", "Parametro: quante consonanti finali confrontare, default 1"},
	{"Lunghezza media parole", "Calcola la lunghezza media delle parole"},
	{"Distribuzione lunghezze", "Conta quante parole hanno 1, 2, 3… lettere"},
	{"Parole con iniziale", "Parametro: una o più lettere iniziali"},
	{"Parole con finale", "Parametro: una o più lettere finali"},
	{"Parole contenenti sequenza", "Parametro: sequenza da cercare dentro le parole"},
	{"Parole di lunghezza N", "Parametro: numero esatto di lettere"},
	{"Doppie / triple", "Rileva lettere consecutive ripetute nelle parole"},
	{"Sequenze ripetute", "Parametro: lunghezza minima della sequenza, default 2"},
	{"Isogrammi", "Parole senza lettere ripetute"},
	{"Parole alfabetiche", "Lettere consecutive dell’alfabeto: AB, ABC, BCD…"},
	{"Parole alfabetiche inverse", "Lettere consecutive dell’alfabeto al contrario: BA, CBA, FED…"},
	{"Elimina duplicati", "Restituisce le parole una sola volta, nell’ordine di apparizione"},
	{"Estrai parole", "Estrae solo le parole dal testo"},
	{"Testo in MAIUSCOLO", "Converte il testo in maiuscolo"},
	{"Testo in minuscolo", "Converte il testo in minuscolo"},
	{"Palindromo inverso", "Parola letta al contrario; mostra solo coppie realmente presenti nel testo"},
	{"Palindromo contrario", "Rileva forme tipo POSSESSO: prima lettera fissa, resto palindromo"},
	{"Inversi", "Ultima lettera fissa; mostra solo coppie realmente presenti nel testo"},
	{"Antipodi", "Prima lettera fissa; mostra solo coppie realmente presenti nel testo"},
	{"Allitterazioni", "Raggruppa parole per lettera iniziale ricorrente"},
	{"Assonanze", "Raggruppa parole per coda vocalica"},
	{"Consonanze", "Raggruppa parole per coda consonantica"},
	{"Ossimori — candidati", "Cerca coppie di termini semanticamente contrari o paradossali vicini"},
}

func normalizeCase(s string, sensitive bool) string {
	if sensitive {
		return s
	}
	return strings.ToLower(s)
}
func normalizeLines(s string) []string {
	return strings.Split(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n"), "\n")
}
func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }
func tokenizeWords(s string) []string {
	rr := []rune(s)
	out := []string{}
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			out = append(out, string(rr[start:end]))
		}
		start = -1
	}
	for i, r := range rr {
		if isWordRune(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if (r == '\'' || r == '’' || r == '-') && start >= 0 && i+1 < len(rr) && isWordRune(rr[i+1]) {
			continue
		}
		flush(i)
	}
	flush(len(rr))
	return out
}
func cleanWord(s string, sensitive bool) string {
	rr := []rune(s)
	a, b := 0, len(rr)
	for a < b && !isWordRune(rr[a]) {
		a++
	}
	for b > a && !isWordRune(rr[b-1]) {
		b--
	}
	return normalizeCase(string(rr[a:b]), sensitive)
}
func reverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
func runeLen(s string) int { return len([]rune(s)) }
func parsePositive(s string, def, min, max int) int {
	n, e := strconv.Atoi(strings.TrimSpace(s))
	if e != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
func limitLines(s string, max int) string {
	// Desktop: nessun taglio dell'output. L'utente deve poter vedere il risultato completo.
	return s
}

type freqPack struct {
	freq    map[string]int
	display map[string]string
	order   []string
}

func freqWords(words []string, sensitive bool) freqPack {
	p := freqPack{map[string]int{}, map[string]string{}, []string{}}
	for _, w := range words {
		k := cleanWord(w, sensitive)
		if k == "" {
			continue
		}
		p.freq[k]++
		if _, ok := p.display[k]; !ok {
			p.display[k] = w
			p.order = append(p.order, k)
		}
	}
	return p
}
func countOccurrences(hay, needle string) int {
	if needle == "" {
		return 0
	}
	n := 0
	for pos := 0; ; {
		i := strings.Index(hay[pos:], needle)
		if i < 0 {
			break
		}
		n++
		pos += i + len(needle)
	}
	return n
}
func runeSuffix(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}
func sortedRunes(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}
func uniqueDisplayWords(words []string, sensitive bool) []string {
	p := freqWords(words, sensitive)
	out := make([]string, 0, len(p.order))
	for _, k := range p.order {
		out = append(out, p.display[k])
	}
	return out
}
func wordLen(w string) int { return runeLen(cleanWord(w, true)) }

func findPalindromes(words []string, sensitive bool, minLen int) string {
	p := freqWords(words, sensitive)
	keys := []string{}
	total := 0
	for k, n := range p.freq {
		if runeLen(k) >= minLen && k == reverseRunes(k) {
			keys = append(keys, k)
			total += n
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if p.freq[keys[i]] != p.freq[keys[j]] {
			return p.freq[keys[i]] > p.freq[keys[j]]
		}
		return keys[i] < keys[j]
	})
	out := []string{fmt.Sprintf("Palindromi distinti: %d", len(keys)), fmt.Sprintf("Occorrenze totali: %d", total), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s = %d", p.display[k], p.freq[k]))
	}
	return strings.Join(out, "\n")
}
func findBifronti(words []string, sensitive bool, minLen int) string {
	p := freqWords(words, sensitive)
	seen := map[string]bool{}
	pairs := [][2]string{}
	for k := range p.freq {
		if runeLen(k) < minLen {
			continue
		}
		rev := reverseRunes(k)
		if rev == k || p.freq[rev] == 0 {
			continue
		}
		a, b := k, rev
		if a > b {
			a, b = b, a
		}
		key := a + "\x00" + b
		if seen[key] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, [2]string{p.display[k], p.display[rev]})
	}
	sort.Slice(pairs, func(i, j int) bool { return strings.ToLower(pairs[i][0]) < strings.ToLower(pairs[j][0]) })
	out := []string{fmt.Sprintf("Bifronti trovati: %d", len(pairs)), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	for _, p := range pairs {
		out = append(out, p[0]+" ↔ "+p[1])
	}
	return strings.Join(out, "\n")
}
func baseAnagramRune(r rune) rune {
	r = unicode.ToLower(r)
	switch r {
	case 'à', 'á', 'â', 'ä', 'ã', 'å':
		return 'a'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ò', 'ó', 'ô', 'ö', 'õ':
		return 'o'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	case 'ç':
		return 'c'
	}
	if r >= 'a' && r <= 'z' {
		return r
	}
	return 0
}

func anagramIdentity(word string) string {
	var b strings.Builder
	for _, r := range word {
		if unicode.IsNumber(r) {
			return ""
		}
		n := baseAnagramRune(r)
		if n != 0 {
			b.WriteRune(n)
		}
	}
	return b.String()
}

func anagramSignatureKey(word string) (string, int) {
	var counts [26]byte
	letters := 0
	for _, r := range word {
		if unicode.IsNumber(r) {
			return "", 0
		}
		b := baseAnagramRune(r)
		if b == 0 {
			continue
		}
		counts[b-'a']++
		letters++
	}
	if letters == 0 {
		return "", 0
	}
	return fmt.Sprintf("%x", counts[:]), letters
}

func dictionaryAnagrams(signatures map[string]bool) (map[string][]string, error) {
	found := make(map[string][]string, len(signatures))
	zr, err := gzip.NewReader(bytes.NewReader(anagramIndexGzip))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		tab := strings.IndexByte(line, '\t')
		if tab <= 0 {
			continue
		}
		sig := line[:tab]
		if !signatures[sig] {
			continue
		}
		vals := strings.Split(line[tab+1:], "|")
		found[sig] = vals
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return found, nil
}

func anagramLexeme(raw string) string {
	w := cleanWord(raw, true)
	// Per gli anagrammi analizziamo la parola lessicale, non l'elisione che la precede:
	// L'Empatia -> Empatia, nell'amare -> amare, un'altra -> altra.
	last := -1
	sepLen := 0
	if i := strings.LastIndex(w, "'"); i > last {
		last, sepLen = i, len("'")
	}
	if i := strings.LastIndex(w, "’"); i > last {
		last, sepLen = i, len("’")
	}
	if last >= 0 && last+sepLen < len(w) {
		w = w[last+sepLen:]
	}
	return cleanWord(w, true)
}

func findAnagrams(words []string, sensitive bool, minLen int) string {
	targets := map[string]bool{}
	wordSig := map[string]string{}
	display := map[string]string{}
	orderedWords := []string{}

	for _, raw := range words {
		lexeme := anagramLexeme(raw)
		if lexeme == "" {
			continue
		}
		k := normalizeCase(lexeme, sensitive)
		if _, seen := display[k]; seen {
			continue
		}
		sig, letters := anagramSignatureKey(k)
		if sig == "" || letters < minLen {
			continue
		}
		targets[sig] = true
		wordSig[k] = sig
		display[k] = lexeme
		orderedWords = append(orderedWords, k)
	}

	if len(targets) == 0 {
		return fmt.Sprintf("Anagrammi trovati: 0\nLunghezza minima: %d\n\nNessuna parola analizzabile.", minLen)
	}

	matches, err := dictionaryAnagrams(targets)
	if err != nil {
		return "Errore nell'indice anagrammatico: " + err.Error()
	}

	out := []string{}
	groups := 0
	for _, k := range orderedWords {
		sig := wordSig[k]
		candidates := matches[sig]
		if len(candidates) == 0 {
			continue
		}

		sourceNorm := anagramIdentity(display[k])
		seen := map[string]bool{}
		filtered := []string{}
		for _, candidate := range candidates {
			cn := anagramIdentity(candidate)
			if cn == "" || cn == sourceNorm || seen[cn] {
				continue
			}
			seen[cn] = true
			filtered = append(filtered, candidate)
		}
		if len(filtered) == 0 {
			continue
		}
		sort.Slice(filtered, func(i, j int) bool { return strings.ToLower(filtered[i]) < strings.ToLower(filtered[j]) })
		if groups > 0 {
			out = append(out, "")
		}
		out = append(out, display[k]+" →")
		for _, candidate := range filtered {
			out = append(out, "    "+candidate)
		}
		groups++
	}

	header := []string{fmt.Sprintf("Anagrammi: %d parole del testo con almeno un anagramma", groups), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	if groups == 0 {
		header = append(header, "Nessun anagramma trovato nel dizionario.")
		return strings.Join(header, "\n")
	}
	return strings.Join(append(header, out...), "\n")
}
func finalWordOfLine(line string, sensitive bool) string {
	w := tokenizeWords(line)
	if len(w) == 0 {
		return ""
	}
	return cleanWord(w[len(w)-1], sensitive)
}
func rhymeLabel(n int) string {
	labels := []rune("xyzwvutsrqponmlkjihgfedcba")
	if n < len(labels) {
		return string(labels[n])
	}
	return fmt.Sprintf("r%d", n+1)
}
func analyzeRhymeScheme(text string, sensitive bool, depth int) string {
	lines := normalizeLines(text)
	stanzas := [][]string{}
	cur := []string{}
	flush := func() {
		if len(cur) > 0 {
			stanzas = append(stanzas, cur)
			cur = []string{}
		}
	}
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			flush()
		} else {
			cur = append(cur, l)
		}
	}
	flush()
	if len(stanzas) == 0 {
		return "Nessun verso rilevato."
	}
	out := []string{fmt.Sprintf("Rima = desinenza grafica · profondità: %d lettere", depth), fmt.Sprintf("Strofe rilevate: %d", len(stanzas)), ""}
	for si, st := range stanzas {
		type row struct{ word, suffix, label string }
		rows := []row{}
		counts := map[string]int{}
		order := []string{}
		seen := map[string]bool{}
		for _, l := range st {
			w := finalWordOfLine(l, sensitive)
			if w == "" {
				continue
			}
			s := runeSuffix(w, depth)
			rows = append(rows, row{word: w, suffix: s})
			counts[s]++
			if !seen[s] {
				seen[s] = true
				order = append(order, s)
			}
		}
		if len(rows) == 0 {
			continue
		}
		groups := map[string]string{}
		if len(order) == 2 {
			single, dom := "", ""
			for _, s := range order {
				if counts[s] == 1 {
					single = s
				} else {
					dom = s
				}
			}
			if single != "" && dom != "" {
				groups[dom] = "x"
				groups[single] = "z"
			}
		}
		if len(groups) == 0 {
			for i, s := range order {
				groups[s] = rhymeLabel(i)
			}
		}
		for i := range rows {
			rows[i].label = groups[rows[i].suffix]
		}
		if len(stanzas) > 1 {
			out = append(out, fmt.Sprintf("STROFA %d", si+1))
		}
		for i, r := range rows {
			out = append(out, fmt.Sprintf("%2d. %-22s → -%-10s → %s", i+1, r.word, r.suffix, r.label))
		}
		labs := []string{}
		for _, r := range rows {
			labs = append(labs, r.label)
		}
		out = append(out, "", "Schema: "+strings.Join(labs, "-"), "Coda progressiva:")
		for d := 1; d <= depth; d++ {
			f := map[string]int{}
			for _, r := range rows {
				f[runeSuffix(r.word, d)]++
			}
			ks := []string{}
			for k, n := range f {
				if n > 1 {
					ks = append(ks, k)
				}
			}
			sort.Strings(ks)
			parts := []string{}
			for _, k := range ks {
				parts = append(parts, fmt.Sprintf("-%s ×%d", k, f[k]))
			}
			if len(parts) > 0 {
				out = append(out, fmt.Sprintf("  %d:", d))
				for _, part := range parts {
					out = append(out, "    "+part)
				}
			}
		}
		if si < len(stanzas)-1 {
			out = append(out, "")
		}
	}
	return limitLines(strings.Join(out, "\n"), 700)
}

func findInclusionChains(words []string, sensitive bool, minLen int) string {
	p := freqWords(words, sensitive)
	keys := []string{}
	for _, k := range p.order {
		if runeLen(k) >= minLen {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if runeLen(keys[i]) != runeLen(keys[j]) {
			return runeLen(keys[i]) < runeLen(keys[j])
		}
		return keys[i] < keys[j]
	})
	children := map[string][]string{}
	incoming := map[string]int{}
	for _, a := range keys {
		la := runeLen(a)
		cand := []string{}
		for _, b := range keys {
			if runeLen(b) > la && strings.HasSuffix(b, a) {
				cand = append(cand, b)
			}
		}
		for _, b := range cand {
			lb := runeLen(b)
			immediate := true
			for _, c := range cand {
				lc := runeLen(c)
				if c != b && lc > la && lc < lb && strings.HasSuffix(c, a) && strings.HasSuffix(b, c) {
					immediate = false
					break
				}
			}
			if immediate {
				children[a] = append(children[a], b)
				incoming[b]++
			}
		}
		sort.Slice(children[a], func(i, j int) bool {
			if runeLen(children[a][i]) != runeLen(children[a][j]) {
				return runeLen(children[a][i]) < runeLen(children[a][j])
			}
			return children[a][i] < children[a][j]
		})
	}
	chains := [][]string{}
	var dfs func(string, []string)
	dfs = func(node string, path []string) {
		q := append(append([]string{}, path...), node)
		ch := children[node]
		if len(ch) == 0 {
			if len(q) >= 2 {
				chains = append(chains, q)
			}
			return
		}
		for _, c := range ch {
			dfs(c, q)
		}
	}
	for _, k := range keys {
		if incoming[k] == 0 && len(children[k]) > 0 {
			dfs(k, nil)
		}
	}
	sort.Slice(chains, func(i, j int) bool {
		if len(chains[i]) != len(chains[j]) {
			return len(chains[i]) > len(chains[j])
		}
		return chains[i][0] < chains[j][0]
	})
	out := []string{"Rima baciata · inclusione progressiva", fmt.Sprintf("Nucleo minimo: %d lettere", minLen), fmt.Sprintf("Catene trovate: %d", len(chains)), ""}
	for i, c := range chains {
		if i >= 100 {
			break
		}
		ds := []string{}
		for _, k := range c {
			ds = append(ds, p.display[k])
		}
		out = append(out, fmt.Sprintf("%2d. %s", i+1, strings.Join(ds, " → ")), fmt.Sprintf("    nucleo: %s · profondità: %d", p.display[c[0]], len(c)))
	}
	if len(chains) == 0 {
		out = append(out, "Nessuna inclusione progressiva rilevata.")
	}
	if len(chains) > 100 {
		out = append(out, "\n… altre catene omesse.")
	}
	return strings.Join(out, "\n")
}
func longestCommonSubstring(a, b string) string {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	bestLen, bestEnd := 0, 0
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		for j := 1; j <= len(br); j++ {
			if ar[i-1] == br[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > bestLen {
					bestLen = cur[j]
					bestEnd = i
				}
			}
		}
		prev = cur
	}
	return string(ar[bestEnd-bestLen : bestEnd])
}
func findLetterTransport(words []string, sensitive bool, minLen int) string {
	if len(words) < 2 {
		return "Servono almeno due parole."
	}
	out := []string{"LetterTransport · parole consecutive", fmt.Sprintf("Trasporto minimo: %d lettere", minLen), ""}
	found := 0
	for i := 0; i+1 < len(words) && found < 200; i++ {
		a0, b0 := words[i], words[i+1]
		a, b := cleanWord(a0, sensitive), cleanWord(b0, sensitive)
		if a == "" || b == "" || a == b {
			continue
		}
		c := longestCommonSubstring(a, b)
		n := runeLen(c)
		if n < minLen {
			continue
		}
		found++
		out = append(out, fmt.Sprintf("%3d. %s → %s", found, a0, b0), fmt.Sprintf("     trasporta: %q (%d lettere)", c, n))
	}
	if found == 0 {
		out = append(out, "Nessun trasporto consecutivo rilevato.")
	}
	if found >= 200 {
		out = append(out, "\n… output limitato a 200 passaggi.")
	}
	return strings.Join(out, "\n")
}
func phraseExists(tokens, parts []string) bool {
	for i := 0; i+len(parts) <= len(tokens); i++ {
		ok := true
		for j := range parts {
			if tokens[i+j] != parts[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func findSciarades(text string, sensitive bool, minSeg int) string {
	raw := tokenizeWords(text)
	tokens := []string{}
	counts := map[string]int{}
	display := map[string]string{}
	for _, w := range raw {
		k := cleanWord(w, sensitive)
		if k == "" {
			continue
		}
		tokens = append(tokens, k)
		counts[k]++
		if _, ok := display[k]; !ok {
			display[k] = w
		}
	}
	targets := []string{}
	for k := range counts {
		if runeLen(k) >= minSeg*2 {
			targets = append(targets, k)
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		if runeLen(targets[i]) != runeLen(targets[j]) {
			return runeLen(targets[i]) > runeLen(targets[j])
		}
		return targets[i] < targets[j]
	})
	out := []string{"Sciarade interne · segmenti presenti nel testo", fmt.Sprintf("Segmento minimo: %d lettere", minSeg), ""}
	found := 0
	for _, target := range targets {
		r := []rune(target)
		set := map[string]bool{}
		segs := []string{}
		for i := minSeg; i <= len(r)-minSeg; i++ {
			a, b := string(r[:i]), string(r[i:])
			if counts[a] == 0 || counts[b] == 0 || !phraseExists(tokens, []string{a, b}) {
				continue
			}
			k := a + "|" + b
			if !set[k] {
				set[k] = true
				segs = append(segs, k)
			}
		}
		for i := minSeg; i <= len(r)-2*minSeg; i++ {
			for j := i + minSeg; j <= len(r)-minSeg; j++ {
				a, b, c := string(r[:i]), string(r[i:j]), string(r[j:])
				if counts[a] == 0 || counts[b] == 0 || counts[c] == 0 || !phraseExists(tokens, []string{a, b, c}) {
					continue
				}
				k := a + "|" + b + "|" + c
				if !set[k] {
					set[k] = true
					segs = append(segs, k)
				}
			}
		}
		if len(segs) == 0 {
			continue
		}
		found++
		head := display[target]
		if counts[target] > 1 {
			head += fmt.Sprintf("  [forma intera ×%d]", counts[target])
		}
		out = append(out, head)
		sort.Strings(segs)
		for _, seg := range segs {
			parts := strings.Split(seg, "|")
			for i, p := range parts {
				if d := display[p]; d != "" {
					parts[i] = d
				}
			}
			out = append(out, "  → "+strings.Join(parts, " + "))
		}
		if counts[target] > 1 {
			out = append(out, "  ↳ possibile salto di dominio: stessa forma intera + segmentazioni diverse; verifica il senso nel contesto.")
		}
		out = append(out, "")
		if found >= 100 {
			out = append(out, "… altri candidati omessi.")
			break
		}
	}
	if found == 0 {
		out = append(out, "Nessuna sciarada interna rilevata con il vocabolario del testo.")
	}
	return strings.Join(out, "\n")
}

func foldItalianVowel(r rune) rune {
	switch unicode.ToLower(r) {
	case 'a', 'à', 'á', 'â', 'ä', 'ã', 'å':
		return 'a'
	case 'e', 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'i', 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'o', 'ò', 'ó', 'ô', 'ö', 'õ':
		return 'o'
	case 'u', 'ù', 'ú', 'û', 'ü':
		return 'u'
	}
	return 0
}
func vowelSkeleton(w string) string {
	var b strings.Builder
	for _, r := range cleanWord(w, false) {
		if v := foldItalianVowel(r); v != 0 {
			b.WriteRune(v)
		}
	}
	return b.String()
}
func consonantSkeleton(w string) string {
	var b strings.Builder
	for _, r := range cleanWord(w, false) {
		if unicode.IsLetter(r) && foldItalianVowel(r) == 0 {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}
func sortedSignature(s string) string { return sortedRunes(s) }
func boundarySignature(s string, n int, end bool) string {
	r := []rune(s)
	if n < 1 {
		n = 1
	}
	if len(r) < n {
		return ""
	}
	if end {
		return string(r[len(r)-n:])
	}
	return string(r[:n])
}
func prettySignature(s string) string {
	r := []rune(strings.ToUpper(s))
	a := make([]string, len(r))
	for i, x := range r {
		a[i] = string(x)
	}
	return strings.Join(a, "-")
}
func groupWordSignatures(words []string, label string, sig func(string) string) string {
	groups := map[string][]string{}
	seen := map[string]map[string]bool{}
	for _, raw := range words {
		clean := cleanWord(raw, false)
		if clean == "" {
			continue
		}
		k := sig(raw)
		if k == "" {
			continue
		}
		if seen[k] == nil {
			seen[k] = map[string]bool{}
		}
		if !seen[k][clean] {
			seen[k][clean] = true
			groups[k] = append(groups[k], raw)
		}
	}
	keys := []string{}
	for k, a := range groups {
		if len(a) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if runeLen(keys[i]) != runeLen(keys[j]) {
			return runeLen(keys[i]) > runeLen(keys[j])
		}
		return keys[i] < keys[j]
	})
	out := []string{fmt.Sprintf("%s: %d gruppi", label, len(keys)), ""}
	for _, k := range keys {
		out = append(out, prettySignature(k)+"  →")
		for _, item := range groups[k] {
			out = append(out, "    "+item)
		}
	}
	if len(keys) == 0 {
		out = append(out, "Nessun gruppo rilevato nel testo.")
	}
	return limitLines(strings.Join(out, "\n"), 700)
}
func materialCountLabel(s string) string {
	counts := map[rune]int{}
	for _, r := range s {
		counts[unicode.ToUpper(r)]++
	}
	keys := make([]rune, 0, len(counts))
	for r := range counts {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(keys))
	for _, r := range keys {
		parts = append(parts, fmt.Sprintf("%c×%d", r, counts[r]))
	}
	return strings.Join(parts, " ")
}

func groupWordMaterials(words []string, label string, orderedSig func(string) string) string {
	groups := map[string][]string{}
	seen := map[string]map[string]bool{}
	for _, raw := range words {
		clean := cleanWord(raw, false)
		if clean == "" {
			continue
		}
		ordered := orderedSig(raw)
		if ordered == "" {
			continue
		}
		material := sortedSignature(ordered)
		if seen[material] == nil {
			seen[material] = map[string]bool{}
		}
		if !seen[material][clean] {
			seen[material][clean] = true
			groups[material] = append(groups[material], raw)
		}
	}
	keys := []string{}
	for k, a := range groups {
		if len(a) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if runeLen(keys[i]) != runeLen(keys[j]) {
			return runeLen(keys[i]) > runeLen(keys[j])
		}
		return keys[i] < keys[j]
	})
	out := []string{fmt.Sprintf("%s: %d gruppi", label, len(keys)), ""}
	for _, k := range keys {
		out = append(out, "Materiale: "+materialCountLabel(k))
		for _, item := range groups[k] {
			out = append(out, fmt.Sprintf("    %s  →  %s", item, prettySignature(orderedSig(item))))
		}
		out = append(out, "")
	}
	if len(keys) == 0 {
		out = append(out, "Nessun gruppo rilevato nel testo.")
	}
	return limitLines(strings.Join(out, "\n"), 700)
}

func repeatedWordsSummary(text string, sensitive bool, maxItems int) string {
	p := freqWords(tokenizeWords(text), sensitive)
	keys := []string{}
	for k, n := range p.freq {
		if n > 1 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if p.freq[keys[i]] != p.freq[keys[j]] {
			return p.freq[keys[i]] > p.freq[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > maxItems {
		keys = keys[:maxItems]
	}
	if len(keys) == 0 {
		return "Nessuna parola ripetuta."
	}
	a := []string{}
	for _, k := range keys {
		a = append(a, fmt.Sprintf("%s = %d", p.display[k], p.freq[k]))
	}
	return strings.Join(a, "\n")
}
func kuntaPhenomena(text string, sensitive bool) string {
	return limitLines(strings.Join([]string{"KUNTA IL TESTO", "==============================", "", "RIPETIZIONI", repeatedWordsSummary(text, sensitive, 20), "", "SCHEMA RIME / DESINENZE", analyzeRhymeScheme(text, sensitive, 3), "", "RIMA BACIATA / INCLUSIONE PROGRESSIVA", findInclusionChains(tokenizeWords(text), sensitive, 3), "", "LETTERTRANSPORT", findLetterTransport(tokenizeWords(text), sensitive, 3), "", "SCIARADE / SALTI DI DOMINIO", findSciarades(text, sensitive, 2), "", "ISOVOCALICHE", groupWordSignatures(tokenizeWords(text), "Scheletro vocalico", vowelSkeleton), "", "ISOCONSONANTICHE", groupWordSignatures(tokenizeWords(text), "Scheletro consonantico", consonantSkeleton)}, "\n"), 1100)
}
func filterWordsBy(words []string, p string, sensitive bool, mode string) string {
	q := normalizeCase(strings.TrimSpace(p), sensitive)
	if q == "" {
		return "Scrivi il parametro da cercare."
	}
	a := []string{}
	for _, w := range uniqueDisplayWords(words, sensitive) {
		k := normalizeCase(cleanWord(w, true), sensitive)
		ok := mode == "start" && strings.HasPrefix(k, q) || mode == "end" && strings.HasSuffix(k, q) || mode == "contains" && strings.Contains(k, q)
		if ok {
			a = append(a, w)
		}
	}
	return fmt.Sprintf("%d parole trovate\n\n%s", len(a), strings.Join(a, "\n"))
}
func findRepeatedRuns(words []string) string {
	out := []string{}
	for _, w := range uniqueDisplayWords(words, false) {
		r := []rune(cleanWord(w, false))
		runs := []string{}
		for i := 0; i < len(r); {
			j := i + 1
			for j < len(r) && r[j] == r[i] {
				j++
			}
			if j-i >= 2 {
				runs = append(runs, string(r[i:j]))
			}
			i = j
		}
		if len(runs) > 0 {
			out = append(out, w+"  →")
			for _, item := range runs {
				out = append(out, "    "+item)
			}
		}
	}
	return fmt.Sprintf("Parole con doppie/triple: %d\n\n%s", len(out), strings.Join(out, "\n"))
}
func findRepeatedSequences(words []string, minLen int) string {
	out := []string{}
	for _, w := range uniqueDisplayWords(words, false) {
		r := []rune(cleanWord(w, false))
		found := map[string]bool{}
		for n := minLen; n <= len(r)/2; n++ {
			for i := 0; i+n <= len(r); i++ {
				seq := string(r[i : i+n])
				c := 0
				for j := 0; j+n <= len(r); j++ {
					if string(r[j:j+n]) == seq {
						c++
					}
				}
				if c > 1 {
					found[seq] = true
				}
			}
		}
		if len(found) > 0 {
			a := []string{}
			for s := range found {
				a = append(a, s)
			}
			sort.Slice(a, func(i, j int) bool {
				if runeLen(a[i]) != runeLen(a[j]) {
					return runeLen(a[i]) > runeLen(a[j])
				}
				return a[i] < a[j]
			})
			out = append(out, w+"  →")
			for _, item := range a {
				out = append(out, "    "+item)
			}
		}
	}
	return fmt.Sprintf("Parole con sequenze ripetute: %d\nLunghezza minima: %d\n\n%s", len(out), minLen, strings.Join(out, "\n"))
}
func isIsogram(w string) bool {
	seen := map[rune]bool{}
	n := 0
	for _, r := range cleanWord(w, false) {
		if !unicode.IsLetter(r) {
			continue
		}
		n++
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	return n > 1
}
func foldAlphabetRune(r rune) rune {
	r = unicode.ToLower(r)
	switch r {
	case 'à', 'á', 'â', 'ä', 'ã', 'å':
		return 'a'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ò', 'ó', 'ô', 'ö', 'õ':
		return 'o'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	}
	return r
}

func alphabeticWord(w string, descending bool) bool {
	a := []rune{}
	for _, r := range cleanWord(w, false) {
		if !unicode.IsLetter(r) {
			continue
		}
		r = foldAlphabetRune(r)
		if r < 'a' || r > 'z' {
			return false
		}
		a = append(a, r)
	}
	if len(a) < 2 {
		return false
	}
	step := rune(1)
	if descending {
		step = -1
	}
	for i := 1; i < len(a); i++ {
		if a[i] != a[i-1]+step {
			return false
		}
	}
	return true
}

func displayCharacter(r rune, sensitive bool) string {
	if !sensitive && unicode.IsLetter(r) {
		r = unicode.ToUpper(r)
	}
	switch r {
	case ' ':
		return "[spazio]"
	case '\n':
		return "[a capo]"
	case '\r':
		return "[ritorno carrello]"
	case '\t':
		return "[tab]"
	}
	return string(r)
}

func characterFrequency(text string, sensitive bool) string {
	freq := map[rune]int{}
	for _, r := range text {
		k := r
		if !sensitive && unicode.IsLetter(k) {
			k = unicode.ToLower(k)
		}
		freq[k]++
	}
	keys := make([]rune, 0, len(freq))
	for r := range freq {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool {
		if freq[keys[i]] != freq[keys[j]] {
			return freq[keys[i]] > freq[keys[j]]
		}
		return displayCharacter(keys[i], sensitive) < displayCharacter(keys[j], sensitive)
	})
	out := []string{fmt.Sprintf("Caratteri: %d", runeLen(text)), ""}
	for _, r := range keys {
		out = append(out, fmt.Sprintf("%d: %s", freq[r], displayCharacter(r, sensitive)))
	}
	return strings.Join(out, "\n")
}

func wordFrequencySummary(words []string, sensitive bool) string {
	p := freqWords(words, sensitive)
	keys := append([]string{}, p.order...)
	sort.Slice(keys, func(i, j int) bool {
		if p.freq[keys[i]] != p.freq[keys[j]] {
			return p.freq[keys[i]] > p.freq[keys[j]]
		}
		return keys[i] < keys[j]
	})
	out := []string{fmt.Sprintf("Parole: %d", len(words)), fmt.Sprintf("Parole distinte: %d", len(keys)), ""}
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d: %s", p.freq[k], p.display[k]))
	}
	return strings.Join(out, "\n")
}
func findContraryPalindromes(words []string, minLen int) string {
	out := []string{}
	for _, w := range uniqueDisplayWords(words, false) {
		k := cleanWord(w, false)
		r := []rune(k)
		if len(r) < minLen+1 {
			continue
		}
		tail := string(r[1:])
		if tail == reverseRunes(tail) {
			out = append(out, w)
		}
	}
	return fmt.Sprintf("Palindromi contrari: %d\n\n%s", len(out), strings.Join(out, "\n"))
}

func findReversePairs(words []string, sensitive bool, minLen int, label string) string {
	p := freqWords(words, sensitive)
	seen := map[string]bool{}
	pairs := [][2]string{}
	for k := range p.freq {
		if runeLen(k) < minLen {
			continue
		}
		rev := reverseRunes(k)
		if rev == k || p.freq[rev] == 0 {
			continue
		}
		a, b := k, rev
		if a > b {
			a, b = b, a
		}
		key := a + "\x00" + b
		if seen[key] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, [2]string{p.display[k], p.display[rev]})
	}
	sort.Slice(pairs, func(i, j int) bool { return strings.ToLower(pairs[i][0]) < strings.ToLower(pairs[j][0]) })
	out := []string{fmt.Sprintf("%s trovati: %d", label, len(pairs)), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	for _, pair := range pairs {
		out = append(out, pair[0]+" ↔ "+pair[1])
	}
	if len(pairs) == 0 {
		out = append(out, "Nessuna coppia presente nel testo.")
	}
	return strings.Join(out, "\n")
}

func transformInverse(w string) string {
	r := []rune(cleanWord(w, true))
	if len(r) < 2 {
		return ""
	}
	prefix := append([]rune(nil), r[:len(r)-1]...)
	for i, j := 0, len(prefix)-1; i < j; i, j = i+1, j-1 {
		prefix[i], prefix[j] = prefix[j], prefix[i]
	}
	return string(prefix) + string(r[len(r)-1])
}

func transformAntipode(w string) string {
	r := []rune(cleanWord(w, true))
	if len(r) < 2 {
		return ""
	}
	tail := append([]rune(nil), r[1:]...)
	for i, j := 0, len(tail)-1; i < j; i, j = i+1, j-1 {
		tail[i], tail[j] = tail[j], tail[i]
	}
	return string(r[0]) + string(tail)
}

func findTransformPairs(words []string, sensitive bool, minLen int, label string, transform func(string) string) string {
	p := freqWords(words, sensitive)
	seen := map[string]bool{}
	pairs := [][2]string{}
	for k := range p.freq {
		if runeLen(k) < minLen {
			continue
		}
		other := transform(k)
		if other == "" || other == k || p.freq[other] == 0 {
			continue
		}
		a, b := k, other
		if a > b {
			a, b = b, a
		}
		key := a + "\x00" + b
		if seen[key] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, [2]string{p.display[k], p.display[other]})
	}
	sort.Slice(pairs, func(i, j int) bool { return strings.ToLower(pairs[i][0]) < strings.ToLower(pairs[j][0]) })
	out := []string{fmt.Sprintf("%s trovati: %d", label, len(pairs)), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	for _, pair := range pairs {
		out = append(out, pair[0]+" ↔ "+pair[1])
	}
	if len(pairs) == 0 {
		out = append(out, "Nessuna coppia presente nel testo.")
	}
	return strings.Join(out, "\n")
}

func suffixChars(s string, n int) string {
	r := []rune(s)
	if len(r) < n {
		n = len(r)
	}
	return string(r[len(r)-n:])
}
func groupByKey(words []string, label string, keyFn func(string) string) string {
	g := map[string][]string{}
	seen := map[string]map[string]bool{}
	for _, w := range words {
		clean := cleanWord(w, false)
		if clean == "" {
			continue
		}
		k := keyFn(w)
		if k == "" {
			continue
		}
		if seen[k] == nil {
			seen[k] = map[string]bool{}
		}
		if !seen[k][clean] {
			seen[k][clean] = true
			g[k] = append(g[k], w)
		}
	}
	keys := []string{}
	for k, a := range g {
		if len(a) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(g[keys[i]]) != len(g[keys[j]]) {
			return len(g[keys[i]]) > len(g[keys[j]])
		}
		return keys[i] < keys[j]
	})
	out := []string{fmt.Sprintf("%s: %d gruppi", label, len(keys)), ""}
	for _, k := range keys {
		out = append(out, strings.ToUpper(k)+"  →")
		for _, item := range g[k] {
			out = append(out, "    "+item)
		}
	}
	return strings.Join(out, "\n")
}
func findOxymoronCandidates(text string) string {
	pairs := [][2]string{{"vivo", "morto"}, {"vita", "morte"}, {"caldo", "freddo"}, {"luce", "buio"}, {"chiaro", "scuro"}, {"vero", "falso"}, {"pieno", "vuoto"}, {"grande", "piccolo"}, {"alto", "basso"}, {"forte", "debole"}, {"ricco", "povero"}, {"aperto", "chiuso"}, {"vicino", "lontano"}, {"vecchio", "nuovo"}, {"pace", "guerra"}, {"amore", "odio"}, {"ordine", "caos"}, {"presenza", "assenza"}, {"sacro", "profano"}, {"naturale", "artificiale"}, {"silenzio", "assordante"}, {"urlo", "silenzioso"}, {"ghiaccio", "bollente"}, {"fuoco", "freddo"}, {"dolce", "amaro"}, {"innocente", "colpevole"}}
	t := tokenizeWords(text)
	for i := range t {
		t[i] = cleanWord(t[i], false)
	}
	hits := []string{}
	seen := map[string]bool{}
	count := 0
	for i := 0; i < len(t); i++ {
		end := i + 5
		if end > len(t) {
			end = len(t)
		}
		for j := i + 1; j < end; j++ {
			for _, p := range pairs {
				if (t[i] == p[0] && t[j] == p[1]) || (t[i] == p[1] && t[j] == p[0]) {
					count++
					h := strings.Join(t[i:j+1], " ")
					if !seen[h] {
						seen[h] = true
						hits = append(hits, h)
					}
					break
				}
			}
		}
	}
	return fmt.Sprintf("Ossimori / contrasti candidati: %d\n\n%s", count, strings.Join(hits, "\n"))
}

func analyze(text string, op int, p string, sensitive bool) string {
	words := tokenizeWords(text)
	lines := normalizeLines(text)
	switch op {
	case 0:
		if strings.TrimSpace(p) == "" {
			return "Scrivi nel campo Parametro le lettere da contare."
		}
		seen := map[rune]bool{}
		targets := []rune{}
		for _, r := range p {
			if unicode.IsSpace(r) {
				continue
			}
			k := r
			if !sensitive {
				k = unicode.ToLower(k)
			}
			if !seen[k] {
				seen[k] = true
				targets = append(targets, k)
			}
		}
		counts := map[rune]int{}
		for _, r := range text {
			k := r
			if !sensitive {
				k = unicode.ToLower(k)
			}
			if seen[k] {
				counts[k]++
			}
		}
		out := []string{}
		total := 0
		for _, r := range targets {
			out = append(out, fmt.Sprintf("%c = %d", r, counts[r]))
			total += counts[r]
		}
		return strings.Join(out, "\n") + fmt.Sprintf("\n\nTotale caratteri cercati: %d", total)
	case 1:
		return characterFrequency(text, sensitive)
	case 2:
		return wordFrequencySummary(words, sensitive)
	case 3:
		n := 0
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				n++
			}
		}
		return fmt.Sprintf("Righe: %d\nRighe non vuote: %d", len(lines), n)
	case 4:
		spaces, white := 0, 0
		for _, r := range text {
			if r == ' ' {
				spaces++
			}
			if unicode.IsSpace(r) {
				white++
			}
		}
		return fmt.Sprintf("Spazi semplici: %d\nCaratteri bianchi totali: %d", spaces, white)
	case 5:
		n := 0
		for _, r := range text {
			if unicode.IsNumber(r) {
				n++
			}
		}
		return fmt.Sprintf("Cifre: %d", n)
	case 6:
		n := 0
		for _, r := range text {
			if unicode.IsPunct(r) {
				n++
			}
		}
		return fmt.Sprintf("Segni di punteggiatura: %d", n)
	case 7:
		q := strings.TrimSpace(p)
		if q == "" {
			return "Scrivi una parola o stringa da cercare."
		}
		return fmt.Sprintf("%q\nOccorrenze: %d", q, countOccurrences(normalizeCase(text, sensitive), normalizeCase(q, sensitive)))
	case 8:
		x := freqWords(words, sensitive)
		ks := append([]string{}, x.order...)
		sort.Strings(ks)
		a := []string{}
		for _, k := range ks {
			a = append(a, x.display[k])
		}
		return limitLines(fmt.Sprintf("Parole uniche: %d\n\n%s", len(a), strings.Join(a, "\n")), 700)
	case 9:
		x := freqWords(words, sensitive)
		ks := []string{}
		for k, n := range x.freq {
			if n > 1 {
				ks = append(ks, k)
			}
		}
		sort.Slice(ks, func(i, j int) bool {
			if x.freq[ks[i]] != x.freq[ks[j]] {
				return x.freq[ks[i]] > x.freq[ks[j]]
			}
			return ks[i] < ks[j]
		})
		a := []string{}
		for _, k := range ks {
			a = append(a, fmt.Sprintf("%s = %d", x.display[k], x.freq[k]))
		}
		return limitLines(fmt.Sprintf("Parole ripetute: %d\n\n%s", len(a), strings.Join(a, "\n")), 700)
	case 10:
		f := map[rune]int{}
		for _, r := range normalizeCase(text, sensitive) {
			if unicode.IsLetter(r) {
				f[r]++
			}
		}
		ks := []rune{}
		for k := range f {
			ks = append(ks, k)
		}
		sort.Slice(ks, func(i, j int) bool {
			if f[ks[i]] != f[ks[j]] {
				return f[ks[i]] > f[ks[j]]
			}
			return ks[i] < ks[j]
		})
		a := []string{}
		for _, k := range ks {
			a = append(a, fmt.Sprintf("%c = %d", k, f[k]))
		}
		return strings.Join(a, "\n")
	case 11:
		x := freqWords(words, sensitive)
		ks := append([]string{}, x.order...)
		sort.Slice(ks, func(i, j int) bool {
			if x.freq[ks[i]] != x.freq[ks[j]] {
				return x.freq[ks[i]] > x.freq[ks[j]]
			}
			return ks[i] < ks[j]
		})
		a := []string{}
		for _, k := range ks {
			a = append(a, fmt.Sprintf("%s = %d", x.display[k], x.freq[k]))
		}
		return limitLines(strings.Join(a, "\n"), 700)
	case 12:
		if len(words) == 0 {
			return "Nessuna parola."
		}
		min, max := 1<<30, 0
		mins, maxs := []string{}, []string{}
		for _, w := range words {
			n := wordLen(w)
			if n < min {
				min = n
				mins = []string{w}
			} else if n == min {
				mins = appendUnique(mins, w)
			}
			if n > max {
				max = n
				maxs = []string{w}
			} else if n == max {
				maxs = appendUnique(maxs, w)
			}
		}
		return fmt.Sprintf("Più corta (%d):\n%s\n\nPiù lunga (%d):\n%s", min, strings.Join(mins, "\n"), max, strings.Join(maxs, "\n"))
	case 13:
		return findPalindromes(words, sensitive, parsePositive(p, 3, 1, 100))
	case 14:
		return findBifronti(words, sensitive, parsePositive(p, 3, 1, 100))
	case 15:
		return findAnagrams(words, sensitive, parsePositive(p, 3, 1, 100))
	case 16:
		a := []rune{}
		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			for _, r := range l {
				if isWordRune(r) {
					a = append(a, r)
					break
				}
			}
		}
		return fmt.Sprintf("Acrostico (%d righe):\n\n%s", len(a), string(a))
	case 17:
		a := []rune{}
		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			r := []rune(l)
			for i := len(r) - 1; i >= 0; i-- {
				if isWordRune(r[i]) {
					a = append(a, r[i])
					break
				}
			}
		}
		return fmt.Sprintf("Telestico (%d righe):\n\n%s", len(a), string(a))
	case 18:
		return reverseRunes(text)
	case 19:
		a := append([]string{}, words...)
		for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
			a[i], a[j] = a[j], a[i]
		}
		return strings.Join(a, " ")
	case 20:
		a := append([]string{}, words...)
		sort.Slice(a, func(i, j int) bool { return normalizeCase(a[i], sensitive) < normalizeCase(a[j], sensitive) })
		return strings.Join(a, "\n")
	case 21:
		nums := extractNumbers(text)
		return fmt.Sprintf("Sequenze numeriche: %d\n\n%s", len(nums), strings.Join(nums, "\n"))
	case 22:
		return analyzeRhymeScheme(text, sensitive, parsePositive(p, 3, 1, 12))
	case 23:
		return findInclusionChains(words, sensitive, parsePositive(p, 3, 2, 20))
	case 24:
		return findLetterTransport(words, sensitive, parsePositive(p, 3, 2, 20))
	case 25:
		return findSciarades(text, sensitive, parsePositive(p, 2, 1, 8))
	case 26:
		return kuntaPhenomena(text, sensitive)
	case 27:
		return groupWordSignatures(words, "Isovocaliche · stesso scheletro vocalico", vowelSkeleton)
	case 28:
		return groupWordSignatures(words, "Isoconsonantiche · stesso scheletro consonantico", consonantSkeleton)
	case 29:
		return groupWordMaterials(words, "Omovocaliche · stesso materiale vocalico", vowelSkeleton)
	case 30:
		return groupWordMaterials(words, "Omoconsonantiche · stesso materiale consonantico", consonantSkeleton)
	case 31:
		n := parsePositive(p, 1, 1, 8)
		return groupWordSignatures(words, fmt.Sprintf("Omovocaliche iniziali · prime %d vocali", n), func(w string) string { return boundarySignature(vowelSkeleton(w), n, false) })
	case 32:
		n := parsePositive(p, 1, 1, 8)
		return groupWordSignatures(words, fmt.Sprintf("Omovocaliche finali · ultime %d vocali", n), func(w string) string { return boundarySignature(vowelSkeleton(w), n, true) })
	case 33:
		n := parsePositive(p, 1, 1, 8)
		return groupWordSignatures(words, fmt.Sprintf("Omoconsonantiche iniziali · prime %d consonanti", n), func(w string) string { return boundarySignature(consonantSkeleton(w), n, false) })
	case 34:
		n := parsePositive(p, 1, 1, 8)
		return groupWordSignatures(words, fmt.Sprintf("Omoconsonantiche finali · ultime %d consonanti", n), func(w string) string { return boundarySignature(consonantSkeleton(w), n, true) })
	case 35:
		if len(words) == 0 {
			return "Nessuna parola."
		}
		sum := 0
		for _, w := range words {
			sum += wordLen(w)
		}
		return fmt.Sprintf("Parole: %d\nLunghezza media: %.2f lettere", len(words), float64(sum)/float64(len(words)))
	case 36:
		m := map[int]int{}
		for _, w := range words {
			n := wordLen(w)
			if n > 0 {
				m[n]++
			}
		}
		ks := []int{}
		for k := range m {
			ks = append(ks, k)
		}
		sort.Ints(ks)
		a := []string{}
		for _, k := range ks {
			a = append(a, fmt.Sprintf("%d lettere = %d", k, m[k]))
		}
		return strings.Join(a, "\n")
	case 37:
		return filterWordsBy(words, p, sensitive, "start")
	case 38:
		return filterWordsBy(words, p, sensitive, "end")
	case 39:
		return filterWordsBy(words, p, sensitive, "contains")
	case 40:
		n := parsePositive(p, 3, 1, 100)
		a := []string{}
		for _, w := range uniqueDisplayWords(words, sensitive) {
			if wordLen(w) == n {
				a = append(a, w)
			}
		}
		return fmt.Sprintf("Parole di %d lettere: %d\n\n%s", n, len(a), strings.Join(a, "\n"))
	case 41:
		return findRepeatedRuns(words)
	case 42:
		return findRepeatedSequences(words, parsePositive(p, 2, 2, 20))
	case 43:
		a := []string{}
		for _, w := range uniqueDisplayWords(words, false) {
			if isIsogram(w) {
				a = append(a, w)
			}
		}
		return fmt.Sprintf("Isogrammi: %d\n\n%s", len(a), strings.Join(a, "\n"))
	case 44:
		a := []string{}
		for _, w := range uniqueDisplayWords(words, false) {
			if alphabeticWord(w, false) {
				a = append(a, w)
			}
		}
		return fmt.Sprintf("Parole alfabetiche: %d\n\n%s", len(a), strings.Join(a, "\n"))
	case 45:
		a := []string{}
		for _, w := range uniqueDisplayWords(words, false) {
			if alphabeticWord(w, true) {
				a = append(a, w)
			}
		}
		return fmt.Sprintf("Parole alfabetiche inverse: %d\n\n%s", len(a), strings.Join(a, "\n"))
	case 46:
		return strings.Join(uniqueDisplayWords(words, sensitive), "\n")
	case 47:
		return strings.Join(words, "\n")
	case 48:
		return strings.ToUpper(text)
	case 49:
		return strings.ToLower(text)
	case 50:
		return findReversePairs(words, sensitive, parsePositive(p, 3, 2, 100), "Palindromi inversi")
	case 51:
		return findContraryPalindromes(words, parsePositive(p, 3, 1, 100))
	case 52:
		return findTransformPairs(words, sensitive, parsePositive(p, 3, 2, 100), "Inversi", transformInverse)
	case 53:
		return findTransformPairs(words, sensitive, parsePositive(p, 3, 2, 100), "Antipodi", transformAntipode)
	case 54:
		return groupByKey(words, "Allitterazioni", func(w string) string {
			r := []rune(cleanWord(w, false))
			if len(r) > 0 {
				return string(r[0])
			}
			return ""
		})
	case 55:
		return groupByKey(words, "Assonanze", func(w string) string { return suffixChars(vowelSkeleton(w), 2) })
	case 56:
		return groupByKey(words, "Consonanze", func(w string) string { return suffixChars(consonantSkeleton(w), 2) })
	case 57:
		return findOxymoronCandidates(text)
	}
	return "Operazione non riconosciuta."
}
func appendUnique(a []string, s string) []string {
	for _, x := range a {
		if x == s {
			return a
		}
	}
	return append(a, s)
}
func extractNumbers(s string) []string {
	out := []string{}
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
