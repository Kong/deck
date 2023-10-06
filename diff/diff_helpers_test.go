package diff

import (
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
			if got := MaskEnvVarValue(tt.args); got != tt.want {
				t.Errorf("maskEnvVarValue() = %v\nwant %v", got, tt.want)
			}
		})
	}
}
