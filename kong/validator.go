package kong

// Validator is an interface that
// wraps Valid method.
type Validator interface {
	Valid() bool
}
