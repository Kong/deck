package utils

import (
	"github.com/google/uuid"
)

// UUID will generate a random v4 unique identifier
func UUID() string {
	return uuid.NewString()
}
