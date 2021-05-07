package solver

import (
	"encoding/json"
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var (
	differ = gojsondiff.New()
)

func getDocumentDiff(a, b *state.Document) (string, error) {
	aCopy := a.ShallowCopy()
	bCopy := a.ShallowCopy()
	aContent := *a.Content
	bContent := *b.Content
	aCopy.Content = nil
	bCopy.Content = nil
	objDiff, err := getDiff(aCopy, bCopy)
	if err != nil {
		return "", err
	}
	var contentDiff string
	if json.Valid([]byte(aContent)) {
		var aObj, bObj interface{}
		err = json.Unmarshal([]byte(aContent), &aObj)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal([]byte(bContent), &bObj)
		if err != nil {
			return "", err
		}
		aBytes, err := json.MarshalIndent(aObj, "", "\t")
		if err != nil {
			return "", err
		}
		bBytes, err := json.MarshalIndent(bObj, "", "\t")
		if err != nil {
			return "", err
		}
		aContent = string(aBytes)
		bContent = string(bBytes)
	}
	edits := myers.ComputeEdits(span.URIFromPath("old"), aContent, bContent)
	contentDiff = fmt.Sprint(gotextdiff.ToUnified("old", "new", aContent, edits))

	return objDiff + contentDiff, nil
}

func getDiff(a, b interface{}) (string, error) {
	utils.ZeroOutTimestamps(a)
	utils.ZeroOutTimestamps(b)
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
