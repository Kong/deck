package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteSNIs() error {
	currentSNIs, err := sc.currentState.SNIs.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching snis from state: %w", err)
	}

	for _, sni := range currentSNIs {
		n, err := sc.deleteSNI(sni)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) deleteSNI(sni *state.SNI) (*Event, error) {
	_, err := sc.targetState.SNIs.Get(*sni.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "sni",
			Obj:  sni,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up sni %q: %w", *sni.Name, err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateSNIs() error {
	sniSNIs, err := sc.targetState.SNIs.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching snis from state: %w", err)
	}

	for _, sni := range sniSNIs {
		n, err := sc.createUpdateSNI(sni)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateSNI(sni *state.SNI) (*Event, error) {
	sni = &state.SNI{SNI: *sni.DeepCopy()}
	currentSNI, err := sc.currentState.SNIs.Get(*sni.ID)
	if err == state.ErrNotFound {
		// sni not present, create it

		return &Event{
			Op:   crud.Create,
			Kind: "sni",
			Obj:  sni,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up sni %q: %w", *sni.Name, err)
	}
	// found, check if update needed

	if !currentSNI.EqualWithOpts(sni, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "sni",
			Obj:    sni,
			OldObj: currentSNI,
		}, nil
	}
	return nil, nil
}
