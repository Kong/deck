package diff

import (
	"reflect"
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

func Test_PrettyPrintJSONString(t *testing.T) {
	type args struct {
		jstring string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic JSON string",
			args: args{
				jstring: `{"foo":"foo","bar":{"a": 1, "b": 2}}`,
			},
			want: `{
	"bar": {
		"a": 1,
		"b": 2
	},
	"foo": "foo"
}`,
			wantErr: false,
		},
		{
			name: "invalid JSON string",
			args: args{
				jstring: "a large swarm of bees",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prettyPrintJSONString(tt.args.jstring)
			if (err != nil) != tt.wantErr {
				t.Errorf("prettyPrintJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prettyPrintJSONString() = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func Test_GetDocumentDiff(t *testing.T) {
	type args struct {
		docA *state.Document
		docB *state.Document
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "JSON",
			args: args{
				docA: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String(`{"foo":"foo","bar":"bar"}`),
					},
				},
				docB: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String(`{"foo":"foo","bar":"bar","baz":"baz"}`),
					},
				},
			},
			want: ` {
   "path": "foo"
 }
--- old
+++ new
@@ -1,4 +1,5 @@
 {
 	"bar": "bar",
+	"baz": "baz",
 	"foo": "foo"
 }
\ No newline at end of file
`,
		},
		{
			name: "not JSON",
			args: args{
				docA: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String(`foo
`),
					},
				},
				docB: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String(`foo
bar
`),
					},
				},
			},
			want: ` {
   "path": "foo"
 }
--- old
+++ new
@@ -1 +1,2 @@
 foo
+bar
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := getDocumentDiff(tt.args.docA, tt.args.docB); got != tt.want {
				t.Errorf("getDocumentDiff() = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func Test_MaskEnvVarsValues(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		envVars map[string]string
	}{
		{
			name: "JSON",
			envVars: map[string]string{
				"DECK_BAR": "barbar",
				"DECK_BAZ": "bazbaz",
			},
			args: `{"foo":"foo","bar":"barbar","baz":"bazbaz"}`,
			want: `{"foo":"foo","bar":"[masked]","baz":"[masked]"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			if got := maskEnvVarValue(tt.args); got != tt.want {
				t.Errorf("maskEnvVarValue() = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func Test_diffObjects(t *testing.T) {
	// Test with two identical objects.
	obj1 := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	obj2 := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	want := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}
	// Test with two different objects.
	obj1 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar1",
		"baz": "baz1",
	}
	obj2 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar2",
		"baz": "baz2",
	}
	want = map[string]interface{}{
		"bar": map[string]interface{}{
			"old": "bar1",
			"new": "bar2",
		},
		"baz": map[string]interface{}{
			"old": "baz1",
			"new": "baz2",
		},
		"foo": "foo",
	}
	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}

	// Test with a missing property in obj2.
	obj1 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	obj2 = map[string]interface{}{
		"foo": "foo",
		"baz": "baz",
	}
	want = map[string]interface{}{
		"bar": "bar",
		"foo": "foo",
		"baz": "baz",
	}
	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}

	// Test with a missing property in obj1.
	obj1 = map[string]interface{}{
		"foo": "foo",
		"baz": "baz",
	}
	obj2 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	want = map[string]interface{}{
		"bar": "bar",
		"baz": "baz",
		"foo": "foo",
	}
	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}

	// Test with a nil value in obj1.
	obj1 = map[string]interface{}{
		"foo": nil,
		"bar": "bar",
	}
	obj2 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
	}
	want = map[string]interface{}{
		"foo": map[string]interface{}{
			"old": nil,
			"new": "foo",
		},
		"bar": "bar",
	}

	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}

	// Test with a nil value in obj2.
	obj1 = map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
	}
	obj2 = map[string]interface{}{
		"foo": "foo",
		"bar": nil,
	}
	want = map[string]interface{}{
		"foo": "foo",
		"bar": map[string]interface{}{
			"old": "bar",
			"new": nil,
		},
	}

	if got := diffObjects(obj1, obj2); !reflect.DeepEqual(got, want) {
		t.Errorf("diffObjects() = %v\nwant %v", got, want)
	}
}
