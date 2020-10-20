package kong

import (
	"fmt"
)

type kongAPIError struct {
	httpCode int
	message  string
}

func (e *kongAPIError) Error() string {
	return fmt.Sprintf("HTTP status %d (message: %q)", e.httpCode, e.message)
}

// IsNotFoundErr returns true if the error or it's cause is
// a 404 response from Kong.
func IsNotFoundErr(e error) bool {
	switch e := e.(type) {
	case *kongAPIError:
		return e.httpCode == 404
	default:
		return false
	}
}
