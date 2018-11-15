package diff

import "github.com/hbagdi/deck/crud"

type Node struct {
	Op     crud.Op
	Kind   crud.Kind
	Obj    interface{}
	OldObj interface{}
}
