package solver

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps int
	UpdateOps int
	DeleteOps int
}

// Solve generates a diff and walks the graph.
func Solve(ctx context.Context, syncer *diff.Syncer,
	client *kong.Client, konnectClient *konnect.Client,
	parallelism int, dry bool) (Stats, []error) {

	r := buildRegistry(client, konnectClient)

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

	errs := syncer.Run(ctx, parallelism, func(e diff.Event) (crud.Arg, error) {
		var err error
		var result crud.Arg

		c := e.Obj.(state.ConsoleString)
		switch e.Op {
		case crud.Create:
			print.CreatePrintln("creating", e.Kind, c.Console())
		case crud.Update:
			var diffString string
			if oldObj, ok := e.OldObj.(*state.Document); ok {
				diffString, err = getDocumentDiff(oldObj, e.Obj.(*state.Document))
			} else {
				diffString, err = getDiff(e.OldObj, e.Obj)
			}
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
				return nil, errors.Wrapf(err, "%v %v %v failed", e.Op, e.Kind, c.Console())
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

func buildRegistry(client *kong.Client, konnectClient *konnect.Client) *crud.Registry {
	var r crud.Registry
	r.MustRegister("service", &serviceCRUD{client: client})
	r.MustRegister("route", &routeCRUD{client: client})
	r.MustRegister("upstream", &upstreamCRUD{client: client})
	r.MustRegister("target", &targetCRUD{client: client})
	r.MustRegister("certificate", &certificateCRUD{client: client})
	r.MustRegister("sni", &sniCRUD{client: client})
	r.MustRegister("ca_certificate", &caCertificateCRUD{client: client})
	r.MustRegister("plugin", &pluginCRUD{client: client})
	r.MustRegister("consumer", &consumerCRUD{client: client})
	r.MustRegister("key-auth", &keyAuthCRUD{client: client})
	r.MustRegister("hmac-auth", &hmacAuthCRUD{client: client})
	r.MustRegister("jwt-auth", &jwtAuthCRUD{client: client})
	r.MustRegister("basic-auth", &basicAuthCRUD{client: client})
	r.MustRegister("acl-group", &aclGroupCRUD{client: client})
	r.MustRegister("oauth2-cred", &oauth2CredCRUD{client: client})
	r.MustRegister("mtls-auth", &mtlsAuthCRUD{client: client})
	r.MustRegister("rbac-role", &rbacRoleCRUD{client: client})
	r.MustRegister("rbac-endpointpermission", &rbacEndpointPermissionCRUD{client: client})

	r.MustRegister("service-package", &servicePackageCRUD{client: konnectClient})
	r.MustRegister("service-version", &serviceVersionCRUD{client: konnectClient})
	r.MustRegister("document", &documentCRUD{client: konnectClient})
	return &r
}
