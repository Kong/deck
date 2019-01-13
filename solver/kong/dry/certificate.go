package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
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
	diff, err := getDiff(oldCertificate, &certificate.Certificate)
	if err != nil {
		return nil, err
	}
	print.UpdatePrintf("updating certificate %s\n%s", *certificate.Cert, diff)
	return certificate, nil
}
