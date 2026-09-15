//go:build windows

package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode"
	"unicode/utf16"
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
		out = append(out, "• "+strings.Join(a, "\n• "))
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
		return strings.Join(uniqueDisplayWords(words, sensitive), "\n")
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

// ---- Native Win32 GUI v3: single EXE, async engine, no browser/server ----
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procPostMessageW         = user32.NewProc("PostMessageW")
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
	procDrawIconEx           = user32.NewProc("DrawIconEx")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procRoundRect        = gdi32.NewProc("RoundRect")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkColor       = gdi32.NewProc("SetBkColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procStretchDIBits    = gdi32.NewProc("StretchDIBits")

	procCoInitializeEx        = ole32.NewProc("CoInitializeEx")
	procCoUninitialize        = ole32.NewProc("CoUninitialize")
	procCoCreateInstance      = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree         = ole32.NewProc("CoTaskMemFree")
	procSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	procDragAcceptFiles       = shell32.NewProc("DragAcceptFiles")
	procDragQueryFileW        = shell32.NewProc("DragQueryFileW")
	procDragFinish            = shell32.NewProc("DragFinish")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_BORDER           = 0x00800000
	WS_CLIPCHILDREN     = 0x02000000
	WS_EX_CLIENTEDGE    = 0x00000200
	ES_MULTILINE        = 0x0004
	ES_AUTOVSCROLL      = 0x0040
	ES_AUTOHSCROLL      = 0x0080
	ES_WANTRETURN       = 0x1000
	ES_READONLY         = 0x0800
	ES_NOHIDESEL        = 0x0100
	BS_OWNERDRAW        = 0x000B
	BS_AUTOCHECKBOX     = 0x0003
	CBS_DROPDOWNLIST    = 0x0003
	CBS_HASSTRINGS      = 0x0200
	SS_LEFT             = 0x0000

	WM_DESTROY           = 0x0002
	WM_SIZE              = 0x0005
	WM_PAINT             = 0x000F
	WM_ERASEBKGND        = 0x0014
	WM_GETMINMAXINFO     = 0x0024
	WM_COMMAND           = 0x0111
	WM_DRAWITEM          = 0x002B
	WM_MEASUREITEM       = 0x002C
	WM_SETFONT           = 0x0030
	WM_CTLCOLORMSGBOX    = 0x0132
	WM_CTLCOLOREDIT      = 0x0133
	WM_CTLCOLORLISTBOX   = 0x0134
	WM_CTLCOLORBTN       = 0x0135
	WM_CTLCOLORDLG       = 0x0136
	WM_CTLCOLORSCROLLBAR = 0x0137
	WM_CTLCOLORSTATIC    = 0x0138
	WM_DROPFILES         = 0x0233
	WM_PASTE             = 0x0302
	WM_COPY              = 0x0301
	WM_SETREDRAW         = 0x000B
	WM_APP_ANALYSIS_DONE = 0x8001

	EM_SETSEL        = 0x00B1
	EM_SETLIMITTEXT  = 0x00C5
	EM_SETMARGINS    = 0x00D3
	EC_LEFTMARGIN    = 0x0001
	EC_RIGHTMARGIN   = 0x0002
	CB_ADDSTRING     = 0x0143
	CB_GETCOUNT      = 0x0146
	CB_GETLBTEXT     = 0x0148
	CB_GETLBTEXTLEN  = 0x0149
	CB_GETCURSEL     = 0x0147
	CB_SETCURSEL     = 0x014E
	CB_SETITEMHEIGHT = 0x0153
	CB_SETMINVISIBLE = 0x1701
	BM_GETCHECK      = 0x00F0
	BST_CHECKED      = 1
	CW_USEDEFAULT    = 0x80000000

	ODT_BUTTON       = 4
	ODT_COMBOBOX     = 3
	ODS_SELECTED     = 0x0001
	ODS_FOCUS        = 0x0010
	ODS_COMBOBOXEDIT = 0x1000
	DT_LEFT          = 0x0000
	DT_CENTER        = 0x0001
	DT_VCENTER       = 0x0004
	DT_SINGLELINE    = 0x0020
	DT_END_ELLIPSIS  = 0x8000
	TRANSPARENT      = 1
	PS_SOLID         = 0
	DI_NORMAL        = 0x0003

	ID_TITLE           = 90
	ID_TAGLINE         = 91
	ID_EDIT_INPUT      = 101
	ID_PASTE           = 102
	ID_OPEN            = 103
	ID_CLEAR           = 104
	ID_QUICK_AZ        = 105
	ID_QUICK_REP       = 106
	ID_QUICK_KUNTA     = 107
	ID_QUICK_VOW       = 108
	ID_QUICK_CONS      = 109
	ID_COMBO           = 110
	ID_PARAM           = 111
	ID_CASE            = 112
	ID_RUN             = 113
	ID_HINT            = 114
	ID_OUTPUT          = 115
	ID_COPY            = 116
	ID_STATUS          = 117
	ID_LABEL_OPERATION = 118
	ID_LABEL_PARAM     = 119
	ID_FOOTER          = 120
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
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}
type COMDLG_FILTERSPEC struct {
	Name *uint16
	Spec *uint16
}
type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}
type BITMAPINFO struct {
	Header BITMAPINFOHEADER
	Colors [1]uint32
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
type MEASUREITEMSTRUCT struct {
	CtlType, CtlID, ItemID, ItemWidth, ItemHeight uint32
	ItemData                                      uintptr
}
type MINMAXINFO struct {
	PtReserved, PtMaxSize, PtMaxPosition, PtMinTrackSize, PtMaxTrackSize POINT
}

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }

