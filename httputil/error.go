package httputil

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func isSuccess(resp *http.Response) bool {
	return resp.StatusCode/100 == 2
}

type httpError struct {
	StatusCode int
	Status     string
	Body       string
}

func (err *httpError) Error() string {
	if err.Body != "" {
		return fmt.Sprintf("%s - %s", err.Status, err.Body)
	}
	return err.Status
}

// ErrorStatusCode returns the status code is it is an HTTP error.
func ErrorStatusCode(err error) int {
	if err == nil {
		return 0
	}
	herr, ok := err.(*httpError)
	if !ok {
		return 0
	}
	return herr.StatusCode
}

// IsNotFound checks if err is an http StatusNotFound error.
func IsNotFound(err error) bool {
	return ErrorStatusCode(err) == http.StatusNotFound
}

// RespError returns the error from an HTTP response.
func RespError(resp *http.Response) error {
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	herr := &httpError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       strings.TrimSpace(string(bs)),
	}
	return herr
}
