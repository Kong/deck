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
}

func BaseURL() string {
	const baseURL = "https://konnect.konghq.com"
	return baseURL
}

// ServicePackage service package model
// +k8s:deepcopy-gen=true
type ServicePackage struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description"`

	Versions []ServiceVersion `json:"versions,omitempty"`
}

func (p *ServicePackage) URL() string {
	return fmt.Sprintf("/api/service_packages/%s", *p.ID)
}

func (p *ServicePackage) Key() string {
	return "ServicePackage" + ":" + *p.ID
}

// ServiceVersion service version model
// +k8s:deepcopy-gen=true
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

// ShallowCopyInto is a shallowcopy function, copying the receiver, writing into out. d must be non-nil.
func (d *Document) ShallowCopyInto(out *Document) {
	*out = *d
	if d.ID != nil {
		d, out := &d.ID, &out.ID
		*out = new(string)
		**out = **d
	}
	if d.Path != nil {
		d, out := &d.Path, &out.Path
		*out = new(string)
		**out = **d
	}
	if d.Content != nil {
		d, out := &d.Content, &out.Content
		*out = new(string)
		**out = **d
	}
	if d.Published != nil {
		d, out := &d.Published, &out.Published
		*out = new(bool)
		**out = **d
	}
	if d.Parent != nil {
		out.Parent = d.Parent
	}
}

// ShallowCopy is a shallowcopy function, copying the receiver, creating a new Document.
func (d *Document) ShallowCopy() *Document {
	if d == nil {
		return nil
	}
	out := new(Document)
	d.ShallowCopyInto(out)
	return out
}

// ControlPlaneServiceRelation control plane service relation model
// +k8s:deepcopy-gen=true
type ControlPlaneServiceRelation struct {
	ID                   *string       `json:"id,omitempty"`
	ControlPlaneEntityID *string       `json:"control_plane_entity_id,omitempty"`
	ControlPlane         *ControlPlane `json:"control_plane,omitempty"`
}

// ControlPlane identifies a specific control plane in Konnect.
// +k8s:deepcopy-gen=true
type ControlPlane struct {
	ID   *string           `json:"id"`
	Type *ControlPlaneType `json:"type"`
}

// ControlPlaneType represents control plane associated information.
// +k8s:deepcopy-gen=true
type ControlPlaneType struct {
	Name *string `json:"name"`
}

// AuthResponse is authentication response wrapper for login.
type AuthResponse struct {
	Organization   string `json:"org_name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	OrganizationID string `json:"org_id"`
}
