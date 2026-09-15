package analysis

import (
    "fmt"
    "sort"
    "strings"
    "unicode"
)

type Group struct { Key string; Words []string }

type Summary struct {
    Characters int
    Letters int
    Words int
    Lines int
    Spaces int
    Digits int
    Punctuation int
    UniqueWords int
    RepeatedWords int
}

func NormalizeWord(s string) string {
    var b strings.Builder
    for _, r := range strings.ToLower(s) {
        if unicode.IsLetter(r) || unicode.IsDigit(r) { b.WriteRune(r) }
    }
    return b.String()
}

func Words(text string) []string {
    return strings.FieldsFunc(text, func(r rune) bool { return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r=='\'' || r=='’') })
}

func Lines(text string) []string {
    text = strings.ReplaceAll(text, "\r\n", "\n")
    text = strings.ReplaceAll(text, "\r", "\n")
    return strings.Split(text, "\n")
}

func SummaryOf(text string) Summary {
    s := Summary{Characters: len([]rune(text)), Lines: len(Lines(text))}
    freq := map[string]int{}
    for _, r := range text {
        switch {
        case unicode.IsLetter(r): s.Letters++
        case unicode.IsDigit(r): s.Digits++
        case unicode.IsSpace(r): s.Spaces++
        case unicode.IsPunct(r): s.Punctuation++
        }
    }
    ws := Words(text); s.Words = len(ws)
    for _, w := range ws { n:=NormalizeWord(w); if n!="" { freq[n]++ } }
    s.UniqueWords = len(freq)
    for _, n := range freq { if n > 1 { s.RepeatedWords++ } }
    return s
}

func CountLetters(text, wanted string, caseSensitive bool) map[rune]int {
    set := map[rune]bool{}
    if !caseSensitive { wanted = strings.ToLower(wanted) }
    for _, r := range wanted { if unicode.IsLetter(r) { set[r]=true } }
    out:=map[rune]int{}
    if !caseSensitive { text = strings.ToLower(text) }
    for _, r := range text { if set[r] { out[r]++ } }
    return out
}

func Occurrences(text, needle string, caseSensitive bool) int {
    if needle=="" { return 0 }
    if !caseSensitive { text=strings.ToLower(text); needle=strings.ToLower(needle) }
    return strings.Count(text, needle)
}

func WordFrequency(text string, caseSensitive bool) []Group {
    m:=map[string]int{}
    for _, w := range Words(text) {
        k:=w; if !caseSensitive { k=strings.ToLower(k) }
        k=strings.Trim(k, "'’"); if k!="" { m[k]++ }
    }
    keys:=make([]string,0,len(m)); for k:=range m { keys=append(keys,k) }
    sort.Slice(keys,func(i,j int)bool{ if m[keys[i]]==m[keys[j]] {return keys[i]<keys[j]}; return m[keys[i]]>m[keys[j]] })
    out:=make([]Group,0,len(keys)); for _,k:=range keys { out=append(out,Group{fmt.Sprintf("%d",m[k]),[]string{k}}) }
    return out
}

func LetterFrequency(text string, caseSensitive bool) []Group {
    m:=map[rune]int{}; if !caseSensitive { text=strings.ToLower(text) }
    for _,r:=range text { if unicode.IsLetter(r) {m[r]++} }
    keys:=make([]rune,0,len(m)); for k:=range m {keys=append(keys,k)}
    sort.Slice(keys,func(i,j int)bool{if m[keys[i]]==m[keys[j]]{return keys[i]<keys[j]};return m[keys[i]]>m[keys[j]]})
    out:=[]Group{}; for _,k:=range keys{out=append(out,Group{fmt.Sprintf("%d",m[k]),[]string{string(k)}})}; return out
}

func ShortestLongest(text string) (shortest,longest []string) {
    ws:=Words(text); if len(ws)==0{return}
    min,max:=1<<30,0; seenS:=map[string]bool{}; seenL:=map[string]bool{}
    for _,w:=range ws { n:=len([]rune(NormalizeWord(w))); if n==0{continue}; if n<min{min=n;shortest=nil;seenS=map[string]bool{}}; if n==min&&!seenS[w]{shortest=append(shortest,w);seenS[w]=true}; if n>max{max=n;longest=nil;seenL=map[string]bool{}}; if n==max&&!seenL[w]{longest=append(longest,w);seenL[w]=true} }
    return
}

func reverse(s string) string { r:=[]rune(s); for i,j:=0,len(r)-1;i<j;i,j=i+1,j-1{r[i],r[j]=r[j],r[i]}; return string(r) }

func Palindromes(text string,min int) []string { out:=[]string{};seen:=map[string]bool{};for _,w:=range Words(text){n:=NormalizeWord(w);if len([]rune(n))>=min&&n==reverse(n)&&!seen[n]{seen[n]=true;out=append(out,w)}};sort.Strings(out);return out }
func Bifronts(text string,min int) []Group { present:=map[string]string{};for _,w:=range Words(text){n:=NormalizeWord(w);if len([]rune(n))>=min{if _,ok:=present[n];!ok{present[n]=w}}};out:=[]Group{};done:=map[string]bool{};for n,w:=range present{r:=reverse(n);if r!=n{if w2,ok:=present[r];ok&&!done[n]&&!done[r]{a,b:=w,w2;if strings.ToLower(a)>strings.ToLower(b){a,b=b,a};out=append(out,Group{"",[]string{a,b}});done[n]=true;done[r]=true}}};sort.Slice(out,func(i,j int)bool{return strings.Join(out[i].Words," ")<strings.Join(out[j].Words," ")});return out}
func Anagrams(text string,min int) []Group { return groupWords(text,min,func(w string)string{r:=[]rune(w);sort.Slice(r,func(i,j int)bool{return r[i]<r[j]});return string(r)},false) }

