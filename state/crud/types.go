package crud

import "github.com/kong/deck/crud"

// Callback represnts a Callback function.
type Callback func(crud.Arg, error) error
