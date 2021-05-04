package reset

import (
	"context"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Reset deletes all entities in Kong.
func Reset(ctx context.Context, state *utils.KongRawState, client *kong.Client) error {
	if state == nil {
		return errors.New("state cannot be empty")
	}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		// Delete routes before services
		for _, r := range state.Routes {
			err := client.Routes.Delete(ctx, r.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		for _, c := range state.Consumers {
			err := client.Consumers.Delete(ctx, c.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		// Upstreams also removes Targets
		for _, u := range state.Upstreams {
			err := client.Upstreams.Delete(ctx, u.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		for _, u := range state.CACertificates {
			err := client.CACertificates.Delete(nil, u.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		for _, p := range state.Plugins {
			// Delete global plugins explicitly since those will not
			// DELETE ON CASCADE
			if p.Consumer == nil && p.Service == nil &&
				p.Route == nil {
				err := client.Plugins.Delete(ctx, p.ID)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	// Routes must be delted before services can be deleted
	for _, s := range state.Services {
		err := client.Services.Delete(nil, s.ID)
		if err != nil {
			return err
		}
	}

	// Services must be deleted before certificates can be deleted
	// Certificates also removes SNIs
	for _, u := range state.Certificates {
		err := client.Certificates.Delete(nil, u.ID)
		if err != nil {
			return err
		}
	}

	// Deleting RBAC roles also deletes their associated permissions
	for _, r := range state.RBACRoles {
		err := client.RBACRoles.Delete(nil, r.ID)
		if err != nil {
			return err
		}
	}

	// TODO handle custom entities
	return nil
}
