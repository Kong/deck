package crud

import "github.com/kong/deck/crud"

type Callback func(crud.Arg, error) error
