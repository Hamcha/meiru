package email

import "strings"

func SplitAddress(addr string) (string, string) {
	at := strings.LastIndexByte(addr, '@')
	return addr[:at], addr[at+1:]
}

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
