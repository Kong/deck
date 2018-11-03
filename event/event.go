package event

import "github.com/hbagdi/doko/crud"

type Node struct {
	Op     crud.Op
	Kind   crud.Kind
	Obj    interface{}
	OldObj interface{}
}
