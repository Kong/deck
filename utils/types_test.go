package utils

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrArrayString(t *testing.T) {
	assert := assert.New(t)
	var err ErrArray
	assert.Equal("nil", err.Error())

	err.Errors = append(err.Errors, errors.New("foo failed"))

	assert.Equal(err.Error(), "1 errors occured:\n\tfoo failed\n")

	err.Errors = append(err.Errors, errors.New("bar failed"))

	assert.Equal(err.Error(), "2 errors occured:\n\tfoo failed\n\tbar failed\n")
}
