package main

import (
	"fmt"

	"github.com/hamcha/meiru/lib/email"
)

func HandleLocalAuthRequest(username, password string) bool {
	// Reject invalid addresses
	if !email.IsValidAddress(username) {
		return false
	}

	name, host := email.SplitAddress(username)
	query := fmt.Sprintf("domain:0=%s user:0=%s password", host, name)

	result, err := conf.Query(query)
	if err != nil || len(result) < 1 || len(result[0].Values) < 1 {
		return false
	}

	return checkPassword(result[0].Values, password)
}
