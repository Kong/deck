package diff

import (
	"sync/atomic"
	"time"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

// Syncer takes in a current and target state of Kong,
// diffs them, generating a Graph to get Kong from current
// to target state.
type Syncer struct {
	currentState *state.KongState
	targetState  *state.KongState
	eventChan    chan Node
	finChan      chan bool

	postProcess crud.Registry

	InFlightOps int32
}

// NewSyncer constructs a Syncer.
func NewSyncer(current, target *state.KongState) (*Syncer, error) {
	s := &Syncer{}
	s.eventChan = make(chan Node)
	s.currentState, s.targetState = current, target
	err := s.postProcess.Register("service", &servicePostAction{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'service' crud")
	}
	err = s.postProcess.Register("route", &routePostAction{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'route' crud")
	}
	return s, nil
}

// Run diffs the current and target states and returns two graphs.
// The first graph contains all the entities which should be deleted from Kong
// and the second graph contains the entities which should be created or
// updated to get Kong to target state.
func (sc *Syncer) Run() error {
	var err error
	err = sc.createUpdate()
	if err != nil {
		return err
	}
	err = sc.delete()
	if err != nil {
		return err
	}

	close(sc.eventChan)
	return nil
}

func (sc *Syncer) delete() error {
	// routes should be deleted before services
	err := sc.deleteRoutes()
	if err != nil {
		return err
	}
	sc.wait()
	err = sc.deleteServices()
	if err != nil {
		return err
	}
	sc.wait()
	return nil
}

func (sc *Syncer) createUpdate() error {
	// TODO write an interface and register by types,
	// then execute in a particular order

	// services should be created before routes
	err := sc.createUpdateServices()
	if err != nil {
		return err
	}
	sc.wait()
	err = sc.createUpdateRoutes()
	if err != nil {
		return err
	}
	sc.wait()
	return nil
}

// EventChan returns the events
func (sc *Syncer) EventChan() chan Node {
	return sc.eventChan
}

func (sc *Syncer) sendEvent(n Node) {
	// fmt.Println("got event", n)
	atomic.AddInt32(&sc.InFlightOps, 1)
	sc.eventChan <- n
}

func (sc *Syncer) sendFin() {
	atomic.AddInt32(&sc.InFlightOps, -1)
}

func (sc *Syncer) wait() {
	for atomic.LoadInt32(&sc.InFlightOps) != 0 {
		// XXX TODO fix hack
		time.Sleep(10 * time.Millisecond)
	}
}

// Process walks a graph and executes actions.
func (sc *Syncer) Process(registry *crud.Registry, client *kong.Client) error {
	for {
		n, ok := <-sc.eventChan
		if !ok {
			return nil
		}
		// every Node will need to add a few things to arg:
		// *kong.Client to use
		// callbacks to execute
		res, err := registry.Do(n.Kind, n.Op, ArgStruct{
			Obj:    n.Obj,
			OldObj: n.OldObj,

			CurrentState: sc.currentState,
			TargetState:  sc.targetState,
			Client:       client,
		})
		if err != nil {
			return errors.Wrapf(err, "while processing event: %v", n)
		}
		if res == nil {
			return errors.New("result of event is nil")
		}
		_, err = sc.postProcess.Do(n.Kind, n.Op, sc.currentState, res)
		if err != nil {
			return errors.Wrap(err, "while post processing")
		}
		sc.sendFin()
	}
}
