package main

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var (
	BSErrorEmptySize             = errors.New("byte size passed was empty")
	BSErrorUnknownByteMultiplier = errors.New("unknown byte multiplier")
)

// parseByteSize parses a human readable byte size to its byte count
// ex. 10M -> 10 * 1024 * 1024 -> 10485760
func parseByteSize(size string) (uint64, error) {
	if len(size) < 0 {
		return 0, BSErrorEmptySize
	}

	lastChar := strings.ToUpper(size)[len(size)-1]
	if unicode.IsLetter(rune(lastChar)) {
		num, err := strconv.ParseUint(size[0:len(size)-1], 10, 64)
		if err != nil {
			return 0, err
		}

		const units = "KMGTPE"
		multiplier := strings.IndexByte(units, lastChar)
		if multiplier < 0 {
			return 0, BSErrorUnknownByteMultiplier
		}

		return num << (10 * uint(multiplier)), nil
	} else {
		return strconv.ParseUint(size, 10, 64)
	}
}
