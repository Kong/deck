package konnect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kong/go-kong/kong"
)

type PageOpt struct {
	// request
	Size   int  `url:"size,omitempty"`
	Number *int `url:"number,omitempty"`

	// response
	NextPageNum *int `url:"next_page_num,omitempty"`
	TotalCount  *int `url:"total_count,omitempty"`
}

type KonnectListOpt struct { //nolint:revive
	// Size of the page
	Page *PageOpt

	// Tags to use for filtering the list.
	Tags []*string `url:"tags,omitempty"`
}

type RLAOverride struct {
	ID    *string            `json:"id,omitempty" yaml:"id,omitempty"`
	Value kong.Configuration `json:"value,omitempty" yaml:"value,omitempty"`
}

type konnectResponseObj struct {
	Item kong.ConsumerGroup `json:"item,omitempty" yaml:"item,omitempty"`
	Page *PageOpt
}

type konnectRLAObj struct {
	ConsumerGroupID     *string  `json:"consumer_group_id,omitempty" yaml:"consumer_group_id,omitempty"`
	ID                  *string  `json:"id,omitempty" yaml:"id,omitempty"`
	Limit               []*int32 `json:"limit,omitempty" yaml:"limit,omitempty"`
	WindowSize          []*int32 `json:"window_size,omitempty" yaml:"window_size,omitempty"`
	WindowType          *string  `json:"window_type,omitempty" yaml:"window_type,omitempty"`
	RetryAfterJitterMax *int32   `json:"retry_after_jitter_max,omitempty" yaml:"retry_after_jitter_max,omitempty"`
}

type konnectRLAResponseObj struct {
	Item konnectRLAObj `json:"item,omitempty" yaml:"item,omitempty"`
}

func isEmptyString(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == ""
}

func CreateConsumerGroup(ctx context.Context, client *kong.Client, entity interface{}) (*kong.ConsumerGroup, error) {
	endpoint := "/v1/consumer-groups"
	req, err := client.NewRequest(http.MethodPost, endpoint, nil, entity)
	if err != nil {
		return nil, err
	}
	var cg konnectResponseObj
	_, err = client.Do(ctx, req, &cg)
	if err != nil {
		return nil, err
	}
	return &cg.Item, nil
}

func UpdateConsumerGroup(ctx context.Context, client *kong.Client,
	cgID *string, entity interface{},
) (*kong.ConsumerGroup, error) {
	if isEmptyString(cgID) {
		return nil, fmt.Errorf("update consumer-group: consumer-group ID cannot be nil")
	}
	endpoint := fmt.Sprintf("/v1/consumer-groups/%v", *cgID)
	req, err := client.NewRequest(http.MethodPut, endpoint, nil, entity)
	if err != nil {
		return nil, err
	}
	var cg konnectResponseObj
	_, err = client.Do(ctx, req, &cg)
	if err != nil {
		return nil, err
	}
	return &cg.Item, nil
}

