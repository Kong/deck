package kong

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundErr(T *testing.T) {

	assert := assert.New(T)
	var e error = &kongAPIError{httpCode: 404}
	assert.True(IsNotFoundErr(e))
	assert.False(IsNotFoundErr(nil))

	err := errors.New("not a 404")
	assert.False(IsNotFoundErr(err))
}

func TestIsNotFoundErrE2E(T *testing.T) {

	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	consumer, err := client.Consumers.Get(defaultCtx, String("does-not-exists"))
	assert.Nil(consumer)
	assert.NotNil(err)
	assert.True(IsNotFoundErr(err))
}
