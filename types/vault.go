package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// vaultCRUD implements crud.Actions interface.
type vaultCRUD struct {
	client *kong.Client
}

func vaultFromStruct(arg crud.Event) *state.Vault {
	vault, ok := arg.Obj.(*state.Vault)
	if !ok {
		panic("unexpected type, expected *state.Vault")
	}
	return vault
}

// Create creates a Vault in Kong.
// The arg should be of type crud.Event, containing the vault to be created,
// else the function will panic.
// It returns a the created *state.Vault.
func (s *vaultCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	vault := vaultFromStruct(event)
	createdVault, err := s.client.Vaults.Create(ctx, &vault.Vault)
	if err != nil {
		return nil, err
	}
	return &state.Vault{Vault: *createdVault}, nil
}

// Delete deletes a Vault in Kong.
// The arg should be of type crud.Event, containing the vault to be deleted,
// else the function will panic.
// It returns a the deleted *state.Vault.
func (s *vaultCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	vault := vaultFromStruct(event)
	err := s.client.Vaults.Delete(ctx, vault.ID)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// Update updates a Vault in Kong.
// The arg should be of type crud.Event, containing the vault to be updated,
// else the function will panic.
// It returns a the updated *state.Vault.
func (s *vaultCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	vault := vaultFromStruct(event)

	updatedVault, err := s.client.Vaults.Create(ctx, &vault.Vault)
	if err != nil {
		return nil, err
	}
	return &state.Vault{Vault: *updatedVault}, nil
}

type vaultDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

// Deletes generates a memdb CRUD DELETE event for Vaults
// which is then consumed by the differ and used to gate Kong client calls.
func (d *vaultDiffer) Deletes(handler func(crud.Event) error) error {
	currentVaults, err := d.currentState.Vaults.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching vaults from state: %w", err)
	}

	for _, vault := range currentVaults {
		n, err := d.deleteVault(vault)
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

func (d *vaultDiffer) deleteVault(vault *state.Vault) (*crud.Event, error) {
	_, err := d.targetState.Vaults.Get(*vault.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "vault",
			Obj:  vault,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up vault %q: %w",
			*vault.Prefix, err)
	}
	return nil, nil
}

// CreateAndUpdates generates a memdb CRUD CREATE/UPDATE event for Vaults
// which is then consumed by the differ and used to gate Kong client calls.
func (d *vaultDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetVaults, err := d.targetState.Vaults.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching vaults from state: %w", err)
	}

	for _, vault := range targetVaults {
		n, err := d.createUpdateVault(vault)
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

func (d *vaultDiffer) createUpdateVault(vault *state.Vault) (*crud.Event,
	error,
) {
	vaultCopy := &state.Vault{Vault: *vault.DeepCopy()}
	currentVault, err := d.currentState.Vaults.Get(*vault.Prefix)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "vault",
			Obj:  vaultCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up vault %v: %w",
			*vault.Prefix, err)
	}

	// found, check if update needed
	if !currentVault.EqualWithOpts(vaultCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "vault",
			Obj:    vaultCopy,
			OldObj: currentVault,
		}, nil
	}
	return nil, nil
}
