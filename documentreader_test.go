package documentreader

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadLimited_Success(t *testing.T) {
	var tests = []struct {
		document    string
		limit       int64
		contentPath string
		checker     func(xml.StartElement) bool
		golden      string
	}{
		{"odt/document1.odt", 700, contentPathODT, isODT, "odt/title1.golden"},
		{"odt/document3.odt", 700, contentPathODT, isODT, "odt/title3.golden"},
		{"docx/document1.docx", 700, contentPathDOCX, isDOCX, "docx/title1.golden"},
		{"docx/document2.docx", 700, contentPathDOCX, isDOCX, "docx/title2.golden"},
		{"docx/document3.docx", 700, contentPathDOCX, isDOCX, "docx/title3.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.document, func(t *testing.T) {
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

			got, err := readLimited(file, size, tt.limit, tt.contentPath, tt.checker)
			if err != nil {
				t.Fatalf("Unexpected error reading %s: %v", tt.document, err)
			}

			want, err := os.ReadFile("testdata/" + tt.golden)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.document, err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Logf("Got:\n%s", got)
				t.Errorf("ReadLimited_Success() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReadLimited_Error_UnexpectedEOF(t *testing.T) {
	document := "odt/document4.odt"
	golden := "odt/long4.golden"

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

	got, err := readLimited(file, size, 99999, contentPathODT, isODT)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected error reading %s: %v", document, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", document, err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadLimited_UnexpectedEOF() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadContentLimited_Success(t *testing.T) {
	var tests = []struct {
		document string
		limit    int64
		checker  func(xml.StartElement) bool
		golden   string
	}{
		{"odt/content1.xml", 700, isODT, "odt/title1.golden"},
		{"odt/content3.xml", 700, isODT, "odt/title3.golden"},
		{"docx/content1.xml", 700, isDOCX, "docx/title1.golden"},
		{"docx/content2.xml", 700, isDOCX, "docx/title2.golden"},
		{"docx/content3.xml", 700, isDOCX, "docx/title3.golden"},
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

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("ReadContentLimited_Success() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestReadContentLimited_Error_UnexpectedEOF(t *testing.T) {
	content := "odt/content4.xml"
	golden := "odt/long4.golden"

	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", content, err)
	}
	defer file.Close()

	got, err := readContentLimited(file, 99999, isODT)

	if err != io.ErrUnexpectedEOF {
		t.Fatalf("Unexpected error reading %s: %v", content, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", golden, err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadContentLimited_UnexpectedEOF() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadContentLimited_Error_Syntax(t *testing.T) {
	content := "odt/invalid.xml"
	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", content, err)
	}
	defer file.Close()

	_, err = readContentLimited(file, 700, isODT)

	var syntaxErr *xml.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("readContentLimited() error = %v, want an *xml.SyntaxError", err)
	}
}
