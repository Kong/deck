package reset

import (
	"errors"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/utils"
)

// Reset deletes all entities in Kong.
func Reset(state *utils.KongRawState, client *kong.Client) error {

	if state == nil {
		return errors.New("state cannot be empty")
	}
	// Delete routes before services
	for _, r := range state.Routes {
		err := client.Routes.Delete(nil, r.ID)
		if err != nil {
			return err
		}
	}
	for _, s := range state.Services {
		err := client.Services.Delete(nil, s.ID)
		if err != nil {
			return err
		}
	}
	// TODO uncomment as development progresses
	// for _, c := range state.Consumers {
	// 	err := client.Consumers.Delete(nil, c.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// for _, u := range state.Upstreams {
	// 	err := client.Consumers.Delete(nil, u.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// for _, u := range state.Certificates {
	// 	err := client.Consumers.Delete(nil, u.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// for _, p := range state.Plugins {
	// 	// Delete global plugins
	// 	if p.APIID == nil && p.ConsumerID == nil && p.ServiceID == nil &&
	// 		p.RouteID == nil {
	// 		err := client.Plugins.Delete(nil, p.ID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// Certificates will delete SNIs
	// Plugins will be removed, except Global plugins
	// Upstreams will remove Targets
	// TODO handle custom entities
	return nil
}
