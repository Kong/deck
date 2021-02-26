package konnect

import (
	"context"
	"encoding/json"
)

type ControlPlaneService service

// List fetches a list of control planes.
// No pagination is being performed because the number of control planes
// is expected to be very small.
func (s *ControlPlaneService) List(ctx context.Context,
	opt *ListOpt) ([]ControlPlane, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/api/control_planes", opt)
	if err != nil {
		return nil, nil, err
	}
	var controlPlanes []ControlPlane

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var controlPlane ControlPlane
		err = json.Unmarshal(b, &controlPlane)
		if err != nil {
			return nil, nil, err
		}
		controlPlanes = append(controlPlanes, controlPlane)
	}

	return controlPlanes, next, nil
}
