package konnect

import (
	"context"
	"encoding/json"
	"fmt"
)

type ControlPlaneRelationsService service

type ControlPlaneServiceRelationCreateRequest struct {
	ServiceVersionID     string `json:"service_version"`
	ControlPlaneEntityID string `json:"control_plane_entity_id"`
	ControlPlane         string `json:"control_plane"`
}

type ControlPlaneServiceRelationUpdateRequest struct {
	ID string
	ControlPlaneServiceRelationCreateRequest
}

// Create creates a ControlPlaneServiceRelation in Konnect.
func (s *ControlPlaneRelationsService) Create(ctx context.Context,
	relation *ControlPlaneServiceRelationCreateRequest) (*ControlPlaneServiceRelation, error) {

	if relation == nil {
		return nil, fmt.Errorf("cannot create a nil ControlPlaneServiceRelation")
	}
	relation.ControlPlane = s.controlPlaneID

	endpoint := "/api/control_plane_service_relations"
	method := "POST"

	req, err := s.client.NewRequest(method, endpoint, nil, relation)
	if err != nil {
		return nil, err
	}

	var createdRelation ControlPlaneServiceRelation
	_, err = s.client.Do(ctx, req, &createdRelation)
	if err != nil {
		return nil, err
	}
	return &createdRelation, nil
}

// Delete deletes a ControlPlaneServiceRelation in Konnect.
func (s *ControlPlaneRelationsService) Delete(ctx context.Context,
	relationID *string) error {
	if emptyString(relationID) {
		return fmt.Errorf("id cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/api/control_plane_service_relations/%v",
		*relationID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// Update updates a ControlPlaneServiceRelation in Konnect.
func (s ControlPlaneRelationsService) Update(ctx context.Context,
	relation *ControlPlaneServiceRelationUpdateRequest) (*ServiceVersion, error) {

	if relation == nil {
		return nil, fmt.Errorf("cannot update a nil ControlPlaneServiceRelation")
	}

	if relation.ID == "" {
		return nil, fmt.Errorf("ID cannot be nil for Update operation")
	}
	relation.ControlPlane = s.controlPlaneID

	endpoint := fmt.Sprintf("/api/control_plane_service_relations/%v", relation.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, relation)
	if err != nil {
		return nil, err
	}

	var updatedSP ServiceVersion
	_, err = s.client.Do(ctx, req, &updatedSP)
	if err != nil {
		return nil, err
	}
	return &updatedSP, nil
}

// List fetches a list of control_plane_service_relations.
func (s *ControlPlaneRelationsService) List(ctx context.Context,
	opt *ListOpt) ([]*ControlPlaneServiceRelation, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/api/control_plane_service_relations", opt)
	if err != nil {
		return nil, nil, err
	}
	var relations []*ControlPlaneServiceRelation

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var relation ControlPlaneServiceRelation
		err = json.Unmarshal(b, &relation)
		if err != nil {
			return nil, nil, err
		}
		relations = append(relations, &relation)
	}

	return relations, next, nil
}

// ListAll fetches all control_plane_service_relations.
func (s *ControlPlaneRelationsService) ListAll(ctx context.Context) ([]*ControlPlaneServiceRelation,
	error) {
	var relations, data []*ControlPlaneServiceRelation
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		relations = append(relations, data...)
	}
	return relations, nil
}
