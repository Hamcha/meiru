package config

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ParseErrorIndentMismatch = errors.New("cfg parse error: indent mismatch")
	ParseErrorUnmatchedQuote = errors.New("cfg parse error: missing ending quote")
)

func parseConfig(configfile string) (Block, error) {
	var block Block
	lines := strings.Split(configfile, "\n")
	scope := []*Block{&block}
	for _, line := range lines {
		// Remove comments (find # without preceding \)
		linecopy := line
		copyoffset := 0
		for {
			i := strings.IndexRune(linecopy, '#')
			if i < 0 {
				break
			}
			if i == 0 || linecopy[i-1] != '\\' {
				line = line[0 : copyoffset+i-1]
				break
			}
			linecopy = line[i+1:]
			copyoffset += i + 1
		}

		// Unescape escaped #
		line = strings.Replace(line, "\\#", "#", -1)

		// Trim space on the right
		line = strings.TrimRightFunc(line, unicode.IsSpace)

		// Skip empty lines
		trimline := strings.TrimSpace(line)
		trimlen := len(trimline)
		if trimlen < 1 {
			continue
		}

		// Check if it contains a block
		isBlock := strings.HasSuffix(trimline, ":")
		if isBlock {
			trimline = strings.TrimRight(trimline, ":")
		}

		// Read indent
		indent := len(line) - trimlen

		// Check for indent mismatch
		if indent >= len(scope) {
			return block, ParseErrorIndentMismatch
		}

		// To avoid scope issues, pop unused scope levels
		if indent < len(scope)-1 {
			scope = scope[:indent+1]
		}

		// Split line into parts
		parts := strings.Split(trimline, " ")

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
					return block, ParseErrorUnmatchedQuote
				}
				if parts[rightIndex][len(parts[rightIndex])-1] == '"' {
					break
				}
			}

			fullAtom := strings.Trim(strings.Join(parts[leftIndex:rightIndex+1], " "), "\"")
			atoms = append(atoms, fullAtom)
			leftIndex = rightIndex
		}

		key := atoms[0]

		// If we have values, add them
		var values []string
		if len(atoms) > 0 {
			values = atoms[1:]
		}

		// Add property to current scope
		*scope[indent] = append(*scope[indent], Property{
			Key:    key,
			Values: values,
		})

		// If we are a block, create it and add it to the scope
		if isBlock {
			index := len(*scope[indent]) - 1
			current := (*scope[indent])[index]
			scope = append(scope, &current.Block)
		}
	}

	return block, nil
}