var (
	hwndMain                                                                          uintptr
	hTitle, hTagline, hInput, hParam, hCombo, hCase, hHint, hOutput, hStatus, hFooter uintptr
	appIcon                                                                           uintptr
	controlsByID                                                                      = map[int]uintptr{}
	panelStatics                                                                      = map[uintptr]bool{}

	fontUI, fontSmall, fontTitle, fontMono, fontBold                   uintptr
	brushBG, brushPanel, brushInput, brushAccent, brushSoft, brushLine uintptr

	colBG        = rgb(16, 20, 15)
	colPanel     = rgb(23, 28, 22)
	colInput     = rgb(18, 24, 18)
	colText      = rgb(238, 242, 235)
	colMuted     = rgb(174, 184, 170)
	colAccent    = rgb(124, 186, 47)
	colAccentInk = rgb(13, 22, 6)
	colSoft      = rgb(32, 42, 29)
	colLine      = rgb(57, 67, 55)

	controlPanel  RECT
	currentDPI    = 96
	busyMu        sync.Mutex
	busy          bool
	resultMu      sync.Mutex
	pendingResult string

	botoloPixels []byte
	botoloWidth  int
	botoloHeight int
	botoloStride int
)

// Botolo + ShiduLab è incorporato direttamente nel sorgente: nessun file esterno a runtime.
const botoloBMPBase64 = `Qk2OaAAAAAAAADYAAAAoAAAAVAAAAGoAAAABABgAAAAAAFhoAADEDgAAxA4AAAAAAAAAAAAADxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEBUREhcTEhcTEhcTEhcTEhcTEBURDxQQERYSEhcTEhcTEhcTEBURERYSEhcTEhcTEhcTEBUREhcTEhcTEhcTERYSDxQQEBUREhcTEhcTEBURERYSEhgTERYSDxQQDxQQERYSEhcTEhcTEBUREhcTEhcTEBURERYSEhcTEhcTEhcTEhcTEhcTEhcTEhcTDxQQERYSEhgTEhcTEBUREhcTExgUEBURERYSERYSERYSEhcTEhcTERYSDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDBENAwgEAAUBAAUBAAUBAwgEDhMPERcTBwsHAAQAAgcDAAQACg8LBQoGAAUBAgYDAQYCCA0JAAUBAgcDAAQBBgoHEhcTCg8LAAUBAAQACg8LBgsHAAQBBQoGDRIOEBURBQoGAAQBAgYDDhMPAwgEAQYCCQ4KBgsHAAQAAgYDAQYCAAUBAAUBAAUBAwgEEBURBwsIAAQAAQYCDhMPBAkFAAQACA0JBgsHBgsHBwwIAAMAAAQABgsHERYSDxQQ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DxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQERYSEhcTDhMPBQoGAgcDAQUCAQUCAgcDCQ4KEhcTEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEBURERYSEhcTEhcTEhcTEhcTEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQ`

var botoloBMP = func() []byte {
	b, err := base64.StdEncoding.DecodeString(botoloBMPBase64)
	if err != nil {
		return nil
	}
	return b
}()

