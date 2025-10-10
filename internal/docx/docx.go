package docx

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/trust-me-im-an-engineer/documentreader/internal/runes"
)

const ContentPath = "word/document.xml"

var spaceRegex = regexp.MustCompile(`\s+`)

func TextLimited(r io.Reader, limit int) ([]byte, error) {
	decoder := xml.NewDecoder(r)
	text := make([]byte, 0, limit)
	runeCount := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return text, io.ErrUnexpectedEOF
		}

		startElem, ok := token.(xml.StartElement)

		// Only interested in tokens that are [StartEmement] and have local name "t"
		if !ok || startElem.Name.Local != "t" {
			continue
		}

		paragraph := []byte{}
		if err := decoder.DecodeElement(&paragraph, &startElem); err != nil {
			panic(err)
		}

		// Normalize into continuous sequence of words separated by single spaces with no spaces at the end
		n := spaceRegex.ReplaceAll(paragraph, []byte(" "))
		normalized := bytes.TrimSpace(n)

		runeLen := utf8.RuneCount(normalized)

		if runeCount+runeLen >= limit {
			sliced, err := runes.Take(normalized, limit-runeCount)
			if err != nil {
				// Should never happen since size to take is checked and xml validates runes
				panic(err)
			}
			text = append(text, sliced...)
			return text, nil
		}

		text = append(text, normalized...)
		text = append(text, byte(' '))
	}
}
