package smtp

import (
	"bytes"
	"errors"
)

var (
	ServerErrInvalidAuthPlainString = errors.New("Invalid or malformed AUTH PLAIN string")
)

func decodePlainResponse(resp []byte) (string, string, error) {
	fields := bytes.Split(resp, []byte{0})
	if len(fields) < 3 {
		return "", "", ServerErrInvalidAuthPlainString
	}
	return string(fields[1]), string(fields[2]), nil
}
