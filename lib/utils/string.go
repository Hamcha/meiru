package utils

import (
	"errors"
	"strings"
)

var (
	ErrSplitUnmatchedQuote = errors.New("str split error: unmatched quote")
)

// SplitQuotes parses a space-separated line allowing spaces inside quotes
func SplitQuotes(str string) ([]string, error) {
	// Split string into parts
	parts := strings.Split(strings.TrimSpace(str), " ")

	// Merge parts contained within quotes
	var atoms []string
	for leftIndex := 0; leftIndex < len(parts); leftIndex++ {
		if parts[leftIndex][0] != '"' {
			atoms = append(atoms, parts[leftIndex])
			continue
		}

		rightIndex := leftIndex
		for ; ; rightIndex++ {
			if rightIndex >= len(parts) {
				return parts, ErrSplitUnmatchedQuote
			}
			if parts[rightIndex][len(parts[rightIndex])-1] == '"' {
				break
			}
		}

		fullAtom := strings.Trim(strings.Join(parts[leftIndex:rightIndex+1], " "), "\"")
		atoms = append(atoms, fullAtom)
		leftIndex = rightIndex
	}

	return atoms, nil
}
