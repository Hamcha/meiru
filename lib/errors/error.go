package errors

import (
	"fmt"
	"strings"
)

type ErrorSource string

type Error struct {
	Type      *ErrorType
	ExtraInfo []string
	SubError  error
}

type ErrorType struct {
	Source  ErrorSource
	Message string
}

func NewType(errsrc ErrorSource, errmsg string) *ErrorType {
	return &ErrorType{
		Source:  errsrc,
		Message: errmsg,
	}
}

func NewError(errtype *ErrorType) *Error {
	return &Error{
		Type:      errtype,
		ExtraInfo: nil,
		SubError:  nil,
	}
}

func (a *Error) Error() string {
	extra := make([]string, len(a.ExtraInfo))
	if a.ExtraInfo != nil {
		copy(extra, a.ExtraInfo)
	}
	if a.SubError != nil {
		extra = append(extra, "Underlying error: "+a.SubError.Error())
	}
	extraMsg := ""
	if len(extra) > 0 {
		extraMsg = "\n\t" + strings.Join(extra, "\n\t")
	}
	return fmt.Sprintf("%s err: %s%s", a.Type.Source, a.Type.Message, extraMsg)
}

func (a *Error) WithError(err error) *Error {
	a.SubError = err
	return a
}

func (a *Error) WithInfo(extraFmt string, args ...interface{}) *Error {
	a.ExtraInfo = append(a.ExtraInfo, fmt.Sprintf(extraFmt, args...))
	return a
}
