package dry

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
)

// PluginCRUD implements Actions interfaces
// from the github.com/kong/crud package for the Plugin entitiy of Kong.
type PluginCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func pluginFromStuct(a diff.Event) *state.Plugin {
	plugin, ok := a.Obj.(*state.Plugin)
	if !ok {
		panic("unexpected type, expected *state.plugin")
	}

	return plugin
}

func formatPluginMessage(plugin *state.Plugin) (message string) {
	if plugin.Service == nil && plugin.Route == nil && plugin.Consumer == nil {
		message = "global plugin " + *plugin.Name
		return
	}
	message = "plugin " + *plugin.Name
	if plugin.Service != nil && plugin.Route != nil {
		message += " on service " + *plugin.Service.Name +
			" and route " + *plugin.Route.Name
		return
	}
	if plugin.Service != nil {
		message += " on service " + *plugin.Service.Name
		return
	}
	if plugin.Route != nil {
		message += " on route " + *plugin.Route.Name
		return
	}
	return
}

// Create creates a fake plugin.
// The arg should be of type diff.Event, containing the plugin to be created,
// else the function will panic.
// It returns a the created *state.Plugin.
func (s *PluginCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)

	print.CreatePrintln("creating", formatPluginMessage(plugin))
	plugin.ID = kong.String(utils.UUID())
	return plugin, nil
}

// Delete deletes a fake plugin.
// The arg should be of type diff.Event, containing the plugin to be deleted,
// else the function will panic.
// It returns a the deleted *state.Plugin.
func (s *PluginCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)

	print.DeletePrintln("deleting", formatPluginMessage(plugin))
	return plugin, nil
}

// Update updates a fake plugin.
// The arg should be of type diff.Event, containing the plugin to be updated,
// else the function will panic.
// It returns a the updated *state.Plugin.
func (s *PluginCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	plugin := pluginFromStuct(event)
	oldPluginObj, ok := event.OldObj.(*state.Plugin)
	if !ok {
		panic("unexpected type, expected *state.plugin")
	}
	oldPlugin := oldPluginObj.DeepCopy()
	// TODO remove this hack
	oldPlugin.CreatedAt = nil
	if oldPlugin.Service != nil {
		oldPlugin.Service = &kong.Service{Name: oldPlugin.Service.Name}
	}
	if plugin.Service != nil {
		plugin.Service = &kong.Service{Name: plugin.Service.Name}
	}
	if oldPlugin.Route != nil {
		oldPlugin.Route = &kong.Route{Name: oldPlugin.Route.Name}
	}
	if plugin.Route != nil {
		plugin.Route = &kong.Route{Name: plugin.Route.Name}
	}
	diff := getDiff(oldPlugin, &plugin.Plugin)
	print.UpdatePrintf("updating %s\n%s", formatPluginMessage(plugin), diff)
	return plugin, nil
}
