package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	if utils.Empty(plugin.Name) {
		return nil, errors.New("'name' attribute for a plugin cannot be nil")
	}
	plugin = &state.Plugin{Plugin: *plugin.DeepCopy()}
	if plugin.Service != nil {
		id := ""
		if plugin.Service.Name != nil {
			id = *plugin.Service.Name
		}
		if plugin.Service.ID != nil {
			id = *plugin.Service.ID
		}
		svc, err := sc.currentState.Services.Get(id)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find service '%v' for plugin %+v",
				id, *plugin.Name)
		}
		plugin.Service = &svc.Service
	}
	if plugin.Route != nil {
		id := ""
		if plugin.Route.Name != nil {
			id = *plugin.Route.Name
		}
		if plugin.Route.ID != nil {
			id = *plugin.Route.ID
		}
		r, err := sc.currentState.Routes.Get(id)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find route '%v' for plugin %+v",
				id, *plugin.Name)
		}
		plugin.Route = &r.Route
	}
	if plugin.Consumer != nil {
		id := ""
		if plugin.Consumer.Username != nil {
			id = *plugin.Consumer.Username
		}
		if plugin.Consumer.ID != nil {
			id = *plugin.Consumer.ID
		}
		c, err := sc.currentState.Consumers.Get(id)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not find Consumer '%v' for plugin %+v",
				id, *plugin.Name)
		}
		plugin.Consumer = &c.Consumer
	}
	name := *plugin.Name
	serviceName, routeName, consumerName := foreignNames(plugin)
	_, err := sc.targetState.Plugins.GetByProp(name, serviceName, routeName,
		consumerName)
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
	serviceName, routeName, consumerName := foreignNames(plugin)
	currentPlugin, err := sc.currentState.Plugins.GetByProp(name,
		serviceName, routeName, consumerName)
	if err == state.ErrNotFound {
		// plugin not present, create it

		// XXX fill foreign
		if plugin.Service != nil {
			svc, err := sc.currentState.Services.Get(*plugin.Service.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find service '%v' for plugin %+v",
					*plugin.Service.Name, *plugin.Name)
			}
			plugin.Service = &svc.Service
		}
		if plugin.Route != nil {
			svc, err := sc.currentState.Routes.Get(*plugin.Route.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find route '%v' for plugin %+v",
					*plugin.Route.Name, *plugin.Name)
			}
			plugin.Route = &svc.Route
		}
		if plugin.Consumer != nil {
			svc, err := sc.currentState.Consumers.Get(*plugin.Consumer.Username)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find consumer '%v' for plugin %+v",
					*plugin.Consumer.Username, *plugin.Name)
			}
			plugin.Consumer = &svc.Consumer
		}
		// XXX

		plugin.ID = nil
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

	if currentPlugin.Service != nil {
		currentPlugin.Service = &kong.Service{Name: currentPlugin.Service.Name}
	}
	if plugin.Service != nil {
		plugin.Service = &kong.Service{Name: plugin.Service.Name}
	}
	if currentPlugin.Route != nil {
		currentPlugin.Route = &kong.Route{Name: currentPlugin.Route.Name}
	}
	if plugin.Route != nil {
		plugin.Route = &kong.Route{Name: plugin.Route.Name}
	}
	if currentPlugin.Consumer != nil {
		currentPlugin.Consumer = &kong.Consumer{Username: currentPlugin.Consumer.Username}
	}
	if plugin.Consumer != nil {
		plugin.Consumer = &kong.Consumer{Username: plugin.Consumer.Username}
	}
	if !currentPlugin.EqualWithOpts(plugin, true, true, false) {
		plugin.ID = kong.String(*currentPlugin.ID)

		// XXX fill foreign
		if plugin.Service != nil {
			svc, err := sc.currentState.Services.Get(*plugin.Service.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find service '%v' for plugin %+v",
					*plugin.Service.Name, *plugin.Name)
			}
			plugin.Service = &svc.Service
		}
		if plugin.Route != nil {
			route, err := sc.currentState.Routes.Get(*plugin.Route.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find route '%v' for plugin %+v",
					*plugin.Route.Name, *plugin.Name)
			}
			plugin.Route = &route.Route
		}
		if plugin.Consumer != nil {
			consumer, err := sc.currentState.Consumers.Get(*plugin.Consumer.Username)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not find consumer '%v' for plugin %+v",
					*plugin.Consumer.Username, *plugin.Name)
			}
			plugin.Consumer = &consumer.Consumer
		}
		// XXX
		return &Event{
			Op:     crud.Update,
			Kind:   "plugin",
			Obj:    plugin,
			OldObj: currentPlugin,
		}, nil
	}
	return nil, nil
}

func foreignNames(p *state.Plugin) (serviceName, routeName,
	consumerUsername string) {
	if p == nil {
		return
	}
	if p.Service != nil && p.Service.Name != nil {
		serviceName = *p.Service.Name
	}
	if p.Route != nil && p.Route.Name != nil {
		routeName = *p.Route.Name
	}
	if p.Consumer != nil && p.Consumer.Username != nil {
		consumerUsername = *p.Consumer.Username
	}
	return
}
