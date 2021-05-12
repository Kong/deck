package solver

import (
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

func Test_PprintJSONString(t *testing.T) {
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
				jstring: "{\"foo\":\"foo\",\"bar\":{\"a\": 1, \"b\": 2}}",
			},
			want:    "{\n\t\"bar\": {\n\t\t\"a\": 1,\n\t\t\"b\": 2\n\t},\n\t\"foo\": \"foo\"\n}",
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
			got, err := pprintJSONString(tt.args.jstring)
			if (err != nil) != tt.wantErr {
				t.Errorf("pprintJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("pprintJSONString() = %v\nwant %v", got, tt.want)
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
