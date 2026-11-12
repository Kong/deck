package validate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/utils"
	"github.com/kong/go-database-reconciler/pkg/state"
	reconcilerUtils "github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

type Validator struct {
	ctx                  context.Context
	state                *state.KongState
	client               *kong.Client
	configClient         *kong.Client
	configChunkSize      int
	parallelism          int
	rbacResourcesOnly    bool
	onlineEntitiesFilter []string
}

type ValidatorOpts struct {
	Ctx    context.Context
	State  *state.KongState
	Client *kong.Client
	// ConfigClient sends the config validation requests. These can take longer than other requests.
	// If it is nil, Client sends them.
	ConfigClient *kong.Client
	// ConfigChunkSize is the maximum size in bytes of one config validation request.
	// If it is 0, DefaultConfigChunkSize applies.
	ConfigChunkSize      int
	Parallelism          int
	RBACResourcesOnly    bool
	OnlineEntitiesFilter []string
}

// Define a map of entity object field names and their corresponding string names
var EntityMap = map[string]string{
	"ACLGroups":               "acls",
	"BasicAuths":              "basicauth_credentials",
	"CACertificates":          "ca_certificates",
	"Certificates":            "certificates",
	"Consumers":               "consumers",
	"Documents":               "documents",
	"FilterChains":            "filter_chains",
	"HMACAuths":               "hmacauth_credentials",
	"JWTAuths":                "jwt_secrets",
	"KeyAuths":                "keyauth_credentials",
	"Oauth2Creds":             "oauth2_credentials",
	"Partials":                "partials",
	"Plugins":                 "plugins",
	"RBACEndpointPermissions": "rbac-endpointpermission",
	"RBACRoles":               "rbac-role",
	"Routes":                  "routes",
	"SNIs":                    "snis",
	"Services":                "services",
	"Targets":                 "targets",
	"Upstreams":               "upstreams",
	"Vaults":                  "vaults",
}

func NewValidator(opt ValidatorOpts) *Validator {
	configClient := opt.ConfigClient
	if configClient == nil {
		configClient = opt.Client
	}
	configChunkSize := opt.ConfigChunkSize
	if configChunkSize <= 0 {
		configChunkSize = DefaultConfigChunkSize
	}
	return &Validator{
		ctx:                  opt.Ctx,
		state:                opt.State,
		client:               opt.Client,
		configClient:         configClient,
		configChunkSize:      configChunkSize,
		parallelism:          opt.Parallelism,
		rbacResourcesOnly:    opt.RBACResourcesOnly,
		onlineEntitiesFilter: opt.OnlineEntitiesFilter,
	}
}

type ErrorsWrapper struct {
	Errors []error
}

func (v ErrorsWrapper) Error() string {
	var errStr string
	for _, e := range v.Errors {
		errStr += e.Error()
		if !errors.Is(e, v.Errors[len(v.Errors)-1]) {
			errStr += "\n"
		}
	}
	return errStr
}

func getEntityNameOrID(entity interface{}) string {
	value := reflect.ValueOf(entity).Elem()
	nameOrID := value.FieldByName("Name")
	if !nameOrID.IsValid() {
		nameOrID = value.FieldByName("ID")
	}
	return nameOrID.Elem().String()
}

func (v *Validator) validateEntity(entityType string, entity interface{}) (bool, error) {
	nameOrID := getEntityNameOrID(entity)
	errWrap := "validate entity '%s (%s)': %s"
	endpoint := fmt.Sprintf("/schemas/%s/validate", entityType)
	var req *http.Request
	var err error
	if v.client.IsKonnectMode() {
		req, err = v.client.NewKonnectWorkspaceRequest(http.MethodPost, endpoint, nil, entity)
	} else {
		req, err = v.client.NewRequest(http.MethodPost, endpoint, nil, entity)
	}
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, nameOrID, err)
	}
	resp, err := v.client.Do(v.ctx, req, nil)
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, nameOrID, err)
	}
	return resp.StatusCode == http.StatusOK, nil
}

