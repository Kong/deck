package diff

import (
	"encoding/json"
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var differ = gojsondiff.New()

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
	if json.Valid([]byte(aContent)) && json.Valid([]byte(bContent)) {
		aContent, err = prettyPrintJSONString(aContent)
		if err != nil {
			return "", err
		}
		bContent, err = prettyPrintJSONString(bContent)
		if err != nil {
			return "", err
		}
	}
	edits := myers.ComputeEdits(span.URIFromPath("old"), aContent, bContent)
	contentDiff = fmt.Sprint(gotextdiff.ToUnified("old", "new", aContent, edits))

	return objDiff + contentDiff, nil
}

func prettyPrintJSONString(JSONString string) (string, error) {
	jBlob := []byte(JSONString)
	var obj interface{}
	err := json.Unmarshal(jBlob, &obj)
	if err != nil {
		return "", err
	}
	bytes, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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
