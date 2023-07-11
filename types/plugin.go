package types

import (
	"context"
	"errors"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// pluginCRUD implements crud.Actions interface.
type pluginCRUD struct {
	client *kong.Client
}

// kong and konnect APIs only require IDs for referenced entities.
func stripPluginReferencesName(plugin *state.Plugin) {
	if plugin.Plugin.Service != nil && plugin.Plugin.Service.Name != nil {
		plugin.Plugin.Service.Name = nil
	}
	if plugin.Plugin.Route != nil && plugin.Plugin.Route.Name != nil {
		plugin.Plugin.Route.Name = nil
	}
	if plugin.Plugin.Consumer != nil && plugin.Plugin.Consumer.Username != nil {
		plugin.Plugin.Consumer.Username = nil
	}
	if plugin.Plugin.ConsumerGroup != nil && plugin.Plugin.ConsumerGroup.Name != nil {
		plugin.Plugin.ConsumerGroup.Name = nil
	}
}

func pluginFromStruct(arg crud.Event) *state.Plugin {
	plugin, ok := arg.Obj.(*state.Plugin)
	if !ok {
		panic("unexpected type, expected *state.Plugin")
	}
	stripPluginReferencesName(plugin)
	return plugin
}

// Create creates a Plugin in Kong.
// The arg should be of type crud.Event, containing the plugin to be created,
// else the function will panic.
// It returns a the created *state.Plugin.
func (s *pluginCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	plugin := pluginFromStruct(event)

	createdPlugin, err := s.client.Plugins.Create(ctx, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	return &state.Plugin{Plugin: *createdPlugin}, nil
}

// Delete deletes a Plugin in Kong.
// The arg should be of type crud.Event, containing the plugin to be deleted,
// else the function will panic.
// It returns a the deleted *state.Plugin.
func (s *pluginCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	plugin := pluginFromStruct(event)
	err := s.client.Plugins.Delete(ctx, plugin.ID)
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

// Update updates a Plugin in Kong.
// The arg should be of type crud.Event, containing the plugin to be updated,
// else the function will panic.
// It returns a the updated *state.Plugin.
func (s *pluginCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	plugin := pluginFromStruct(event)

	updatedPlugin, err := s.client.Plugins.Create(ctx, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	return &state.Plugin{Plugin: *updatedPlugin}, nil
}

type pluginDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *pluginDiffer) Deletes(handler func(crud.Event) error) error {
	currentPlugins, err := d.currentState.Plugins.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching plugins from state: %w", err)
	}

	for _, plugin := range currentPlugins {
		n, err := d.deletePlugin(plugin)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *pluginDiffer) deletePlugin(plugin *state.Plugin) (*crud.Event, error) {
	plugin = &state.Plugin{Plugin: *plugin.DeepCopy()}
	name := *plugin.Name
	serviceID, routeID, consumerID, consumerGroupID := foreignNames(plugin)
	_, err := d.targetState.Plugins.GetByProp(
		name, serviceID, routeID, consumerID, consumerGroupID,
	)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  plugin,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up plugin %q: %w", *plugin.ID, err)
	}
	return nil, nil
}

func (d *pluginDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetPlugins, err := d.targetState.Plugins.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching plugins from state: %w", err)
	}

	for _, plugin := range targetPlugins {
		n, err := d.createUpdatePlugin(plugin)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *pluginDiffer) createUpdatePlugin(plugin *state.Plugin) (*crud.Event, error) {
	plugin = &state.Plugin{Plugin: *plugin.DeepCopy()}
	name := *plugin.Name
	serviceID, routeID, consumerID, consumerGroupID := foreignNames(plugin)
	currentPlugin, err := d.currentState.Plugins.GetByProp(
		name, serviceID, routeID, consumerID, consumerGroupID,
	)
	if errors.Is(err, state.ErrNotFound) {
		// plugin not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  plugin,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up plugin %q: %w",
			*plugin.Name, err)
	}
	currentPlugin = &state.Plugin{Plugin: *currentPlugin.DeepCopy()}
	// found, check if update needed

	if !currentPlugin.EqualWithOpts(plugin, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    plugin,
			OldObj: currentPlugin,
		}, nil
	}
	return nil, nil
}

func foreignNames(p *state.Plugin) (serviceID, routeID, consumerID, consumerGroupID string) {
	if p == nil {
		return
	}
	if p.Service != nil && p.Service.ID != nil {
		serviceID = *p.Service.ID
	}
	if p.Route != nil && p.Route.ID != nil {
		routeID = *p.Route.ID
	}
	if p.Consumer != nil && p.Consumer.ID != nil {
		consumerID = *p.Consumer.ID
	}
	if p.ConsumerGroup != nil && p.ConsumerGroup.ID != nil {
		consumerGroupID = *p.ConsumerGroup.ID
	}
	return
}
