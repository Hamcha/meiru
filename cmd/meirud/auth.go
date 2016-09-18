package main

import (
	"fmt"

	"github.com/hamcha/meiru/lib/email"
	"github.com/hamcha/meiru/lib/smtp"
)

func HandleLocalAuthRequest(authRequest smtp.ServerAuthRequest) bool {
	address := authRequest.Username

	// Reject invalid addresses
	if !email.IsValidAddress(address) {
		return false
	}

	name, host := email.SplitAddress(address)
	query := fmt.Sprintf("domain:0=%s user:0=%s password", host, name)

	result, err := conf.Query(query)
	if err != nil || len(result) < 1 || len(result[0].Values) < 1 {
		return false
	}

	return checkPassword(result[0].Values, authRequest.Password)
}
