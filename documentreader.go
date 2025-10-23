// Package documentreader implements odt and docx reading.
package documentreader

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/trust-me-im-an-engineer/documentreader/internal/runes"
)

const (
	contentPathODT  = "content.xml"
	contentPathDOCX = "word/document.xml"
)

var spaceRegex = regexp.MustCompile(`\s+`)

// readLimitedODT reads text from ODT document.
//
// document expected to be odt; totalSize should match document size in bytes; limit is set in runes;
//
// Returned text is normalized into continious sequence of words separated by single spaces.
// limitRunes counts normalized text (so sequence of spaces counts as one rune).
//
// Note: text may have single space at the end.
//
// If text length is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimitedODT(document io.ReaderAt, totalSize, limitRunes int64) ([]byte, error) {
	return readLimited(document, totalSize, limitRunes, contentPathODT, isODT)
}

// readLimited reads text from DOCX document.
//
// document expected to be docx; totalSize should match document size in bytes; limit is set in runes;
//
// Returned text is normalized into continious sequence of words separated by single spaces.
// limitRunes counts normalized text (so sequence of spaces counts as one rune).
//
// Note: text may have single space at the end.
//
// If text length is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimitedDOCX(document io.ReaderAt, totalSize, limitRunes int64) ([]byte, error) {
	return readLimited(document, totalSize, limitRunes, contentPathDOCX, isDOCX)
}

func readLimited(document io.ReaderAt, totalSize, limitRunes int64, contentPath string, isText func(xml.StartElement) bool) ([]byte, error) {
	zr, err := zip.NewReader(document, totalSize)
	if err != nil {
		return []byte{}, fmt.Errorf("invalid document: %w", err)
	}

	for _, f := range zr.File {
		if f.Name == contentPath {
			rc, err := f.Open()
			if err != nil {
				return []byte{}, fmt.Errorf("invalid document: %w", err)
			}
			defer rc.Close()

			text, err := readContentLimited(rc, limitRunes, isText)
			if err == io.ErrUnexpectedEOF {
				return text, err
			}
			if err != nil {
				return []byte{}, err
			}

			return text, nil
		}
	}

	return []byte{}, fmt.Errorf("invalid document: %s not found", contentPath)
}

// isODT checks if [xml.StartElement] is one of odt's text tags
func isODT(se xml.StartElement) bool {
	return (se.Name.Local == "p" || se.Name.Local == "h" || se.Name.Local == "span")
}

// isDOCX checks if [xml.StartElement] is docx's text tag
func isDOCX(se xml.StartElement) bool {
	return se.Name.Local == "t"
}

// readContentLimited extracts text from reader.
// Reader expected to be either odt's content.xml or docx's word/document.xml.
//
// Returned text is normalized into continious sequence of words separated by single spaces.
//
// Note: text may have single space at the end.
//
// limitRunes set in runes and counts normalized text (so sequence of spaces counts as one rune).
// If text is less than limit, returns all text and [io.ErrUnexpectedEOF].
func readContentLimited(r io.Reader, limitRunes int64, isText func(xml.StartElement) bool) ([]byte, error) {
	decoder := xml.NewDecoder(r)
	text := make([]byte, 0, limitRunes)
	runeLen := int64(0)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return text, io.ErrUnexpectedEOF
		}
		if err != nil {
			return text, fmt.Errorf("invalid content: %w", err)
		}

		startElem, ok := token.(xml.StartElement)

		// Only interested in tokens that are [StartEmement] and text tag
		// (different for odt and docx, so isText function is used)
		if !ok || !isText(startElem) {
			continue
		}

		paragraph, err := extractText(decoder, startElem)
		if err != nil {
			return text, fmt.Errorf("failed to extract text: %w", err)
		}

		// Normalize into continuous sequence of words separated by single spaces with no spaces at the end
		n := spaceRegex.ReplaceAll(paragraph, []byte(" "))
		normalized := bytes.TrimSpace(n)
		if len(normalized) == 0 {
			continue
		}

		paragraphRuneLen := int64(utf8.RuneCount(normalized))

		if runeLen+paragraphRuneLen >= limitRunes {
			sliced, err := runes.Take(normalized, limitRunes-runeLen)
			if err != nil {
				// Should never happen since size to take is checked and xml validates runes
				panic(err)
			}
			text = append(text, sliced...)
			return text, nil
		}

		text = append(text, normalized...)
		text = append(text, byte(' '))
		runeLen += paragraphRuneLen
	}
}

// extractText recursivelly decodes given start element.
// It's similar to xml decoder.DecodeElement() but with nested text too.
func extractText(decoder *xml.Decoder, start xml.StartElement) ([]byte, error) {
	var buf bytes.Buffer

	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {

		// collect text content
		case xml.CharData:
			buf.Write([]byte(t))
			buf.WriteByte(byte(' '))

		// recursively handle nested elements
		case xml.StartElement:
			child, err := extractText(decoder, t)
			if err != nil {
				return nil, err
			}
			buf.Write(child)

		// return collected text
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return buf.Bytes(), nil
			}
		}
	}
}
