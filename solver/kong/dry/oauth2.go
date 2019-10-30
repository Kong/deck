package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// Oauth2CredCRUD implements Actions interface
// from the github.com/kong/crud package for the oauth2 cred of Kong.
type Oauth2CredCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func oauth2CredFromStruct(arg diff.Event) *state.Oauth2Credential {
	oauth2Cred, ok := arg.Obj.(*state.Oauth2Credential)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return oauth2Cred
}

// Create creates a fake oauth2 cred.
// The arg should be of type diff.Event, containing the oauth2Cred to be created,
// else the function will panic.
// It returns a the created *state.Oauth2Cred.
func (s *Oauth2CredCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)
	print.CreatePrintln("creating oauth2 cred", *oauth2Cred.Name,
		"for consumer", *oauth2Cred.Consumer.ID)
	return oauth2Cred, nil
}

// Delete deletes a fake Route.
// The arg should be of type diff.Event, containing the oauth2Cred to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *Oauth2CredCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)
	print.DeletePrintln("deleting oauth2 cred", *oauth2Cred.Name,
		"for consumer", *oauth2Cred.Consumer.ID)
	return oauth2Cred, nil
}

// Update updates a fake Route.
// The arg should be of type diff.Event, containing the oauth2Cred to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *Oauth2CredCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	oauth2Cred := oauth2CredFromStruct(event)
	oldRoute, ok := event.OldObj.(*state.Oauth2Credential)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}
	// TODO remove this hack
	oldRoute.CreatedAt = nil

	diffString, err := getDiff(oldRoute.Oauth2Credential, oauth2Cred.Oauth2Credential)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating oauth2 cred %s\n%s", *oauth2Cred.Name,
		diffString)
	return oauth2Cred, nil
}