func vowelsOnly(w string) string { var b strings.Builder; for _,r:=range strings.ToLower(w){if strings.ContainsRune("aeiouàèéìíîòóùú",r){b.WriteRune(r)}};return b.String() }
func consonantsOnly(w string) string { var b strings.Builder; for _,r:=range strings.ToLower(w){if unicode.IsLetter(r)&&!strings.ContainsRune("aeiouàèéìíîòóùú",r){b.WriteRune(r)}};return b.String() }
func runeSorted(s string)string{r:=[]rune(s);sort.Slice(r,func(i,j int)bool{return r[i]<r[j]});return string(r)}
func edgeKey(s string,n int,tail bool)string{r:=[]rune(s);if len(r)<n{return ""};if tail{return string(r[len(r)-n:])};return string(r[:n])}
func groupWords(text string,min int,keyfn func(string)string,distinct bool) []Group { m:=map[string][]string{}; seen:=map[string]map[string]bool{};for _,raw:=range Words(text){w:=NormalizeWord(raw);if len([]rune(w))<min{continue};k:=keyfn(w);if k==""{continue};if seen[k]==nil{seen[k]=map[string]bool{}};if !seen[k][w]{seen[k][w]=true;m[k]=append(m[k],raw)}};out:=[]Group{};for k,ws:=range m{if len(ws)>1{if distinct{u:=map[string]bool{};for _,x:=range ws{u[NormalizeWord(x)]=true};if len(u)<2{continue}};sort.Slice(ws,func(i,j int)bool{return strings.ToLower(ws[i])<strings.ToLower(ws[j])});out=append(out,Group{k,ws})}};sort.Slice(out,func(i,j int)bool{return out[i].Key<out[j].Key});return out }

// Isovocaliche: stessa sequenza vocalica, nello stesso ordine e con la stessa molteplicità.
func Isovocalic(text string,min int) []Group { return groupWords(text,min,vowelsOnly,true) }
// Isoconsonantiche: stessa sequenza consonantica, nello stesso ordine e con la stessa molteplicità.
func Isoconsonantic(text string,min int) []Group { return groupWords(text,min,consonantsOnly,true) }
// Omovocaliche: stesso patrimonio vocalico, indipendentemente dall'ordine.
func Homovocalic(text string,min int) []Group { return groupWords(text,min,func(w string)string{return runeSorted(vowelsOnly(w))},true) }
// Omoconsonantiche: stesso patrimonio consonantico, indipendentemente dall'ordine.
func Homoconsonantic(text string,min int) []Group { return groupWords(text,min,func(w string)string{return runeSorted(consonantsOnly(w))},true) }
func HomovocalicInitial(text string,min,edge int) []Group { return groupWords(text,min,func(w string)string{return edgeKey(vowelsOnly(w),edge,false)},true) }
func HomovocalicFinal(text string,min,edge int) []Group { return groupWords(text,min,func(w string)string{return edgeKey(vowelsOnly(w),edge,true)},true) }
func HomoconsonanticInitial(text string,min,edge int) []Group { return groupWords(text,min,func(w string)string{return edgeKey(consonantsOnly(w),edge,false)},true) }
func HomoconsonanticFinal(text string,min,edge int) []Group { return groupWords(text,min,func(w string)string{return edgeKey(consonantsOnly(w),edge,true)},true) }

func Acrostic(text string) string { var b strings.Builder; for _,ln:=range Lines(text){ln=strings.TrimSpace(ln);if ln==""{continue};r:=[]rune(ln);b.WriteRune(r[0])};return b.String() }
func Telestic(text string) string { var b strings.Builder; for _,ln:=range Lines(text){ln=strings.TrimSpace(ln);if ln==""{continue};r:=[]rune(ln);b.WriteRune(r[len(r)-1])};return b.String() }
func ReverseText(text string) string { return reverse(text) }
func ReverseWords(text string) string { ws:=Words(text);for i,j:=0,len(ws)-1;i<j;i,j=i+1,j-1{ws[i],ws[j]=ws[j],ws[i]};return strings.Join(ws," ") }
func SortWords(text string) string { ws:=Words(text);sort.Slice(ws,func(i,j int)bool{return strings.ToLower(ws[i])<strings.ToLower(ws[j])});return strings.Join(ws,"\n") }
func Numbers(text string) []string { var out []string; var b strings.Builder; flush:=func(){if b.Len()>0{out=append(out,b.String());b.Reset()}};for _,r:=range text{if unicode.IsDigit(r)||((r==','||r=='.')&&b.Len()>0){b.WriteRune(r)}else{flush()}};flush();return out }
