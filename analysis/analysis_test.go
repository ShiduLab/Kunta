package analysis
import "testing"
func TestIsoVocalic(t *testing.T){g:=Isovocalic("casa rama fata mela",2);found:=false;for _,x:=range g{if x.Key=="aa"&&len(x.Words)==3{found=true}};if !found{t.Fatalf("atteso gruppo aa: %#v",g)}}
func TestAcroTele(t *testing.T){s:="Casa\nOrma\nRete";if Acrostic(s)!="COR"{t.Fatal(Acrostic(s))};if Telestic(s)!="aae"{t.Fatal(Telestic(s))}}
func TestEdges(t *testing.T){g:=HomovocalicInitial("aiuola aiuto paura",2,2);if len(g)==0{t.Fatalf("nessun gruppo: %#v",g)}}
func TestPal(t *testing.T){p:=Palindromes("ossesso anna casa",3);if len(p)!=2{t.Fatalf("%v",p)}}
