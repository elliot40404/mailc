package util

import (
	"unicode"
	"unicode/utf8"
)

// UpperFirst uppercases the first rune of the provided string.
func UpperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// LowerFirst lowercases the first rune of the provided string.
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

// MakeExportedName converts an arbitrary string (e.g., filename) into a valid
// exported Go identifier by splitting on non-alphanumeric characters, title‑
// casing each chunk, and concatenating them. If the result starts with a
// non-letter, it is prefixed with 'X'.
func MakeExportedName(s string) string {
	// Build words of letters/digits, splitting on anything else
	words := make([]string, 0, 4)
	current := make([]rune, 0, len(s))
	flush := func() {
		if len(current) == 0 {
			return
		}
		// Uppercase first, keep rest as-is (filenames are usually lowercase)
		r0 := unicode.ToUpper(current[0])
		word := string(r0) + string(current[1:])
		words = append(words, word)
		current = current[:0]
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
		} else {
			flush()
		}
	}
	flush()
	// Join words
	out := ""
	for _, w := range words {
		out += w
	}
	if out == "" {
		return "X"
	}
	// Ensure starts with a letter
	r, _ := utf8.DecodeRuneInString(out)
	if !unicode.IsLetter(r) {
		out = "X" + out
	}
	return out
}
