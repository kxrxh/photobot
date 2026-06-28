package maputil

import "testing"

func TestParseFloatString(t *testing.T) {
	if v, ok := ParseFloatString("1.5"); !ok || v != 1.5 {
		t.Fatalf("float: %v %v", v, ok)
	}
	if v, ok := ParseFloatString("3,25"); !ok || v != 3.25 {
		t.Fatalf("comma decimal: %v %v", v, ok)
	}
	if _, ok := ParseFloatString(""); ok {
		t.Fatal("empty")
	}
	if _, ok := ParseFloatString("-"); ok {
		t.Fatal("dash")
	}
}
