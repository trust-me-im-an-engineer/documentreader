package runes

import (
	"errors"
	"io"
	"unicode/utf8"
)

var ErrInvalidRune = errors.New("invalid utf-8 rune encountered")

// Take returns x runes from []byte.
//
// If data contains less than x runes returns [io.UnexpectedEOF].
// If invalid rune encountered returns [ErrInvalidRune].
// In both cases already read runes returned.
func Take(data []byte, x int) ([]byte, error) {
	result := make([]byte, 0, x)

	currentByteIndex := 0
	runeCount := 0

	for runeCount < x {
		if currentByteIndex >= len(data) {
			return result, io.ErrUnexpectedEOF
		}

		r, size := utf8.DecodeRune(data[currentByteIndex:])
		if r == utf8.RuneError && size == 0 {
			return result, io.ErrUnexpectedEOF
		}
		if r == utf8.RuneError && size == 1 {
			return nil, ErrInvalidRune
		}

		result = append(result, data[currentByteIndex:currentByteIndex+size]...)
		currentByteIndex += size
		runeCount++
	}

	return result, nil
}
