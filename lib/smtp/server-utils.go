package smtp

import (
	"bytes"

	"github.com/hamcha/meiru/lib/errors"
)

var (
	ServerErrInvalidAuthPlainString = errors.NewType(ErrSrcServer, "Invalid or malformed AUTH PLAIN string")
)

func decodePlainResponse(resp []byte) (string, string, error) {
	fields := bytes.Split(resp, []byte{0})
	if len(fields) < 3 {
		return "", "", errors.NewError(ServerErrInvalidAuthPlainString)
	}
	return string(fields[1]), string(fields[2]), nil
}