var (
	clsidFileOpenDialog = GUID{0xDC1C5A9C, 0xE88A, 0x4DDE, [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	iidIFileOpenDialog  = GUID{0xD57C7288, 0xD4AD, 0x4768, [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func textOf(h uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(h)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, int(n)+1)
	procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func windowsText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}
func setText(h uintptr, s string) {
	procSetWindowTextW.Call(h, uintptr(unsafe.Pointer(u16(windowsText(s)))))
}
func replaceOutput(s string) {
	if hOutput == 0 {
		return
	}
	// Replace atomically and force an erase/repaint so old glyphs can never remain under the new result.
	procSendMessageW.Call(hOutput, WM_SETREDRAW, 0, 0)
	procSetWindowTextW.Call(hOutput, uintptr(unsafe.Pointer(u16(""))))
	procSetWindowTextW.Call(hOutput, uintptr(unsafe.Pointer(u16(windowsText(s)))))
	procSendMessageW.Call(hOutput, EM_SETSEL, 0, 0)
	procSendMessageW.Call(hOutput, WM_SETREDRAW, 1, 0)
	procInvalidateRect.Call(hOutput, 0, 1)
	procUpdateWindow.Call(hOutput)
}
func move(h uintptr, x, y, w, hh int) {
	if h != 0 && w > 0 && hh > 0 {
		procMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1)
	}
}
func message(s string) {
	procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(u16(s))), uintptr(unsafe.Pointer(u16("Kunta"))), 0x40)
}
func scale(v int) int { return v * currentDPI / 96 }
func setDarkTheme(h uintptr) {
	if h == 0 {
		return
	}
	procSetWindowTheme.Call(h, uintptr(unsafe.Pointer(u16("DarkMode_Explorer"))), 0)
}
func setImmersiveDarkTitlebar(hwnd uintptr) {
	on := int32(1)
	// 20 is DWMWA_USE_IMMERSIVE_DARK_MODE on current Windows 10/11; 19 on older builds.
	if r, _, _ := procDwmSetWindowAttribute.Call(hwnd, 20, uintptr(unsafe.Pointer(&on)), unsafe.Sizeof(on)); r != 0 {
		procDwmSetWindowAttribute.Call(hwnd, 19, uintptr(unsafe.Pointer(&on)), unsafe.Sizeof(on))
	}
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
		if f := fontForID(id); f != 0 {
			procSendMessageW.Call(r, WM_SETFONT, f, 1)
		}
		setDarkTheme(r)
	}
	return r
}
func addButton(text string, id int) uintptr {
	return create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, 0, id)
}
func addStatic(text string, id int, panel bool) uintptr {
	h := create("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, 0, id)
	if panel {
		panelStatics[h] = true
	}
	return h
}

func decodeTextBytes(b []byte) string {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return string(b[3:])
	}
	if len(b) >= 2 && ((b[0] == 0xFF && b[1] == 0xFE) || (b[0] == 0xFE && b[1] == 0xFF)) {
		le := b[0] == 0xFF
		b = b[2:]
		u := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			if le {
				u = append(u, binary.LittleEndian.Uint16(b[i:i+2]))
			} else {
				u = append(u, binary.BigEndian.Uint16(b[i:i+2]))
			}
		}
		return string(utf16.Decode(u))
	}
	return string(b)
}
func loadTextPath(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		message("Impossibile leggere il file.")
		return
	}
	setText(hInput, decodeTextBytes(b))
	setText(hStatus, fmt.Sprintf("Testo caricato · %.1f KB", float64(len(b))/1024.0))
	procSetFocus.Call(hInput)
}
func hresultFailed(hr uintptr) bool { return int32(uint32(hr)) < 0 }
func comCall(obj uintptr, method int, args ...uintptr) uintptr {
	if obj == 0 {
		return uintptr(0x80004003)
	}
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(method)*unsafe.Sizeof(uintptr(0))))
	callArgs := make([]uintptr, 0, len(args)+1)
	callArgs = append(callArgs, obj)
	callArgs = append(callArgs, args...)
	r, _, _ := syscall.SyscallN(fn, callArgs...)
	return r
}
func utf16PtrToString(p uintptr) string {
	if p == 0 {
		return ""
	}
	a := (*[1 << 28]uint16)(unsafe.Pointer(p))
	n := 0
	for n < len(a) && a[n] != 0 {
		n++
	}
	return syscall.UTF16ToString(a[:n])
}
func openTextFile() {
	var dlg uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
		uintptr(unsafe.Pointer(&dlg)),
	)
	if hresultFailed(hr) || dlg == 0 {
		message(fmt.Sprintf("Impossibile aprire il selettore file.\nHRESULT: 0x%08X", uint32(hr)))
		return
	}
	defer comCall(dlg, 2) // IUnknown::Release

	n1, s1 := syscall.StringToUTF16("Testi"), syscall.StringToUTF16("*.txt;*.md;*.log;*.csv;*.tsv")
	n2, s2 := syscall.StringToUTF16("Tutti i file"), syscall.StringToUTF16("*.*")
	filters := []COMDLG_FILTERSPEC{{&n1[0], &s1[0]}, {&n2[0], &s2[0]}}
	comCall(dlg, 4, uintptr(len(filters)), uintptr(unsafe.Pointer(&filters[0]))) // SetFileTypes
	comCall(dlg, 9, uintptr(0x40|0x800|0x1000))                                  // FOS_FORCEFILESYSTEM | PATHMUSTEXIST | FILEMUSTEXIST
	title := syscall.StringToUTF16("Apri testo in Kunta")
	comCall(dlg, 17, uintptr(unsafe.Pointer(&title[0]))) // SetTitle

	hr = comCall(dlg, 3, hwndMain) // IModalWindow::Show
	if hresultFailed(hr) {
		// 0x800704C7 = annullato dall'utente: non è un errore.
		if uint32(hr) != 0x800704C7 {
			message(fmt.Sprintf("Il selettore file non si è aperto.\nHRESULT: 0x%08X", uint32(hr)))
		}
		return
	}

	var item uintptr
	hr = comCall(dlg, 20, uintptr(unsafe.Pointer(&item))) // IFileDialog::GetResult
	if hresultFailed(hr) || item == 0 {
		return
	}
	defer comCall(item, 2)

	var pathPtr uintptr
	hr = comCall(item, 5, uintptr(0x80058000), uintptr(unsafe.Pointer(&pathPtr))) // IShellItem::GetDisplayName(SIGDN_FILESYSPATH)
	if hresultFailed(hr) || pathPtr == 0 {
		return
	}
	path := utf16PtrToString(pathPtr)
	procCoTaskMemFree.Call(pathPtr)
	if path != "" {
		loadTextPath(path)
	}
}
func handleDrop(hDrop uintptr) {
	n, _, _ := procDragQueryFileW.Call(hDrop, 0xFFFFFFFF, 0, 0)
	if n > 0 {
		ln, _, _ := procDragQueryFileW.Call(hDrop, 0, 0, 0)
		buf := make([]uint16, int(ln)+1)
		procDragQueryFileW.Call(hDrop, 0, uintptr(unsafe.Pointer(&buf[0])), ln+1)
		loadTextPath(syscall.UTF16ToString(buf))
	}
	procDragFinish.Call(hDrop)
}
func updateHint() {
	i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0)
	if int(i) >= 0 && int(i) < len(operations) {
		setText(hHint, operations[int(i)].Hint)
	}
}
func setBusy(v bool) {
	busyMu.Lock()
	busy = v
	busyMu.Unlock()
	ids := []int{ID_RUN, ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS}
	enabled := uintptr(1)
	if v {
		enabled = 0
	}
	for _, id := range ids {
		procEnableWindow.Call(controlsByID[id], enabled)
	}
	procEnableWindow.Call(hCombo, enabled)
	if v {
		setText(hStatus, "Kunta sta lavorando…")
	} else {
		setText(hStatus, fmt.Sprintf("Locale · %d operazioni · nessun invio esterno", len(operations)))
	}
}
func runOp(op int) {
	busyMu.Lock()
	already := busy
	busyMu.Unlock()
	if already {
		return
	}
	txt := textOf(hInput)
	if txt == "" {
		replaceOutput("Non c’è testo da analizzare.")
		return
	}
	param := textOf(hParam)
	checked, _, _ := procSendMessageW.Call(hCase, BM_GETCHECK, 0, 0)
	sensitive := checked == BST_CHECKED
	replaceOutput("")
	setBusy(true)
	go func() {
		res := analyze(txt, op, param, sensitive)
		resultMu.Lock()
		pendingResult = res
		resultMu.Unlock()
		procPostMessageW.Call(hwndMain, WM_APP_ANALYSIS_DONE, 0, 0)
	}()
}

