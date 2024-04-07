package http

import (
	"errors"
	"fmt"
	xjson "github.com/lnashier/goarc/x/json"
	"net/http"
)

// ConvertError translates ordinary error to Error with http.StatusInternalServerError.
// If err is already of type Error, it is returned without modifications.
func ConvertError(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err)
}

// Error represents an HTTP error with an underlying error cause
type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

// WriteJSON writes the error to ResponseWriter in JSON format.
func (err *Error) WriteJSON(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(err.Status)
	return w.Write(xjson.Marshal(err))
}

// WriteText writes the error to ResponseWriter in text format.
func (err *Error) WriteText(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(err.Status)
	return w.Write([]byte(err.Message))
}

func (err *Error) String() string {
	return err.Error()
}

func (err *Error) Error() string {
	if err.Cause != nil {
		return fmt.Sprintf("%d %s caused by %s", err.Status, err.Message, err.Cause.Error())
	}
	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

func (err *Error) Unwrap() error {
	return err.Cause
}

// NewError builds a new Error instance
func NewError(status int, message string, cause error) *Error {
	return &Error{
		Status:  status,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorf builds a new HTTPError instance.
// Similar usage as NewError.
// The 'message' param is a format specifier.
func NewErrorf(status int, cause error, message string, args ...any) *Error {
	return &Error{
		Status:  status,
		Message: fmt.Sprintf(message, args...),
		Cause:   cause,
	}
}

func Is4xx(err error) (int, bool) {
	herr, ok := err.(*Error)
	if !ok {
		return 0, false
	}
	return herr.Status, herr.Status >= 400 && herr.Status < 500
}

func NotFound(err error) error {
	return &Error{
		Status:  http.StatusNotFound,
		Message: http.StatusText(http.StatusNotFound),
		Cause:   err,
	}
}

func NotFoundf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func Conflict(err error) error {
	return &Error{
		Status:  http.StatusConflict,
		Message: http.StatusText(http.StatusConflict),
		Cause:   err,
	}
}

func Conflictf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusConflict,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func BadRequest(err error) error {
	return &Error{
		Status:  http.StatusBadRequest,
		Message: http.StatusText(http.StatusBadRequest),
		Cause:   err,
	}
}

func BadRequestf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusBadRequest,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func UnprocessableEntity(err error) error {
	return &Error{
		Status:  http.StatusUnprocessableEntity,
		Message: http.StatusText(http.StatusUnprocessableEntity),
		Cause:   err,
	}
}

func UnprocessableEntityf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusUnprocessableEntity,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func PreconditionFailed(err error) error {
	return &Error{
		Status:  http.StatusPreconditionFailed,
		Message: http.StatusText(http.StatusPreconditionFailed),
		Cause:   err,
	}
}

func PreconditionFailedf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusPreconditionFailed,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func PreconditionRequired(err error) error {
	return &Error{
		Status:  http.StatusPreconditionRequired,
		Message: http.StatusText(http.StatusPreconditionRequired),
		Cause:   err,
	}
}

func PreconditionRequiredf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusPreconditionRequired,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func Internal(err error) error {
	return &Error{
		Status:  http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Cause:   err,
	}
}

func Internalf(err error, format string, v ...any) error {
	return &Error{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}
