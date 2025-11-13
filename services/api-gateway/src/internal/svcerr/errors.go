package svcerr

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrBadField = errors.New("bad field")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadField)
}
