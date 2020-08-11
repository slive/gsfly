/*
 * Author:slive
 * DATE:2020/8/7
 */
package common

import "errors"

type GError interface {
	error
	GetErrCode() string
	GetErr() error
}

type innerErr struct {
	error   error
	errCode string
}

func NewError0(err error) GError {
	return &innerErr{
		error:   err,
		errCode: "",
	}
}

func NewError1(errCode string, err error) GError {
	return &innerErr{
		error:   err,
		errCode: errCode,
	}
}

func NewError2(errCode string, errStr string) GError {
	return &innerErr{
		error:   errors.New(errStr),
		errCode: errCode,
	}
}

func (err *innerErr) Error() string {
	return err.errCode + ":" + err.error.Error()
}
func (err *innerErr) GetErrCode() string {
	return err.errCode
}

func (err *innerErr) GetErr() error {
	return err.error
}
