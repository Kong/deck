package kong

import (
	"context"
	"encoding/json"
)

// ListOpt aids in paginating through list endpoints
type ListOpt struct {
	Size   int    `url:"size,omitempty"`
	Offset string `url:"offset,omitempty"`
}

// list fetches a list of an entity in Kong.
// opt can be used to control pagination.
func (c *Client) list(ctx context.Context, endpoint string, opt *ListOpt) ([]json.RawMessage, *ListOpt, error) {

	req, err := c.newRequest("GET", endpoint, opt, nil)
	if err != nil {
		return nil, nil, err
	}
	var list struct {
		Data []json.RawMessage `json:"data"`
		Next *string           `json:"offset"`
	}

	_, err = c.Do(ctx, req, &list)
	if err != nil {
		return nil, nil, err
	}

	// convinient for end user to use this opt till it's nil
	var next *ListOpt
	if list.Next != nil {
		next = &ListOpt{
			Offset: *list.Next,
		}
		if opt != nil && next != nil {
			next.Size = opt.Size
		}
	}

	return list.Data, next, nil
}
