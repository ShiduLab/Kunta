//go:build windows

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unsafe"
)

// Kunta Desktop native: single Windows executable, no browser, no local HTTP server.

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
	{"Omovocaliche", "Stesso materiale vocalico, ordine non rilevante"},
	{"Omoconsonantiche", "Stesso materiale consonantico, ordine non rilevante"},
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
	{"Parole alfabetiche", "Lettere in ordine alfabetico crescente"},
	{"Parole alfabetiche inverse", "Lettere in ordine alfabetico decrescente"},
	{"Elimina duplicati", "Restituisce le parole una sola volta, nell’ordine di apparizione"},
	{"Estrai parole", "Estrae solo le parole dal testo"},
	{"Testo in MAIUSCOLO", "Converte il testo in maiuscolo"},
	{"Testo in minuscolo", "Converte il testo in minuscolo"},
	{"Palindromo inverso", "Inverte ogni parola: Rima → Amir"},
	{"Palindromo contrario", "Rileva forme tipo POSSESSO: prima lettera fissa, resto palindromo"},
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
	a := strings.Split(s, "\n")
	if len(a) <= max {
		return s
	}
	return strings.Join(a[:max], "\n") + fmt.Sprintf("\n\n… output limitato a %d righe.", max)
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
func findAnagrams(words []string, sensitive bool, minLen int) string {
	p := freqWords(words, sensitive)
	g := map[string][]string{}
	for _, k := range p.order {
		if runeLen(k) < minLen {
			continue
		}
		sig := sortedRunes(k)
		g[sig] = append(g[sig], p.display[k])
	}
	groups := [][]string{}
	for _, a := range g {
		if len(a) > 1 {
			sort.Slice(a, func(i, j int) bool { return strings.ToLower(a[i]) < strings.ToLower(a[j]) })
			groups = append(groups, a)
		}
	}
	sort.Slice(groups, func(i, j int) bool { return strings.ToLower(groups[i][0]) < strings.ToLower(groups[j][0]) })
	out := []string{fmt.Sprintf("Gruppi di anagrammi: %d", len(groups)), fmt.Sprintf("Lunghezza minima: %d", minLen), ""}
	for _, a := range groups {
		out = append(out, strings.Join(a, " · "))
	}
	return limitLines(strings.Join(out, "\n"), 700)
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
				out = append(out, fmt.Sprintf("  %d: %s", d, strings.Join(parts, " · ")))
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
		out = append(out, fmt.Sprintf("%s  →  %s", prettySignature(k), strings.Join(groups[k], " · ")))
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
			out = append(out, w+"  →  "+strings.Join(runs, " · "))
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
			out = append(out, w+"  →  "+strings.Join(a, " · "))
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
func alphabeticWord(w string, descending bool) bool {
	a := []rune{}
	for _, r := range cleanWord(w, false) {
		if unicode.IsLetter(r) {
			a = append(a, unicode.ToLower(r))
		}
	}
	if len(a) < 2 {
		return false
	}
	for i := 1; i < len(a); i++ {
		if descending {
			if a[i-1] < a[i] {
				return false
			}
		} else {
			if a[i-1] > a[i] {
				return false
			}
		}
	}
	return true
}
func reverseWordDisplay(w string) string {
	raw := cleanWord(w, true)
	rev := reverseRunes(strings.ToLower(raw))
	r := []rune(raw)
	q := []rune(rev)
	if len(r) > 0 && unicode.IsUpper(r[0]) && len(q) > 0 {
		q[0] = unicode.ToUpper(q[0])
		return string(q)
	}
	return rev
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
		out = append(out, strings.ToUpper(k)+"  →  "+strings.Join(g[k], " · "))
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
		return fmt.Sprintf("Caratteri (Unicode): %d\nByte UTF-8: %d", runeLen(text), len([]byte(text)))
	case 2:
		return fmt.Sprintf("Parole: %d\nParole distinte: %d", len(words), len(freqWords(words, sensitive).freq))
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
		return fmt.Sprintf("Più corta (%d): %s\n\nPiù lunga (%d): %s", min, strings.Join(mins, ", "), max, strings.Join(maxs, ", "))
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
		return groupWordSignatures(words, "Omovocaliche · stesso materiale vocalico (ordine non rilevante)", func(w string) string { return sortedSignature(vowelSkeleton(w)) })
	case 30:
		return groupWordSignatures(words, "Omoconsonantiche · stesso materiale consonantico (ordine non rilevante)", func(w string) string { return sortedSignature(consonantSkeleton(w)) })
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
		return strings.Join(uniqueDisplayWords(words, sensitive), " ")
	case 47:
		return strings.Join(words, "\n")
	case 48:
		return strings.ToUpper(text)
	case 49:
		return strings.ToLower(text)
	case 50:
		a := uniqueDisplayWords(words, true)
		o := []string{fmt.Sprintf("Palindromo inverso · %d trasformazioni", len(a)), ""}
		for _, w := range a {
			o = append(o, w+" → "+reverseWordDisplay(w))
		}
		return strings.Join(o, "\n")
	case 51:
		return findContraryPalindromes(words, parsePositive(p, 3, 1, 100))
	case 52:
		return groupByKey(words, "Allitterazioni", func(w string) string {
			r := []rune(cleanWord(w, false))
			if len(r) > 0 {
				return string(r[0])
			}
			return ""
		})
	case 53:
		return groupByKey(words, "Assonanze", func(w string) string { return suffixChars(vowelSkeleton(w), 2) })
	case 54:
		return groupByKey(words, "Consonanze", func(w string) string { return suffixChars(consonantSkeleton(w), 2) })
	case 55:
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

// ---- Native Win32 GUI ----
var (
    user32   = syscall.NewLazyDLL("user32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")
    comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
    gdi32    = syscall.NewLazyDLL("gdi32.dll")

    procRegisterClassExW     = user32.NewProc("RegisterClassExW")
    procCreateWindowExW      = user32.NewProc("CreateWindowExW")
    procDefWindowProcW       = user32.NewProc("DefWindowProcW")
    procGetMessageW          = user32.NewProc("GetMessageW")
    procTranslateMessage     = user32.NewProc("TranslateMessage")
    procDispatchMessageW     = user32.NewProc("DispatchMessageW")
    procPostQuitMessage      = user32.NewProc("PostQuitMessage")
    procMoveWindow           = user32.NewProc("MoveWindow")
    procSendMessageW         = user32.NewProc("SendMessageW")
    procSetWindowTextW       = user32.NewProc("SetWindowTextW")
    procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
    procGetWindowTextW       = user32.NewProc("GetWindowTextW")
    procGetClientRect        = user32.NewProc("GetClientRect")
    procSetFocus             = user32.NewProc("SetFocus")
    procMessageBoxW          = user32.NewProc("MessageBoxW")
    procLoadCursorW          = user32.NewProc("LoadCursorW")
    procLoadIconW            = user32.NewProc("LoadIconW")
    procBeginPaint           = user32.NewProc("BeginPaint")
    procEndPaint             = user32.NewProc("EndPaint")
    procFillRect             = user32.NewProc("FillRect")
    procFrameRect            = user32.NewProc("FrameRect")
    procDrawTextW            = user32.NewProc("DrawTextW")
    procInvalidateRect       = user32.NewProc("InvalidateRect")

    procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
    procCreateFontW      = gdi32.NewProc("CreateFontW")
    procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
    procSetTextColor     = gdi32.NewProc("SetTextColor")
    procSetBkColor       = gdi32.NewProc("SetBkColor")
    procSetBkMode        = gdi32.NewProc("SetBkMode")
    procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
)

const (
    WS_OVERLAPPEDWINDOW = 0x00CF0000
    WS_VISIBLE          = 0x10000000
    WS_CHILD            = 0x40000000
    WS_TABSTOP          = 0x00010000
    WS_VSCROLL          = 0x00200000
    WS_HSCROLL          = 0x00100000
    WS_BORDER           = 0x00800000
    ES_MULTILINE        = 0x0004
    ES_AUTOVSCROLL      = 0x0040
    ES_AUTOHSCROLL      = 0x0080
    ES_WANTRETURN       = 0x1000
    ES_READONLY         = 0x0800
    BS_OWNERDRAW        = 0x000B
    BS_AUTOCHECKBOX     = 0x0003
    CBS_DROPDOWNLIST    = 0x0003
    SS_LEFT             = 0x0000
    SS_CENTER           = 0x0001

    WM_DESTROY       = 0x0002
    WM_SIZE          = 0x0005
    WM_PAINT         = 0x000F
    WM_COMMAND       = 0x0111
    WM_DRAWITEM      = 0x002B
    WM_SETFONT       = 0x0030
    WM_CTLCOLORMSGBOX = 0x0132
    WM_CTLCOLOREDIT   = 0x0133
    WM_CTLCOLORLISTBOX= 0x0134
    WM_CTLCOLORBTN    = 0x0135
    WM_CTLCOLORDLG    = 0x0136
    WM_CTLCOLORSCROLLBAR = 0x0137
    WM_CTLCOLORSTATIC = 0x0138
    WM_PASTE         = 0x0302
    WM_COPY          = 0x0301
    EM_SETSEL        = 0x00B1
    EM_SETMARGINS    = 0x00D3
    EC_LEFTMARGIN    = 0x0001
    EC_RIGHTMARGIN   = 0x0002
    CB_ADDSTRING     = 0x0143
    CB_SETCURSEL     = 0x014E
    CB_GETCURSEL     = 0x0147
    BM_GETCHECK      = 0x00F0
    BST_CHECKED      = 1
    SW_SHOW          = 5
    CW_USEDEFAULT    = 0x80000000

    ODS_SELECTED = 0x0001
    DT_CENTER = 0x0001
    DT_VCENTER = 0x0004
    DT_SINGLELINE = 0x0020
    DT_END_ELLIPSIS = 0x8000
    TRANSPARENT = 1

    ID_TITLE            = 90
    ID_TAGLINE          = 91
    ID_EDIT_INPUT       = 101
    ID_PASTE            = 102
    ID_OPEN             = 103
    ID_CLEAR            = 104
    ID_QUICK_AZ         = 105
    ID_QUICK_REP        = 106
    ID_QUICK_KUNTA      = 107
    ID_QUICK_VOW        = 108
    ID_QUICK_CONS       = 109
    ID_COMBO            = 110
    ID_PARAM            = 111
    ID_CASE             = 112
    ID_RUN              = 113
    ID_HINT             = 114
    ID_OUTPUT           = 115
    ID_COPY             = 116
    ID_STATUS           = 117
    ID_LABEL_OPERATION  = 118
    ID_LABEL_PARAM      = 119
    ID_FOOTER           = 120
)

type POINT struct{ X, Y int32 }
type MSG struct {
    Hwnd           uintptr
    Message        uint32
    WParam, LParam uintptr
    Time           uint32
    Pt             POINT
}
type RECT struct{ Left, Top, Right, Bottom int32 }
type WNDCLASSEX struct {
    CbSize                                   uint32
    Style                                    uint32
    LpfnWndProc                              uintptr
    CbClsExtra, CbWndExtra                   int32
    HInstance, HIcon, HCursor, HbrBackground uintptr
    LpszMenuName, LpszClassName              *uint16
    HIconSm                                  uintptr
}
type OPENFILENAME struct {
    LStructSize                        uint32
    HwndOwner, HInstance               uintptr
    LpstrFilter, LpstrCustomFilter     *uint16
    NMaxCustFilter, NFilterIndex       uint32
    LpstrFile                          *uint16
    NMaxFile                           uint32
    LpstrFileTitle                     *uint16
    NMaxFileTitle                      uint32
    LpstrInitialDir, LpstrTitle        *uint16
    Flags, NFileOffset, NFileExtension uint32
    LpstrDefExt                        *uint16
    LCustData, LpfnHook                uintptr
    LpTemplateName                     *uint16
    PvReserved                         uintptr
    DwReserved, FlagsEx                uint32
}
type PAINTSTRUCT struct {
    Hdc         uintptr
    FErase      int32
    RcPaint     RECT
    FRestore    int32
    FIncUpdate  int32
    RgbReserved [32]byte
}
type DRAWITEMSTRUCT struct {
    CtlType    uint32
    CtlID      uint32
    ItemID     uint32
    ItemAction uint32
    ItemState  uint32
    HwndItem   uintptr
    HDC        uintptr
    RcItem     RECT
    ItemData   uintptr
}

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }

var (
    hwndMain uintptr
    hTitle, hTagline, hInput, hParam, hCombo, hCase, hHint, hOutput, hStatus, hFooter uintptr
    controlsByID = map[int]uintptr{}

    fontUI, fontSmall, fontTitle, fontMono, fontBold uintptr
    brushBG, brushPanel, brushInput, brushAccent, brushSoft, brushLine uintptr

    colBG     = rgb(16, 20, 15)
    colPanel  = rgb(23, 28, 22)
    colInput  = rgb(18, 24, 18)
    colText   = rgb(238, 242, 235)
    colMuted  = rgb(174, 184, 170)
    colAccent = rgb(124, 186, 47)
    colAccentInk = rgb(13, 22, 6)
    colSoft   = rgb(32, 42, 29)
    colLine   = rgb(57, 67, 55)
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func textOf(h uintptr) string {
    n, _, _ := procGetWindowTextLengthW.Call(h)
    buf := make([]uint16, n+1)
    procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
    return syscall.UTF16ToString(buf)
}
func setText(h uintptr, s string) { procSetWindowTextW.Call(h, uintptr(unsafe.Pointer(u16(s)))) }
func move(h uintptr, x, y, w, hh int) {
    if h != 0 { procMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1) }
}
func message(s string) {
    procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(u16(s))), uintptr(unsafe.Pointer(u16("Kunta"))), 0x40)
}
func fontForID(id int) uintptr {
    switch id {
    case ID_TITLE:
        return fontTitle
    case ID_TAGLINE, ID_HINT, ID_STATUS, ID_LABEL_OPERATION, ID_LABEL_PARAM, ID_FOOTER:
        return fontSmall
    case ID_EDIT_INPUT, ID_OUTPUT:
        return fontMono
    case ID_RUN:
        return fontBold
    default:
        return fontUI
    }
}
func create(class, text string, style, ex uint32, id int) uintptr {
    r, _, _ := procCreateWindowExW.Call(uintptr(ex), uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))), uintptr(style), 0, 0, 10, 10, hwndMain, uintptr(id), 0, 0)
    if r != 0 {
        controlsByID[id] = r
        f := fontForID(id)
        if f != 0 { procSendMessageW.Call(r, WM_SETFONT, f, 1) }
    }
    return r
}
func addButton(text string, id int) uintptr {
    return create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, 0, id)
}
func addStatic(text string, id int) uintptr {
    return create("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, 0, id)
}
func openTextFile() {
    buf := make([]uint16, 32768)
    filter := syscall.StringToUTF16("Testi (*.txt;*.md;*.log;*.csv;*.tsv)\x00*.txt;*.md;*.log;*.csv;*.tsv\x00Tutti i file (*.*)\x00*.*\x00\x00")
    ofn := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: hwndMain, LpstrFilter: &filter[0], LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), Flags: 0x00001000 | 0x00000800 | 0x00000008}
    r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
    if r == 0 { return }
    path := syscall.UTF16ToString(buf)
    b, err := os.ReadFile(path)
    if err != nil { message("Impossibile leggere il file."); return }
    s := strings.TrimPrefix(string(b), "\ufeff")
    setText(hInput, s)
    setText(hStatus, "Testo caricato.")
}
func updateHint() {
    i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0)
    if int(i) >= 0 && int(i) < len(operations) { setText(hHint, operations[int(i)].Hint) }
}
func runOp(op int) {
    txt := textOf(hInput)
    if txt == "" { setText(hOutput, "Non c’è testo da analizzare."); return }
    checked, _, _ := procSendMessageW.Call(hCase, BM_GETCHECK, 0, 0)
    setText(hOutput, analyze(txt, op, textOf(hParam), checked == BST_CHECKED))
    setText(hStatus, "Locale · nessun invio esterno")
}
func layout() {
    var r RECT
    procGetClientRect.Call(hwndMain, uintptr(unsafe.Pointer(&r)))
    w, h := int(r.Right-r.Left), int(r.Bottom-r.Top)
    if w < 760 { w = 760 }
    if h < 600 { h = 600 }

    pad, gap := 16, 9
    y := 14
    move(hTitle, pad, y, 260, 32)
    move(hTagline, pad, y+31, 360, 22)
    move(hFooter, w-190, y+5, 170, 32)
    y += 64

    btnH := 38
    move(controlsByID[ID_PASTE], pad, y, 105, btnH)
    move(controlsByID[ID_OPEN], pad+114, y, 130, btnH)
    move(controlsByID[ID_CLEAR], pad+253, y, 105, btnH)
    y += btnH + gap

    reservedBottom := 272
    inputH := (h - y - reservedBottom) * 46 / 100
    if inputH < 125 { inputH = 125 }
    move(hInput, pad, y, w-2*pad, inputH)
    y += inputH + gap

    quickW := (w - 2*pad - 4*gap) / 5
    qs := []int{ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS}
    x := pad
    for _, id := range qs { move(controlsByID[id], x, y, quickW, btnH); x += quickW + gap }
    y += btnH + gap

    panelTop := y
    labelH := 18
    comboW := w*46/100
    paramW := w*18/100
    rightW := w - 2*pad - comboW - paramW - 3*gap

    move(controlsByID[ID_LABEL_OPERATION], pad+10, y+7, comboW-20, labelH)
    move(controlsByID[ID_LABEL_PARAM], pad+comboW+gap+10, y+7, paramW-20, labelH)
    y += 28
    move(hCombo, pad+10, y, comboW-20, 240)
    move(hParam, pad+comboW+gap+10, y, paramW-20, 36)
    move(hCase, pad+comboW+paramW+2*gap, y, rightW-112, 36)
    move(controlsByID[ID_RUN], w-pad-108, y, 108, 36)
    y += 41
    move(hHint, pad+10, y, w-2*pad-20, 22)
    y += 29
    panelBottom := y
    _ = panelTop
    _ = panelBottom

    outH := h - y - 62
    if outH < 110 { outH = 110 }
    move(hOutput, pad, y, w-2*pad, outH)
    y += outH + gap
    move(controlsByID[ID_COPY], pad, y, 145, 36)
    move(hStatus, pad+158, y+8, w-pad-158, 24)
}
func buttonStyle(id uint32) (bg, fg uintptr) {
    switch id {
    case ID_RUN:
        return colAccent, colAccentInk
    case ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS:
        return colSoft, colText
    default:
        return colPanel, colText
    }
}
func drawOwnerButton(di *DRAWITEMSTRUCT) {
    bg, fg := buttonStyle(di.CtlID)
    if di.ItemState&ODS_SELECTED != 0 {
        if di.CtlID == ID_RUN { bg = rgb(103, 160, 37) } else { bg = rgb(43, 53, 40) }
    }
    b, _, _ := procCreateSolidBrush.Call(bg)
    procFillRect.Call(di.HDC, uintptr(unsafe.Pointer(&di.RcItem)), b)
    procFrameRect.Call(di.HDC, uintptr(unsafe.Pointer(&di.RcItem)), brushLine)
    procSetBkMode.Call(di.HDC, TRANSPARENT)
    procSetTextColor.Call(di.HDC, fg)
    txt := textOf(di.HwndItem)
    rr := di.RcItem
    procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16(txt))), ^uintptr(0), uintptr(unsafe.Pointer(&rr)), DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}
