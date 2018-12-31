package state

// KongState is an in-memory database representation
// of Kong's configuration.
type KongState struct {
	Services  *ServicesCollection
	Routes    *RoutesCollection
	Upstreams *UpstreamsCollection
	Targets   *TargetsCollection
}

// NewKongState creates a new in-memory KongState.
func NewKongState() (*KongState, error) {
	services, err := NewServicesCollection()
	if err != nil {
		return nil, err
	}
	routes, err := NewRoutesCollection()
	if err != nil {
		return nil, err
	}
	upstreams, err := NewUpstreamsCollection()
	if err != nil {
		return nil, err
	}
	targets, err := NewTargetsCollection()
	if err != nil {
		return nil, err
	}
	return &KongState{
		Services:  services,
		Routes:    routes,
		Upstreams: upstreams,
		Targets:   targets,
	}, nil
}
