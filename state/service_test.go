package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestBasicService(t *testing.T) {
	assert := assert.New(t)
	state, err := NewKongState()
	assert.Nil(err)
	assert.NotNil(state)

	var service Service
	// service.Name = kong.String("foo")
	service.Name = kong.String("first")
	service.ID = kong.String("first")
	err = state.AddService(service)
	assert.Nil(err)

	se, err := state.GetService("first")
	assert.Nil(err)
	assert.NotNil(se)

	se.Host = kong.String("example.com")
	err = state.UpdateService(*se)
	assert.Nil(err)

	se, err = state.GetService("first")
	assert.Nil(err)
	assert.NotNil(se)
	assert.Equal("example.com", *se.Host)
}
