package documentreader

import (
	"encoding/xml"
	"io"
	"os"
	"slices"
	"testing"
)

func TestReadLimited(t *testing.T) {
	var tests = []struct {
		document    string
		limit       int64
		contentPath string
		checker     func(xml.StartElement) bool
		golden      string
	}{
		{"odt/document1.odt", 700, OdtContentPath, IsOdtText, "odt/title1.golden"},
		{"odt/document3.odt", 700, OdtContentPath, IsOdtText, "odt/title3.golden"},
		{"odt/document4.odt", 700, OdtContentPath, IsOdtText, "odt/title4.golden"},
		{"odt/document4.odt", 9999, OdtContentPath, IsOdtText, "odt/title4.golden"},
		{"docx/document1.docx", 700, DocxContentPath, IsDocxText, "docx/title1.golden"},
		{"docx/document2.docx", 700, DocxContentPath, IsDocxText, "docx/title2.golden"},
		{"docx/document3.docx", 700, DocxContentPath, IsDocxText, "docx/title3.golden"},
	}

	for _, tt := range tests {
		file, err := os.Open("testdata/" + tt.document)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tt.document, err)
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			t.Fatalf("Failed to get %s info: %v", tt.document, err)
		}
		size := fi.Size()

		want, err := os.ReadFile("testdata/" + tt.golden)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tt.document, err)
		}

		got, err := ReadLimited(file, size, tt.limit, tt.contentPath, tt.checker)
		if err != nil && err != io.ErrUnexpectedEOF {
			t.Fatalf("Failed to parse text in %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}

func TestReadContentLimited(t *testing.T) {
	var tests = []struct {
		document string
		limit    int64
		checker  func(xml.StartElement) bool
		golden   string
	}{
		{"odt/content1.xml", 700, IsOdtText, "odt/title1.golden"},
		{"odt/content3.xml", 700, IsOdtText, "odt/title3.golden"},
		{"odt/content4.xml", 700, IsOdtText, "odt/title4.golden"},
		{"odt/content4.xml", 9999, IsOdtText, "odt/title4.golden"},
		{"docx/content1.xml", 700, IsDocxText, "docx/title1.golden"},
		{"docx/content2.xml", 700, IsDocxText, "docx/title2.golden"},
		{"docx/content3.xml", 700, IsDocxText, "docx/title3.golden"},
	}

	for _, tt := range tests {
		file, err := os.Open("testdata/" + tt.document)
		if err != nil {
			t.Fatalf("Failed to open %s: %v", tt.document, err)
		}
		defer file.Close()

		want, err := os.ReadFile("testdata/" + tt.golden)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tt.document, err)
		}

		got, err := readContentLimited(file, tt.limit, tt.checker)
		if err != nil && err != io.ErrUnexpectedEOF {
			t.Fatalf("Failed to parse text in %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}
