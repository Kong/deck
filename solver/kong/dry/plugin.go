package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	message = "plugin " + *plugin.Name + " on"
	associations := []string{}
	if plugin.Service != nil {
		associations = append(associations, " service "+*plugin.Service.ID)
	}
	if plugin.Route != nil {
		associations = append(associations, " route "+*plugin.Route.ID)
	}
	if plugin.Consumer != nil {
		associations = append(associations, " consumer "+*plugin.Consumer.ID)
	}
	for i := 0; i < len(associations); i++ {
		message += associations[i]
		if i < len(associations)-1 {
			message += ","
		}
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

	diff, err := getDiff(oldPlugin, &plugin.Plugin)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating %s\n%s", formatPluginMessage(plugin), diff)
	return plugin, nil
}
