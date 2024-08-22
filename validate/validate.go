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
	ctx               context.Context
	state             *state.KongState
	client            *kong.Client
	parallelism       int
	rbacResourcesOnly bool
}

type ValidatorOpts struct {
	Ctx               context.Context
	State             *state.KongState
	Client            *kong.Client
	Parallelism       int
	RBACResourcesOnly bool
}

func NewValidator(opt ValidatorOpts) *Validator {
	return &Validator{
		ctx:               opt.Ctx,
		state:             opt.State,
		client:            opt.Client,
		parallelism:       opt.Parallelism,
		rbacResourcesOnly: opt.RBACResourcesOnly,
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
	req, err := v.client.NewRequest(http.MethodPost, endpoint, nil, entity)
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, nameOrID, err)
	}
	resp, err := v.client.Do(v.ctx, req, nil)
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, nameOrID, err)
	}
	return resp.StatusCode == http.StatusOK, nil
}

func (v *Validator) entities(obj interface{}, entityType string) []error {
	entities, err := reconcilerUtils.CallGetAll(obj)
	if err != nil {
		return []error{err}
	}
	errors := []error{}

	// create a buffer of channels. Creation of new coroutines
	// are allowed only if the buffer is not full.
	chanBuff := make(chan struct{}, v.parallelism)

	var wg sync.WaitGroup
	wg.Add(entities.Len())
	// each coroutine will append on a slice of errors.
	// since slices are not thread-safe, let's add a mutex
	// to handle access to the slice.
	mu := &sync.Mutex{}
	for i := 0; i < entities.Len(); i++ {
		// reserve a slot
		chanBuff <- struct{}{}
		go func(i int) {
			defer wg.Done()
			// release a slot when completed
			defer func() { <-chanBuff }()
			_, err := v.validateEntity(entityType, entities.Index(i).Interface())
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

func (v *Validator) Validate(formatVersion semver.Version) []error {
	allErr := []error{}

	// validate RBAC resources first.
	if err := v.entities(v.state.RBACEndpointPermissions, "rbac-endpointpermission"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.RBACRoles, "rbac-role"); err != nil {
		allErr = append(allErr, err...)
	}
	if v.rbacResourcesOnly {
		return allErr
	}

	if err := v.entities(v.state.Services, "services"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.ACLGroups, "acls"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.BasicAuths, "basicauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.CACertificates, "ca_certificates"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Certificates, "certificates"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Consumers, "consumers"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Documents, "documents"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.HMACAuths, "hmacauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.JWTAuths, "jwt_secrets"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.KeyAuths, "keyauth_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Oauth2Creds, "oauth2_credentials"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Plugins, "plugins"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Routes, "routes"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.SNIs, "snis"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Targets, "targets"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.Upstreams, "upstreams"); err != nil {
		allErr = append(allErr, err...)
	}
	if err := v.entities(v.state.FilterChains, "filter_chains"); err != nil {
		allErr = append(allErr, err...)
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
			unsupportedRoutes = append(unsupportedRoutes, *r.Route.ID)
		}
	}
	if len(unsupportedRoutes) > 0 {
		reconcilerUtils.PrintRouteRegexWarning(unsupportedRoutes)
	}
}
