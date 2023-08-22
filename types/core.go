package types

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

type Differ interface {
	Deletes(func(crud.Event) error) error
	CreateAndUpdates(func(crud.Event) error) error
}

type DuplicatesDeleter interface {
	// DuplicatesDeletes returns delete events for entities that have duplicates in the current and target state.
	// A duplicate is defined as an entity with the same name but different ID.
	DuplicatesDeletes() ([]crud.Event, error)
}

type Entity interface {
	Type() EntityType
	CRUDActions() crud.Actions
	PostProcessActions() crud.Actions
	Differ() Differ
}

type entityImpl struct {
	typ                EntityType
	crudActions        crud.Actions // needs to set client
	postProcessActions crud.Actions // needs currentstate Set
	differ             Differ
}

func (e entityImpl) Type() EntityType {
	return e.typ
}

func (e entityImpl) CRUDActions() crud.Actions {
	return e.crudActions
}

func (e entityImpl) PostProcessActions() crud.Actions {
	return e.postProcessActions
}

func (e entityImpl) Differ() Differ {
	return e.differ
}

type EntityOpts struct {
	CurrentState  *state.KongState
	TargetState   *state.KongState
	KongClient    *kong.Client
	KonnectClient *konnect.Client

	IsKonnect bool
}

// EntityType defines a type of entity that is managed by decK.
type EntityType string

const (
	// Service identifies a Service in Kong.
	Service EntityType = "service"
	// Route identifies a Route in Kong.
	Route EntityType = "route"
	// Plugin identifies a Plugin in Kong.
	Plugin EntityType = "plugin"

	// Certificate identifies a Certificate in Kong.
	Certificate EntityType = "certificate"
	// SNI identifies a SNI in Kong.
	SNI EntityType = "sni"
	// CACertificate identifies a CACertificate in Kong.
	CACertificate EntityType = "ca-certificate"

	// Upstream identifies a Upstream in Kong.
	Upstream EntityType = "upstream"
	// Target identifies a Target in Kong.
	Target EntityType = "target"

	// Consumer identifies a Consumer in Kong.
	Consumer EntityType = "consumer"
	// ConsumerGroup identifies a ConsumerGroup in Kong.
	ConsumerGroup EntityType = "consumer-group"
	// ConsumerGroupConsumer identifies a ConsumerGroupConsumer in Kong.
	ConsumerGroupConsumer EntityType = "consumer-group-consumer"
	// ConsumerGroupPlugin identifies a ConsumerGroupPlugin in Kong.
	ConsumerGroupPlugin EntityType = "consumer-group-plugin"
	// ACLGroup identifies a ACLGroup in Kong.
	ACLGroup EntityType = "acl-group"
	// BasicAuth identifies a BasicAuth in Kong.
	BasicAuth EntityType = "basic-auth"
	// HMACAuth identifies a HMACAuth in Kong.
	HMACAuth EntityType = "hmac-auth"
	// JWTAuth identifies a JWTAuth in Kong.
	JWTAuth EntityType = "jwt-auth"
	// MTLSAuth identifies a MTLSAuth in Kong.
	MTLSAuth EntityType = "mtls-auth"
	// KeyAuth identifies aKeyAuth in Kong.
	KeyAuth EntityType = "key-auth"
	// OAuth2Cred identifies a OAuth2Cred in Kong.
	OAuth2Cred EntityType = "oauth2-cred" //nolint:gosec

	// RBACRole identifies a RBACRole in Kong Enterprise.
	RBACRole EntityType = "rbac-role"
	// RBACEndpointPermission identifies a RBACEndpointPermission in Kong Enterprise.
	RBACEndpointPermission EntityType = "rbac-endpoint-permission" //nolint:gosec

	// ServicePackage identifies a ServicePackage in Konnect.
	ServicePackage EntityType = "service-package"
	// ServiceVersion identifies a ServiceVersion in Konnect.
	ServiceVersion EntityType = "service-version"
	// Document identifies a Document in Konnect.
	Document EntityType = "document"

	// Vault identifies a Vault in Kong.
	Vault EntityType = "vault"
)

