package state

import (
	"fmt"

	"github.com/kong/deck/utils"
)

// Get builds a KongState from a raw representation of Kong.
func Get(raw *utils.KongRawState) (*KongState, error) {
	kongState, err := NewKongState()
	if err != nil {
		return nil, fmt.Errorf("creating new in-memory state of Kong: %w", err)
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
			return fmt.Errorf("inserting service into state: %w", err)
		}
	}
	for _, r := range raw.Routes {
		err := kongState.Routes.Add(Route{Route: *r})
		if err != nil {
			return fmt.Errorf("inserting route into state: %w", err)
		}
	}
	for _, c := range raw.Consumers {
		err := kongState.Consumers.Add(Consumer{Consumer: *c})
		if err != nil {
			return fmt.Errorf("inserting consumer into state: %w", err)
		}
	}
	ensureConsumer := func(consumerID string) (bool, error) {
		_, err := kongState.Consumers.Get(consumerID)
		if err != nil {
			if err == ErrNotFound {
				return false, nil
			}
			return false, fmt.Errorf("looking up consumer %q: %w", consumerID, err)

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
			return fmt.Errorf("inserting key-auth into state: %w", err)
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
			return fmt.Errorf("inserting hmac-auth into state: %w", err)
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
			return fmt.Errorf("inserting jwt into state: %w", err)
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
			return fmt.Errorf("inserting basic-auth into state: %w", err)
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
			return fmt.Errorf("inserting oauth2-cred into state: %w", err)
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
			return fmt.Errorf("inserting basic-auth into state: %w", err)
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
			return fmt.Errorf("inserting mtls-auth into state: %w", err)
		}
	}
	for _, u := range raw.Upstreams {
		err := kongState.Upstreams.Add(Upstream{Upstream: *u})
		if err != nil {
			return fmt.Errorf("inserting upstream into state: %w", err)
		}
	}
	for _, t := range raw.Targets {
		err := kongState.Targets.Add(Target{Target: *t})
		if err != nil {
			return fmt.Errorf("inserting target into state: %w", err)
		}
	}

	for _, c := range raw.Certificates {
		err := kongState.Certificates.Add(Certificate{Certificate: *c})
		if err != nil {
			return fmt.Errorf("inserting certificate into state: %w", err)
		}
	}

	for _, s := range raw.SNIs {
		err := kongState.SNIs.Add(SNI{SNI: *s})
		if err != nil {
			return fmt.Errorf("inserting sni into state: %w", err)
		}
	}

	for _, c := range raw.CACertificates {
		err := kongState.CACertificates.Add(CACertificate{
			CACertificate: *c,
		})
		if err != nil {
			return fmt.Errorf("inserting ca_certificate into state: %w", err)
		}
	}

	for _, p := range raw.Plugins {
		err := kongState.Plugins.Add(Plugin{Plugin: *p})
		if err != nil {
			return fmt.Errorf("inserting plugins into state: %w", err)
		}
	}

	for _, r := range raw.RBACRoles {
		err := kongState.RBACRoles.Add(RBACRole{RBACRole: *r})
		if err != nil {
			return fmt.Errorf("inserting rbac roles into state: %w", err)
		}
	}
	for _, r := range raw.RBACEndpointPermissions {
		err := kongState.RBACEndpointPermissions.Add(RBACEndpointPermission{RBACEndpointPermission: *r})
		if err != nil {
			return fmt.Errorf("inserting rbac endpoint permissions into state: %w", err)
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
			return fmt.Errorf("inserting service-package into state: %w", err)
		}

		for _, v := range s.Versions {
			v = *v.DeepCopy()
			v.ServicePackage = servicePackage.DeepCopy()
			err := kongState.ServiceVersions.Add(ServiceVersion{
				ServiceVersion: v,
			})
			if err != nil {
				return fmt.Errorf("inserting service-version into state: %w", err)
			}
		}
	}
	for _, d := range raw.Documents {
		document := d.ShallowCopy()
		err := kongState.Documents.Add(Document{
			Document: *document,
		})
		if err != nil {
			return fmt.Errorf("inserting document into state: %w", err)
		}
	}
	return nil
}

func GetKonnectState(rawKong *utils.KongRawState,
	rawKonnect *utils.KonnectRawState,
) (*KongState, error) {
	kongState, err := NewKongState()
	if err != nil {
		return nil, fmt.Errorf("creating new in-memory state of Kong: %w", err)
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
