package docx

import (
	"os"
	"slices"
	"testing"
)

func TestRead(t *testing.T) {
	var tests = []struct {
		document string
		limit    int
		golden   string
	}{
		{"document1.xml", 700, "title1.golden"},
		{"document2.xml", 700, "title2.golden"},
		{"document3.xml", 700, "title3.golden"},
	}

	for _, tt := range tests {
		file, err := os.Open("testdata/" + tt.document)
		if err != nil {
			t.Fatalf("Failed to open %s: %v", tt.document, err)
		}
		defer file.Close()

		want, err := os.ReadFile("testdata/" + tt.golden)
		if err != nil {
			t.Fatalf("Failed to open %s: %v", tt.document, err)
		}

		got, err := TextLimited(file, tt.limit)
		if err != nil {
			t.Fatalf("Failed to parse text in %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}
