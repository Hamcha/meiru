package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func checkPassword(passwordData []string, otherPassword string) bool {
	pwdType := "plain"
	originalPassword := ""
	if len(passwordData) > 1 {
		pwdType = strings.ToLower(passwordData[0])
		originalPassword = passwordData[1]
	} else {
		originalPassword = passwordData[0]
	}

	switch pwdType {
	case "plain":
		return originalPassword == otherPassword
	case "sha256":
		shabytes := sha256.Sum256([]byte(otherPassword))
		return originalPassword == hex.EncodeToString(shabytes[:])
	}

	return false
}
