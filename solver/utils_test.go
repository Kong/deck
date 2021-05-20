package solver

import (
	"testing"

	"github.com/kong/deck/file"
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

func Test_IsDocument(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "document",
			args: args{
				obj: &state.Document{},
			},
			want: true,
		},
		{
			name: "not document",
			args: args{
				obj: &state.Service{},
			},
			want: false,
		},
		{
			name: "the wrong sort of document",
			args: args{
				obj: &file.FDocument{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDocument(tt.args.obj); got != tt.want {
				t.Errorf("isDocument() = %v\nwant %v", got, tt.want)
			}
		})
	}
}
