package kong

type id interface {
	id() *string
}

// KeyAuth represents a key-auth credential in Kong.
// +k8s:deepcopy-gen=true
type KeyAuth struct {
	Consumer  *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Key       *string   `json:"key,omitempty" yaml:"key,omitempty"`
	TTL       *int      `json:"ttl,omitempty" yaml:"ttl,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c KeyAuth) id() *string {
	return c.ID
}

// BasicAuth represents a basic-auth credential in Kong.
// +k8s:deepcopy-gen=true
type BasicAuth struct {
	Consumer  *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Username  *string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password  *string   `json:"password,omitempty" yaml:"password,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c BasicAuth) id() *string {
	return c.ID
}

// HMACAuth represents a hmac-auth credential in Kong.
// +k8s:deepcopy-gen=true
type HMACAuth struct {
	Consumer  *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Username  *string   `json:"username,omitempty" yaml:"username,omitempty"`
	Secret    *string   `json:"secret,omitempty" yaml:"secret,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c HMACAuth) id() *string {
	return c.ID
}

// Oauth2Credential represents a Oauth2 credential in Kong.
// +k8s:deepcopy-gen=true
type Oauth2Credential struct {
	Consumer     *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt    *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID           *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name         *string   `json:"name,omitempty" yaml:"name,omitempty"`
	ClientID     *string   `json:"client_id,omitempty" yaml:"client_id,omitempty"`
	ClientSecret *string   `json:"client_secret,omitempty" yaml:"client_secret,omitempty"`
	RedirectURIs []*string `json:"redirect_uris,omitempty" yaml:"redirect_uris,omitempty"`
	Tags         []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c Oauth2Credential) id() *string {
	return c.ID
}

// JWTAuth represents a JWT credential in Kong.
// +k8s:deepcopy-gen=true
type JWTAuth struct {
	Consumer     *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt    *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID           *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Algorithm    *string   `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
	Key          *string   `json:"key,omitempty" yaml:"key,omitempty"`
	RSAPublicKey *string   `json:"rsa_public_key,omitempty" yaml:"rsa_public_key,omitempty"`
	Secret       *string   `json:"secret,omitempty" yaml:"secret,omitempty"`
	Tags         []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c JWTAuth) id() *string {
	return c.ID
}

// MTLSAuth represents a MTLS credential in Kong.
// +k8s:deepcopy-gen=true
type MTLSAuth struct {
	Consumer      *Consumer      `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt     *int           `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID            *string        `json:"id,omitempty" yaml:"id,omitempty"`
	SubjectName   *string        `json:"subject_name,omitempty" yaml:"subject_name,omitempty"`
	CACertificate *CACertificate `json:"ca_certificate,omitempty" yaml:"ca_certificate,omitempty"`
	Tags          []*string      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c MTLSAuth) id() *string {
	return c.ID
}

// ACLGroup represents an ACL group for a consumer in Kong.
// +k8s:deepcopy-gen=true
type ACLGroup struct {
	Consumer  *Consumer `json:"consumer,omitempty" yaml:"consumer,omitempty"`
	CreatedAt *int      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        *string   `json:"id,omitempty" yaml:"id,omitempty"`
	Group     *string   `json:"group,omitempty" yaml:"group,omitempty"`
	Tags      []*string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (c ACLGroup) id() *string {
	return c.ID
}
