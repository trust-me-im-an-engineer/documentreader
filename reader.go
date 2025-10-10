package documentread

import (
	"archive/zip"
	"errors"
	"io"

	"github.com/trust-me-im-an-engineer/documentreader/internal/docx"
)

var ErrUnsupportedFormat = errors.New("unsupported document format (only .docx and .odt are supported)")

func Read(document io.ReaderAt, totalSize, limit int64) ([]byte, error) {
	zr, err := zip.NewReader(document, totalSize)
	if err != nil {
		return []byte{}, ErrUnsupportedFormat
	}

	for _, f := range zr.File {
		if f.Name == docx.ContentPath {
			rc, err := f.Open()
			defer rc.Close()
			if err != nil {
				return []byte{}, ErrUnsupportedFormat
			}

			if f.Name == docx.ContentPath {
				text, err := docx.TextLimited(rc, limit)
				if err != nil {
					return []byte{}, err
				}

				return text, nil
			}
		}
	}

	return []byte{}, ErrUnsupportedFormat
}