func layout() {
	var r RECT
	procGetClientRect.Call(hwndMain, uintptr(unsafe.Pointer(&r)))
	w, h := int(r.Right-r.Left), int(r.Bottom-r.Top)
	pad, gap := scale(14), scale(9)
	if w < scale(820) {
		w = scale(820)
	}
	if h < scale(680) {
		h = scale(680)
	}

	y := scale(13)
	iconW := scale(50)
	move(hTitle, pad+iconW+scale(10), y+scale(2), scale(320), scale(31))
	move(hTagline, pad+iconW+scale(10), y+scale(31), scale(440), scale(20))
	y += scale(62)

	btnH := scale(40)
	move(controlsByID[ID_PASTE], pad, y, scale(105), btnH)
	move(controlsByID[ID_OPEN], pad+scale(114), y, scale(138), btnH)
	move(controlsByID[ID_CLEAR], pad+scale(261), y, scale(105), btnH)
	y += btnH + gap

	bottomReserve := scale(385)
	inputH := (h - y - bottomReserve) * 46 / 100
	if inputH < scale(135) {
		inputH = scale(135)
	}
	move(hInput, pad, y, w-2*pad, inputH)
	y += inputH + gap

	quickW := (w - 2*pad - 4*gap) / 5
	qs := []int{ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS}
	x := pad
	for _, id := range qs {
		move(controlsByID[id], x, y, quickW, btnH)
		x += quickW + gap
	}
	y += btnH + gap

	panelX, panelY := pad, y
	panelW, panelH := w-2*pad, scale(112)
	controlPanel = RECT{int32(panelX), int32(panelY), int32(panelX + panelW), int32(panelY + panelH)}
	inner := scale(10)
	labelH := scale(18)
	comboW := panelW * 46 / 100
	paramW := panelW * 18 / 100
	rightX := panelX + comboW + paramW + 2*gap
	rightW := panelX + panelW - inner - rightX

	move(controlsByID[ID_LABEL_OPERATION], panelX+inner, panelY+scale(8), comboW-inner*2, labelH)
	move(controlsByID[ID_LABEL_PARAM], panelX+comboW+gap+inner, panelY+scale(8), paramW-inner*2, labelH)
	fieldY := panelY + scale(31)
	move(hCombo, panelX+inner, fieldY, comboW-inner*2, scale(620))
	move(hParam, panelX+comboW+gap+inner, fieldY, paramW-inner*2, scale(36))
	runW := scale(105)
	caseW := rightW - runW - gap
	if caseW < scale(190) {
		caseW = scale(190)
	}
	move(hCase, rightX, fieldY, caseW, scale(36))
	move(controlsByID[ID_RUN], panelX+panelW-inner-runW, fieldY, runW, scale(36))
	move(hHint, panelX+inner, panelY+scale(75), panelW-inner*2, scale(25))

	y = panelY + panelH + gap
	footerH := scale(118)
	outH := h - y - footerH - pad
	if outH < scale(125) {
		outH = scale(125)
	}
	move(hOutput, pad, y, w-2*pad, outH)
	y += outH + gap
	move(controlsByID[ID_COPY], pad, y, scale(155), scale(36))
	move(hStatus, pad+scale(169), y+scale(8), w-pad*2-scale(330), scale(24))
	move(hFooter, w-pad-scale(150), y+scale(8), scale(45), scale(24))
	procInvalidateRect.Call(hwndMain, 0, 0)
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
func drawRoundedBox(hdc uintptr, rc RECT, bg, border uintptr, radius int) {
	b, _, _ := procCreateSolidBrush.Call(bg)
	p, _, _ := procCreatePen.Call(PS_SOLID, 1, border)
	oldB, _, _ := procSelectObject.Call(hdc, b)
	oldP, _, _ := procSelectObject.Call(hdc, p)
	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), uintptr(radius), uintptr(radius))
	procSelectObject.Call(hdc, oldB)
	procSelectObject.Call(hdc, oldP)
	procDeleteObject.Call(b)
	procDeleteObject.Call(p)
}
func drawOwnerButton(di *DRAWITEMSTRUCT) {
	bg, fg := buttonStyle(di.CtlID)
	if di.ItemState&ODS_SELECTED != 0 {
		if di.CtlID == ID_RUN {
			bg = rgb(103, 160, 37)
		} else {
			bg = rgb(43, 53, 40)
		}
	}
	drawRoundedBox(di.HDC, di.RcItem, bg, colLine, scale(7))
	procSetBkMode.Call(di.HDC, TRANSPARENT)
	procSetTextColor.Call(di.HDC, fg)
	txt := textOf(di.HwndItem)
	rr := di.RcItem
	procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16(txt))), ^uintptr(0), uintptr(unsafe.Pointer(&rr)), DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}
