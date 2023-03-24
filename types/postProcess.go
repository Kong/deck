package types

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

type servicePostAction struct {
	currentState *state.KongState
}

func (crud *servicePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Add(*args[0].(*state.Service))
}

func (crud *servicePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Delete(*((args[0].(*state.Service)).ID))
}

func (crud *servicePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Services.Update(*args[0].(*state.Service))
}

type routePostAction struct {
	currentState *state.KongState
}

func (crud *routePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Add(*args[0].(*state.Route))
}

func (crud *routePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Delete(*((args[0].(*state.Route)).ID))
}

func (crud *routePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Routes.Update(*args[0].(*state.Route))
}

type upstreamPostAction struct {
	currentState *state.KongState
}

func (crud *upstreamPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Add(*args[0].(*state.Upstream))
}

func (crud *upstreamPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Delete(*((args[0].(*state.Upstream)).ID))
}

func (crud *upstreamPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Upstreams.Update(*args[0].(*state.Upstream))
}

type targetPostAction struct {
	currentState *state.KongState
}

func (crud *targetPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Targets.Add(*args[0].(*state.Target))
}

func (crud *targetPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	target := args[0].(*state.Target)
	return nil, crud.currentState.Targets.Delete(*target.Upstream.ID, *target.ID)
}

func (crud *targetPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Targets.Update(*args[0].(*state.Target))
}

type certificatePostAction struct {
	currentState *state.KongState
}

func (crud *certificatePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Add(*args[0].(*state.Certificate))
}

func (crud *certificatePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Delete(*((args[0].(*state.Certificate)).ID))
}

func (crud *certificatePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Certificates.Update(*args[0].(*state.Certificate))
}

type sniPostAction struct {
	currentState *state.KongState
}

func (crud *sniPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.SNIs.Add(*args[0].(*state.SNI))
}

func (crud *sniPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	sni := args[0].(*state.SNI)
	return nil, crud.currentState.SNIs.Delete(*sni.ID)
}

func (crud *sniPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.SNIs.Update(*args[0].(*state.SNI))
}

type caCertificatePostAction struct {
	currentState *state.KongState
}

func (crud *caCertificatePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Add(*args[0].(*state.CACertificate))
}

func (crud *caCertificatePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Delete(*((args[0].(*state.CACertificate)).ID))
}

func (crud *caCertificatePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.CACertificates.Update(*args[0].(*state.CACertificate))
}

type pluginPostAction struct {
	currentState *state.KongState
}

func (crud *pluginPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Add(*args[0].(*state.Plugin))
}

func (crud *pluginPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Delete(*((args[0].(*state.Plugin)).ID))
}

func (crud *pluginPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Plugins.Update(*args[0].(*state.Plugin))
}

type consumerPostAction struct {
	currentState *state.KongState
}

func (crud *consumerPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Add(*args[0].(*state.Consumer))
}

func (crud *consumerPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Delete(*((args[0].(*state.Consumer)).ID))
}

func (crud *consumerPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Consumers.Update(*args[0].(*state.Consumer))
}

type consumerGroupPostAction struct {
	currentState *state.KongState
}

func (crud *consumerGroupPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroups.Add(*args[0].(*state.ConsumerGroup))
}

func (crud *consumerGroupPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroups.Delete(*((args[0].(*state.ConsumerGroup)).ID))
}

func (crud *consumerGroupPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroups.Update(*args[0].(*state.ConsumerGroup))
}

type consumerGroupConsumerPostAction struct {
	currentState *state.KongState
}

func (crud *consumerGroupConsumerPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupConsumers.Add(*args[0].(*state.ConsumerGroupConsumer))
}

func (crud *consumerGroupConsumerPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupConsumers.Delete(
		*((args[0].(*state.ConsumerGroupConsumer)).Consumer.ID),
		*((args[0].(*state.ConsumerGroupConsumer)).ConsumerGroup.ID),
	)
}

func (crud *consumerGroupConsumerPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupConsumers.Update(*args[0].(*state.ConsumerGroupConsumer))
}

type consumerGroupPluginPostAction struct {
	currentState *state.KongState
}

func (crud *consumerGroupPluginPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupPlugins.Add(*args[0].(*state.ConsumerGroupPlugin))
}

func (crud *consumerGroupPluginPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupPlugins.Delete(
		*((args[0].(*state.ConsumerGroupPlugin)).ID),
		*((args[0].(*state.ConsumerGroupConsumer)).ConsumerGroup.ID),
	)
}

func (crud *consumerGroupPluginPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ConsumerGroupPlugins.Update(*args[0].(*state.ConsumerGroupPlugin))
}

type keyAuthPostAction struct {
	currentState *state.KongState
}

func (crud *keyAuthPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Add(*args[0].(*state.KeyAuth))
}

func (crud *keyAuthPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Delete(*((args[0].(*state.KeyAuth)).ID))
}

