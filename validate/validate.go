package validate

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

type Validator struct {
	ctx               context.Context
	state             *state.KongState
	client            *kong.Client
	parallelism       int
	rbacResourcesOnly bool
}

func NewValidator(
	ctx context.Context,
	ks *state.KongState,
	client *kong.Client,
	parallelism int,
	rbacResourcesOnly bool,
) *Validator {
	return &Validator{
		ctx:               ctx,
		state:             ks,
		client:            client,
		parallelism:       parallelism,
		rbacResourcesOnly: rbacResourcesOnly,
	}
}

type ErrorsWrapper struct {
	Errors []error
}

func (v ErrorsWrapper) Error() string {
	var errStr string
	for _, e := range v.Errors {
		errStr += fmt.Sprintf("%s\n", e.Error())
	}
	return errStr
}

func (v *Validator) validateEntity(entityType string, entity interface{}) (bool, error) {
	errWrap := "validate entity '%s': %s"
	endpoint := fmt.Sprintf("/schemas/%s/validate", entityType)
	req, err := v.client.NewRequest(http.MethodPost, endpoint, nil, entity)
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, err)
	}
	resp, err := v.client.Do(v.ctx, req, nil)
	if err != nil {
		return false, fmt.Errorf(errWrap, entityType, err)
	}
	return resp.StatusCode == http.StatusOK, nil
}

func (v *Validator) entities(obj interface{}, entityType string) []error {
	entities, err := utils.CallGetAll(obj)
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

func (v *Validator) Validate() []error {
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
	return allErr
}
