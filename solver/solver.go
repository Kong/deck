package solver

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	cruds "github.com/hbagdi/deck/solver/kong"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps int
	UpdateOps int
	DeleteOps int
}

// Solve generates a diff and walks the graph.
func Solve(doneCh chan struct{}, syncer *diff.Syncer,
	client *kong.Client, parallelism int, dry bool) (Stats, []error) {
	var r *crud.Registry
	var err error

	r, err = buildRegistry(client)
	if err != nil {
		return Stats{}, append([]error{},
			errors.Wrapf(err, "cannot build registry"))
	}

	var stats Stats
	recordOp := func(op crud.Op) {
		switch op {
		case crud.Create:
			stats.CreateOps = stats.CreateOps + 1
		case crud.Update:
			stats.UpdateOps = stats.UpdateOps + 1
		case crud.Delete:
			stats.DeleteOps = stats.DeleteOps + 1
		}
	}

	errs := syncer.Run(doneCh, parallelism, func(e diff.Event) (crud.Arg, error) {
		var err error
		var result crud.Arg

		c := e.Obj.(state.ConsoleString)
		switch e.Op {
		case crud.Create:
			print.CreatePrintln("creating", e.Kind, c.Console())
		case crud.Update:
			diffString, err := getDiff(e.OldObj, e.Obj)
			if err != nil {
				return nil, err
			}
			print.UpdatePrintln("updating", e.Kind, c.Console(), diffString)
		case crud.Delete:
			print.DeletePrintln("deleting", e.Kind, c.Console())
		default:
			panic("unknown operation " + e.Op.String())
		}

		if !dry {
			// sync mode
			// fire the request to Kong
			result, err = r.Do(e.Kind, e.Op, e)
			if err != nil {
				return nil, err
			}
		} else {
			// diff mode
			// return the new obj as is
			result = e.Obj
		}
		// record operation in both: diff and sync commands
		recordOp(e.Op)

		return result, nil
	})
	return stats, errs
}

func buildRegistry(client *kong.Client) (*crud.Registry, error) {
	var r crud.Registry
	var err error
	service, err := cruds.NewServiceCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a service CRUD")
	}
	err = r.Register("service", service)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'service' crud")
	}
	route, err := cruds.NewRouteCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a route CRUD")
	}
	err = r.Register("route", route)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'route' crud")
	}
	upstream, err := cruds.NewUpstreamCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a upstream CRUD")
	}
	err = r.Register("upstream", upstream)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'upstream' crud")
	}
	target, err := cruds.NewTargetCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a target CRUD")
	}
	err = r.Register("target", target)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'target' crud")
	}
	certificate, err := cruds.NewCertificateCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a certificate CRUD")
	}
	err = r.Register("certificate", certificate)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'certificate' crud")
	}
	caCertificate, err := cruds.NewCACertificateCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a caCertificate CRUD")
	}
	err = r.Register("ca_certificate", caCertificate)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'ca_certificate' crud")
	}
	plugin, err := cruds.NewPluginCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a plugin CRUD")
	}
	err = r.Register("plugin", plugin)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'plugin' crud")
	}
	consumer, err := cruds.NewConsumerCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a 'consumer' CRUD")
	}
	err = r.Register("consumer", consumer)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'consumer' crud")
	}
	keyAuth, err := cruds.NewKeyAuthCRUD(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating a 'key-auth' CRUD")
	}
	err = r.Register("key-auth", keyAuth)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'key-auth' crud")
	}
	hmacAuth, err := cruds.NewHMACAuthCRUD(client)
	if err != nil {
		return nil, errors.Wrapf(err, "creating 'hmac-auth' crud")
	}
	err = r.Register("hmac-auth", hmacAuth)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'hmac-auth' crud")
	}
	jwtAuth, err := cruds.NewJWTAuthCRUD(client)
	if err != nil {
		return nil, errors.Wrapf(err, "creating 'jwt-auth' crud")
	}
	err = r.Register("jwt-auth", jwtAuth)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'jwt-auth' crud")
	}
	basicAuth, err := cruds.NewBasicAuthCRUD(client)
	if err != nil {
		return nil, errors.Wrapf(err, "creating 'basic-auth' crud")
	}
	err = r.Register("basic-auth", basicAuth)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'basic-auth' crud")
	}
	aclGroups, err := cruds.NewACLGroupCRUD(client)
	if err != nil {
		return nil, errors.Wrapf(err, "creating 'acl' crud")
	}
	err = r.Register("acl-group", aclGroups)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'acl-group' crud")
	}
	oauth2Cred, err := cruds.NewOauth2CredCRUD(client)
	if err != nil {
		return nil, errors.Wrapf(err, "creating 'oauth2-cred' crud")
	}
	err = r.Register("oauth2-cred", oauth2Cred)
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'oauth2-cred' crud")
	}
	return &r, nil
}
