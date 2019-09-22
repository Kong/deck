package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	state, err := NewKongState()
	assert := assert.New(t)
	assert.Nil(err)
	assert.NotNil(state)
}

func state() *KongState {
	s, err := NewKongState()
	if err != nil {
		panic(err)
	}
	return s
}
