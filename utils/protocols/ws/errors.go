package ws

import "fmt"

const prefix = "ws:"

type ErrCode int

const (
	ErrWriteMsg ErrCode = iota
	ErrReadMsg
)

type wsError struct {
	code ErrCode
	msg  string
}

func (e *wsError) Error() string {
	return fmt.Sprintf("%s: %s", prefix, e.msg)
}

func NewError(code ErrCode, msg string) *wsError {
	return &wsError{
		code: code,
		msg:  msg,
	}
}
