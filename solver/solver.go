package solver

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
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

	r = buildRegistry(client)

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

func buildRegistry(client *kong.Client) *crud.Registry {
	var r crud.Registry
	r.MustRegister("service", &serviceCRUD{client: client})
	r.MustRegister("route", &routeCRUD{client: client})
	r.MustRegister("upstream", &upstreamCRUD{client: client})
	r.MustRegister("target", &targetCRUD{client: client})
	r.MustRegister("certificate", &certificateCRUD{client: client})
	r.MustRegister("ca_certificate", &caCertificateCRUD{client: client})
	r.MustRegister("plugin", &pluginCRUD{client: client})
	r.MustRegister("consumer", &consumerCRUD{client: client})
	r.MustRegister("key-auth", &keyAuthCRUD{client: client})
	r.MustRegister("hmac-auth", &hmacAuthCRUD{client: client})
	r.MustRegister("jwt-auth", &jwtAuthCRUD{client: client})
	r.MustRegister("basic-auth", &basicAuthCRUD{client: client})
	r.MustRegister("acl-group", &aclGroupCRUD{client: client})
	r.MustRegister("oauth2-cred", &oauth2CredCRUD{client: client})
	return &r
}
