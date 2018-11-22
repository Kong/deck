package diff

import "github.com/hbagdi/deck/crud"

// Node represents an imperative operation
// that gets Kong closer to the target state.
type Node struct {
	Op     crud.Op
	Kind   crud.Kind
	Obj    interface{}
	OldObj interface{}
}

// func (n *Node) String() string {
// 	return n.Op.String() + " " + string(n.Kind)
// }
