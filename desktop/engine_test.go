package main

import (
	"strings"
	"testing"
)

func TestOperationsComplete(t *testing.T) {
	if len(operations) != 58 {
		t.Fatalf("operazioni = %d, voglio 58 con Palindromo inverso, Palindromo contrario, Inversi e Antipodi distinti", len(operations))
	}
}

func TestAllOperationsProduceOutput(t *testing.T) {
	sample := `Chi mi ama mi chiama, chi mi odia mi insegua.
AMA radar anna otto.
luce buio amore odio vivo morto.
Rima Amir ACETONE ENOTECA POSSESSO cima prima tremA.
organo onagro citerei eretici cotenna canneto.
Energia esistenza efficace nell'amare devastante.
alpha beta beta gamma gamma gamma.
123 456 123.`
	params := map[int]string{0: "aeiou", 7: "ama", 13: "3", 14: "3", 15: "3", 22: "3", 23: "3", 24: "3", 25: "2", 31: "1", 32: "1", 33: "1", 34: "1", 37: "a", 38: "a", 39: "am", 40: "3", 42: "2", 50: "3", 51: "3", 52: "3", 53: "3"}
	for i := range operations {
		got := analyze(sample, i, params[i], false)
		if strings.TrimSpace(got) == "" {
			t.Fatalf("operazione %d %q: output vuoto", i, operations[i].Name)
		}
	}
}

func TestSignaturesKeepOriginalOrder(t *testing.T) {
	if got := vowelSkeleton("Energia"); got != "eeia" {
		t.Fatalf("Energia vocali = %q, voglio eeia", got)
	}
	if got := vowelSkeleton("esistenza"); got != "eiea" {
		t.Fatalf("esistenza vocali = %q, voglio eiea", got)
	}
	if got := consonantSkeleton("organo"); got != "rgn" {
		t.Fatalf("organo consonanti = %q, voglio rgn", got)
	}
	if got := consonantSkeleton("onagro"); got != "ngr" {
		t.Fatalf("onagro consonanti = %q, voglio ngr", got)
	}
}

