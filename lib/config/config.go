package config

import (
	"errors"
	"io/ioutil"
	"strings"
	"unicode"
)

/*

An attempt to make a relodable and flexible yet dead simple configuration file format

Example config:

bind localhost local.domain 127.0.0.1

default:
	box /mail/:name
	limit 100M

user admin:
	limit none

user ext:
	@include users/ext.conf

@include rest.conf

*/

type Config struct {
	Data Block
}

type Block map[string]Property

type Property struct {
	Values []string
	Block  Block
}

type QueryResult struct {
	Single   string
	Property *Property
}

var (
	ErrPIndentMismatch  = errors.New("cfg parse error: indent mismatch")
	ErrPUnmatchedQuote  = errors.New("cfg parse error: missing ending quote")
	ErrQNotFound        = errors.New("query cfg error: property not found")
	ErrQDifferentFormat = errors.New("query cfg error: format mismatch")
)

func LoadConfig(path string) (Config, error) {
	cfg := Config{
		Data: make(Block),
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	lines := strings.Split(string(data), "\n")
	scope := []Block{cfg.Data}
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
			return cfg, ErrPIndentMismatch
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

			rightIndex := leftIndex + 1
			currentAtom := parts[leftIndex]
			for ; ; rightIndex++ {
				if rightIndex >= len(parts) {
					return cfg, ErrPUnmatchedQuote
				}
				currentAtom += " " + parts[rightIndex]
				if parts[rightIndex][len(parts[rightIndex])-1] == '"' {
					break
				}
			}

			currentAtom = strings.Trim(currentAtom, "\"")
			atoms = append(atoms, currentAtom)
			leftIndex = rightIndex
		}

		// Get key (first atom)
		key := atoms[0]

		// Create property
		property := Property{}

		// If we have values, add them
		if len(atoms) > 0 {
			property.Values = atoms[1:]
		}

		// If we are a block, create it and add it to the scope
		if isBlock {
			property.Block = make(Block)
			scope = append(scope, property.Block)
		}

		// Add property to current scope
		scope[indent][key] = property
	}

	return cfg, nil
}

func (cfg Config) Query(path string) (QueryResult, error) {
	/*
		parts := strings.Split(path, " ")

		curNode := cfg.Data
		for _, v := range parts {

		}
	*/
	return QueryResult{}, ErrQNotFound
}

func (cfg Config) QuerySingle(path string) (string, error) {
	_, err := cfg.Query(path)
	return "", err
}
