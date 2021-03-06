package konnect

const (
	authEndpoint = "/api/auth"
)

func BaseURL() string {
	const baseURL = "https://konnect.konghq.com"
	// TODO override the baseURL using configuration
	return baseURL
}

// +k8s:deepcopy-gen=true
type ServicePackage struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`

	Versions []ServiceVersion `json:"versions,omitempty"`
}

// +k8s:deepcopy-gen=true
type ServiceVersion struct {
	ID      *string `json:"id,omitempty"`
	Version *string `json:"version,omitempty"`

	ServicePackage *ServicePackage `json:"service_package,omitempty"`

	ControlPlaneServiceRelation *ControlPlaneServiceRelation `json:"control_plane_service_relation,omitempty"`
}

// +k8s:deepcopy-gen=true
type ControlPlaneServiceRelation struct {
	ID                   *string       `json:"id,omitempty"`
	ControlPlaneEntityID *string       `json:"control_plane_entity_id,omitempty"`
	ControlPlane         *ControlPlane `json:"control_plane,omitempty"`
}

// +k8s:deepcopy-gen=true
type ControlPlane struct {
	ID   *string           `json:"id"`
	Type *ControlPlaneType `json:"type"`
}

// +k8s:deepcopy-gen=true
type ControlPlaneType struct {
	Name *string `json:"name"`
}

type AuthResponse struct {
	Organization   string `json:"org_name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	OrganizationID string `json:"org_id"`
}
