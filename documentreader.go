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
	DocxContentPath = "word/document.xml"
	OdtContentPath  = "content.xml"
)

var spaceRegex = regexp.MustCompile(`\s+`)

// ReadLimited reads text from odt or docx document.
//
// ReaderAt expected to be odt or docx; totalSize should match document size in bytes; limit is set in runes;
// contentPath is path to content inside document; isText is type specific checker (use IsOdtText or IsDocxText).
//
// Returned text is normalized into continious sequence of words separated by single spaces.
// Note that text may end with space.
// limitRunes counts normalized text (so sequence of spaces counts as one rune).
//
// If text length is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimited(document io.ReaderAt, totalSize, limitRunes int64, contentPath string, isText func(xml.StartElement) bool) ([]byte, error) {
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

// IsOdtText checks if [xml.StartElement] is one of odt's text tags
func IsOdtText(se xml.StartElement) bool {
	return (se.Name.Local == "p" || se.Name.Local == "h" || se.Name.Local == "span")
}

// IsDocxText checks if [xml.StartElement] is docx's text tag
func IsDocxText(se xml.StartElement) bool {
	return se.Name.Local == "t"
}

// readContentLimited extracts text from reader.
// Reader expected to be either odt's content.xml or docx's word/document.xml.
//
// Returned text is normalized into continious sequence of words separated by single spaces.
// Note that text may end with space.
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

		paragraph := []byte{}
		if err := decoder.DecodeElement(&paragraph, &startElem); err != nil {
			panic(err)
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
