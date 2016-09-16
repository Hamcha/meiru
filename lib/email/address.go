package email

import "strings"

func IsValidAddress(addr string) bool {
	at := strings.LastIndexByte(addr, '@')
	if at < 0 {
		return false
	}

	if len(addr) < 2 || at == len(addr)-1 {
		return false
	}

	return true
}