func getEmbeddedTypeFromWrappedType(wrappedType interface{}, entityType string) interface{} {
	switch entityType {
	case "basicauth_credentials":
		if wrappedType == nil {
			return wrappedType
		}
		// Prefer the pointer case (most common)
		if baPtr, ok := wrappedType.(*state.BasicAuth); ok && baPtr != nil {
			return &baPtr.BasicAuth
		}
		// Also handle a value (non-pointer) just in case
		if baVal, ok := wrappedType.(state.BasicAuth); ok {
			return &baVal.BasicAuth
		}
		// Fallback: unknown shape, return as-is to avoid panics
		return wrappedType
	}

	return wrappedType
}

func (v *Validator) entities(obj interface{}, entityType string) []error {
	entities, err := reconcilerUtils.CallGetAll(obj)
	if err != nil {
		return []error{err}
	}
	list := make([]interface{}, 0, entities.Len())
	for i := 0; i < entities.Len(); i++ {
		list = append(list, entities.Index(i).Interface())
	}
	return v.validateEntities(entityType, list)
}

// validateEntities sends one validation request for each entity.
func (v *Validator) validateEntities(entityType string, entities []interface{}) []error {
	errors := []error{}

	// create a buffer of channels. Creation of new coroutines
	// are allowed only if the buffer is not full.
	chanBuff := make(chan struct{}, v.parallelism)

	var wg sync.WaitGroup
	wg.Add(len(entities))
	// each coroutine will append on a slice of errors.
	// since slices are not thread-safe, let's add a mutex
	// to handle access to the slice.
	mu := &sync.Mutex{}
	for i := range entities {
		// reserve a slot
		chanBuff <- struct{}{}
		go func(i int) {
			defer wg.Done()
			// release a slot when completed
			defer func() { <-chanBuff }()
			embeddedEntity := getEmbeddedTypeFromWrappedType(entities[i], entityType)
			_, err := v.validateEntity(entityType, embeddedEntity)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return errors
}

// validateStateField validates each entity of the KongState field fieldName.
func (v *Validator) validateStateField(fieldName, entityType string) []error {
	valueOfState := reflect.ValueOf(v.state)
	if valueOfState.Kind() == reflect.Pointer {
		valueOfState = valueOfState.Elem()
	}
	fieldValue := valueOfState.FieldByName(fieldName)
	if !fieldValue.IsValid() || !fieldValue.CanInterface() {
		return []error{fmt.Errorf("invalid field '%s' in state", fieldName)}
	}
	return v.entities(fieldValue.Interface(), entityType)
}

func (v *Validator) Validate(formatVersion semver.Version) []error {
	allErr := []error{}

	if v.rbacResourcesOnly {
		// validate RBAC resources first.
		if err := v.entities(v.state.RBACEndpointPermissions, "rbac-endpointpermission"); err != nil {
			allErr = append(allErr, err...)
		}
		if err := v.entities(v.state.RBACRoles, "rbac-role"); err != nil {
			allErr = append(allErr, err...)
		}
		return allErr
	}

	// Create a copy of entityMap with only the specififed resources to check online.
	filteredEntityMap := make(map[string]string)
	if len(v.onlineEntitiesFilter) > 0 {
		for _, value := range v.onlineEntitiesFilter {
			for key, entityName := range EntityMap {
				if value == key {
					filteredEntityMap[key] = entityName
				}
			}
		}
	} else {
		// If no filter is specified, use the original entityMap.
		filteredEntityMap = EntityMap
	}

	if errs, ok := v.validateWithConfigEndpoint(filteredEntityMap); ok {
		allErr = append(allErr, errs...)
	} else {
		allErr = append(allErr, v.localConflictErrors(filteredEntityMap)...)
		for fieldName, entityName := range filteredEntityMap {
			allErr = append(allErr, v.validateStateField(fieldName, entityName)...)
		}
	}

	// validate routes format with Kong 3.x
	parsed30, err := semver.ParseTolerant(utils.FormatVersion30)
	if err != nil {
		allErr = append(allErr, err)
	}
	if parsed30.LTE(formatVersion) {
		validate3xRoutes(v.state.Routes)
	}
	return allErr
}

func validate3xRoutes(routes *state.RoutesCollection) {
	results, _ := routes.GetAll()
	unsupportedRoutes := []string{}
	for _, r := range results {
		if reconcilerUtils.HasPathsWithRegex300AndAbove(r.Route) {
			unsupportedRoutes = append(unsupportedRoutes, *r.ID)
		}
	}
	if len(unsupportedRoutes) > 0 {
		reconcilerUtils.PrintRouteRegexWarning(unsupportedRoutes)
	}
}
