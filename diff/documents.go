package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteDocuments() error {
	currentDocuments, err := sc.currentState.Documents.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching documents from state: %w", err)
	}

	for _, d := range currentDocuments {
		n, err := sc.deleteDocument(d)
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

func (sc *Syncer) deleteDocument(d *state.Document) (*Event, error) {
	_, err := sc.targetState.Documents.GetByParent(d.Parent, *d.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "document",
			Obj:  d,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up document %q: %w",
			d.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateDocuments() error {
	targetDocuments, err := sc.targetState.Documents.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching documents from state: %w", err)
	}

	for _, d := range targetDocuments {
		n, err := sc.createUpdateDocument(d)
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

func (sc *Syncer) createUpdateDocument(d *state.Document) (*Event, error) {
	dCopy := &state.Document{Document: *d.ShallowCopy()}
	currentd, err := sc.currentState.Documents.GetByParent(d.Parent, *d.ID)

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "document",
			Obj:  dCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up document %q: %w",
			d.Identifier(), err)
	}

	// found, check if update needed
	// Service Package-attached Documents fail equality checks if ignoreForeign
	// is disabled. This appears to be related to an invalid diff detection for
	// Service Versions attached to the package.
	if !currentd.EqualWithOpts(dCopy, false, true, true) {
		return &Event{
			Op:     crud.Update,
			Kind:   "document",
			Obj:    dCopy,
			OldObj: currentd,
		}, nil
	}
	return nil, nil
}
