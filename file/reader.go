package file

import (
	"fmt"
	"strconv"

	"github.com/hbagdi/deck/counter"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

var count counter.Counter

// GetStateFromFile reads in a file with filename and constructs
// a state. If filename is `-`, then it will read from os.Stdin.
// If filename represents a directory, it will traverse the tree
// rooted at filename, read all the files with .yaml and .yml extensions
// and generate a state after a merge of the content from all the files.
//
// It will return an error if the file representation is invalid
// or if there is any error during processing.
// All entities without an ID will get a `placeholder-{iota}` ID
// assigned to them.
// TODO return a struct than 4 args
func GetStateFromFile(filename string) (*state.KongState,
	[]string, string, error) {
	if filename == "" {
		return nil, nil, "", errors.New("filename cannot be empty")
	}

	fileContent, err := getContent(filename)
	if err != nil {
		return nil, nil, "", err
	}
	return GetStateFromContent(fileContent)
}

// GetStateFromContent takes the serialized state and returns a Kong.
// It will return an error if the file representation is invalid
// or if there is any error during processing.
// All entities without an ID will get a `placeholder-{iota}` ID
// assigned to them.
func GetStateFromContent(fileContent *Content) (*state.KongState,
	[]string, string, error) {
	count.Reset()
	d, err := utils.GetKongDefaulter()
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "creating defaulter")
	}
	selectTags := fileContent.Info.SelectorTags
	workspace := fileContent.Workspace
	kongState, err := state.NewKongState()
	if err != nil {
		return nil, nil, "", err
	}
	for _, s := range fileContent.Services {
		if utils.Empty(s.ID) {
			s.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(s.Service.Name) {
			return nil, nil, "", errors.New("all services in the file must be named")
		}
		_, err := kongState.Services.Get(*s.Service.Name)
		if err != state.ErrNotFound {
			return nil, nil, "", errors.Errorf("duplicate service definitions"+
				" found for: '%s'", *s.Service.Name)
		}
		if err = utils.MergeTags(&s.Service, selectTags); err != nil {
			return nil, nil, "", errors.Wrap(err,
				"merging selector tag with object")
		}
		err = d.Set(&s.Service)
		if err != nil {
			return nil, nil, "", errors.Wrap(err, "filling in defaults for service")
		}
		err = kongState.Services.Add(state.Service{Service: s.Service})
		if err != nil {
			return nil, nil, "", err
		}
		for _, p := range s.Plugins {
			if ok, err := processPlugin(p, selectTags); !ok {
				return nil, nil, "", err
			}
			p.Service = s.Service.DeepCopy()
			if err = utils.MergeTags(&p.Plugin, selectTags); err != nil {
				return nil, nil, "", errors.Wrap(err,
					"merging selector tag with object")
			}
			err = kongState.Plugins.Add(state.Plugin{Plugin: p.Plugin})
			if err != nil {
				return nil, nil, "", err
			}
		}

		for _, r := range s.Routes {
			if utils.Empty(r.ID) {
				r.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(r.Name) {
				return nil, nil, "", errors.New("all routes in the file must be named")
			}
			_, err := kongState.Routes.Get(*r.Name)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate route definitions"+
					" found for: '%s'", *r.Name)
			}
			r.Service = s.Service.DeepCopy()
			if err = utils.MergeTags(&r.Route, selectTags); err != nil {
				return nil, nil, "", errors.Wrap(err,
					"merging selector tag with object")
			}
			err = d.Set(&r.Route)
			if err != nil {
				return nil, nil, "", errors.Wrap(err, "filling in defaults for route")
			}
			err = kongState.Routes.Add(state.Route{Route: r.Route})
			if err != nil {
				return nil, nil, "", err
			}
			for _, p := range r.Plugins {
				if ok, err := processPlugin(p, selectTags); !ok {
					return nil, nil, "", err
				}
				p.Route = r.Route.DeepCopy()
				err = kongState.Plugins.Add(state.Plugin{Plugin: p.Plugin})
				if err != nil {
					return nil, nil, "", err
				}
			}
		}
	}

	for _, p := range fileContent.Plugins {
		if ok, err := processPlugin(&p, selectTags); !ok {
			return nil, nil, "", err
		}
		err = kongState.Plugins.Add(state.Plugin{Plugin: p.Plugin})
		if err != nil {
			return nil, nil, "", err
		}
	}

	for _, u := range fileContent.Upstreams {
		if utils.Empty(u.ID) {
			u.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(u.Name) {
			return nil, nil, "", errors.New("all upstreams in the file must be named")
		}
		_, err := kongState.Upstreams.Get(*u.Name)
		if err != state.ErrNotFound {
			return nil, nil, "", errors.Errorf("duplicate upstream definitions"+
				" found for: '%s'", *u.Name)
		}
		if err = utils.MergeTags(&u.Upstream, selectTags); err != nil {
			return nil, nil, "", errors.Wrap(err,
				"merging selector tag with object")
		}
		err = d.Set(&u.Upstream)
		if err != nil {
			return nil, nil, "", errors.Wrap(err, "filling in defaults for upstream")
		}
		err = kongState.Upstreams.Add(state.Upstream{Upstream: u.Upstream})
		if err != nil {
			return nil, nil, "", err
		}

		for _, t := range u.Targets {
			if utils.Empty(t.ID) {
				t.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			_, err := kongState.Targets.GetByUpstreamNameAndTarget(*u.Name, *t.Target.Target)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate target definitions"+
					" found for: '%s'", *t.Target.Target)
			}
			t.Upstream = u.Upstream.DeepCopy()
			err = d.Set(&t.Target)
			if err != nil {
				return nil, nil, "", errors.Wrap(err, "setting defaults in target")
			}
			if err = utils.MergeTags(&t.Target, selectTags); err != nil {
				return nil, nil, "", errors.Wrap(err,
					"merging selector tag with object")
			}
			if err != nil {
				return nil, nil, "", errors.Wrap(err, "filling in defaults for target")
			}
			err = kongState.Targets.Add(state.Target{Target: t.Target})
			if err != nil {
				return nil, nil, "", err
			}
		}
	}

	allSNIs := make(map[string]bool)
	for _, c := range fileContent.Certificates {
		if utils.Empty(c.ID) {
			c.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(c.Cert) || utils.Empty(c.Key) {
			return nil, nil, "", errors.Errorf("all certificates must have a cert" +
				" and a key")
		}
		// check if an SNI is present in multiple certificates
		for _, s := range c.SNIs {
			if allSNIs[*s] {
				return nil, nil, "", errors.Errorf("duplicate sni found: '%s'", *s)
			}
			allSNIs[*s] = true
		}

		_, err := kongState.Certificates.GetByCertKey(*c.Cert, *c.Key)
		if err != state.ErrNotFound {
			return nil, nil, "", errors.Errorf("duplicate certificate definitions"+
				" found for the following certificate:\n'%s'", *c.Cert)
		}
		if err = utils.MergeTags(&c.Certificate, selectTags); err != nil {
			return nil, nil, "", errors.Wrap(err,
				"merging selector tag with object")
		}
		err = kongState.Certificates.Add(state.Certificate{
			Certificate: c.Certificate,
		})
		if err != nil {
			return nil, nil, "", err
		}
	}
	for _, c := range fileContent.CACertificates {
		if utils.Empty(c.ID) {
			c.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(c.Cert) {
			return nil, nil, "",
				errors.Errorf("all ca_certificates must have a cert")
		}

		_, err := kongState.CACertificates.Get(*c.Cert)
		if err != state.ErrNotFound {
			return nil, nil, "",
				errors.Errorf("duplicate ca_certificate definitions"+
					" found for the following ca_certificate:\n'%s'", *c.Cert)
		}
		if err = utils.MergeTags(&c.CACertificate, selectTags); err != nil {
			return nil, nil, "",
				errors.Wrap(err, "merging selector tag with object")
		}
		err = kongState.CACertificates.Add(state.CACertificate{
			CACertificate: c.CACertificate,
		})
		if err != nil {
			return nil, nil, "", err
		}
	}

	for _, c := range fileContent.Consumers {
		if utils.Empty(c.ID) {
			c.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(c.Consumer.Username) {
			return nil, nil, "",
				errors.New("all consumers in the file must have a username")
		}
		_, err := kongState.Consumers.Get(*c.Consumer.Username)
		if err != state.ErrNotFound {
			return nil, nil, "", errors.Errorf("duplicate consumer definitions"+
				" found for: '%v'", *c.Consumer.Username)
		}
		if err = utils.MergeTags(&c.Consumer, selectTags); err != nil {
			return nil, nil, "", errors.Wrap(err,
				"merging selector tag with object")
		}
		err = kongState.Consumers.Add(state.Consumer{Consumer: c.Consumer})
		if err != nil {
			return nil, nil, "", err
		}
		for _, p := range c.Plugins {
			if ok, err := processPlugin(p, selectTags); !ok {
				return nil, nil, "", err
			}
			p.Consumer = c.Consumer.DeepCopy()
			err = kongState.Plugins.Add(state.Plugin{Plugin: p.Plugin})
			if err != nil {
				return nil, nil, "", err
			}
		}
		for _, cred := range c.KeyAuths {
			if utils.Empty(cred.ID) {
				cred.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(cred.Key) {
				return nil, nil, "", errors.New("key field is missing" +
					"for keyauth_credential")
			}
			_, err := kongState.KeyAuths.Get(*cred.Key)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate key-auth definitions"+
					" found for apikey: '%s'", *cred.Key)
			}
			cred.Consumer = c.Consumer.DeepCopy()
			err = kongState.KeyAuths.Add(state.KeyAuth{KeyAuth: *cred})
			if err != nil {
				return nil, nil, "", err
			}
		}
		for _, cred := range c.HMACAuths {
			if utils.Empty(cred.ID) {
				cred.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(cred.Username) {
				return nil, nil, "", errors.New("username field is missing" +
					"for keyauth_credential")
			}
			_, err := kongState.HMACAuths.Get(*cred.Username)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate hmac-auth "+
					"definitions found for username: '%s'", *cred.Username)
			}
			cred.Consumer = c.Consumer.DeepCopy()
			err = kongState.HMACAuths.Add(state.HMACAuth{HMACAuth: *cred})
			if err != nil {
				return nil, nil, "", err
			}
		}
		for _, cred := range c.JWTAuths {
			if utils.Empty(cred.ID) {
				cred.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(cred.Key) {
				return nil, nil, "", errors.New("key field is missing" +
					"for jwt_secret")
			}
			_, err := kongState.JWTAuths.Get(*cred.Key)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate jwt definitions"+
					" found for key: '%s'", *cred.Key)
			}
			cred.Consumer = c.Consumer.DeepCopy()
			err = kongState.JWTAuths.Add(state.JWTAuth{JWTAuth: *cred})
			if err != nil {
				return nil, nil, "", err
			}
		}
		for _, cred := range c.BasicAuths {
			if utils.Empty(cred.ID) {
				cred.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(cred.Username) {
				return nil, nil, "", errors.New("username field is missing" +
					"for keyauth_credential")
			}
			_, err := kongState.BasicAuths.Get(*cred.Username)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate basic-auth definitions"+
					" found for username: '%s'", *cred.Username)
			}
			cred.Consumer = c.Consumer.DeepCopy()
			err = kongState.BasicAuths.Add(state.BasicAuth{BasicAuth: *cred})
			if err != nil {
				return nil, nil, "", err
			}
		}
		for _, cred := range c.ACLGroups {
			if utils.Empty(cred.ID) {
				cred.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(cred.Group) {
				return nil, nil, "", errors.New("group field is missing" +
					"for acl")
			}
			_, err := kongState.ACLGroups.Get(*c.Username, *cred.Group)
			if err != state.ErrNotFound {
				return nil, nil, "", errors.Errorf("duplicate acl definitions"+
					" found for username: '%s'", *c.Username)
			}
			cred.Consumer = c.Consumer.DeepCopy()
			err = kongState.ACLGroups.Add(state.ACLGroup{ACLGroup: *cred})
			if err != nil {
				return nil, nil, "", err
			}
		}
	}

	return kongState, fileContent.Info.SelectorTags, workspace, nil
}

