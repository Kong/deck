package kong

type err404 struct {
}

func (e err404) Error() string {
	return "Not found"
}

// IsNotFoundErr returns true if the error or it's cause is
// a 404 response from Kong.
func IsNotFoundErr(e error) bool {
	if e == nil {
		return false
	}
	_, ok := e.(err404)
	return ok
}
