# documentreader
Package documentreader implements odt and docx reading.

Documentation could be found on https://pkg.go.dev/github.com/trust-me-im-an-engineer/documentreader

# example
Call documentreader.ReadlimitedODT with file, file size and limit like this:
```go
text, err := documentreader.ReadLimitedODT(file, size, 100)
if err == io.ErrUnexpectedEOF {
  // io.ErrUnexpectedEOF means document text is shorter than runeLimit.
  fmt.Printf("Document was shorter than 100 runes:\n%s", string(text))
} else if err != nil {
  fmt.Fprintln(os.Stderr, "unexpected error reading example.odt: ", err)
} else {
  fmt.Println(string(text))
}
```
Full example on https://pkg.go.dev/github.com/trust-me-im-an-engineer/documentreader#example-ReadLimitedODT
