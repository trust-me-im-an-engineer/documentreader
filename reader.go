package documentread

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/trust-me-im-an-engineer/documentreader/internal/docx"
)

// var ErrUnsupportedFormat = errors.New("unsupported document format (only .docx and .odt are supported)")

func ReadLimitedDocx(document io.ReaderAt, totalSize, limit int64) ([]byte, error) {
	zr, err := zip.NewReader(document, totalSize)
	if err != nil {
		return []byte{}, fmt.Errorf("invalid document: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == docx.ContentPath {
			rc, err := f.Open()
			if err != nil {
				return []byte{}, fmt.Errorf("invalid document: %v", err)
			}
			defer rc.Close()

			text, err := docx.ReadLimited(rc, limit)
			if err != nil {
				return []byte{}, err
			}

			return text, nil
		}
	}

	return []byte{}, fmt.Errorf("invalid document: %s not found", docx.ContentPath)
}

func ReadLimitedOdt(document io.ReaderAt, totalSize, limit int64) ([]byte, error) {
	zr, err := zip.NewReader(document, totalSize)
	if err != nil {
		return []byte{}, fmt.Errorf("invalid document: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == docx.ContentPath {
			rc, err := f.Open()
			if err != nil {
				return []byte{}, fmt.Errorf("invalid document: %v", err)
			}
			defer rc.Close()

			text, err := docx.ReadLimited(rc, limit)
			if err != nil {
				return []byte{}, err
			}

			return text, nil
		}
	}

	return []byte{}, fmt.Errorf("invalid document: %s not found", docx.ContentPath)
}