func paintBackground(hwnd uintptr) {
    var ps PAINTSTRUCT
    hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
    if hdc != 0 {
        var r RECT
        procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
        procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brushBG)
    }
    procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
    switch msg {
    case WM_DESTROY:
        procPostQuitMessage.Call(0); return 0
    case WM_SIZE:
        layout(); return 0
    case WM_PAINT:
        paintBackground(hwnd); return 0
    case WM_DRAWITEM:
        if lParam != 0 {
            di := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
            drawOwnerButton(di)
            return 1
        }
    case WM_CTLCOLOREDIT:
        hdc := wParam
        procSetTextColor.Call(hdc, colText)
        procSetBkColor.Call(hdc, colInput)
        return brushInput
    case WM_CTLCOLORLISTBOX:
        hdc := wParam
        procSetTextColor.Call(hdc, colText)
        procSetBkColor.Call(hdc, colPanel)
        return brushPanel
    case WM_CTLCOLORSTATIC:
        hdc := wParam
        procSetBkMode.Call(hdc, TRANSPARENT)
        ctl := lParam
        if ctl == hTitle { procSetTextColor.Call(hdc, colText) } else { procSetTextColor.Call(hdc, colMuted) }
        return brushBG
    case WM_CTLCOLORBTN:
        hdc := wParam
        procSetTextColor.Call(hdc, colText)
        procSetBkColor.Call(hdc, colBG)
        return brushBG
    case WM_COMMAND:
        id := int(wParam & 0xffff)
        code := int((wParam >> 16) & 0xffff)
        switch id {
        case ID_PASTE:
            procSetFocus.Call(hInput); procSendMessageW.Call(hInput, WM_PASTE, 0, 0)
        case ID_OPEN:
            openTextFile()
        case ID_CLEAR:
            setText(hInput, ""); setText(hOutput, ""); setText(hParam, ""); procSetFocus.Call(hInput)
        case ID_QUICK_AZ:
            procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0); updateHint(); runOp(20)
        case ID_QUICK_REP:
            procSendMessageW.Call(hCombo, CB_SETCURSEL, 9, 0); updateHint(); runOp(9)
        case ID_QUICK_KUNTA:
            procSendMessageW.Call(hCombo, CB_SETCURSEL, 26, 0); updateHint(); runOp(26)
        case ID_QUICK_VOW:
            procSendMessageW.Call(hCombo, CB_SETCURSEL, 27, 0); updateHint(); runOp(27)
        case ID_QUICK_CONS:
            procSendMessageW.Call(hCombo, CB_SETCURSEL, 28, 0); updateHint(); runOp(28)
        case ID_COMBO:
            if code == 1 { updateHint() }
        case ID_RUN:
            i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0); runOp(int(i))
        case ID_COPY:
            procSendMessageW.Call(hOutput, EM_SETSEL, 0, ^uintptr(0)); procSendMessageW.Call(hOutput, WM_COPY, 0, 0); setText(hStatus, "Risultato copiato.")
        }
        return 0
    }
    r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
    return r
}
func main() {
    hInst, _, _ := procGetModuleHandleW.Call(0)
    cur, _, _ := procLoadCursorW.Call(0, 32512)
    appIcon, _, _ := procLoadIconW.Call(hInst, 1)

    brushBG, _, _ = procCreateSolidBrush.Call(colBG)
    brushPanel, _, _ = procCreateSolidBrush.Call(colPanel)
    brushInput, _, _ = procCreateSolidBrush.Call(colInput)
    brushAccent, _, _ = procCreateSolidBrush.Call(colAccent)
    brushSoft, _, _ = procCreateSolidBrush.Call(colSoft)
    brushLine, _, _ = procCreateSolidBrush.Call(colLine)

    fontUI, _, _ = procCreateFontW.Call(^uintptr(17), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))
    fontSmall, _, _ = procCreateFontW.Call(^uintptr(14), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))
    fontTitle, _, _ = procCreateFontW.Call(^uintptr(28), 0, 0, 0, 600, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))
    fontMono, _, _ = procCreateFontW.Call(^uintptr(15), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Consolas"))))
    fontBold, _, _ = procCreateFontW.Call(^uintptr(17), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))

    class := u16("KuntaNativeWindow")
    wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HIcon: appIcon, HCursor: cur, HbrBackground: brushBG, LpszClassName: class, HIconSm: appIcon}
    if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 { return }

    hwndMain, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(u16("Kunta"))), WS_OVERLAPPEDWINDOW|WS_VISIBLE, CW_USEDEFAULT, CW_USEDEFAULT, 1100, 820, 0, 0, hInst, 0)
    if hwndMain == 0 { return }

    hTitle = addStatic("Kunta", ID_TITLE)
    hTagline = addStatic("Conta · osserva · ordina · rac-conta", ID_TAGLINE)
    hFooter = addStatic("ShiduLab", ID_FOOTER)

    addButton("Incolla", ID_PASTE)
    addButton("Apri testo…", ID_OPEN)
    addButton("Pulisci", ID_CLEAR)

    hInput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN, 0, ID_EDIT_INPUT)
    procSendMessageW.Call(hInput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(10|(10<<16)))

    addButton("A–Z", ID_QUICK_AZ)
    addButton("Ripetizioni", ID_QUICK_REP)
    addButton("Kunta il testo", ID_QUICK_KUNTA)
    addButton("Isovocaliche", ID_QUICK_VOW)
    addButton("Isoconsonantiche", ID_QUICK_CONS)

    create("STATIC", "Operazione", WS_CHILD|WS_VISIBLE|SS_LEFT, 0, ID_LABEL_OPERATION)
    create("STATIC", "Parametro", WS_CHILD|WS_VISIBLE|SS_LEFT, 0, ID_LABEL_PARAM)
    hCombo = create("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST, 0, ID_COMBO)
    for _, op := range operations { procSendMessageW.Call(hCombo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(op.Name)))) }
    procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0)

    hParam = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL, 0, ID_PARAM)
    procSendMessageW.Call(hParam, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(8|(8<<16)))
    hCase = create("BUTTON", "Maiuscole/minuscole distinte", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 0, ID_CASE)
    addButton("KUNTA!", ID_RUN)
    hHint = addStatic("", ID_HINT)
    updateHint()

    hOutput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN|ES_READONLY, 0, ID_OUTPUT)
    procSendMessageW.Call(hOutput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(10|(10<<16)))
    addButton("Copia risultato", ID_COPY)
    hStatus = addStatic("Locale · nessun invio esterno", ID_STATUS)

    layout()
    procInvalidateRect.Call(hwndMain, 0, 1)

    var msg MSG
    for {
        r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
        if int32(r) <= 0 { break }
        procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
        procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
    }
}
