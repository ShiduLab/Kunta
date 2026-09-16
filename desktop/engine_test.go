package main

import (
	"strings"
	"testing"
)

func TestOperationsComplete(t *testing.T) {
	if len(operations) != 56 {
		t.Fatalf("operazioni = %d, voglio 56", len(operations))
	}
}

func TestAllOperationsProduceOutput(t *testing.T) {
	sample := `Chi mi ama mi chiama, chi mi odia mi insegua.
AMA radar anna otto.
luce buio amore odio vivo morto.
Rima cima prima tremA.
organo onagro citerei eretici cotenna canneto.
Energia esistenza efficace nell'amare devastante.
alpha beta beta gamma gamma gamma.
123 456 123.`
	params := map[int]string{0: "aeiou", 7: "ama", 13: "3", 14: "3", 15: "3", 22: "3", 23: "3", 24: "3", 25: "2", 31: "1", 32: "1", 33: "1", 34: "1", 37: "a", 38: "a", 39: "am", 40: "3", 42: "2", 50: "3", 51: "3"}
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
	got := analyze("citerei eretici chi mi ama", 50, "3", false)
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
	got := analyze("cotenna canneto chi mi ama", 51, "3", false)
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
