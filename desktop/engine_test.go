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
alpha beta beta gamma gamma gamma.
123 456 123.`
	params := map[int]string{0: "aeiou", 7: "ama", 13: "3", 14: "3", 15: "3", 22: "3", 23: "3", 24: "3", 25: "2", 31: "1", 32: "1", 33: "1", 34: "1", 37: "a", 38: "a", 39: "am", 40: "3", 42: "2", 51: "3"}
	for i := range operations {
		p := params[i]
		got := analyze(sample, i, p, false)
		if strings.TrimSpace(got) == "" {
			t.Fatalf("operazione %d %q: output vuoto", i, operations[i].Name)
		}
	}
}

func TestColumnOutputs(t *testing.T) {
	sample := "beta alfa beta gamma alfa beta"
	for _, op := range []int{8, 9, 10, 11, 20, 36, 40, 46, 47, 50} {
		got := analyze(sample, op, "3", false)
		if !strings.Contains(got, "\n") {
			t.Fatalf("operazione %d %q: output non verticale: %q", op, operations[op].Name, got)
		}
	}
}
