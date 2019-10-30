package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// PluginCRUD implements Actions interface
// from the github.com/kong/crud package for the Plugin entitiy of Kong.
type PluginCRUD struct {
	client *kong.Client
}

// NewPluginCRUD creates a new PluginCRUD. Client is required.
func NewPluginCRUD(client *kong.Client) (*PluginCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &PluginCRUD{
		client: client,
	}, nil
}

func pluginFromStuct(arg diff.Event) *state.Plugin {
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
func (s *PluginCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)

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
func (s *PluginCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)
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
func (s *PluginCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)

	updatedPlugin, err := s.client.Plugins.Create(nil, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	return &state.Plugin{Plugin: *updatedPlugin}, nil
}
