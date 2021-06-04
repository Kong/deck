package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
)

// documentCRUD implements crud.Actions interface.
type documentCRUD struct {
	client *konnect.Client
}

func documentFromStruct(arg crud.Event) *state.Document {
	d, ok := arg.Obj.(*state.Document)
	if !ok {
		panic("unexpected type, expected *state.Document")
	}
	return d
}

func oldDocumentFromStruct(arg crud.Event) *state.Document {
	d, ok := arg.OldObj.(*state.Document)
	if !ok {
		panic("unexpected type, expected *state.Document")
	}
	return d
}

// Create creates a document in Konnect.
// The arg should be of type crud.Event, containing the document to be created,
// else the function will panic.
// It returns the created *state.Document.
func (s *documentCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	d := documentFromStruct(event)
	createdDoc, err := s.client.Documents.Create(ctx, &d.Document)
	if err != nil {
		return nil, err
	}
	return &state.Document{Document: *createdDoc}, nil
}

// Delete deletes a document in Konnect.
// The arg should be of type crud.Event, containing the document to be deleted,
// else the function will panic.
// It returns a the deleted *state.Document.
func (s *documentCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	d := documentFromStruct(event)
	err := s.client.Documents.Delete(ctx, &d.Document)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Update updates a document in Konnect.
// The arg should be of type crud.Event, containing the document to be updated,
// else the function will panic.
// It returns a the updated *state.Document.
func (s *documentCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	var (
		err        error
		updatedDoc *konnect.Document
	)
	event := crud.EventFromArg(arg[0])
	document := documentFromStruct(event)
	oldDocument := oldDocumentFromStruct(event)

	// if there is a change in document entity, make a PATCH
	if !document.EqualWithOpts(oldDocument, false, true, true) {
		documentCopy := &state.Document{Document: *document.ShallowCopy()}
		updatedDoc, err = s.client.Documents.Update(ctx, &documentCopy.Document)
		if err != nil {
			return nil, err
		}
	} else {
		updatedDoc = &document.Document
	}

	return &state.Document{Document: *updatedDoc}, nil
}

type documentDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *documentDiffer) Deletes(handler func(crud.Event) error) error {
	currentDocuments, err := d.currentState.Documents.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching documents from state: %w", err)
	}

	for _, doc := range currentDocuments {
		n, err := d.deleteDocument(doc)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (d *documentDiffer) deleteDocument(doc *state.Document) (*crud.Event, error) {
	_, err := d.targetState.Documents.GetByParent(doc.Parent, *doc.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "document",
			Obj:  doc,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up document %q: %w",
			doc.Identifier(), err)
	}
	return nil, nil
}

func (d *documentDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetDocuments, err := d.targetState.Documents.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching documents from state: %w", err)
	}

	for _, doc := range targetDocuments {
		n, err := d.createUpdateDocument(doc)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *documentDiffer) createUpdateDocument(doc *state.Document) (*crud.Event, error) {
	dCopy := &state.Document{Document: *doc.ShallowCopy()}
	currentDoc, err := d.currentState.Documents.GetByParent(doc.Parent, *doc.ID)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "document",
			Obj:  dCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up document %q: %w",
			doc.Identifier(), err)
	}

	// found, check if update needed
	// Service Package-attached Documents fail equality checks if ignoreForeign
	// is disabled. This appears to be related to an invalid diff detection for
	// Service Versions attached to the package.
	if !currentDoc.EqualWithOpts(dCopy, false, true, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "document",
			Obj:    dCopy,
			OldObj: currentDoc,
		}, nil
	}
	return nil, nil
}
