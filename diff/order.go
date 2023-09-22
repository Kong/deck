package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/types"
)

/*
                                       Root
                                         |
         +----------+----------+---------+------------+---------------+
         |          |          |         |            |               |
         v          v          v         v            v               v
L1    Service    RbacRole  Upstream  Certificate  CACertificate  Consumer ---+
      Package        |         |        |     |      |                |      |
        |            v         v        v     |      v                v      |
L2      |        RBACRole   Target     SNI    +-> Service       Credentials  |
        |        Endpoint                         |  |              (7)      |
        |                                         |  |                       |
        |                                         |  |                       |
L3      +---------------------------> Service <---+  +-> Route               |
        |                             Version        |     |                 |
        |                                 |          |     |                 |
        |                                 |          |     v                 |
L4      +----------> Document   <---------+          +-> Plugins  <----------+
*/

// dependencyOrder defines the order in which entities will be synced by decK.
// Entities at the same level are processed concurrently.
// Entities at level n will only be processed after all entities at level n-1
// have been processed.
// The processing order for create and update stage is top-down while that
// for delete stage is bottom-up.
var dependencyOrder = [][]types.EntityType{
	{
		types.ServicePackage,
		types.RBACRole,
		types.Certificate,
		types.CACertificate,
		types.Consumer,
		types.Vault,
		types.KeySet,
	},
	{
		types.ConsumerGroup,
		types.RBACEndpointPermission,
		types.SNI,
		types.Service,
		types.Upstream,
		types.Key,

		types.KeyAuth, types.HMACAuth, types.JWTAuth,
		types.BasicAuth, types.OAuth2Cred, types.ACLGroup,
		types.MTLSAuth,
	},
	{
		types.ServiceVersion,
		types.Route,
		types.Target,
		types.ConsumerGroupConsumer,
		types.ConsumerGroupPlugin,
	},
	{
		types.Plugin,
		types.Document,
	},
}

func order() [][]types.EntityType {
	return deepCopy(dependencyOrder)
}

func reverseOrder() [][]types.EntityType {
	order := deepCopy(dependencyOrder)
	return reverse(order)
}

func reverse(src [][]types.EntityType) [][]types.EntityType {
	src = deepCopy(src)
	i := 0
	j := len(src) - 1
	for i < j {
		temp := src[i]
		src[i] = src[j]
		src[j] = temp
		i++
		j--
	}
	return src
}

func deepCopy(src [][]types.EntityType) [][]types.EntityType {
	res := make([][]types.EntityType, len(src))
	for i := range src {
		res[i] = make([]types.EntityType, len(src[i]))
		copy(res[i], src[i])
	}
	return res
}

func eventsInOrder(events []crud.Event, order [][]types.EntityType) [][]crud.Event {
	// kindToLevel maps a Kind to its level in the order to avoid repeated lookups.
	kindToLevel := make(map[crud.Kind]int)

	// eventsByLevel is a slice of slices of events, where each slice of events is at the same level and can be
	// processed concurrently.
	eventsByLevel := make([][]crud.Event, len(order))

	for _, event := range events {
		level, ok := kindToLevel[event.Kind]
		if !ok {
			level = levelForEvent(event, order)
			kindToLevel[event.Kind] = level
		}

		eventsByLevel[level] = append(eventsByLevel[level], event)
	}

	return eventsByLevel
}

func levelForEvent(event crud.Event, order [][]types.EntityType) int {
	for i, level := range order {
		for _, entityType := range level {
			if event.Kind == crud.Kind(entityType) {
				return i
			}
		}
	}

	// This should never happen.
	return -1
}