func (crud *keyAuthPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.KeyAuths.Update(*args[0].(*state.KeyAuth))
}

type hmacAuthPostAction struct {
	currentState *state.KongState
}

func (crud hmacAuthPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Add(*args[0].(*state.HMACAuth))
}

func (crud hmacAuthPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Delete(*((args[0].(*state.HMACAuth)).ID))
}

func (crud hmacAuthPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.HMACAuths.Update(*args[0].(*state.HMACAuth))
}

type jwtAuthPostAction struct {
	currentState *state.KongState
}

func (crud jwtAuthPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Add(*args[0].(*state.JWTAuth))
}

func (crud jwtAuthPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Delete(*((args[0].(*state.JWTAuth)).ID))
}

func (crud jwtAuthPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.JWTAuths.Update(*args[0].(*state.JWTAuth))
}

type basicAuthPostAction struct {
	currentState *state.KongState
}

func (crud basicAuthPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Add(*args[0].(*state.BasicAuth))
}

func (crud basicAuthPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Delete(*((args[0].(*state.BasicAuth)).ID))
}

func (crud basicAuthPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.BasicAuths.Update(*args[0].(*state.BasicAuth))
}

type aclGroupPostAction struct {
	currentState *state.KongState
}

func (crud aclGroupPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Add(*args[0].(*state.ACLGroup))
}

func (crud aclGroupPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Delete(*((args[0].(*state.ACLGroup)).ID))
}

func (crud aclGroupPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ACLGroups.Update(*args[0].(*state.ACLGroup))
}

type oauth2CredPostAction struct {
	currentState *state.KongState
}

func (crud oauth2CredPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Add(*args[0].(*state.Oauth2Credential))
}

func (crud oauth2CredPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Delete(*((args[0].(*state.Oauth2Credential)).ID))
}

func (crud oauth2CredPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Oauth2Creds.Update(*args[0].(*state.Oauth2Credential))
}

type mtlsAuthPostAction struct {
	currentState *state.KongState
}

func (crud *mtlsAuthPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Add(*args[0].(*state.MTLSAuth))
}

func (crud *mtlsAuthPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Delete(*((args[0].(*state.MTLSAuth)).ID))
}

func (crud *mtlsAuthPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.MTLSAuths.Update(*args[0].(*state.MTLSAuth))
}

type rbacRolePostAction struct {
	currentState *state.KongState
}

func (crud *rbacRolePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACRoles.Add(*args[0].(*state.RBACRole))
}

func (crud *rbacRolePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACRoles.Delete(*((args[0].(*state.RBACRole)).ID))
}

func (crud *rbacRolePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACRoles.Update(*args[0].(*state.RBACRole))
}

type rbacEndpointPermissionPostAction struct {
	currentState *state.KongState
}

func (crud *rbacEndpointPermissionPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACEndpointPermissions.Add(*args[0].(*state.RBACEndpointPermission))
}

func (crud *rbacEndpointPermissionPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACEndpointPermissions.Delete(args[0].(*state.RBACEndpointPermission).FriendlyName())
}

func (crud *rbacEndpointPermissionPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.RBACEndpointPermissions.Update(*args[0].(*state.RBACEndpointPermission))
}

type servicePackagePostAction struct {
	currentState *state.KongState
}

func (crud servicePackagePostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ServicePackages.Add(*args[0].(*state.ServicePackage))
}

func (crud servicePackagePostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ServicePackages.Delete(*((args[0].(*state.ServicePackage)).ID))
}

func (crud servicePackagePostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ServicePackages.Update(*args[0].(*state.ServicePackage))
}

type serviceVersionPostAction struct {
	currentState *state.KongState
}

func (crud serviceVersionPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ServiceVersions.Add(*args[0].(*state.ServiceVersion))
}

func (crud serviceVersionPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	sv := args[0].(*state.ServiceVersion)
	return nil, crud.currentState.ServiceVersions.Delete(*sv.ServicePackage.ID, *sv.ID)
}

func (crud serviceVersionPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.ServiceVersions.Update(*args[0].(*state.ServiceVersion))
}

type documentPostAction struct {
	currentState *state.KongState
}

func (crud documentPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Documents.Add(*args[0].(*state.Document))
}

func (crud documentPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	d := args[0].(*state.Document)
	return nil, crud.currentState.Documents.DeleteByParent(d.Parent, *d.ID)
}

func (crud documentPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Documents.Update(*args[0].(*state.Document))
}

type vaultPostAction struct {
	currentState *state.KongState
}

func (crud vaultPostAction) Create(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Vaults.Add(*args[0].(*state.Vault))
}

func (crud vaultPostAction) Delete(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Vaults.Delete(*((args[0].(*state.Vault)).ID))
}

func (crud vaultPostAction) Update(_ context.Context, args ...crud.Arg) (crud.Arg, error) {
	return nil, crud.currentState.Vaults.Update(*args[0].(*state.Vault))
}
