package solver

import (
	"context"
	"fmt"
	"sync"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/types"
	"github.com/kong/go-kong/kong"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps *AtomicInt32Counter
	UpdateOps *AtomicInt32Counter
	DeleteOps *AtomicInt32Counter
}

type AtomicInt32Counter struct {
	counter int32
	lock    sync.RWMutex
}

func (a *AtomicInt32Counter) Increment(delta int32) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.counter += delta
}

func (a *AtomicInt32Counter) Count() int32 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.counter
}

// Solve generates a diff and walks the graph.
func Solve(ctx context.Context, syncer *diff.Syncer,
	client *kong.Client, konnectClient *konnect.Client,
	parallelism int, dry bool) (Stats, []error) {

	r := buildRegistry(client, konnectClient)

	stats := Stats{
		CreateOps: &AtomicInt32Counter{},
		UpdateOps: &AtomicInt32Counter{},
		DeleteOps: &AtomicInt32Counter{},
	}
	recordOp := func(op crud.Op) {
		switch op {
		case crud.Create:
			stats.CreateOps.Increment(1)
		case crud.Update:
			stats.UpdateOps.Increment(1)
		case crud.Delete:
			stats.DeleteOps.Increment(1)
		}
	}

	errs := syncer.Run(ctx, parallelism, func(e diff.Event) (crud.Arg, error) {
		var err error
		var result crud.Arg

		c := e.Obj.(state.ConsoleString)
		switch e.Op {
		case crud.Create:
			cprint.CreatePrintln("creating", e.Kind, c.Console())
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
			cprint.UpdatePrintln("updating", e.Kind, c.Console(), diffString)
		case crud.Delete:
			cprint.DeletePrintln("deleting", e.Kind, c.Console())
		default:
			panic("unknown operation " + e.Op.String())
		}

		if !dry {
			// sync mode
			// fire the request to Kong
			result, err = r.Do(ctx, e.Kind, e.Op, e)
			if err != nil {
				return nil, fmt.Errorf("%v %v %v failed: %w", e.Op, e.Kind, c.Console(), err)
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
	opts := types.EntityOpts{
		KongClient:    client,
		KonnectClient: konnectClient,
	}
	service, err := types.NewEntity(types.Service, opts)
	if err != nil {
		panic(err)
	}
	var r crud.Registry
	r.MustRegister("service", service.CRUDActions())
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
