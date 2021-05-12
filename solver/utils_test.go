package solver

import (
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

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
						Content: kong.String("{\"foo\":\"foo\",\"bar\":\"bar\"}"),
					},
				},
				docB: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String("{\"foo\":\"foo\",\"bar\":\"bar\",\"baz\":\"baz\"}"),
					},
				},
			},
			want: " {\n   \"path\": \"foo\"\n }\n--- old\n+++ new\n@@ -1,4 +1,5 @@\n {\n \t\"bar\": \"bar\",\n+\t\"baz\": \"baz\",\n \t\"foo\": \"foo\"\n }\n\\ No newline at end of file\n",
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
						Content: kong.String("foo\n"),
					},
				},
				docB: &state.Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
						Content: kong.String("foo\nbar\n"),
					},
				},
			},
			want: " {\n   \"path\": \"foo\"\n }\n--- old\n+++ new\n@@ -1 +1,2 @@\n foo\n+bar\n",
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