// GetConsumerGroup fetches a ConsumerGroup from Konnect.
func GetConsumerGroup(ctx context.Context,
	client *kong.Client, nameOrID *string,
) (*kong.ConsumerGroup, error) {
	if isEmptyString(nameOrID) {
		return nil, fmt.Errorf("getting consumer-group: nameOrID cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumer-groups/%v", *nameOrID)
	req, err := client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var cg konnectResponseObj
	_, err = client.Do(ctx, req, &cg)
	if err != nil {
		return nil, err
	}
	return &cg.Item, nil
}

// ListAllConsumerGroupMembers fetches all ConsumerGroups members from Konnect.
func ListAllConsumerGroupMembers(
	ctx context.Context, client *kong.Client, cgID *string,
) ([]*kong.Consumer, error) {
	if isEmptyString(cgID) {
		return nil, fmt.Errorf("list consumer-group-members: consumer-group ID cannot be nil")
	}
	var members, data []*kong.Consumer
	var err error
	opt := &KonnectListOpt{Page: &PageOpt{Size: 100}}
	for opt != nil {
		endpoint := fmt.Sprintf("/v1/consumer-groups/%v/members", *cgID)
		data, opt, err = ListConsumerGroupMembers(ctx, client, endpoint, opt)
		if err != nil {
			return nil, err
		}
		members = append(members, data...)
	}
	return members, nil
}

func upsertRateLimitingAdvancedPlugin(
	ctx context.Context, client *kong.Client, id string, config kong.Configuration, method string,
) (*kong.ConsumerGroupRLA, error) {
	endpoint := fmt.Sprintf("/v1/consumer-groups/%v/rate-limiting-advanced-config", id)
	req, err := client.NewRequest(method, endpoint, nil, config)
	if err != nil {
		return nil, err
	}
	var rla konnectRLAResponseObj
	_, err = client.Do(ctx, req, &rla)
	if err != nil {
		return nil, err
	}
	rlaConfig := kong.Configuration{}
	if rla.Item.Limit != nil {
		rlaConfig["limit"] = rla.Item.Limit
	}
	if rla.Item.WindowSize != nil {
		rlaConfig["window_size"] = rla.Item.WindowSize
	}
	if rla.Item.WindowType != nil {
		rlaConfig["window_type"] = rla.Item.WindowType
	}
	if rla.Item.RetryAfterJitterMax != nil {
		rlaConfig["retry_after_jitter_max"] = rla.Item.RetryAfterJitterMax
	}
	return &kong.ConsumerGroupRLA{
		Plugin:        kong.String("rate-limiting-advanced"),
		ConsumerGroup: rla.Item.ConsumerGroupID,
		Config:        rlaConfig,
	}, nil
}

func CreateRateLimitingAdvancedPlugin(
	ctx context.Context, client *kong.Client, cgID *string, config kong.Configuration,
) (*kong.ConsumerGroupRLA, error) {
	if isEmptyString(cgID) {
		return nil, fmt.Errorf("create consumer-group override: consumer-group ID cannot be nil")
	}
	return upsertRateLimitingAdvancedPlugin(
		ctx, client, *cgID, config, http.MethodPost,
	)
}

func UpdateRateLimitingAdvancedPlugin(
	ctx context.Context, client *kong.Client, cgID *string, config kong.Configuration,
) (*kong.ConsumerGroupRLA, error) {
	if isEmptyString(cgID) {
		return nil, fmt.Errorf("update consumer-group override: consumer-group ID cannot be nil")
	}
	return upsertRateLimitingAdvancedPlugin(
		ctx, client, *cgID, config, http.MethodPut,
	)
}

// GetConsumerGroupRateLimitingAdvancedPlugin fetches the RLA override for
// a ConsumerGroup from Konnect.
func GetConsumerGroupRateLimitingAdvancedPlugin(
	ctx context.Context, client *kong.Client, cgID *string,
) (*kong.ConsumerGroupPlugin, error) {
	if isEmptyString(cgID) {
		return nil, fmt.Errorf("get consumer-group override: consumer-group ID cannot be nil")
	}
	endpoint := fmt.Sprintf("/v1/consumer-groups/%v/rate-limiting-advanced-config", *cgID)
	req, err := client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var rla konnectRLAResponseObj
	res, err := client.Do(ctx, req, &rla)
	if err != nil {
		// Konnect returns a 404 if no plugin exists yet.
		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	config := kong.Configuration{}
	if rla.Item.Limit != nil {
		config["limit"] = rla.Item.Limit
	}
	if rla.Item.WindowSize != nil {
		config["window_size"] = rla.Item.WindowSize
	}
	if rla.Item.WindowType != nil {
		config["window_type"] = rla.Item.WindowType
	}
	if rla.Item.RetryAfterJitterMax != nil {
		config["retry_after_jitter_max"] = rla.Item.RetryAfterJitterMax
	}
	return &kong.ConsumerGroupPlugin{
		ID:   rla.Item.ID,
		Name: kong.String("rate-limiting-advanced"),
		ConsumerGroup: &kong.ConsumerGroup{
			ID: cgID,
		},
		Config: config,
	}, nil
}

// DeleteRateLimitingAdvancedPlugin deletes a ConsumerGroup plugin in Kong
func DeleteRateLimitingAdvancedPlugin(
	ctx context.Context, client *kong.Client, cgID *string,
) error {
	if isEmptyString(cgID) {
		return fmt.Errorf("deleting consumer-group plugin: id cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumer-groups/%v/rate-limiting-advanced-config", *cgID)
	req, err := client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(ctx, req, nil)
	return err
}

// ListConsumerGroupMembers fetches a page members for a ConsumerGroup from Konnect.
func ListConsumerGroupMembers(ctx context.Context,
	client *kong.Client, endpoint string, opt *KonnectListOpt,
) ([]*kong.Consumer, *KonnectListOpt, error) {
	data, next, err := list(ctx, client, endpoint, opt)
	if err != nil {
		return nil, nil, err
	}

	var consumers []*kong.Consumer

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var consumer kong.Consumer
		err = json.Unmarshal(b, &consumer)
		if err != nil {
			return nil, nil, err
		}
		consumers = append(consumers, &consumer)
	}

	return consumers, next, nil
}

// GetConsumerGroupObject Get fetches a ConsumerGroup from Kong.
func GetConsumerGroupObject(ctx context.Context,
	client *kong.Client, cgID *string,
) (*kong.ConsumerGroupObject, error) {
	r, err := GetConsumerGroup(ctx, client, cgID)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m, err := ListAllConsumerGroupMembers(ctx, client, cgID)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var plugins []*kong.ConsumerGroupPlugin
	p, err := GetConsumerGroupRateLimitingAdvancedPlugin(ctx, client, cgID)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p != nil {
		plugins = append(plugins, p)
	}

	group := &kong.ConsumerGroupObject{
		ConsumerGroup: r,
		Consumers:     m,
		Plugins:       plugins,
	}
	return group, nil
}

// DeleteConsumerGroup deletes a ConsumerGroup in Kong
func DeleteConsumerGroup(
	ctx context.Context, client *kong.Client, cgID *string,
) error {
	if isEmptyString(cgID) {
		return fmt.Errorf("delete consumer-group: ID cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumer-groups/%v", *cgID)
	req, err := client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(ctx, req, nil)
	return err
}

func DeleteConsumerGroupMember(
	ctx context.Context, client *kong.Client, cgID, consumer *string,
) error {
	if isEmptyString(cgID) {
		return fmt.Errorf("delete consumer-group-member: ID cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumers/%s/groups/%s/members", *consumer, *cgID)
	req, err := client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(ctx, req, nil)
	return err
}

func CreateConsumerGroupMember(
	ctx context.Context, client *kong.Client, cgID, consumer *string,
) error {
	if isEmptyString(consumer) {
		return fmt.Errorf("create consumer-group-member: consumer cannot be nil")
	} else if isEmptyString(cgID) {
		return fmt.Errorf("create consumer-group-member: consumer group ID cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumers/%s/groups/%s/members", *consumer, *cgID)
	req, err := client.NewRequest("POST", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(ctx, req, nil)
	return err
}

func UpdateConsumerGroupMember(
	ctx context.Context, client *kong.Client, cgID, consumer *string,
) error {
	if isEmptyString(consumer) {
		return fmt.Errorf("create consumer-group-member: consumer cannot be nil")
	} else if isEmptyString(cgID) {
		return fmt.Errorf("create consumer-group-member: consumer group ID cannot be nil")
	}

	endpoint := fmt.Sprintf("/v1/consumers/%s/groups/%s/members", *consumer, *cgID)
	req, err := client.NewRequest("PUT", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(ctx, req, nil)
	return err
}

// list fetches a list of an entity in Kong.
// opt can be used to control pagination.
func list(ctx context.Context,
	client *kong.Client, endpoint string, opt *KonnectListOpt,
) ([]json.RawMessage, *KonnectListOpt, error) {
	req, err := client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	var list struct {
		Items []json.RawMessage `json:"items"`
		KonnectListOpt
	}

	_, err = client.Do(ctx, req, &list)
	if err != nil {
		return nil, nil, err
	}

	var next *KonnectListOpt
	if list.Page != nil && list.Page.NextPageNum != nil {
		next = &KonnectListOpt{
			Page: &PageOpt{
				Size:   opt.Page.Size,
				Number: list.Page.NextPageNum,
			},
			Tags: opt.Tags,
		}
	}

	return list.Items, next, nil
}

// List fetches a list of ConsumerGroup in Kong.
// opt can be used to control pagination.
func ListConsumerGroups(ctx context.Context,
	client *kong.Client, opt *KonnectListOpt,
) ([]*kong.ConsumerGroup, *KonnectListOpt, error) {
	data, next, err := list(ctx, client, "/v1/consumer-groups", opt)
	if err != nil {
		return nil, nil, err
	}

	var consumers []*kong.ConsumerGroup

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var consumer kong.ConsumerGroup
		err = json.Unmarshal(b, &consumer)
		if err != nil {
			return nil, nil, err
		}
		consumers = append(consumers, &consumer)
	}

	return consumers, next, nil
}

// ListAll fetches all ConsumerGroup in Kong.
func ListAllConsumerGroups(ctx context.Context, client *kong.Client, tags []*string) ([]*kong.ConsumerGroup, error) {
	var consumerGroups, data []*kong.ConsumerGroup
	var err error
	opt := &KonnectListOpt{Page: &PageOpt{Size: 100}}
	if tags != nil {
		opt.Tags = tags
	}

	for opt != nil {
		data, opt, err = ListConsumerGroups(ctx, client, opt)
		if err != nil {
			return nil, err
		}
		consumerGroups = append(consumerGroups, data...)
	}
	return consumerGroups, nil
}