func processPlugin(p *Plugin, tags []string) (bool, error) {
	if utils.Empty(p.ID) {
		p.ID = kong.String("placeholder-" +
			strconv.FormatUint(count.Inc(), 10))
	}
	if utils.Empty(p.Name) {
		return false, errors.New("plugin does not have a name")
	}
	if p.Route != nil || p.Service != nil || p.Consumer != nil {
		return false, errors.New("plugin " + *p.Name +
			" has foreign relations " +
			"defined. Plugins in config file " +
			"cannot define foreign relations (yet).")
	}
	if p.Config == nil {
		p.Config = make(map[string]interface{})
	}
	p.Config = ensureJSON(p.Config)
	if err := utils.MergeTags(&p.Plugin, tags); err != nil {
		return false, errors.Wrap(err,
			"merging selector tag with object")
	}
	// TODO error out on consumer/route not nil
	return true, nil
}

func ensureJSON(m map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = yamlToJSON(v2)
		case []interface{}:
			var array []interface{}
			for _, element := range v2 {
				switch el := element.(type) {
				case map[interface{}]interface{}:
					array = append(array, yamlToJSON(el))
				default:
					array = append(array, el)
				}
			}
			if array != nil {
				res[fmt.Sprint(k)] = array
			} else {
				res[fmt.Sprint(k)] = v
			}
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}

func yamlToJSON(m map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = yamlToJSON(v2)
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}
