package diff

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/pkg/errors"
)

var (
	errEnqueueFailed = errors.New("failed to queue event")
)

// TODO get rid of the syncer struct and simply have a func for it

// Syncer takes in a current and target state of Kong,
// diffs them, generating a Graph to get Kong from current
// to target state.
type Syncer struct {
	currentState *state.KongState
	targetState  *state.KongState
	postProcess  crud.Registry

	eventChan chan Event
	errChan   chan error
	stopChan  chan struct{}

	inFlightOps int32

	SilenceWarnings bool
	StageDelaySec   int

	once sync.Once
}

// NewSyncer constructs a Syncer.
func NewSyncer(current, target *state.KongState) (*Syncer, error) {
	s := &Syncer{}
	s.currentState, s.targetState = current, target

	s.postProcess.MustRegister("service", &servicePostAction{current})
	s.postProcess.MustRegister("route", &routePostAction{current})
	s.postProcess.MustRegister("upstream", &upstreamPostAction{current})
	s.postProcess.MustRegister("target", &targetPostAction{current})
	s.postProcess.MustRegister("certificate", &certificatePostAction{current})
	s.postProcess.MustRegister("sni", &sniPostAction{current})
	s.postProcess.MustRegister("ca_certificate", &caCertificatePostAction{current})
	s.postProcess.MustRegister("plugin", &pluginPostAction{current})
	s.postProcess.MustRegister("consumer", &consumerPostAction{current})
	s.postProcess.MustRegister("key-auth", &keyAuthPostAction{current})
	s.postProcess.MustRegister("hmac-auth", &hmacAuthPostAction{current})
	s.postProcess.MustRegister("jwt-auth", &jwtAuthPostAction{current})
	s.postProcess.MustRegister("basic-auth", &basicAuthPostAction{current})
	s.postProcess.MustRegister("acl-group", &aclGroupPostAction{current})
	s.postProcess.MustRegister("oauth2-cred", &oauth2CredPostAction{current})
	return s, nil
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
	err = sc.deleteTargets()
	if err != nil {
		return err
	}
	err = sc.deleteSNIs()
	if err != nil {
		return err
	}
	err = sc.deleteCACertificates()
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

	// finish createUpdate before returning
	sc.wait()

	return nil
}

func (sc *Syncer) queueEvent(e Event) error {
	atomic.AddInt32(&sc.inFlightOps, 1)
	select {
	case sc.eventChan <- e:
		return nil
	case <-sc.stopChan:
		return errEnqueueFailed
	}
}

func (sc *Syncer) eventCompleted() {
	atomic.AddInt32(&sc.inFlightOps, -1)
}

func (sc *Syncer) wait() {
	time.Sleep(time.Duration(sc.StageDelaySec) * time.Second)
	for atomic.LoadInt32(&sc.inFlightOps) != 0 {
		// TODO hack?
		time.Sleep(1 * time.Millisecond)
	}
}

// Run starts a diff and invokes d for every diff.
func (sc *Syncer) Run(done <-chan struct{}, parallelism int, d Do) []error {
	if parallelism < 1 {
		return append([]error{}, errors.New("parallelism can not be negative"))
	}

	var wg sync.WaitGroup
	const eventBuffer = 10

	sc.eventChan = make(chan Event, eventBuffer)
	sc.stopChan = make(chan struct{})
	sc.errChan = make(chan error)

	// run rabbit run
	// start the consumers
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			err := sc.eventLoop(d)
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
	case <-done:
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
// TODO remove crud.Arg
type Do func(a Event) (crud.Arg, error)

func (sc *Syncer) eventLoop(d Do) error {
	for event := range sc.eventChan {
		err := sc.handleEvent(d, event)
		sc.eventCompleted()
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) handleEvent(d Do, event Event) error {
	res, err := d(event)
	if err != nil {
		return errors.Wrapf(err, "while processing event")
	}
	if res == nil {
		return errors.New("result of event is nil")
	}
	_, err = sc.postProcess.Do(event.Kind, event.Op, res)
	if err != nil {
		return errors.Wrap(err, "while post processing event")
	}
	return nil
}
