package documentreader

import (
	"encoding/xml"
	"io"
	"os"
	"slices"
	"testing"
)

func TestReadContentLimited(t *testing.T) {
	var tests = []struct {
		document string
		limit    int64
		checker  func(xml.StartElement) bool
		golden   string
	}{
		{"docx/content1.xml", 700, IsDocxText, "docx/title1.golden"},
		{"docx/content2.xml", 700, IsDocxText, "docx/title2.golden"},
		{"docx/content3.xml", 700, IsDocxText, "docx/title3.golden"},
		{"odt/content1.xml", 700, IsOdtText, "odt/title1.golden"},
		{"odt/content3.xml", 700, IsOdtText, "odt/title3.golden"},
		{"odt/content4.xml", 700, IsOdtText, "odt/title4.golden"},
		{"odt/content4.xml", 9999, IsOdtText, "odt/title4.golden"},
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

		got, err := readContentLimited(file, tt.limit, tt.checker)
		if err != nil  && err != io.ErrUnexpectedEOF{
			t.Fatalf("Failed to parse text in %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}
