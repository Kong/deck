package dry

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
)

// CertificateCRUD implements Actions interface
// from the github.com/kong/crud package for the Certificate entitiy of Kong.
type CertificateCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func certificateFromStuct(a diff.Event) *state.Certificate {
	certificate, ok := a.Obj.(*state.Certificate)
	if !ok {
		panic("unexpected type, expected *state.certificate")
	}

	return certificate
}

// Create creates a fake certificate.
// The arg should be of type diff.Event, containing the certificate to be created,
// else the function will panic.
// It returns a the created *state.Certificate.
func (s *CertificateCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := certificateFromStuct(event)

	print.CreatePrintln("creating certificate", *certificate.Cert)
	certificate.ID = kong.String(utils.UUID())
	return certificate, nil
}

// Delete deletes a fake certificate.
// The arg should be of type diff.Event, containing the certificate to be deleted,
// else the function will panic.
// It returns a the deleted *state.Certificate.
func (s *CertificateCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := certificateFromStuct(event)

	print.DeletePrintln("deleting certificate", *certificate.Cert)
	return certificate, nil
}

// Update updates a fake certificate.
// The arg should be of type diff.Event, containing the certificate to be updated,
// else the function will panic.
// It returns a the updated *state.Certificate.
func (s *CertificateCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	certificate := certificateFromStuct(event)
	oldCertificateObj, ok := event.OldObj.(*state.Certificate)
	if !ok {
		panic("unexpected type, expected *state.certificate")
	}
	oldCertificate := oldCertificateObj.DeepCopy()
	// TODO remove this hack
	oldCertificate.CreatedAt = nil
	diff := getDiff(oldCertificate, &certificate.Certificate)
	print.UpdatePrintln("updating certificate", *certificate.Cert)
	print.UpdatePrintf("%s", diff)
	return certificate, nil
}
