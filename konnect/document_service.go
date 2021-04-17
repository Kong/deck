package konnect

import (
	"context"
	"encoding/json"
	"fmt"
)

type DocumentService service

// TODO move this to types.go. It lives here for now to avoid accidentally
// committing the base URL change for the test env

// +k8s:deepcopy-gen=true
type Document struct {
	ID             *string         `json:"id,omitempty"`
	Path           *string         `json:"path,omitempty"`
	Content        *string         `json:"content,omitempty"`
	Published      *bool           `json:"published,omitempty"`
	ServicePackage *ServicePackage `json:"-"`
	ServiceVersion *ServiceVersion `json:"-"`
}

// TODO move to types.go

// ParentEndpoint returns the path of the Service Package or Service Version a Document is attached to.
// It returns an error if the Document has neither or both set
func (d *Document) ParentEndpoint() (string, error) {
	if d.ServicePackage != nil && d.ServiceVersion != nil {
		return "", fmt.Errorf("Document cannot be attached to both a ServicePackage and ServiceVersion")
	}
	if d.ServiceVersion != nil {
		return "service_versions/" + *d.ServiceVersion.ID, nil
	}
	if d.ServicePackage != nil {
		return "service_packages/" + *d.ServiceVersion.ID, nil
	}
	return "", fmt.Errorf("Document must be attached to either a ServicePackage and ServiceVersion")
}

// Create creates a Document in Konnect.
func (d *DocumentService) Create(ctx context.Context, doc *Document) (*Document, error) {

	if doc == nil {
		return nil, fmt.Errorf("cannot create a nil document")
	}

	parent, err := doc.ParentEndpoint()
	if err != nil {
		return nil, err
	}

	endpoint := "/api/" + parent + "/documents"
	method := "POST"
	if doc.ID != nil {
		method = "PUT"
		endpoint = endpoint + "/" + *doc.ID
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
	createdDoc.ServicePackage = doc.ServicePackage
	createdDoc.ServiceVersion = doc.ServiceVersion
	return &createdDoc, nil
}

// Delete deletes a Document in Konnect.
func (d *DocumentService) Delete(ctx context.Context, doc *Document) error {

	if emptyString(doc.ID) {
		return fmt.Errorf("id cannot be nil for Delete operation")
	}

	parent, err := doc.ParentEndpoint()
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/api/"+parent+"/documents/%v", *doc.ID)
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

	parent, err := doc.ParentEndpoint()
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/api/"+parent+"/documents/%v", *doc.ID)
	req, err := d.client.NewRequest("PATCH", endpoint, nil, doc)
	if err != nil {
		return nil, err
	}

	var updatedDoc Document
	_, err = d.client.Do(ctx, req, &updatedDoc)
	if err != nil {
		return nil, err
	}
	updatedDoc.ServicePackage = doc.ServicePackage
	updatedDoc.ServiceVersion = doc.ServiceVersion
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

// ListAllForServicePackage fetches all Documents in Kong enabled for a Service Package.
func (d *DocumentService) ListAllForServicePackage(ctx context.Context, sp *ServicePackage) ([]*Document, error) {
	if sp == nil {
		return nil, fmt.Errorf("ServicePackage cannot be nil")
	}
	docs, err := d.listAllByPath(ctx, "/api/service_packages/"+*sp.ID+"/documents")
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		doc.ServicePackage = sp
	}
	return docs, nil
}

// ListAllForServiceVersion fetches all Documents in Kong enabled for a Service Version.
func (d *DocumentService) ListAllForServiceVersion(ctx context.Context, sv *ServiceVersion) ([]*Document, error) {
	if sv == nil {
		return nil, fmt.Errorf("ServiceVersion cannot be nil")
	}
	docs, err := d.listAllByPath(ctx, "/api/service_versions/"+*sv.ID+"/documents")
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		doc.ServiceVersion = sv
	}
	return docs, nil
}
