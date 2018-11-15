package utils

// Overrider can override properties
// in any type.
type Overrider interface {
	Override(interface{}) (interface{}, error)
}
