// Package documentreader implements odt and docx reading.
package documentreader

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

const (
	contentPathODT  = "content.xml"
	contentPathDOCX = "word/document.xml"
)

var (
	ErrInvalidDocument = errors.New("invalid document")
	ErrContentNotFound = errors.New("content path not found")

	spaceRegex = regexp.MustCompile(`\s+`)
)

// ReadLimitedODT reads text from ODT document limited by specified number of bytes.
//
// document expected to be odt; totalSize should match document size in bytes; limit is set in bytes.
//
// Text is normalized into continious sequence of words separated by single spaces.
// limit counts normalized text (e.g. sequence of spaces in document would count as one byte).
//
// Note: returned text may be up to 3 bytes shorter than limit without error for the sake of rune integrity.
//
// If text's length is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimitedODT(document io.ReaderAt, totalSize, limit int64) ([]byte, error) {
	return readLimited(document, totalSize, limit, contentPathODT, isODT)
}

// ReadLimitedDOCX reads text from DOCX document limited by specified number of bytes.
//
// document expected to be docx; totalSize should match document size in bytes; limit is set in bytes.
//
// Text is normalized into continious sequence of words separated by single spaces.
// limit counts normalized text (e.g. sequence of spaces in document would count as one byte).
//
// Note: returned text may be up to 3 bytes shorter than limit without error for the sake of rune integrity.
//
// If text's length is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimitedDOCX(document io.ReaderAt, totalSize, limit int64) ([]byte, error) {
	return readLimited(document, totalSize, limit, contentPathDOCX, isDOCX)
}

func readLimited(document io.ReaderAt, totalSize, limit int64, contentPath string, isText func(xml.StartElement) bool) ([]byte, error) {
	zr, err := zip.NewReader(document, totalSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}

	for _, f := range zr.File {
		if f.Name == contentPath {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
			}
			defer rc.Close()

			text, err := readContentLimited(rc, limit, isText)
			if err == io.ErrUnexpectedEOF {
				return text, err
			}
			if err != nil {
				return nil, err
			}

			return text, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrContentNotFound, contentPath)
}

// isODT checks if [xml.StartElement] is one of odt's text tags
func isODT(se xml.StartElement) bool {
	return (se.Name.Local == "p" || se.Name.Local == "h" || se.Name.Local == "span")
}

// isDOCX checks if [xml.StartElement] is docx's text tag
func isDOCX(se xml.StartElement) bool {
	return se.Name.Local == "t"
}

func readContentLimited(r io.Reader, limit int64, isText func(xml.StartElement) bool) ([]byte, error) {
	decoder := xml.NewDecoder(r)
	text := make([]byte, 0, limit)
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

		// Check if text is long enough and slice oversized part
		if int64(len(text))+int64(len(normalized)) >= limit {
			text = append(text, normalized[:limit-int64(len(text))]...)
			return trimIncompleteRune(text), nil
		}

		text = append(text, normalized...)
		text = append(text, byte(' '))
	}
}

// extractText recursivelly decodes given start element.
// It's similar to xml decoder.DecodeElement() but with nested text too.
func extractText(decoder *xml.Decoder, start xml.StartElement) ([]byte, error) {
	var buf bytes.Buffer

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				// Unexpected EOF — document structure broken
				return nil, fmt.Errorf("%w: unexpected EOF inside <%s>", ErrInvalidDocument, start.Name.Local)
			}
			// Wrap all other tokenization errors
			return nil, fmt.Errorf("%w: failed to read token inside <%s>: %v", ErrInvalidDocument, start.Name.Local, err)
		}

		switch t := tok.(type) {
		case xml.CharData:
			buf.Write([]byte(t))
			buf.WriteByte(' ')

		case xml.StartElement:
			child, err := extractText(decoder, t)
			if err != nil {
				return nil, fmt.Errorf("%w: nested element <%s>: %v", ErrInvalidDocument, t.Name.Local, err)
			}
			buf.Write(child)

		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return buf.Bytes(), nil
			}
		}
	}
}

func trimIncompleteRune(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	// Check the last rune
	i := len(b)
	for i > 0 && (b[i-1]&0xC0) == 0x80 {
		// It's a UTF-8 continuation byte (10xxxxxx)
		i--
	}

	if i == 0 {
		// All bytes were continuation bytes — invalid start
		return []byte{}
	}

	r, _ := utf8.DecodeRune(b[i-1:])
	if r == utf8.RuneError {
		// Invalid/incomplete rune at the end
		return b[:i-1]
	}

	return b
}