func comboItemText(hwnd uintptr, itemID uint32) string {
	if itemID == 0xFFFFFFFF {
		i, _, _ := procSendMessageW.Call(hwnd, CB_GETCURSEL, 0, 0)
		if int(i) < 0 {
			return ""
		}
		itemID = uint32(i)
	}
	ln, _, _ := procSendMessageW.Call(hwnd, CB_GETLBTEXTLEN, uintptr(itemID), 0)
	if int32(ln) < 0 {
		return ""
	}
	buf := make([]uint16, int(ln)+1)
	procSendMessageW.Call(hwnd, CB_GETLBTEXT, uintptr(itemID), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}
func drawOwnerCombo(di *DRAWITEMSTRUCT) {
	bg := colPanel
	if di.ItemState&ODS_SELECTED != 0 {
		bg = colSoft
	}
	b, _, _ := procCreateSolidBrush.Call(bg)
	procFillRect.Call(di.HDC, uintptr(unsafe.Pointer(&di.RcItem)), b)
	procDeleteObject.Call(b)
	procSetBkMode.Call(di.HDC, TRANSPARENT)
	procSetTextColor.Call(di.HDC, colText)
	rr := di.RcItem
	rr.Left += int32(scale(9))
	rr.Right -= int32(scale(6))
	txt := comboItemText(di.HwndItem, di.ItemID)
	procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16(txt))), ^uintptr(0), uintptr(unsafe.Pointer(&rr)), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}
