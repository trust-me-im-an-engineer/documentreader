package documentreader

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"slices"
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
		{"odt/document1.odt", 1200, contentPathODT, isODT, "odt/title1.golden"},
		{"odt/document3.odt", 1200, contentPathODT, isODT, "odt/title3.golden"},
		{"docx/document1.docx", 1200, contentPathDOCX, isDOCX, "docx/title1.golden"},
		{"docx/document2.docx", 1200, contentPathDOCX, isDOCX, "docx/title2.golden"},
		{"docx/document3.docx", 1400, contentPathDOCX, isDOCX, "docx/title3.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.document, func(t *testing.T) {
			file, err := os.Open("testdata/" + tt.document)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.document, err)
			}
			defer file.Close()

			fi, err := file.Stat()
			if err != nil {
				t.Fatalf("failed to get %s info: %v", tt.document, err)
			}
			size := fi.Size()

			got, err := readLimited(file, size, tt.limit, tt.contentPath, tt.checker)
			if err != nil {
				t.Fatalf("unexpected error reading %s: %v", tt.document, err)
			}

			want, err := os.ReadFile("testdata/" + tt.golden)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.document, err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s\ngot:\n%s\n", diff, got)
			}
		})
	}
}

func TestReadLimited_Error_InvalidDocument(t *testing.T) {
	document := "odt/invalid_format.doc"

	file, err := os.Open("testdata/" + document)
	if err != nil {
		t.Fatalf("failed to read %s: %v", document, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to get %s info: %v", document, err)
	}
	size := fi.Size()

	if _, err := readLimited(file, size, 100, contentPathODT, isODT); !errors.Is(err, ErrContentNotFound) {
		t.Fatalf("expected ErrContentNotFound, got %v", err)
	}
}

func TestReadLimited_Error_UnexpectedEOF(t *testing.T) {
	document := "odt/document4.odt"
	golden := "odt/long4.golden"

	file, err := os.Open("testdata/" + document)
	if err != nil {
		t.Fatalf("failed to read %s: %v", document, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to get %s info: %v", document, err)
	}
	size := fi.Size()

	got, err := readLimited(file, size, 99999, contentPathODT, isODT)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected error reading %s: %v", document, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("failed to read %s: %v", document, err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("output mismatch (-want +got):\n%s\ngot:\n%s\n", diff, got)
	}
}

func TestReadContentLimited_Success(t *testing.T) {
	var tests = []struct {
		document string
		limit    int64
		checker  func(xml.StartElement) bool
		golden   string
	}{
		{"odt/content1.xml", 1200, isODT, "odt/title1.golden"},
		{"odt/content3.xml", 1200, isODT, "odt/title3.golden"},
		{"docx/content1.xml", 1200, isDOCX, "docx/title1.golden"},
		{"docx/content2.xml", 1200, isDOCX, "docx/title2.golden"},
		{"docx/content3.xml", 1400, isDOCX, "docx/title3.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.document, func(t *testing.T) {
			file, err := os.Open("testdata/" + tt.document)
			if err != nil {
				t.Fatalf("failed to open %s: %v", tt.document, err)
			}
			defer file.Close()

			got, err := readContentLimited(file, tt.limit, tt.checker)
			if err != nil {
				t.Fatalf("failed to read content from %v: %v", tt.document, err)
			}

			want, err := os.ReadFile("testdata/" + tt.golden)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.document, err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s\ngot:\n%s\n", diff, got)
			}
		})
	}
}

func TestReadContentLimited_Error_UnexpectedEOF(t *testing.T) {
	content := "odt/content4.xml"
	golden := "odt/long4.golden"

	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("failed to open %s: %v", content, err)
	}
	defer file.Close()

	got, err := readContentLimited(file, 99999, isODT)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected error reading %s: %v", content, err)
	}

	want, err := os.ReadFile("testdata/" + golden)
	if err != nil {
		t.Fatalf("failed to read %s: %v", golden, err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("output mismatch (-want +got):\n%s\ngot:\n%s\n", diff, got)
	}
}

func TestReadContentLimited_Error_Syntax(t *testing.T) {
	content := "odt/invalid.xml"
	file, err := os.Open("testdata/" + content)
	if err != nil {
		t.Fatalf("failed to open %s: %v", content, err)
	}
	defer file.Close()

	_, err = readContentLimited(file, 700, isODT)

	var syntaxErr *xml.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("readContentLimited() error = %v, want an *xml.SyntaxError", err)
	}
}

func TestTrimIncompleteRune(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want []byte
	}{
		{
			name: "empty input",
			in:   []byte{},
			want: []byte{},
		},
		{
			name: "pure ASCII",
			in:   []byte("hello"),
			want: []byte("hello"),
		},
		{
			name: "valid multi-byte UTF-8 string",
			in:   []byte("привет"),
			want: []byte("привет"),
		},
		{
			name: "cut in middle of 2-byte rune (Cyrillic)",
			in:   []byte("привет")[:5],
			want: []byte("привет")[:4],
		},
		{
			name: "cut in middle of 2-byte rune (é incomplete)",
			in:   []byte{0xC3},
			want: []byte{},
		},
		{
			name: "complete 2-byte rune (é complete)",
			in:   []byte{0xC3, 0xA9},
			want: []byte{0xC3, 0xA9},
		},
		{
			name: "cut in middle of 4-byte rune (emoji)",
			in:   []byte("😀")[:2],
			want: []byte{},
		},
		{
			name: "mix of valid and incomplete at end",
			in:   append([]byte("ok"), 0xE3, 0x81),
			want: []byte("ok"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimIncompleteRune(tt.in)

			if !slices.Equal(got, tt.want) {
				t.Errorf("\nname: %s\ngot:  %q\nwant: %q\n", tt.name, got, tt.want)
			}
		})
	}
}
