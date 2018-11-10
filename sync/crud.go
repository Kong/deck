package sync

import (
	"fmt"

	"github.com/hbagdi/doko/crud"
)

type Callback func(crud.Arg, error) error

type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func (s *ServiceCRUD) Create(arg crud.Arg) (crud.Arg, error) {
	fmt.Println("Create called with ", arg)
	return nil, nil
}
func (s *ServiceCRUD) Delete(arg crud.Arg) (crud.Arg, error) {
	fmt.Println("Delete called with ", arg)
	return nil, nil
}
func (s *ServiceCRUD) Update(arg crud.Arg) (crud.Arg, error) {
	fmt.Println("Update called with ", arg)
	return nil, nil
}