func initBotolo() {
	if len(botoloBMP) < 54 || string(botoloBMP[:2]) != "BM" {
		return
	}
	off := int(binary.LittleEndian.Uint32(botoloBMP[10:14]))
	w := int(int32(binary.LittleEndian.Uint32(botoloBMP[18:22])))
	h := int(int32(binary.LittleEndian.Uint32(botoloBMP[22:26])))
	bpp := int(binary.LittleEndian.Uint16(botoloBMP[28:30]))
	if off <= 0 || w <= 0 || h == 0 || bpp != 24 {
		return
	}
	if h < 0 {
		h = -h
	}
	stride := ((w*3 + 3) / 4) * 4
	need := off + stride*h
	if need > len(botoloBMP) {
		return
	}
	pix := append([]byte(nil), botoloBMP[off:need]...)
	bgR, bgG, bgB := byte(colBG&0xFF), byte((colBG>>8)&0xFF), byte((colBG>>16)&0xFF)
	for y := 0; y < h; y++ {
		row := pix[y*stride:]
		for x := 0; x < w; x++ {
			b, g, r := row[x*3], row[x*3+1], row[x*3+2]
			if r > 238 && g > 238 && b > 238 {
				row[x*3], row[x*3+1], row[x*3+2] = bgB, bgG, bgR
			}
		}
	}
	botoloPixels, botoloWidth, botoloHeight, botoloStride = pix, w, h, stride
}
func drawBotolo(hdc uintptr, client RECT) {
	if len(botoloPixels) == 0 || botoloWidth <= 0 || botoloHeight <= 0 {
		return
	}
	// Asset già ritagliato: mantienilo interamente visibile nel footer in basso a destra.
	dw, dh := scale(botoloWidth), scale(botoloHeight)
	x := int(client.Right) - dw - scale(14)
	y := int(client.Bottom) - dh - scale(6)
	if x < scale(10) {
		x = scale(10)
	}
	if y < scale(10) {
		y = scale(10)
	}
	bmi := BITMAPINFO{Header: BITMAPINFOHEADER{
		Size: uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), Width: int32(botoloWidth), Height: int32(botoloHeight),
		Planes: 1, BitCount: 24, Compression: 0, SizeImage: uint32(botoloStride * botoloHeight),
	}}
	procStretchDIBits.Call(hdc, uintptr(x), uintptr(y), uintptr(dw), uintptr(dh), 0, 0,
		uintptr(botoloWidth), uintptr(botoloHeight), uintptr(unsafe.Pointer(&botoloPixels[0])), uintptr(unsafe.Pointer(&bmi)), 0, 0x00CC0020)
}

