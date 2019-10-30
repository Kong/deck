package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
)

// CACertificateCRUD implements Actions interface
// from the github.com/kong/crud package for the Certificate entitiy of Kong.
type CACertificateCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func caCertcFromStuct(a diff.Event) *state.CACertificate {
	certificate, ok := a.Obj.(*state.CACertificate)
	if !ok {
		panic("unexpected type, expected *state.CACertificate")
	}

	return certificate
}

// Create creates a fake certificate.
// The arg should be of type diff.Event, containing the certificate to be created,
// else the function will panic.
// It returns a the created *state.CACertificate.
func (s *CACertificateCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertcFromStuct(event)

	print.CreatePrintln("creating ca_certificate", certificate.Identifier())
	return certificate, nil
}

// Delete deletes a fake CACertificate.
// The arg should be of type diff.Event, containing the CACertificate to be deleted,
// else the function will panic.
// It returns a the deleted *state.CACertificate.
func (s *CACertificateCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertcFromStuct(event)

	print.DeletePrintln("deleting ca_certificate", certificate.Identifier())
	return certificate, nil
}

// Update updates a fake certificate.
// The arg should be of type diff.Event, containing the certificate to be updated,
// else the function will panic.
// It returns a the updated *state.CACertificate.
func (s *CACertificateCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := caCertcFromStuct(event)
	oldCertificateObj, ok := event.OldObj.(*state.CACertificate)
	if !ok {
		panic("unexpected type, expected *state.CACertificate")
	}
	oldCertificate := oldCertificateObj.DeepCopy()
	// TODO remove this hack
	oldCertificate.CreatedAt = nil
	diff, err := getDiff(oldCertificate, &certificate.CACertificate)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating ca_certificate %s\n%s", certificate.Identifier(), diff)
	return certificate, nil
}
