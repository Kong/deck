package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// pluginCRUD implements crud.Actions interface.
type pluginCRUD struct {
	client *kong.Client
}

func pluginFromStruct(arg diff.Event) *state.Plugin {
	plugin, ok := arg.Obj.(*state.Plugin)
	if !ok {
		panic("unexpected type, expected *state.Plugin")
	}

	return plugin
}

// Create creates a Plugin in Kong.
// The arg should be of type diff.Event, containing the plugin to be created,
// else the function will panic.
// It returns a the created *state.Plugin.
func (s *pluginCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStruct(event)

	createdPlugin, err := s.client.Plugins.Create(nil, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	return &state.Plugin{Plugin: *createdPlugin}, nil
}

// Delete deletes a Plugin in Kong.
// The arg should be of type diff.Event, containing the plugin to be deleted,
// else the function will panic.
// It returns a the deleted *state.Plugin.
func (s *pluginCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStruct(event)
	err := s.client.Plugins.Delete(nil, plugin.ID)
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

// Update updates a Plugin in Kong.
// The arg should be of type diff.Event, containing the plugin to be updated,
// else the function will panic.
// It returns a the updated *state.Plugin.
func (s *pluginCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStruct(event)

	updatedPlugin, err := s.client.Plugins.Create(nil, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	return &state.Plugin{Plugin: *updatedPlugin}, nil
}