func paintBackground(hwnd uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc != 0 {
		var r RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brushBG)
		if controlPanel.Right > controlPanel.Left {
			drawRoundedBox(hdc, controlPanel, colPanel, colLine, scale(9))
		}
		if appIcon != 0 {
			procDrawIconEx.Call(hdc, uintptr(scale(14)), uintptr(scale(14)), appIcon, uintptr(scale(48)), uintptr(scale(48)), 0, 0, DI_NORMAL)
		}
		drawBotolo(hdc, r)
	}
	procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}
func applyDPI(hwnd uintptr) {
	if dpi, _, _ := procGetDpiForWindow.Call(hwnd); dpi >= 72 && dpi <= 384 {
		currentDPI = int(dpi)
	}
}
func setControlFonts() {
	h := func(px int, weight int, face string) uintptr {
		height := -scale(px)
		r, _, _ := procCreateFontW.Call(uintptr(int64(height)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16(face))))
		return r
	}
	fontUI = h(17, 400, "Segoe UI")
	fontSmall = h(14, 400, "Segoe UI")
	fontTitle = h(27, 600, "Segoe UI")
	fontMono = h(15, 400, "Consolas")
	fontBold = h(17, 700, "Segoe UI")
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_GETMINMAXINFO:
		if lParam != 0 {
			m := (*MINMAXINFO)(unsafe.Pointer(lParam))
			m.PtMinTrackSize.X = int32(scale(820))
			m.PtMinTrackSize.Y = int32(scale(680))
			return 0
		}
	case WM_SIZE:
		layout()
		return 0
	case WM_PAINT:
		paintBackground(hwnd)
		return 0
	case WM_DROPFILES:
		handleDrop(wParam)
		return 0
	case WM_DRAWITEM:
		if lParam != 0 {
			di := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
			if di.CtlType == ODT_BUTTON {
				drawOwnerButton(di)
				return 1
			}
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
		ctl := lParam
		procSetBkMode.Call(hdc, TRANSPARENT)
		if ctl == hTitle {
			procSetTextColor.Call(hdc, colText)
		} else {
			procSetTextColor.Call(hdc, colMuted)
		}
		if panelStatics[ctl] {
			return brushPanel
		}
		return brushBG
	case WM_CTLCOLORBTN:
		hdc := wParam
		procSetTextColor.Call(hdc, colText)
		procSetBkColor.Call(hdc, colPanel)
		return brushPanel
	case WM_APP_ANALYSIS_DONE:
		resultMu.Lock()
		res := pendingResult
		pendingResult = ""
		resultMu.Unlock()
		replaceOutput(res)
		setBusy(false)
		return 0
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		code := int((wParam >> 16) & 0xffff)
		switch id {
		case ID_PASTE:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSetFocus.Call(hInput)
			procSendMessageW.Call(hInput, WM_PASTE, 0, 0)
		case ID_OPEN:
			if code != 0 {
				return 0
			} // BN_CLICKED
			openTextFile()
		case ID_CLEAR:
			if code != 0 {
				return 0
			} // BN_CLICKED
			setText(hInput, "")
			replaceOutput("")
			setText(hParam, "")
			setText(hStatus, fmt.Sprintf("Locale · %d operazioni · nessun invio esterno", len(operations)))
			procSetFocus.Call(hInput)
		case ID_QUICK_AZ:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0)
			updateHint()
			runOp(20)
		case ID_QUICK_REP:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 9, 0)
			updateHint()
			runOp(9)
		case ID_QUICK_KUNTA:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 26, 0)
			updateHint()
			runOp(26)
		case ID_QUICK_VOW:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 27, 0)
			updateHint()
			runOp(27)
		case ID_QUICK_CONS:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 28, 0)
			updateHint()
			runOp(28)
		case ID_COMBO:
			if code == 1 {
				updateHint()
			}
		case ID_RUN:
			if code != 0 {
				return 0
			} // BN_CLICKED
			i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0)
			runOp(int(i))
		case ID_COPY:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hOutput, EM_SETSEL, 0, ^uintptr(0))
			procSendMessageW.Call(hOutput, WM_COPY, 0, 0)
			setText(hStatus, "Risultato copiato.")
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func main() {
	runtime.LockOSThread()
	hrCOM, _, _ := procCoInitializeEx.Call(0, 0x2) // COINIT_APARTMENTTHREADED
	if !hresultFailed(hrCOM) {
		defer procCoUninitialize.Call()
	}
	initBotolo()
	hInst, _, _ := procGetModuleHandleW.Call(0)
	cur, _, _ := procLoadCursorW.Call(0, 32512)
	appIcon, _, _ = procLoadIconW.Call(hInst, 1)

	brushBG, _, _ = procCreateSolidBrush.Call(colBG)
	brushPanel, _, _ = procCreateSolidBrush.Call(colPanel)
	brushInput, _, _ = procCreateSolidBrush.Call(colInput)
	brushAccent, _, _ = procCreateSolidBrush.Call(colAccent)
	brushSoft, _, _ = procCreateSolidBrush.Call(colSoft)
	brushLine, _, _ = procCreateSolidBrush.Call(colLine)

	class := u16("KuntaNativeWindowV8")
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HIcon: appIcon, HCursor: cur, HbrBackground: brushBG, LpszClassName: class, HIconSm: appIcon}
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return
	}

	hwndMain, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(u16("Kunta"))), WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN, CW_USEDEFAULT, CW_USEDEFAULT, 1120, 860, 0, 0, hInst, 0)
	if hwndMain == 0 {
		return
	}
	applyDPI(hwndMain)
	setImmersiveDarkTitlebar(hwndMain)
	setControlFonts()
	procDragAcceptFiles.Call(hwndMain, 1)

	hTitle = addStatic("Kunta", ID_TITLE, false)
	hTagline = addStatic("Conta · osserva · ordina · rac-conta", ID_TAGLINE, false)

	addButton("Incolla", ID_PASTE)
	addButton("Apri testo…", ID_OPEN)
	addButton("Pulisci", ID_CLEAR)

	hInput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN|ES_NOHIDESEL, WS_EX_CLIENTEDGE, ID_EDIT_INPUT)
	procSendMessageW.Call(hInput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(10)|(scale(10)<<16)))
	procSendMessageW.Call(hInput, EM_SETLIMITTEXT, 0x7FFFFFFE, 0)

	addButton("A–Z", ID_QUICK_AZ)
	addButton("Ripetizioni", ID_QUICK_REP)
	addButton("Kunta il testo", ID_QUICK_KUNTA)
	addButton("Isovocaliche", ID_QUICK_VOW)
	addButton("Isoconsonantiche", ID_QUICK_CONS)

	addStatic(fmt.Sprintf("Operazione · %d voci", len(operations)), ID_LABEL_OPERATION, true)
	addStatic("Parametro", ID_LABEL_PARAM, true)
	hCombo = create("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST|CBS_HASSTRINGS, 0, ID_COMBO)
	for _, op := range operations {
		procSendMessageW.Call(hCombo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(op.Name))))
	}
	procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0)
	procSendMessageW.Call(hCombo, CB_SETMINVISIBLE, 18, 0)
	if n, _, _ := procSendMessageW.Call(hCombo, CB_GETCOUNT, 0, 0); int(n) != len(operations) {
		message(fmt.Sprintf("Errore interno: operazioni caricate %d su %d.", int(n), len(operations)))
	}

	hParam = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_AUTOHSCROLL, WS_EX_CLIENTEDGE, ID_PARAM)
	procSendMessageW.Call(hParam, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(8)|(scale(8)<<16)))
	hCase = create("BUTTON", "Maiuscole/minuscole distinte", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 0, ID_CASE)
	addButton("KUNTA!", ID_RUN)
	hHint = addStatic("", ID_HINT, true)
	updateHint()

	hOutput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN|ES_READONLY|ES_NOHIDESEL, WS_EX_CLIENTEDGE, ID_OUTPUT)
	procSendMessageW.Call(hOutput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(10)|(scale(10)<<16)))
	procSendMessageW.Call(hOutput, EM_SETLIMITTEXT, 0x7FFFFFFE, 0)
	addButton("Copia risultato", ID_COPY)
	hStatus = addStatic(fmt.Sprintf("Locale · %d operazioni · nessun invio esterno", len(operations)), ID_STATUS, false)
	hFooter = addStatic("", ID_FOOTER, false)

	layout()
	procInvalidateRect.Call(hwndMain, 0, 1)
	procSetFocus.Call(hInput)

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
