package filters

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func buildTestComments(t *testing.T) Comments {
	list, _ := CreateComments("", []byte(`[{
		"patterns": [ "*.txt" ],
		"single": "//",
		"multi-start": "/*",
		"multi-end": "*/"
	}]`))
	if len(list.comments) != 1 {
		t.Fatal("Could not build comment")
	}
	return list
}

func dump(sb *strings.Builder, b []byte) {
	var a [16]byte
	n := (len(b) + 15) &^ 15
	for i := 0; i < n; i++ {
		if i%16 == 0 {
			sb.WriteString(fmt.Sprintf("%4d", i))
		}
		if i%8 == 0 {
			sb.WriteString(" ")
		}
		if i < len(b) {
			sb.WriteString(fmt.Sprintf(" %02X", b[i]))
		} else {
			sb.WriteString("   ")
		}
		if i >= len(b) {
			a[i%16] = ' '
		} else if b[i] < 32 || b[i] > 126 {
			a[i%16] = '.'
		} else {
			a[i%16] = b[i]
		}
		if i%16 == 15 {
			sb.WriteString(fmt.Sprintf("  %s\n", string(a[:])))
		}
	}
}

func TestSearch(t *testing.T) {
	list := buildTestComments(t)
	if search("hello.txt")(list.comments[0]) == false {
		t.Fatal("Did not find expected")
	}
	if search("hello.md")(list.comments[0]) == true {
		t.Fatal("Unexpected find")
	}
}

func TestNormalize(t *testing.T) {
	list := buildTestComments(t)
	for i := range 99 {
		if i > 0 {
			b, err := os.ReadFile(fmt.Sprintf("./testfiles/%02d.txt", i))
			if err != nil {
				break
			}
			content := string(b)
			b, err = os.ReadFile(fmt.Sprintf("./testfiles/%02d-expected.txt", i))
			if err != nil {
				break
			}
			expected := string(b)
			actual := normalize(list.comments[0], content)
			if actual != expected {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Test %02d\nExpected:\n", i))
				dump(&sb, []byte(expected))
				sb.WriteString("\nActual:\n")
				dump(&sb, []byte(actual))
				sb.WriteString("\nInput:\n")
				dump(&sb, []byte(content))
				t.Fatalf(sb.String())
			}
		}
	}
}
