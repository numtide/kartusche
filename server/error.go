package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
)

type errorWithCode struct {
	err  error
	code int
}

func newErrorWithCode(err error, code int) *errorWithCode {
	return &errorWithCode{err, code}
}

func (e *errorWithCode) Error() string {
	return fmt.Errorf("code %d: %w", e.code, e.err).Error()
}

func (e *errorWithCode) As(err any) bool {
	er, ok := err.(*errorWithCode)
	if ok {
		e.code = er.code
		e.err = er.err
	}
	return ok
}

func handleHttpError(w http.ResponseWriter, err error, log logr.Logger) {
	if err == nil {
		return
	}

	log.Error(err, "while handling http request")

	ec := &errorWithCode{}

	if errors.As(err, &ec) {
		http.Error(w, ec.Error(), ec.code)
		return
	}

	http.Error(w, err.Error(), 500)

}
