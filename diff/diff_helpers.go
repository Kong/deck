package diff

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Kong/gojsondiff"
	"github.com/Kong/gojsondiff/formatter"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
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

type EnvVar struct {
	Key   string
	Value string
}

func parseDeckEnvVars() []EnvVar {
	const envVarPrefix = "DECK_"
	var parsedEnvVars []EnvVar

	for _, envVarStr := range os.Environ() {
		envPair := strings.SplitN(envVarStr, "=", 2)
		if strings.HasPrefix(envPair[0], envVarPrefix) {
			envVar := EnvVar{}
			envVar.Key = envPair[0]
			envVar.Value = envPair[1]
			parsedEnvVars = append(parsedEnvVars, envVar)
		}
	}

	sort.Slice(parsedEnvVars, func(i, j int) bool {
		return len(parsedEnvVars[i].Value) > len(parsedEnvVars[j].Value)
	})
	return parsedEnvVars
}

func MaskEnvVarValue(diffString string) string {
	for _, envVar := range parseDeckEnvVars() {
		diffString = strings.Replace(diffString, envVar.Value, "[masked]", -1)
	}
	return diffString
}
