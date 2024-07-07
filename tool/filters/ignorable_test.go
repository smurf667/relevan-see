package filters

import (
	"testing"
)

func TestIgnore(t *testing.T) {
	var modifications = []Modification{
		{"./README.txt", "M", "", ""},
		{"./always-seen", "M", "", ""},
	}
	ignorable := Ignorable{patterns: ToPatterns([]string{"*.txt"})}
	result := ignorable.Filter(modifications)
	if len(result) != 1 {
		t.Fatalf("Unexpected filter result: %q", result)
	}
}
