package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deletePlugins() error {
	currentPlugins, err := sc.currentState.Plugins.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching plugins from state")
	}

	for _, plugin := range currentPlugins {
		n, err := sc.deletePlugin(plugin)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deletePlugin(plugin *state.Plugin) (*Event, error) {
	plugin = &state.Plugin{Plugin: *plugin.DeepCopy()}
	name := *plugin.Name
	serviceID, routeID, consumerID := foreignNames(plugin)
	_, err := sc.targetState.Plugins.GetByProp(name, serviceID, routeID,
		consumerID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "plugin",
			Obj:  plugin,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up plugin '%v'", *plugin.ID)
	}
	return nil, nil
}

func (sc *Syncer) createUpdatePlugins() error {
	targetPlugins, err := sc.targetState.Plugins.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching plugins from state")
	}

	for _, plugin := range targetPlugins {
		n, err := sc.createUpdatePlugin(plugin)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdatePlugin(plugin *state.Plugin) (*Event, error) {
	plugin = &state.Plugin{Plugin: *plugin.DeepCopy()}
	name := *plugin.Name
	serviceID, routeID, consumerID := foreignNames(plugin)
	currentPlugin, err := sc.currentState.Plugins.GetByProp(name,
		serviceID, routeID, consumerID)
	if err == state.ErrNotFound {
		// plugin not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "plugin",
			Obj:  plugin,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up plugin %v",
			*plugin.Name)
	}
	currentPlugin = &state.Plugin{Plugin: *currentPlugin.DeepCopy()}
	// found, check if update needed

	if !currentPlugin.EqualWithOpts(plugin, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "plugin",
			Obj:    plugin,
			OldObj: currentPlugin,
		}, nil
	}
	return nil, nil
}

func foreignNames(p *state.Plugin) (serviceID, routeID, consumerID string) {
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
	return
}
