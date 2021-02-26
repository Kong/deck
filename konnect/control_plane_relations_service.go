package konnect

import (
	"context"
	"encoding/json"
)

type ControlPlaneRelationsService service

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
