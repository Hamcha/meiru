package email

import "strings"

type Email struct {
	Headers string
	Body    string
}

func Parse(data string) Email {
	// Header/Body separator is the first sequence of two newlines
	rnIndex := strings.Index(data, "\r\n\r\n")
	if rnIndex < 0 {
		// No \r\n? Maybe it's just \n's
		rnIndex = strings.Index(data, "\n\n")
		if rnIndex < 0 {
			// No, just headers, no body.
			rnIndex = len(data)
		}
	}

	return Email{
		Headers: data[:rnIndex],
		Body:    data[rnIndex:],
	}
}
