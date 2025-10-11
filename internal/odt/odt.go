package odt

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/trust-me-im-an-engineer/documentreader/internal/runes"
)

const ContentPath = "content.xml"

var spaceRegex = regexp.MustCompile(`\s+`)

// ReadLimited extracts text from reader. Reader expected to be docx's word/document.xml
//
// Returned text is normalized into continious sequence of words separated by single spaces.
// Limit set in runes and count normalized text (so sequence of spaces counts as one rune).
// If text is less than limit, returns all text and [io.ErrUnexpectedEOF].
func ReadLimited(r io.Reader, limitRunes int64) ([]byte, error) {
	decoder := xml.NewDecoder(r)
	text := make([]byte, 0, limitRunes)
	runeLen := int64(0)
	for {
		token, err := decoder.Token()
		if err != nil {
			return text, io.ErrUnexpectedEOF
		}

		startElem, ok := token.(xml.StartElement)

		// Only interested in tokens that are [StartEmement] and have local name "t"
		if !ok { //|| startElem.Name.Space != "text" {
			continue
		}
		if !(startElem.Name.Local == "p" || startElem.Name.Local == "h" || startElem.Name.Local == "span") {
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
