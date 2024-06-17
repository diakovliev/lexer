package lexer

import (
	"errors"
	"io"
	"unicode/utf8"
)

// NextRuneFrom reads the next rune from the given io.Reader and returns it along with the number of bytes read and any error encountered.
//
// Parameters:
// - reader: The io.Reader from which to read the rune.
//
// Returns:
// - ret: The rune read from the reader.
// - w: The number of bytes read from the reader.
// - err: An error if there was a problem reading the rune.
func NextRuneFrom(reader io.Reader) (ret rune, w int, err error) {
	fetch := newFetcher(reader)
	defer fetch.Reset()
	for i := 1; i < 5; i++ {
		_, err = fetch.Fetch(i)
		if i == 1 && errors.Is(err, io.EOF) {
			ret = 0
			break
		}
		ret, w = utf8.DecodeRune(fetch.Buffer())
		if err != nil || ret != utf8.RuneError {
			break
		}
	}
	w = fetch.Len()
	return
}

// NextRunesFrom reads the next n runes from the given io.Reader and returns them along with the number of bytes read and any error encountered.
//
// Parameters:
// - reader: The io.Reader from which to read the runes.
// - n: The number of runes to read.
//
// Returns:
// - ret: The runes read from the reader.
// - w: The number of runes read from the reader.
// - err: An error if there was a problem reading the runes.
func NextRunesFrom(reader io.Reader, n int) (ret []rune, w int, err error) {
	ret = make([]rune, 0, n)
	for len(ret) < n {
		r, rw, rErr := NextRuneFrom(reader)
		if errors.Is(rErr, io.EOF) && rw == 0 && r == 0 {
			err = rErr
			return
		}
		w += rw
		if r == utf8.RuneError {
			err = newErrInvalidRune(w-rw, w)
			return
		}
		ret = append(ret, r)
		if rErr != nil {
			err = rErr
			return
		}
	}
	return
}

// ReadRunesFromUntil reads runes from the given io.Reader until the until function returns true.
//
// Parameters:
// - reader: The io.Reader from which to read the runes.
// - until: The function that determines when to stop reading runes. It takes a rune as input and returns a boolean.
//
// Returns:
// - ret: The slice of runes read from the reader.
// - w: The total number of bytes read from the reader.
// - err: An error if there was a problem reading the runes.
func ReadRunesFromUntil(reader io.Reader, until func(r rune) bool) (ret []rune, w int, err error) {
	for {
		r, rw, rErr := NextRuneFrom(reader)
		if errors.Is(rErr, io.EOF) && rw == 0 && r == 0 {
			err = rErr
			return
		}
		w += rw
		if r == utf8.RuneError {
			err = newErrInvalidRune(w-rw, w)
			return
		}
		ret = append(ret, r)
		if rErr != nil {
			err = rErr
			return
		}
		if until(r) {
			return
		}
	}
}

// ReadRunesFromUntilRunes reads runes from the given io.Reader until the until function returns true.
//
// Parameters:
//   - reader: The io.Reader from which to read the runes.
//   - until: A function that takes a slice of runes and returns a boolean.
//     The function is called with each slice of runes read from the reader.
//     If the function returns true, the reading process is stopped and the function returns the current slice of runes and the number of runes read.
//
// Returns:
// - ret: The slice of runes read from the reader until the until function returns true.
// - w: The number of runes read from the reader.
// - err: An error if there was a problem reading the runes.
func ReadRunesFromUntilRunes(reader io.Reader, until func(r []rune) bool) (ret []rune, w int, err error) {
	for {
		r, rw, rErr := NextRuneFrom(reader)
		if errors.Is(rErr, io.EOF) && rw == 0 && r == 0 {
			err = rErr
			return
		}
		w += rw
		if r == utf8.RuneError {
			err = newErrInvalidRune(w-rw, w)
			return
		}
		ret = append(ret, r)
		if rErr != nil {
			err = rErr
			return
		}
		if until(ret) {
			return
		}
	}
}
