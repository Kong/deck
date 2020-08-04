package solver

import (
	"encoding/json"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var (
	differ = gojsondiff.New()
)

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

func eventFromArg(arg crud.Arg) diff.Event {
	event, ok := arg.(diff.Event)
	if !ok {
		panic("unexpected type, expected diff.Event")
	}
	return event
}
