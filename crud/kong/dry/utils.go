package dry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kong/deck/crud"
	arg "github.com/kong/deck/crud/kong"
	diff "gopkg.in/d4l3k/messagediff.v1"
)

// TODO abstract this out
func argStructFromArg(a crud.Arg) arg.ArgStruct {
	argStruct, ok := a.(arg.ArgStruct)
	if !ok {
		panic("unexpected type, expected ArgStruct")
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
