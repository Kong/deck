package konnect

import (
	"context"
	"encoding/json"
)

// ListOpt aids in paginating through list endpoints.
type ListOpt struct {
	// Size of the page
	Size int `url:"size,omitempty"`
	// Page number to fetch
	Page int `url:"page,omitempty"`
}

const (
	// max page size in Konnect's API is 100
	pageSize = 100
)

// list fetches a list of an entity in Kong.
// opt can be used to control pagination.
func (c *Client) list(ctx context.Context,
	endpoint string, opt *ListOpt,
) ([]json.RawMessage, *ListOpt, error) {
	pageSize := 100
	if opt != nil {
		if opt.Size > 100 {
			opt.Size = pageSize
		} else {
			pageSize = opt.Size
		}
	}

	req, err := c.NewRequest("GET", endpoint, opt, nil)
	if err != nil {
		return nil, nil, err
	}
	var list struct {
		Data      []json.RawMessage `json:"data"`
		Page      int               `json:"page"`
		PageCount int               `json:"pageCount"`
	}

	_, err = c.Do(ctx, req, &list)
	if err != nil {
		return nil, nil, err
	}

	// convenient for end user to use this opt till it's nil
	var next *ListOpt
	if len(list.Data) > 0 && list.Page != list.PageCount {
		next = &ListOpt{
			Page: list.Page + 1,
			Size: pageSize,
		}
	}

	return list.Data, next, nil
}
