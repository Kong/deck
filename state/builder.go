package state

import (
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

// Get builds a KongState from a raw representation of Kong.
func Get(raw *utils.KongRawState) (*KongState, error) {
	kongState, err := NewKongState()
	if err != nil {
		return nil, errors.Wrap(err, "creating new in-memory state of Kong")
	}
	err = buildKong(kongState, raw)
	if err != nil {
		return nil, err
	}
	return kongState, nil
}

func buildKong(kongState *KongState, raw *utils.KongRawState) error {
	for _, s := range raw.Services {
		err := kongState.Services.Add(Service{Service: *s})
		if err != nil {
			return errors.Wrap(err, "inserting service into state")
		}
	}
	for _, r := range raw.Routes {
		err := kongState.Routes.Add(Route{Route: *r})
		if err != nil {
			return errors.Wrap(err, "inserting route into state")
		}
	}
	for _, c := range raw.Consumers {
		err := kongState.Consumers.Add(Consumer{Consumer: *c})
		if err != nil {
			return errors.Wrap(err, "inserting consumer into state")
		}
	}
	ensureConsumer := func(consumerID string) (bool, error) {
		_, err := kongState.Consumers.Get(consumerID)
		if err != nil {
			if err == ErrNotFound {
				return false, nil
			}
			return false, errors.Wrapf(err,
				"looking up consumer '%v'", consumerID)
		}
		return true, nil
	}
	for _, cred := range raw.KeyAuths {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.KeyAuths.Add(KeyAuth{KeyAuth: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting key-auth into state")
		}
	}
	for _, cred := range raw.HMACAuths {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.HMACAuths.Add(HMACAuth{HMACAuth: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting hmac-auth into state")
		}
	}
	for _, cred := range raw.JWTAuths {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.JWTAuths.Add(JWTAuth{JWTAuth: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting jwt into state")
		}
	}
	for _, cred := range raw.BasicAuths {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.BasicAuths.Add(BasicAuth{BasicAuth: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting basic-auth into state")
		}
	}
	for _, cred := range raw.Oauth2Creds {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.Oauth2Creds.Add(Oauth2Credential{Oauth2Credential: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting oauth2-cred into state")
		}
	}
	for _, cred := range raw.ACLGroups {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.ACLGroups.Add(ACLGroup{ACLGroup: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting basic-auth into state")
		}
	}
	for _, cred := range raw.MTLSAuths {
		ok, err := ensureConsumer(*cred.Consumer.ID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		err = kongState.MTLSAuths.Add(MTLSAuth{MTLSAuth: *cred})
		if err != nil {
			return errors.Wrap(err, "inserting mtls-auth into state")
		}
	}
	for _, u := range raw.Upstreams {
		err := kongState.Upstreams.Add(Upstream{Upstream: *u})
		if err != nil {
			return errors.Wrap(err, "inserting upstream into state")
		}
	}
	for _, t := range raw.Targets {
		err := kongState.Targets.Add(Target{Target: *t})
		if err != nil {
			return errors.Wrap(err, "inserting target into state")
		}
	}

	for _, c := range raw.Certificates {
		err := kongState.Certificates.Add(Certificate{Certificate: *c})
		if err != nil {
			return errors.Wrap(err, "inserting certificate into state")
		}
	}

	for _, s := range raw.SNIs {
		err := kongState.SNIs.Add(SNI{SNI: *s})
		if err != nil {
			return errors.Wrap(err, "inserting sni into state")
		}
	}

	for _, c := range raw.CACertificates {
		err := kongState.CACertificates.Add(CACertificate{
			CACertificate: *c,
		})
		if err != nil {
			return errors.Wrap(err, "inserting ca_certificate into state")
		}
	}

	for _, p := range raw.Plugins {
		err := kongState.Plugins.Add(Plugin{Plugin: *p})
		if err != nil {
			return errors.Wrap(err, "inserting plugins into state")
		}
	}

	for _, r := range raw.RBACRoles {
		err := kongState.RBACRoles.Add(RBACRole{RBACRole: *r})
		if err != nil {
			return errors.Wrap(err, "inserting rbac roles into state")
		}
	}
	for _, r := range raw.RBACEndpointPermissions {
		err := kongState.RBACEndpointPermissions.Add(RBACEndpointPermission{RBACEndpointPermission: *r})
		if err != nil {
			return errors.Wrap(err, "inserting rbac endpoint permissions into state")
		}
	}
	return nil
}

func buildKonnect(kongState *KongState, raw *utils.KonnectRawState) error {
	for _, s := range raw.ServicePackages {
		servicePackage := s.DeepCopy()
		servicePackage.Versions = nil
		err := kongState.ServicePackages.Add(ServicePackage{
			ServicePackage: *servicePackage,
		})
		if err != nil {
			return errors.Wrap(err, "inserting service-package into state")
		}

		for _, v := range s.Versions {
			v = *v.DeepCopy()
			v.ServicePackage = servicePackage.DeepCopy()
			err := kongState.ServiceVersions.Add(ServiceVersion{
				ServiceVersion: v,
			})
			if err != nil {
				return errors.Wrap(err, "inserting service-version into state")
			}
		}
	}
	for _, d := range raw.Documents {
		document := d.ShallowCopy()
		err := kongState.Documents.Add(Document{
			Document: *document,
		})
		if err != nil {
			return errors.Wrap(err, "inserting document into state")
		}
	}
	return nil
}

func GetKonnectState(rawKong *utils.KongRawState,
	rawKonnect *utils.KonnectRawState) (*KongState, error) {
	kongState, err := NewKongState()
	if err != nil {
		return nil, errors.Wrap(err, "creating new in-memory state of Kong")
	}

	err = buildKong(kongState, rawKong)
	if err != nil {
		return nil, err
	}

	err = buildKonnect(kongState, rawKonnect)
	if err != nil {
		return nil, err
	}
	return kongState, nil
}
