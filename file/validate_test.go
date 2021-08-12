package file

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func Test_checkDefaults(t *testing.T) {
	serviceID := "serviceID"
	host := "host"
	name := "name"
	port := 45
	id := "id"
	target := "target"

	c := &Content{
		Routes: []FRoute{
			Route: kong.Route{
				Name: &name,
				ID:   &id,
			},
		},

		Services: []FService{
			Service: kong.Service{
				ID:   &serviceID,
				Host: &host,
				Name: &name,
				Port: &port,
			},
		},

		Upstream: []FUpstream{
			UPstream: kong.FUpstream{
				ID:   &id,
				Name: &name,
			},
		},
	},

	table := []struct {
		Content     *c
		unbexpected error
	}{
		{
			Content:     c,
			unbexpected: nil,
		},
	}

	for _, entry := range table {
		err := checkDefaults(*entry.Content)
		assert.NotEqual(t, err, entry.unbexpected)
	}
}

func Test_Check(t *testing.T) {
	serviceID := "service_id"
	table := []struct {
		val1     string
		val2     *string
		expected string
	}{
		{
			val1:     serviceID,
			val2:     &serviceID,
			expected: serviceID,
		},
		{
			val1:     "",
			val2:     &serviceID,
			expected: "",
		},
		{
			val1:     serviceID,
			val2:     nil,
			expected: "",
		},
	}

	for _, entry := range table {
		res := check(entry.val1, entry.val2)
		assert.Equal(t, res, entry.expected)
	}
}
