package dry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kong/deck/crud"
	arg "github.com/kong/deck/diff"
	diff "gopkg.in/d4l3k/messagediff.v1"
)

// TODO abstract this out
func eventFromArg(a crud.Arg) arg.Event {
	argStruct, ok := a.(arg.Event)
	if !ok {
		panic("unexpected type, expected Event")
	}
	return argStruct
}

// TODO add a diff of from to, like Port changed from 80 to 443
func getDiff(a, b interface{}) string {
	d, _ := diff.DeepDiff(a, b)
	var dstr []string
	for path, added := range d.Added {
		dstr = append(dstr, fmt.Sprintf("  added: %s = %#v\n", path.String(), added))
	}
	for path, removed := range d.Removed {
		dstr = append(dstr, fmt.Sprintf("  removed: %s = %#v\n", path.String(), removed))
	}
	for path, modified := range d.Modified {
		dstr = append(dstr, fmt.Sprintf("  modified: %s = %#v\n", path.String(), modified))
	}
	sort.Strings(dstr)
	return strings.Join(dstr, "")
}