// AllTypes represents all types defined in the
// package.
var AllTypes = []EntityType{
	Service, Route, Plugin,

	Certificate, SNI, CACertificate,

	Upstream, Target,

	Consumer,
	ConsumerGroup, ConsumerGroupConsumer, ConsumerGroupPlugin,
	ACLGroup, BasicAuth, KeyAuth,
	HMACAuth, JWTAuth, OAuth2Cred,
	MTLSAuth,

	RBACRole, RBACEndpointPermission,

	ServicePackage, ServiceVersion, Document,

	Vault,
}

func entityTypeToKind(t EntityType) crud.Kind {
	return crud.Kind(t)
}

func NewEntity(t EntityType, opts EntityOpts) (Entity, error) {
	switch t {
	case Service:
		return entityImpl{
			typ: Service,
			crudActions: &serviceCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &servicePostAction{
				currentState: opts.CurrentState,
			},
			differ: &serviceDiffer{
				kind:         entityTypeToKind(Service),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Route:
		return entityImpl{
			typ: Route,
			crudActions: &routeCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &routePostAction{
				currentState: opts.CurrentState,
			},
			differ: &routeDiffer{
				kind:         entityTypeToKind(Route),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Upstream:
		return entityImpl{
			typ: Upstream,
			crudActions: &upstreamCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &upstreamPostAction{
				currentState: opts.CurrentState,
			},
			differ: &upstreamDiffer{
				kind:         entityTypeToKind(Upstream),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Target:
		return entityImpl{
			typ: Target,
			crudActions: &targetCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &targetPostAction{
				currentState: opts.CurrentState,
			},
			differ: &targetDiffer{
				kind:         entityTypeToKind(Target),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Plugin:
		return entityImpl{
			typ: Plugin,
			crudActions: &pluginCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &pluginPostAction{
				currentState: opts.CurrentState,
			},
			differ: &pluginDiffer{
				kind:         entityTypeToKind(Plugin),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Consumer:
		return entityImpl{
			typ: Consumer,
			crudActions: &consumerCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &consumerPostAction{
				currentState: opts.CurrentState,
			},
			differ: &consumerDiffer{
				kind:         entityTypeToKind(Consumer),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ConsumerGroup:
		return entityImpl{
			typ: ConsumerGroup,
			crudActions: &consumerGroupCRUD{
				client:    opts.KongClient,
				isKonnect: opts.IsKonnect,
			},
			postProcessActions: &consumerGroupPostAction{
				currentState: opts.CurrentState,
			},
			differ: &consumerGroupDiffer{
				kind:         entityTypeToKind(ConsumerGroup),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ConsumerGroupConsumer:
		return entityImpl{
			typ: ConsumerGroupConsumer,
			crudActions: &consumerGroupConsumerCRUD{
				client:    opts.KongClient,
				isKonnect: opts.IsKonnect,
			},
			postProcessActions: &consumerGroupConsumerPostAction{
				currentState: opts.CurrentState,
			},
			differ: &consumerGroupConsumerDiffer{
				kind:         entityTypeToKind(ConsumerGroupConsumer),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ConsumerGroupPlugin:
		return entityImpl{
			typ: ConsumerGroupPlugin,
			crudActions: &consumerGroupPluginCRUD{
				client:    opts.KongClient,
				isKonnect: opts.IsKonnect,
			},
			postProcessActions: &consumerGroupPluginPostAction{
				currentState: opts.CurrentState,
			},
			differ: &consumerGroupPluginDiffer{
				kind:         entityTypeToKind(ConsumerGroupPlugin),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ServicePackage:
		return entityImpl{
			typ: ServicePackage,
			crudActions: &servicePackageCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &servicePackagePostAction{
				currentState: opts.CurrentState,
			},
			differ: &servicePackageDiffer{
				kind:         entityTypeToKind(ServicePackage),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ServiceVersion:
		return entityImpl{
			typ: ServiceVersion,
			crudActions: &serviceVersionCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &serviceVersionPostAction{
				currentState: opts.CurrentState,
			},
			differ: &serviceVersionDiffer{
				kind:         entityTypeToKind(ServiceVersion),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Document:
		return entityImpl{
			typ: Document,
			crudActions: &documentCRUD{
				client: opts.KonnectClient,
			},
			postProcessActions: &documentPostAction{
				currentState: opts.CurrentState,
			},
			differ: &documentDiffer{
				kind:         entityTypeToKind(Document),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Certificate:
		return entityImpl{
			typ: Certificate,
			crudActions: &certificateCRUD{
				client:    opts.KongClient,
				isKonnect: opts.IsKonnect,
			},
			postProcessActions: &certificatePostAction{
				currentState: opts.CurrentState,
			},
			differ: &certificateDiffer{
				kind:         entityTypeToKind(Certificate),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
				isKonnect:    opts.IsKonnect,
			},
		}, nil
	case CACertificate:
		return entityImpl{
			typ: CACertificate,
			crudActions: &caCertificateCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &caCertificatePostAction{
				currentState: opts.CurrentState,
			},
			differ: &caCertificateDiffer{
				kind:         entityTypeToKind(CACertificate),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case SNI:
		return entityImpl{
			typ: SNI,
			crudActions: &sniCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &sniPostAction{
				currentState: opts.CurrentState,
			},
			differ: &sniDiffer{
				kind:         entityTypeToKind(SNI),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case RBACEndpointPermission:
		return entityImpl{
			typ: RBACEndpointPermission,
			crudActions: &rbacEndpointPermissionCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &rbacEndpointPermissionPostAction{
				currentState: opts.CurrentState,
			},
			differ: &rbacEndpointPermissionDiffer{
				kind:         entityTypeToKind(RBACEndpointPermission),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case RBACRole:
		return entityImpl{
			typ: RBACRole,
			crudActions: &rbacRoleCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &rbacRolePostAction{
				currentState: opts.CurrentState,
			},
			differ: &rbacRoleDiffer{
				kind:         entityTypeToKind(RBACRole),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case ACLGroup:
		return entityImpl{
			typ: ACLGroup,
			crudActions: &aclGroupCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &aclGroupPostAction{
				currentState: opts.CurrentState,
			},
			differ: &aclGroupDiffer{
				kind:         entityTypeToKind(ACLGroup),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case BasicAuth:
		return entityImpl{
			typ: BasicAuth,
			crudActions: &basicAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &basicAuthPostAction{
				currentState: opts.CurrentState,
			},
			differ: &basicAuthDiffer{
				kind:         entityTypeToKind(BasicAuth),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case KeyAuth:
		return entityImpl{
			typ: KeyAuth,
			crudActions: &keyAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &keyAuthPostAction{
				currentState: opts.CurrentState,
			},
			differ: &keyAuthDiffer{
				kind:         entityTypeToKind(KeyAuth),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case HMACAuth:
		return entityImpl{
			typ: HMACAuth,
			crudActions: &hmacAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &hmacAuthPostAction{
				currentState: opts.CurrentState,
			},
			differ: &hmacAuthDiffer{
				kind:         entityTypeToKind(HMACAuth),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case JWTAuth:
		return entityImpl{
			typ: JWTAuth,
			crudActions: &jwtAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &jwtAuthPostAction{
				currentState: opts.CurrentState,
			},
			differ: &jwtAuthDiffer{
				kind:         entityTypeToKind(JWTAuth),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case MTLSAuth:
		return entityImpl{
			typ: MTLSAuth,
			crudActions: &mtlsAuthCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &mtlsAuthPostAction{
				currentState: opts.CurrentState,
			},
			differ: &mtlsAuthDiffer{
				kind:         entityTypeToKind(MTLSAuth),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case OAuth2Cred:
		return entityImpl{
			typ: OAuth2Cred,
			crudActions: &oauth2CredCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &oauth2CredPostAction{
				currentState: opts.CurrentState,
			},
			differ: &oauth2CredDiffer{
				kind:         entityTypeToKind(OAuth2Cred),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	case Vault:
		return entityImpl{
			typ: Vault,
			crudActions: &vaultCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &vaultPostAction{
				currentState: opts.CurrentState,
			},
			differ: &vaultDiffer{
				kind:         entityTypeToKind(Vault),
				currentState: opts.CurrentState,
				targetState:  opts.TargetState,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown type: %q", t)
	}
}
