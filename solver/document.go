package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
)

// documentCRUD implements crud.Actions interface.
type documentCRUD struct {
	client *konnect.Client
}

func documentFromStruct(arg diff.Event) *state.Document {
	d, ok := arg.Obj.(*state.Document)
	if !ok {
		panic("unexpected type, expected *state.Document")
	}
	return d
}

func oldDocumentFromStruct(arg diff.Event) *state.Document {
	d, ok := arg.OldObj.(*state.Document)
	if !ok {
		panic("unexpected type, expected *state.Document")
	}
	return d
}

// Create creates a document in Konnect.
// The arg should be of type diff.Event, containing the document to be created,
// else the function will panic.
// It returns the created *state.Document.
func (s *documentCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	d := documentFromStruct(event)
	createdDoc, err := s.client.Documents.Create(nil, &d.Document)
	if err != nil {
		return nil, err
	}
	return &state.Document{Document: *createdDoc}, nil
}

// Delete deletes a document in Konnect.
// The arg should be of type diff.Event, containing the document to be deleted,
// else the function will panic.
// It returns a the deleted *state.Document.
func (s *documentCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	d := documentFromStruct(event)
	err := s.client.Documents.Delete(nil, &d.Document)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Update updates a document in Konnect.
// The arg should be of type diff.Event, containing the document to be updated,
// else the function will panic.
// It returns a the updated *state.Document.
func (s *documentCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	var (
		err        error
		updatedDoc *konnect.Document
	)
	event := eventFromArg(arg[0])
	document := documentFromStruct(event)
	oldDocument := oldDocumentFromStruct(event)

	// if there is a change in document entity, make a PATCH
	if !document.EqualWithOpts(oldDocument, false, true, true) {
		documentCopy := &state.Document{Document: *document.ShallowCopy()}
		updatedDoc, err = s.client.Documents.Update(nil, &documentCopy.Document)
		if err != nil {
			return nil, err
		}
	} else {
		updatedDoc = &document.Document
	}

	return &state.Document{Document: *updatedDoc}, nil
}
