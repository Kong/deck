package types

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

func eventFromArg(arg crud.Arg) crud.Event {
	event, ok := arg.(crud.Event)
	if !ok {
		panic("unexpected type, expected diff.Event")
	}
	return event
}

type Entity interface {
	Type() string
	CRUDActions() crud.Actions
	PostProcessActions() crud.Actions
}

type entityImpl struct {
	typ                string
	cRUDActions        crud.Actions // needs to set client
	postProcessActions crud.Actions // needs currentstate Set
}

func (e entityImpl) Type() string {
	return e.typ
}

func (e entityImpl) CRUDActions() crud.Actions {
	return e.cRUDActions
}

func (e entityImpl) PostProcessActions() crud.Actions {
	return e.postProcessActions
}

type EntityOpts struct {
	CurrentState  *state.KongState
	TargetState   *state.KongState
	KongClient    *kong.Client
	KonnectClient *konnect.Client
}

const (
	// Service identifies a Service in Kong.
	Service = "service"
	// Route identifies a Route in Kong.
	Route = "route"
	// Plugin identifies a Plugin in Kong.
	Plugin = "plugin"

	// Certificate identifies a Certificate in Kong.
	Certificate = "certificate"
	// SNI identifies a SNI in Kong.
	SNI = "sni"
	// CACertificate identifies a CACertificate in Kong.
	CACertificate = "ca-certificate"

	// Upstream identifies a Upstream in Kong.
	Upstream = "upstream"
	// Target identifies a Target in Kong.
	Target = "target"

	// Consumer identifies a Consumer in Kong.
	Consumer = "consumer"
	// ACLGroup identifies a ACLGroup in Kong.
	ACLGroup = "acl-group"
	// BasicAuth identifies a BasicAuth in Kong.
	BasicAuth = "basic-auth"
	// HMACAuth identifies a HMACAuth in Kong.
	HMACAuth = "hmac-auth"
	// JWTAuth identifies a JWTAuth in Kong.
	JWTAuth = "jwt-auth"
	// MTLSAuth identifies a MTLSAuth in Kong.
	MTLSAuth = "mtls-auth"
	// KeyAuth identifies aKeyAuth in Kong.
	KeyAuth = "key-auth"
	// OAuth2Cred identifies a OAuth2Cred in Kong.
	OAuth2Cred = "oauth2-cred"

	// RBACRole identifies a RBACRole in Kong Enterprise.
	RBACRole = "rbac-role"
	// RBACEndpointPermission identifies a RBACEndpointPermission in Kong Enterprise.
	RBACEndpointPermission = "rbac-endpoint-permission"

	// ServicePackage identifies a ServicePackage in Konnect.
	ServicePackage = "service-package"
	// ServiceVersion identifies a ServiceVersion in Konnect.
	ServiceVersion = "service-version"
	// Document identifies a Document in Konnect.
	Document = "document"
)

// AllTypes represents all types defined in the
// package.
var AllTypes = []string{
	Service, Route, Plugin,

	Certificate, SNI, CACertificate,

	Upstream, Target,

	Consumer,
	ACLGroup, BasicAuth, KeyAuth,
	HMACAuth, JWTAuth, OAuth2Cred,
	MTLSAuth,

	RBACRole, RBACEndpointPermission,

	ServicePackage, ServiceVersion, Document,
}

func NewEntity(t string, opts EntityOpts) (Entity, error) {
	switch t {
	case Service:
		return entityImpl{
			typ: "service",
			cRUDActions: &serviceCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &servicePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Route:
		return entityImpl{
			typ: "route",
			cRUDActions: &routeCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &routePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Upstream:
		return entityImpl{
			typ: "upstream",
			cRUDActions: &upstreamCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &upstreamPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Target:
		return entityImpl{
			typ: "target",
			cRUDActions: &targetCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &targetPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Plugin:
		return entityImpl{
			typ: "plugin",
			cRUDActions: &pluginCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &pluginPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Consumer:
		return entityImpl{
			typ: "consumer",
			cRUDActions: &consumerCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &consumerPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case ServicePackage:
		return entityImpl{
			typ: "service-package",
			cRUDActions: &servicePackageCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &servicePackagePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case ServiceVersion:
		return entityImpl{
			typ: "service-version",
			cRUDActions: &serviceVersionCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &serviceVersionPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Document:
		return entityImpl{
			typ: "document",
			cRUDActions: &documentCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &documentPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case Certificate:
		return entityImpl{
			typ: "certificate",
			cRUDActions: &certificateCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &certificatePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case CACertificate:
		return entityImpl{
			typ: "ca-certificate",
			cRUDActions: &caCertificateCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &caCertificatePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case SNI:
		return entityImpl{
			typ: "sni",
			cRUDActions: &sniCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &sniPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case RBACEndpointPermission:
		return entityImpl{
			typ: "rbac-endpoint-permission",
			cRUDActions: &rbacEndpointPermissionCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &rbacEndpointPermissionPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case RBACRole:
		return entityImpl{
			typ: "rbac-role",
			cRUDActions: &rbacRoleCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &rbacRolePostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case ACLGroup:
		return entityImpl{
			typ: "acl-group",
			cRUDActions: &aclGroupCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &aclGroupPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case BasicAuth:
		return entityImpl{
			typ: "basic-auth",
			cRUDActions: &basicAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &basicAuthPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case KeyAuth:
		return entityImpl{
			typ: "key-auth",
			cRUDActions: &keyAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &keyAuthPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case HMACAuth:
		return entityImpl{
			typ: "hmac-auth",
			cRUDActions: &hmacAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &hmacAuthPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case JWTAuth:
		return entityImpl{
			typ: "jwt-auth",
			cRUDActions: &jwtAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &jwtAuthPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case MTLSAuth:
		return entityImpl{
			typ: "mtls-auth",
			cRUDActions: &mtlsAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &mtlsAuthPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	case OAuth2Cred:
		return entityImpl{
			typ: "oauth2-cred",
			cRUDActions: &oauth2CredCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &oauth2CredPostAction{
				currentState: opts.CurrentState,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown type: %q", t)
	}
}
