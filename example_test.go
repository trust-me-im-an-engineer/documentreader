package documentreader_test

import (
	"fmt"
	"io"
	"os"

	"github.com/trust-me-im-an-engineer/documentreader"
)

// Basic usage of ReadLimitedODT.
// ReadLimitedDOCX is identical with the only difference of expecting DOCX file instead.
func ExampleReadLimitedODT() {
	// One of the ways to get io.ReaderAt is to open file.
	file, err := os.Open("example.odt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read example.odt: ", err)
	}
	defer file.Close()

	// Get file size
	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get file stats: ", err)
	}
	size := fi.Size()

	text, err := documentreader.ReadLimitedODT(file, size, 100)
	if err == io.ErrUnexpectedEOF {
		// io.ErrUnexpectedEOF means document text is shorter than runeLimit.
		fmt.Printf("Document was shorter than 100 runes:\n%s", string(text))
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "unexpected error reading example.odt: ", err)
	} else {
		fmt.Println(string(text))
	}
}
