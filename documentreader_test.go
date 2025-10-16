package documentreader

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"slices"
	"testing"
)

func TestReadLimited_Success(t *testing.T) {
	var tests = []struct {
		document    string
		limit       int64
		contentPath string
		checker     func(xml.StartElement) bool
		golden      string
	}{
		{"odt/document1.odt", 700, OdtContentPath, IsOdtText, "odt/title1.golden"},
		{"odt/document3.odt", 700, OdtContentPath, IsOdtText, "odt/title3.golden"},
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

		got, err := ReadLimited(file, size, tt.limit, tt.contentPath, tt.checker)
		if err != nil {
			t.Fatalf("Unexpected error reading %s: %v", tt.document, err)
		}

		want, err := os.ReadFile("testdata/" + tt.golden)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}

func TestReadLimited_Error_UnexpectedEOF(t *testing.T) {
	document := "odt/document4.odt"
	golden := "odt/title4.golden"

	file, err := os.Open("testdata/" + document)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", document, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to get %s info: %v", document, err)
	}
	size := fi.Size()

	got, err := ReadLimited(file, size, 99999, OdtContentPath, IsOdtText)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected error reading %s: %v", document, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", document, err)
	}

	if !slices.Equal(got, want) {
		t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", document, got, golden, want)
	}
}

func TestReadContentLimited_Success(t *testing.T) {
	var tests = []struct {
		document string
		limit    int64
		checker  func(xml.StartElement) bool
		golden   string
	}{
		{"odt/content1.xml", 700, IsOdtText, "odt/title1.golden"},
		{"odt/content3.xml", 700, IsOdtText, "odt/title3.golden"},
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

		got, err := readContentLimited(file, tt.limit, tt.checker)
		if err != nil {
			t.Fatalf("Failed to read content from %v: %v", tt.document, err)
		}

		want, err := os.ReadFile("testdata/" + tt.golden)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", tt.document, err)
		}

		if !slices.Equal(got, want) {
			t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", tt.document, got, tt.golden, want)
		}
	}
}

func TestReadContentLimited_Error_UnexpectedEOF(t *testing.T) {
	content := "odt/content4.xml"
	golden := "odt/title4.golden"

	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", content, err)
	}
	defer file.Close()

	got, err := readContentLimited(file, 99999, IsOdtText)

	if err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected error reading %s: %v", content, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", golden, err)
	}

	if !slices.Equal(got, want) {
		t.Errorf("\n%s:\n\t%s\n\n%s:\n\t%s\n", content, got, golden, want)
	}
}

func TestReadContentLimited_Error_Syntax(t *testing.T) {
	content := "odt/invalid.xml"
	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", content, err)
	}
	defer file.Close()

	_, err = readContentLimited(file, 700, IsOdtText)

	var syntaxErr *xml.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("readContentLimited() error = %v, want an *xml.SyntaxError", err)
	}
}