func TestOmovocalicheShowOrderedSignaturePerWord(t *testing.T) {
	got := analyze("Energia esistenza efficace nell'amare devastante", 29, "", false)
	must := []string{
		"Energia  →  E-E-I-A",
		"esistenza  →  E-I-E-A",
		"efficace  →  E-I-A-E",
		"nell'amare  →  E-A-A-E",
		"devastante  →  E-A-A-E",
	}
	for _, want := range must {
		if !strings.Contains(got, want) {
			t.Fatalf("manca %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Energia  →  A-E-E-I") {
		t.Fatalf("Energia è stata alfabetizzata invece di mantenere l'ordine: %s", got)
	}
}

func TestOmoconsonanticheShowOrderedSignaturePerWord(t *testing.T) {
	got := analyze("organo onagro", 30, "", false)
	if !strings.Contains(got, "organo  →  R-G-N") || !strings.Contains(got, "onagro  →  N-G-R") {
		t.Fatalf("ordine consonanti errato:\n%s", got)
	}
}

func TestBifronteOnlyIfCounterpartExists(t *testing.T) {
	got := analyze("organo onagro chi", 14, "3", false)
	if !strings.Contains(strings.ToLower(got), "organo") || !strings.Contains(strings.ToLower(got), "onagro") {
		t.Fatalf("bifronte reale non rilevato:\n%s", got)
	}
	if strings.Contains(strings.ToLower(got), "ihc") {
		t.Fatalf("parola inventata nel bifronte:\n%s", got)
	}
}

func TestInverseOnlyIfCounterpartExists(t *testing.T) {
	got := analyze("citerei eretici chi mi ama", 52, "3", false)
	low := strings.ToLower(got)
	if !strings.Contains(low, "citerei") || !strings.Contains(low, "eretici") {
		t.Fatalf("inverso reale non rilevato:\n%s", got)
	}
	for _, fakeLine := range []string{"chi ↔ ihc", "mi ↔ im", "ama ↔ ama", "chi →", "mi →"} {
		if strings.Contains(low, fakeLine) {
			t.Fatalf("trasformazione inesistente %q presente:\n%s", fakeLine, got)
		}
	}
}

func TestAntipodeOnlyIfCounterpartExists(t *testing.T) {
	got := analyze("cotenna canneto chi mi ama", 53, "3", false)
	low := strings.ToLower(got)
	if !strings.Contains(low, "cotenna") || !strings.Contains(low, "canneto") {
		t.Fatalf("antipodo reale non rilevato:\n%s", got)
	}
	if strings.Contains(low, "chi ↔") {
		t.Fatalf("antipodo inesistente presente:\n%s", got)
	}
}

func TestColumnOutputs(t *testing.T) {
	sample := "beta alfa beta gamma alfa beta"
	for _, op := range []int{8, 9, 10, 11, 20, 36, 40, 46, 47} {
		got := analyze(sample, op, "3", false)
		if !strings.Contains(got, "\n") {
			t.Fatalf("operazione %d %q: output non verticale: %q", op, operations[op].Name, got)
		}
	}
}

func TestPalindromoInversoOnlyIfCounterpartExists(t *testing.T) {
	got := analyze("Rima Amir ACETONE ENOTECA chi mi ama", 50, "3", false)
	low := strings.ToLower(got)
	for _, want := range []string{"rima", "amir", "acetone", "enoteca"} {
		if !strings.Contains(low, want) {
			t.Fatalf("palindromo inverso reale %q non rilevato:\n%s", want, got)
		}
	}
	for _, fakeLine := range []string{"chi ↔ ihc", "mi ↔ im", "ama ↔ ama"} {
		if strings.Contains(low, fakeLine) {
			t.Fatalf("trasformazione inventata %q presente:\n%s", fakeLine, got)
		}
	}
}

func TestPalindromoContrarioRestored(t *testing.T) {
	got := analyze("POSSESSO parola casa", 51, "3", false)
	if !strings.Contains(strings.ToLower(got), "possesso") {
		t.Fatalf("Palindromo contrario POSSESSO non rilevato:\n%s", got)
	}
}

func TestFourDistinctEnigmaticOperations(t *testing.T) {
	want := []string{"Palindromo inverso", "Palindromo contrario", "Inversi", "Antipodi"}
	for i, name := range want {
		if operations[50+i].Name != name {
			t.Fatalf("operazione %d = %q, voglio %q", 50+i, operations[50+i].Name, name)
		}
	}
}

func TestParoleAlfabeticheSonoConsecutive(t *testing.T) {
	got := analyze("ab abc bcd def xyz al ci del", 44, "", false)
	low := strings.ToLower(got)
	for _, want := range []string{"ab", "abc", "bcd", "def", "xyz"} {
		if !strings.Contains(low, "\n"+want+"\n") && !strings.HasSuffix(low, "\n"+want) {
			t.Fatalf("sequenza alfabetica %q non rilevata:\n%s", want, got)
		}
	}
	for _, bad := range []string{"al", "ci", "del"} {
		if strings.Contains(low, "\n"+bad+"\n") || strings.HasSuffix(low, "\n"+bad) {
			t.Fatalf("%q non è una sequenza alfabetica consecutiva:\n%s", bad, got)
		}
	}
}

func TestParoleAlfabeticheInverseSonoConsecutive(t *testing.T) {
	got := analyze("ba cba fed zyx la eda", 45, "", false)
	low := strings.ToLower(got)
	for _, want := range []string{"ba", "cba", "fed", "zyx"} {
		if !strings.Contains(low, "\n"+want+"\n") && !strings.HasSuffix(low, "\n"+want) {
			t.Fatalf("sequenza alfabetica inversa %q non rilevata:\n%s", want, got)
		}
	}
	for _, bad := range []string{"la", "eda"} {
		if strings.Contains(low, "\n"+bad+"\n") || strings.HasSuffix(low, "\n"+bad) {
			t.Fatalf("%q non è una sequenza alfabetica inversa consecutiva:\n%s", bad, got)
		}
	}
}

func TestCaratteriMostraTotaleEFrequenze(t *testing.T) {
	got := analyze("AaA cc", 1, "", false)
	for _, want := range []string{"Caratteri: 6", "3: A", "2: C", "1: [spazio]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("manca %q nell'output Caratteri:\n%s", want, got)
		}
	}
}

func TestParoleMostraTotaleDistinteEFrequenze(t *testing.T) {
	got := analyze("Ciao ciao mondo ciao", 2, "", false)
	for _, want := range []string{"Parole: 4", "Parole distinte: 2", "3: Ciao", "1: mondo"} {
		if !strings.Contains(got, want) {
			t.Fatalf("manca %q nell'output Parole:\n%s", want, got)
		}
	}
}

func TestAnagrammiUsanoDizionarioCompleto(t *testing.T) {
	got := analyze("caos", 15, "3", false)
	low := strings.ToLower(got)
	for _, want := range []string{"caso", "cosa"} {
		if !strings.Contains(low, "    "+want) {
			t.Fatalf("anagramma %q non trovato nel dizionario:\n%s", want, got)
		}
	}
	if strings.Contains(low, "caos →\n    caos") {
		t.Fatalf("la parola sorgente non deve essere proposta come proprio anagramma:\n%s", got)
	}
}

func TestFirmaAnagrammaticaConservaMolteplicita(t *testing.T) {
	a, _ := anagramSignatureKey("casa")
	b, _ := anagramSignatureKey("caas")
	c, _ := anagramSignatureKey("caso")
	if a != b {
		t.Fatalf("casa e caas devono avere la stessa firma")
	}
	if a == c {
		t.Fatalf("casa e caso non devono avere la stessa firma")
	}
}
