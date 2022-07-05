package konnect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DocumentService service

// Create creates a Document in Konnect.
func (d *DocumentService) Create(ctx context.Context, doc *Document) (*Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("cannot create a nil document")
	}

	if doc.Parent == nil {
		return nil, fmt.Errorf("document must have a Parent")
	}

	endpoint := d.client.prefix + doc.Parent.URL() + "/documents/"
	method := http.MethodPost
	if doc.ID != nil {
		method = "PUT"
		endpoint = endpoint + *doc.ID
	}
	req, err := d.client.NewRequest(method, endpoint, nil, doc)
	if err != nil {
		return nil, err
	}

	var createdDoc Document
	_, err = d.client.Do(ctx, req, &createdDoc)
	if err != nil {
		return nil, err
	}
	createdDoc.Parent = doc.Parent
	return &createdDoc, nil
}

// Delete deletes a Document in Konnect.
func (d *DocumentService) Delete(ctx context.Context, doc *Document) error {
	if emptyString(doc.ID) {
		return fmt.Errorf("id cannot be nil for Delete operation")
	}

	if doc.Parent == nil {
		return fmt.Errorf("document must have a Parent")
	}

	endpoint := fmt.Sprintf("%s/%s/documents/%s", d.client.prefix, doc.Parent.URL(), *doc.ID)
	req, err := d.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = d.client.Do(ctx, req, nil)
	return err
}

// Update updates a Document in Konnect.
func (d *DocumentService) Update(ctx context.Context, doc *Document) (*Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("cannot update a nil document")
	}

	if emptyString(doc.ID) {
		return nil, fmt.Errorf("ID cannot be nil for Update operation")
	}

	if doc.Parent == nil {
		return nil, fmt.Errorf("document must have a Parent")
	}

	// Document PATCHes run through POST validation logic. Attempting to PATCH a Published: true
	// document without toggling Published results in a 400, as if you'd tried to POST another
	// Published: true document under the same resource. As such, this PUTs instead.
	endpoint := fmt.Sprintf("%s/%s/documents/%s", d.client.prefix, doc.Parent.URL(), *doc.ID)
	putReq, err := d.client.NewRequest("PUT", endpoint, nil, doc)
	if err != nil {
		return nil, err
	}

	var updatedDoc Document
	_, err = d.client.Do(ctx, putReq, &updatedDoc)
	if err != nil {
		return nil, err
	}
	updatedDoc.Parent = doc.Parent
	return &updatedDoc, nil
}

// listByPath fetches a list of Documents in Konnect on a specific path.
// This is a helper method for listing all documents for specific entities.
func (d *DocumentService) listByPath(ctx context.Context, path string, opt *ListOpt) ([]*Document, *ListOpt, error) {
	data, next, err := d.client.list(ctx, path, opt)
	if err != nil {
		return nil, nil, err
	}
	var docs []*Document

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var doc Document
		err = json.Unmarshal(b, &doc)
		if err != nil {
			return nil, nil, err
		}
		docs = append(docs, &doc)
	}

	return docs, next, nil
}

// ListAll fetches all Documents in Kong.
func (d *DocumentService) listAllByPath(ctx context.Context, path string) ([]*Document, error) {
	var docs, data []*Document
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = d.listByPath(ctx, path, opt)
		if err != nil {
			return nil, err
		}
		docs = append(docs, data...)
	}
	return docs, nil
}

// ListAllForParent fetches all Documents in Konnect for a parent entity.
func (d *DocumentService) ListAllForParent(ctx context.Context, parent ParentInfoer) ([]*Document, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}
	var docs []*Document
	var err error
	docs, err = d.listAllByPath(ctx, d.client.prefix+parent.URL()+"/documents")
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		doc.Parent = parent
	}
	return docs, nil
}
