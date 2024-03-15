package http

import (
	"fmt"
	xjson "github.com/lnashier/goarc/x/json"
	"net/http"
)

// HandleError writes provided error to ResponseWriter in JSON format.
// If error is not of type Error, error is converted to InternalServerError
func HandleError(w http.ResponseWriter, err error) {
	var httpErr *Error
	switch specificError := err.(type) {
	case *Error:
		httpErr = specificError
	default:
		// default for unknown errors is 500 with no client-facing message
		httpErr = NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpErr.Status)
	w.Write(xjson.Marshal(httpErr))
}

// Error represents an HTTP error with an underlying error cause
type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (err Error) String() string {
	return err.Error()
}

func (err Error) Error() string {
	if err.Cause != nil {
		return fmt.Sprintf("%d %s caused by %s", err.Status, err.Message, err.Cause.Error())
	}
	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

func (err Error) Unwrap() error {
	return err.Cause
}

// NewError builds a new http.Error instance
//
// The 'message' parameter is what gets returned in our standard error response.
// Overall, we don't want to give away internal details of most errors;
// but in the case of a 400, it's customary to tell callers what they did wrong.
func NewError(status int, message string, cause error) *Error {
	return &Error{
		Status:  status,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorf builds a new HTTPError instance.
//
// Similar usage as NewError.
//
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

func NotFound(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusNotFound,
		Message: http.StatusText(http.StatusNotFound),
		Cause:   err,
	}
}

func NotFoundf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func Conflict(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusConflict,
		Message: http.StatusText(http.StatusConflict),
		Cause:   err,
	}
}

func Conflictf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusConflict,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func BadRequest(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusBadRequest,
		Message: http.StatusText(http.StatusBadRequest),
		Cause:   err,
	}
}

func BadRequestf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusBadRequest,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func UnprocessableEntity(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusUnprocessableEntity,
		Message: http.StatusText(http.StatusUnprocessableEntity),
		Cause:   err,
	}
}

func UnprocessableEntityf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusUnprocessableEntity,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func PreconditionFailed(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusPreconditionFailed,
		Message: http.StatusText(http.StatusPreconditionFailed),
		Cause:   err,
	}
}

func PreconditionFailedf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusPreconditionFailed,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func PreconditionRequired(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusPreconditionRequired,
		Message: http.StatusText(http.StatusPreconditionRequired),
		Cause:   err,
	}
}

func PreconditionRequiredf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusPreconditionRequired,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}

func Internal(err error) (any, error) {
	return nil, &Error{
		Status:  http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Cause:   err,
	}
}

func Internalf(err error, format string, v ...any) (any, error) {
	return nil, &Error{
		Status:  http.StatusInternalServerError,
		Message: fmt.Sprintf(format, v...),
		Cause:   err,
	}
}
