package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
)

type servicePostAction struct {
	currentState *state.KongState
}

func (crud *servicePostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Add(*args[0].(*state.Service))
}

func (crud *servicePostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Delete(*((args[0].(*state.Service)).ID))
}

func (crud *servicePostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Update(*args[0].(*state.Service))
}

type routePostAction struct {
	currentState *state.KongState
}

func (crud *routePostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Add(*args[0].(*state.Route))
}

func (crud *routePostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Delete(*((args[0].(*state.Route)).ID))
}

func (crud *routePostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Update(*args[0].(*state.Route))
}

type upstreamPostAction struct {
	currentState *state.KongState
}

func (crud *upstreamPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Add(*args[0].(*state.Upstream))
}

func (crud *upstreamPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Delete(*((args[0].(*state.Upstream)).ID))
}

func (crud *upstreamPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Update(*args[0].(*state.Upstream))
}

type targetPostAction struct {
	currentState *state.KongState
}

func (crud *targetPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Targets.Add(*args[0].(*state.Target))
}

func (crud *targetPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	target := args[0].(*state.Target)
	return nil, crud.currentState.Targets.Delete(*target.Upstream.ID, *target.ID)
}

func (crud *targetPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Targets.Update(*args[0].(*state.Target))
}

type certificatePostAction struct {
	currentState *state.KongState
}

func (crud *certificatePostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Add(*args[0].(*state.Certificate))
}

func (crud *certificatePostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Delete(*((args[0].(*state.Certificate)).ID))
}

func (crud *certificatePostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Update(*args[0].(*state.Certificate))
}

type sniPostAction struct {
	currentState *state.KongState
}

func (crud *sniPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.SNIs.Add(*args[0].(*state.SNI))
}

func (crud *sniPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	sni := args[0].(*state.SNI)
	return nil, crud.currentState.SNIs.Delete(*sni.ID)
}

func (crud *sniPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.SNIs.Update(*args[0].(*state.SNI))
}

type caCertificatePostAction struct {
	currentState *state.KongState
}

func (crud *caCertificatePostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Add(*args[0].(*state.CACertificate))
}

func (crud *caCertificatePostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Delete(*((args[0].(*state.CACertificate)).ID))
}

func (crud *caCertificatePostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Update(*args[0].(*state.CACertificate))
}

type pluginPostAction struct {
	currentState *state.KongState
}

func (crud *pluginPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Add(*args[0].(*state.Plugin))
}

func (crud *pluginPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Delete(*((args[0].(*state.Plugin)).ID))
}

func (crud *pluginPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Update(*args[0].(*state.Plugin))
}

type consumerPostAction struct {
	currentState *state.KongState
}

func (crud *consumerPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Add(*args[0].(*state.Consumer))
}

func (crud *consumerPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Delete(*((args[0].(*state.Consumer)).ID))
}

func (crud *consumerPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Update(*args[0].(*state.Consumer))
}

type keyAuthPostAction struct {
	currentState *state.KongState
}

func (crud *keyAuthPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Add(*args[0].(*state.KeyAuth))
}

func (crud *keyAuthPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Delete(*((args[0].(*state.KeyAuth)).ID))
}

func (crud *keyAuthPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Update(*args[0].(*state.KeyAuth))
}

type hmacAuthPostAction struct {
	currentState *state.KongState
}

func (crud hmacAuthPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Add(*args[0].(*state.HMACAuth))
}

func (crud hmacAuthPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Delete(*((args[0].(*state.HMACAuth)).ID))
}

func (crud hmacAuthPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Update(*args[0].(*state.HMACAuth))
}

type jwtAuthPostAction struct {
	currentState *state.KongState
}

func (crud jwtAuthPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Add(*args[0].(*state.JWTAuth))
}

func (crud jwtAuthPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Delete(*((args[0].(*state.JWTAuth)).ID))
}

func (crud jwtAuthPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Update(*args[0].(*state.JWTAuth))
}

type basicAuthPostAction struct {
	currentState *state.KongState
}

func (crud basicAuthPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Add(*args[0].(*state.BasicAuth))
}

func (crud basicAuthPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Delete(*((args[0].(*state.BasicAuth)).ID))
}

func (crud basicAuthPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Update(*args[0].(*state.BasicAuth))
}

type aclGroupPostAction struct {
	currentState *state.KongState
}

func (crud aclGroupPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Add(*args[0].(*state.ACLGroup))
}

func (crud aclGroupPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Delete(*((args[0].(*state.ACLGroup)).ID))
}

func (crud aclGroupPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Update(*args[0].(*state.ACLGroup))
}

type oauth2CredPostAction struct {
	currentState *state.KongState
}

func (crud oauth2CredPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Add(*args[0].(*state.Oauth2Credential))
}

func (crud oauth2CredPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Delete(*((args[0].(*state.Oauth2Credential)).ID))
}

func (crud oauth2CredPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Update(*args[0].(*state.Oauth2Credential))
}

type mtlsAuthPostAction struct {
	currentState *state.KongState
}

func (crud *mtlsAuthPostAction) Create(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Add(*args[0].(*state.MTLSAuth))
}

func (crud *mtlsAuthPostAction) Delete(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Delete(*((args[0].(*state.MTLSAuth)).ID))
}

func (crud *mtlsAuthPostAction) Update(args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Update(*args[0].(*state.MTLSAuth))
}
