package diff

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/types"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

var errEnqueueFailed = errors.New("failed to queue event")

func defaultBackOff() backoff.BackOff {
	// For various reasons, Kong can temporarily fail to process
	// a valid request (e.g. when the database is under heavy load).
	// We retry each request up to 3 times on failure, after around
	// 1 second, 3 seconds, and 9 seconds (randomized exponential backoff).
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = 1 * time.Second
	exponentialBackoff.Multiplier = 3
	return backoff.WithMaxRetries(exponentialBackoff, 4)
}

// Syncer takes in a current and target state of Kong,
// diffs them, generating a Graph to get Kong from current
// to target state.
type Syncer struct {
	currentState *state.KongState
	targetState  *state.KongState

	processor     crud.Registry
	postProcessor crud.Registry

	eventChan chan crud.Event
	errChan   chan error
	stopChan  chan struct{}

	inFlightOps int32

	silenceWarnings bool
	stageDelaySec   int

	once sync.Once

	kongClient    *kong.Client
	konnectClient *konnect.Client
}

type SyncerOpts struct {
	CurrentState *state.KongState
	TargetState  *state.KongState

	KongClient    *kong.Client
	KonnectClient *konnect.Client

	SilenceWarnings bool
	StageDelaySec   int
}

// NewSyncer constructs a Syncer.
func NewSyncer(opts SyncerOpts) (*Syncer, error) {
	s := &Syncer{
		currentState: opts.CurrentState,
		targetState:  opts.TargetState,

		kongClient:    opts.KongClient,
		konnectClient: opts.KonnectClient,

		silenceWarnings: opts.SilenceWarnings,
		stageDelaySec:   opts.StageDelaySec,
	}

	err := s.init()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (sc *Syncer) init() error {
	opts := types.EntityOpts{
		CurrentState: sc.currentState,
		TargetState:  sc.targetState,

		KongClient:    sc.kongClient,
		KonnectClient: sc.konnectClient,
	}

	entities := []string{
		types.Service, types.Route, types.Plugin,

		types.Certificate, types.SNI, types.CACertificate,

		types.Upstream, types.Target,

		types.Consumer,
		types.ACLGroup, types.BasicAuth, types.KeyAuth,
		types.HMACAuth, types.JWTAuth, types.OAuth2Cred,
		types.MTLSAuth,

		types.RBACRole, types.RBACEndpointPermission,

		types.ServicePackage, types.ServiceVersion, types.Document,
	}
	for _, entityType := range entities {
		entity, err := types.NewEntity(entityType, opts)
		if err != nil {
			return err
		}
		sc.postProcessor.MustRegister(crud.Kind(entityType), entity.PostProcessActions())
		sc.processor.MustRegister(crud.Kind(entityType), entity.CRUDActions())
	}
	return nil
}

func (sc *Syncer) diff() error {
	var err error
	err = sc.createUpdate()
	if err != nil {
		return err
	}
	err = sc.delete()
	if err != nil {
		return err
	}
	return nil
}

func (sc *Syncer) delete() error {
	var err error

	err = sc.deletePlugins()
	if err != nil {
		return err
	}
	err = sc.deleteKeyAuths()
	if err != nil {
		return err
	}
	err = sc.deleteHMACAuths()
	if err != nil {
		return err
	}
	err = sc.deleteJWTAuths()
	if err != nil {
		return err
	}
	err = sc.deleteBasicAuths()
	if err != nil {
		return err
	}
	err = sc.deleteOauth2Creds()
	if err != nil {
		return err
	}
	err = sc.deleteACLGroups()
	if err != nil {
		return err
	}
	err = sc.deleteMTLSAuths()
	if err != nil {
		return err
	}
	err = sc.deleteTargets()
	if err != nil {
		return err
	}
	err = sc.deleteSNIs()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// plugins must be deleted before services, routes and consumers
	// routes must be deleted before service can be deleted
	// credentials must be deleted before consumers
	// targets must be deleted before upstream

	// PLEASE NOTE that if the order is not preserved, then decK will error
	// out because deleting a child entity whose parent is already deleted
	// will return a 404
	sc.wait()

	err = sc.deleteRoutes()
	if err != nil {
		return err
	}
	err = sc.deleteConsumers()
	if err != nil {
		return err
	}
	err = sc.deleteUpstreams()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// routes must be deleted before services
	sc.wait()

	err = sc.deleteServices()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// services must be deleted before certificates (client_certificate)
	sc.wait()

	err = sc.deleteCertificates()
	if err != nil {
		return err
	}

	// services must be deleted before ca_certificates
	err = sc.deleteCACertificates()
	if err != nil {
		return err
	}

	err = sc.deleteRBACEndpointPermissions()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// RBAC endpoint permissions must be deleted before RBAC roles
	sc.wait()

	err = sc.deleteRBACRoles()
	if err != nil {
		return err
	}

	err = sc.deleteDocuments()
	if err != nil {
		return err
	}

	err = sc.deleteServiceVersions()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// ServiceVersions must be deleted before ServicePackages
	sc.wait()

	err = sc.deleteServicePackages()
	if err != nil {
		return err
	}

	// finish delete before returning
	sc.wait()

	return nil
}

func (sc *Syncer) createUpdate() error {
	// TODO write an interface and register by types,
	// then execute in a particular order

	err := sc.createUpdateCertificates()
	if err != nil {
		return err
	}
	err = sc.createUpdateCACertificates()
	if err != nil {
		return err
	}
	err = sc.createUpdateConsumers()
	if err != nil {
		return err
	}
	err = sc.createUpdateUpstreams()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// upstreams must be created before targets
	// certificates must be created before SNIs
	// consumers must be created before creds of all kinds
	// certificates must be created before services (client_certificate)
	sc.wait()

	err = sc.createUpdateTargets()
	if err != nil {
		return err
	}
	err = sc.createUpdateSNIs()
	if err != nil {
		return err
	}
	err = sc.createUpdateServices()
	if err != nil {
		return err
	}
	err = sc.createUpdateKeyAuths()
	if err != nil {
		return err
	}
	err = sc.createUpdateHMACAuths()
	if err != nil {
		return err
	}
	err = sc.createUpdateJWTAuths()
	if err != nil {
		return err
	}
	err = sc.createUpdateBasicAuths()
	if err != nil {
		return err
	}
	err = sc.createUpdateOauth2Creds()
	if err != nil {
		return err
	}
	err = sc.createUpdateACLGroups()
	if err != nil {
		return err
	}
	err = sc.createUpdateMTLSAuths()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// services must be created before routes
	sc.wait()

	err = sc.createUpdateRoutes()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// services, routes and consumers must be created before plugins
	sc.wait()

	err = sc.createUpdatePlugins()
	if err != nil {
		return err
	}

	err = sc.createUpdateRBACRoles()
	if err != nil {
		return err
	}

	// barrier for foreign relations
	// RBAC roles must be created before endpoint permissions
	sc.wait()

	err = sc.createUpdateRBACEndpointPermissions()
	if err != nil {
		return err
	}

	err = sc.createUpdateServicePackages()
	if err != nil {
		return err
	}
	// barrier for foreign relations
	// services, routes and consumers must be created before plugins
	sc.wait()

	err = sc.createUpdateServiceVersions()
	if err != nil {
		return err
	}

	err = sc.createUpdateDocuments()
	if err != nil {
		return err
	}

	// finish createUpdate before returning
	sc.wait()

	return nil
}

func (sc *Syncer) queueEvent(e crud.Event) error {
	atomic.AddInt32(&sc.inFlightOps, 1)
	select {
	case sc.eventChan <- e:
		return nil
	case <-sc.stopChan:
		atomic.AddInt32(&sc.inFlightOps, -1)
		return errEnqueueFailed
	}
}

func (sc *Syncer) eventCompleted() {
	atomic.AddInt32(&sc.inFlightOps, -1)
}

func (sc *Syncer) wait() {
	time.Sleep(time.Duration(sc.stageDelaySec) * time.Second)
	for atomic.LoadInt32(&sc.inFlightOps) != 0 {
		select {
		case <-sc.stopChan:
			return
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// Run starts a diff and invokes d for every diff.
func (sc *Syncer) Run(ctx context.Context, parallelism int, d Do) []error {
	if parallelism < 1 {
		return append([]error{}, fmt.Errorf("parallelism can not be negative"))
	}

	var wg sync.WaitGroup
	const eventBuffer = 10

	sc.eventChan = make(chan crud.Event, eventBuffer)
	sc.stopChan = make(chan struct{})
	sc.errChan = make(chan error)

	// run rabbit run
	// start the consumers
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			err := sc.eventLoop(ctx, d)
			if err != nil {
				sc.errChan <- err
			}
			wg.Done()
		}()
	}

	// start the producer
	wg.Add(1)
	go func() {
		err := sc.diff()
		if err != nil {
			sc.errChan <- err
		}
		close(sc.eventChan)
		wg.Done()
	}()

	// close the error chan once all done
	go func() {
		wg.Wait()
		close(sc.errChan)
	}()

	var errs []error
	select {
	case <-ctx.Done():
	case err, ok := <-sc.errChan:
		if ok && err != nil {
			if err != errEnqueueFailed {
				errs = append(errs, err)
			}
		}
	}

	// stop the producer
	close(sc.stopChan)

	// collect errors
	for err := range sc.errChan {
		if err != errEnqueueFailed {
			errs = append(errs, err)
		}
	}

	return errs
}

// Do is the worker function to sync the diff
type Do func(a crud.Event) (crud.Arg, error)

func (sc *Syncer) eventLoop(ctx context.Context, d Do) error {
	for event := range sc.eventChan {
		// Stop if program is terminated
		select {
		case <-sc.stopChan:
			return nil
		default:
		}

		err := sc.handleEvent(ctx, d, event)
		sc.eventCompleted()
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) handleEvent(ctx context.Context, d Do, event crud.Event) error {
	err := backoff.Retry(func() error {
		res, err := d(event)
		if err != nil {
			err = fmt.Errorf("while processing event: %w", err)

			var kongAPIError *kong.APIError
			if errors.As(err, &kongAPIError) &&
				kongAPIError.Code() == http.StatusInternalServerError {
				// Only retry if the request to Kong returned a 500 status code
				return err
			}

			// Do not retry on other status codes
			return backoff.Permanent(err)
		}
		if res == nil {
			// Do not retry empty responses
			return backoff.Permanent(fmt.Errorf("result of event is nil"))
		}
		_, err = sc.postProcessor.Do(ctx, event.Kind, event.Op, res)
		if err != nil {
			// Do not retry program errors
			return backoff.Permanent(fmt.Errorf("while post processing event: %w", err))
		}
		return nil
	}, defaultBackOff())

	return err
}

// Stats holds the stats related to a Solve.
type Stats struct {
	CreateOps *utils.AtomicInt32Counter
	UpdateOps *utils.AtomicInt32Counter
	DeleteOps *utils.AtomicInt32Counter
}

// Solve generates a diff and walks the graph.
func (sc *Syncer) Solve(ctx context.Context, parallelism int, dry bool) (Stats, []error) {
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

	errs := sc.Run(ctx, parallelism, func(e crud.Event) (crud.Arg, error) {
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
			result, err = sc.processor.Do(ctx, e.Kind, e.Op, e)
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
