package sanitize

import (
	"encoding/json"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/tidwall/gjson"
)

func Test_findRelevantFieldNamesWithKey(t *testing.T) {
	jsonSchema := `{
		"fields": [
			{"foo": {"type": "string", "one_of": ["a", "b"]}},
			{"bar": {"type": "set", "elements": {"type": "string", "one_of": ["x", "y"]}}},
			{"baz": {"type": "array", "elements": {"type": "string", "enum": ["z"]}}},
			{"qux": {"type": "string"}}
		]
    }`
	var schema kong.Schema
	_ = json.Unmarshal([]byte(jsonSchema), &schema)
	jsonb, _ := json.Marshal(&schema)
	gjsonSchema := gjson.ParseBytes(jsonb)
	fields := gjsonSchema.Get("fields")

	exempted := make(map[string]bool)
	findRelevantFieldNamesWithKey(exempted, fields, "one_of", "elements")
	findRelevantFieldNamesWithKey(exempted, fields, "enum", "elements")

	if !exempted["foo"] {
		t.Errorf("expected 'foo' to be exempted")
	}
	if !exempted["bar"] {
		t.Errorf("expected 'bar' to be exempted")
	}
	if !exempted["baz"] {
		t.Errorf("expected 'baz' to be exempted")
	}
	if exempted["qux"] {
		t.Errorf("did not expect 'qux' to be exempted")
	}
}
