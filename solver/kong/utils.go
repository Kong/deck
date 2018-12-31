package kong

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
)

func eventFromArg(arg crud.Arg) diff.Event {
	event, ok := arg.(diff.Event)
	if !ok {
		panic("unexpected type, expected diff.Event")
	}
	return event
}
