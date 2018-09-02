package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// PluginService handles Plugins in Kong.
type PluginService service

// Create creates a Plugin in Kong.
// If an ID is specified, it will be used to
// create a plugin in Kong, otherwise an ID
// is auto-generated.
func (s *PluginService) Create(ctx context.Context, plugin *Plugin) (*Plugin, error) {

	queryPath := "/plugins"
	method := "POST"
	// TODO enable PUT support once plugins are migrated to new DAO in Kong
	// if plugin.ID != nil {
	// 	queryPath = queryPath + "/" + *plugin.ID
	// 	method = "PUT"
	// }
	req, err := s.client.newRequest(method, queryPath, nil, plugin)

	if err != nil {
		return nil, err
	}

	var createdPlugin Plugin
	_, err = s.client.Do(ctx, req, &createdPlugin)
	if err != nil {
		return nil, err
	}
	return &createdPlugin, nil
}

// Get fetches a Plugin in Kong.
func (s *PluginService) Get(ctx context.Context, usernameOrID *string) (*Plugin, error) {

	if usernameOrID == nil {
		return nil, errors.New("usernameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/plugins/%v", *usernameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var plugin Plugin
	_, err = s.client.Do(ctx, req, &plugin)
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// Update updates a Plugin in Kong
func (s *PluginService) Update(ctx context.Context, plugin *Plugin) (*Plugin, error) {

	if plugin.ID == nil {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/plugins/%v", *plugin.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, plugin)
	if err != nil {
		return nil, err
	}

	var updatedAPI Plugin
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes a Plugin in Kong
func (s *PluginService) Delete(ctx context.Context, usernameOrID *string) error {

	if usernameOrID == nil {
		return errors.New("usernameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/plugins/%v", *usernameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of Plugins in Kong.
// opt can be used to control pagination.
func (s *PluginService) List(ctx context.Context, opt *ListOpt) ([]*Plugin, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/plugins", opt)
	if err != nil {
		return nil, nil, err
	}
	var plugins []*Plugin

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var plugin Plugin
		err = json.Unmarshal(b, &plugin)
		if err != nil {
			return nil, nil, err
		}
		plugins = append(plugins, &plugin)
	}

	return plugins, next, nil
}
