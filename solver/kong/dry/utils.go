package dry

import (
	"encoding/json"

	"github.com/hbagdi/deck/crud"
	arg "github.com/hbagdi/deck/diff"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var differ *diff.Differ

func init() {
	differ = diff.New()
}

// TODO abstract this out
func eventFromArg(a crud.Arg) arg.Event {
	argStruct, ok := a.(arg.Event)
	if !ok {
		panic("unexpected type, expected Event")
	}
	return argStruct
}

// TODO add a diff of from to, like Port changed from 80 to 443
func getDiff(a, b interface{}) (string, error) {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	d, err := differ.Compare(aJSON, bJSON)
	if err != nil {
		return "", err
	}
	var leftObject map[string]interface{}
	err = json.Unmarshal(aJSON, &leftObject)
	if err != nil {
		return "", err
	}

	formatter := formatter.NewAsciiFormatter(leftObject,
		formatter.AsciiFormatterConfig{})
	diffString, err := formatter.Format(d)
	return diffString, err
}
