package diff

import (
	"context"
	"fmt"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/types"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps *utils.AtomicInt32Counter
	UpdateOps *utils.AtomicInt32Counter
	DeleteOps *utils.AtomicInt32Counter
}

// Solve generates a diff and walks the graph.
func Solve(ctx context.Context, syncer *Syncer,
	client *kong.Client, konnectClient *konnect.Client,
	parallelism int, dry bool) (Stats, []error) {

	r := buildRegistry(client, konnectClient)

	stats := Stats{
		CreateOps: &utils.AtomicInt32Counter{},
		UpdateOps: &utils.AtomicInt32Counter{},
		DeleteOps: &utils.AtomicInt32Counter{},
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

	errs := syncer.Run(ctx, parallelism, func(e crud.Event) (crud.Arg, error) {
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
	var r crud.Registry

	for _, entityType := range types.AllTypes {
		entity, err := types.NewEntity(entityType, opts)
		if err != nil {
			panic(err)
		}
		r.MustRegister(crud.Kind(entityType), entity.CRUDActions())
	}
	return &r
}
