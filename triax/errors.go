package triax

import (
	"errors"
	"fmt"
)

type ErrInvalidEndpoint struct {
	err error
}

func (err *ErrInvalidEndpoint) Error() string {
	return fmt.Sprintf("invalid endpoint: %v", err.err)
}

func (err *ErrInvalidEndpoint) Unwrap() error {
	return err.err
}

var ErrMissingCredentials = errors.New("missing username/password")

type genericError struct{ msg string }

func (err *genericError) Error() string {
	return err.msg
}

type ErrUnexpectedStatus struct {
	Method string
	URL    string
	Status int
	Body   []byte
}

func (err *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status %d for %v %v: %s", err.Status, err.Method, err.URL, err.Body)
}
