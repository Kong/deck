package state

import (
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
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

	se, err := state.GetServiceByName("first")
	assert.Nil(err)
	assert.NotNil(se)
}
