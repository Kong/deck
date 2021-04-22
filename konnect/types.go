package konnect

import (
	"fmt"
)

const (
	authEndpoint = "/api/auth"
)

type ParentInfoer interface {
	URL() string
	Key() string
	DeepCopyParentInfoer() ParentInfoer
}

func BaseURL() string {
	const baseURL = "https://konnect.konghq.com"
	// TODO override the baseURL using configuration
	return baseURL
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/kong/deck/konnect.ParentInfoer
type ServicePackage struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`

	Versions []ServiceVersion `json:"versions,omitempty"`
}

func (p *ServicePackage) URL() string {
	return fmt.Sprintf("/api/service_packages/%s", *p.ID)
}

func (p *ServicePackage) Key() string {
	return "ServicePackage" + ":" + *p.ID
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/kong/deck/konnect.ParentInfoer
type ServiceVersion struct {
	ID      *string `json:"id,omitempty"`
	Version *string `json:"version,omitempty"`

	ServicePackage *ServicePackage `json:"service_package,omitempty"`

	ControlPlaneServiceRelation *ControlPlaneServiceRelation `json:"control_plane_service_relation,omitempty"`
}

func (v *ServiceVersion) URL() string {
	return fmt.Sprintf("/api/service_versions/%s", *v.ID)
}

func (v *ServiceVersion) Key() string {
	return "ServiceVersion" + ":" + *v.ID
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces
type Document struct {
	ID        *string      `json:"id,omitempty"`
	Path      *string      `json:"path,omitempty"`
	Content   *string      `json:"content,omitempty"`
	Published *bool        `json:"published,omitempty"`
	Parent    ParentInfoer `json:"-"`
}

func (d *Document) ParentKey() string {
	return d.Parent.Key()
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
