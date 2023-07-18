package file

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

func Test_configFilesInDir(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "empty directory",
			args:    args{"testdata/emptydir"},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "directory does not exist",
			args:    args{"testdata/does-not-exist"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid directory",
			args: args{"testdata/emptyfiles"},
			want: []string{
				"testdata/emptyfiles/Baz.YamL",
				"testdata/emptyfiles/bar.yaml",
				"testdata/emptyfiles/foo.yml",
				"testdata/emptyfiles/foobar.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ConfigFilesInDir(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("configFilesInDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("configFilesInDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getReaders(t *testing.T) {
	type args struct {
		fileOrDir string
	}
	tests := []struct {
		name string
		args args
		want map[string]io.Reader
		// length of returned array
		wantLen int
		wantErr bool
	}{
		{
			name: "read from standard input",
			args: args{"-"},
			want: map[string]io.Reader{
				"STDIN": os.Stdin,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "directory does not exist",
			args:    args{"testdata/does-not-exist"},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "valid directory",
			args:    args{"testdata/emptyfiles"},
			want:    nil,
			wantLen: 4,
			wantErr: false,
		},
		{
			name:    "valid file",
			args:    args{"testdata/file.yaml"},
			want:    nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "valid JSON file",
			args:    args{"testdata/file.json"},
			want:    nil,
			wantLen: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getReaders(tt.args.fileOrDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getReaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantLen != len(got) {
				t.Errorf("getReaders() mismatch in returned length: "+
					"want = %v, got = %v", tt.wantLen, len(got))
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getReaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func sortSlices(x, y interface{}) bool {
	var xName, yName string
	switch xEntity := x.(type) {
	case FService:
		yEntity := y.(FService)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case FRoute:
		yEntity := y.(FRoute)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case FConsumer:
		yEntity := y.(FConsumer)
		xName = *xEntity.Username
		yName = *yEntity.Username
	case FPlugin:
		yEntity := y.(FPlugin)
		xName = *xEntity.Name
		yName = *yEntity.Name
	}
	return xName < yName
}

func Test_getContent(t *testing.T) {
	type args struct {
		filenames []string
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    *Content
		wantErr bool
	}{
		{
			name:    "directory does not exist",
			args:    args{[]string{"testdata/does-not-exist"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty directory",
			args:    args{[]string{"testdata/emptydir"}},
			want:    &Content{},
			wantErr: false,
		},
		{
			name:    "directory with empty files",
			args:    args{[]string{"testdata/emptyfiles"}},
			want:    &Content{},
			wantErr: false,
		},
		{
			name:    "bad yaml",
			args:    args{[]string{"testdata/badyaml"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "bad JSON",
			args:    args{[]string{"testdata/badjson"}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "single file",
			args: args{[]string{"testdata/file.yaml"}},
			envVars: map[string]string{
				"DECK_SVC2_HOST": "2.example.com",
				"DECK_FILE_LOG_FUNCTION": `
function parse_traceid(str)str = string.sub(str,1,8)
  local uint = 0
  for i = 1, #str do
    uint = uint + str:byte(i) * 0x100^(i-1)
  end
  return string.format("%.0f", uint)
end

kong.log.set_serialize_value("trace_id", parse_traceid(ngx.ctx.KONG_SPANS[1].trace_id))
kong.log.set_serialize_value("span_id", parse_traceid(ngx.ctx.KONG_SPANS[1].span_id))`,
			},
			want: &Content{
				Services: []FService{
					{
						Service: kong.Service{
							Name: kong.String("svc2"),
							Host: kong.String("2.example.com"),
							Tags: kong.StringSlice("<"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r2"),
									Paths: kong.StringSlice("/r2"),
								},
							},
						},
					},
				},
				Plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("prometheus"),
						},
					},
					{
						Plugin: kong.Plugin{
							Name: kong.String("pre-function"),
							Config: kong.Configuration{
								"log": `
function parse_traceid(str)str = string.sub(str,1,8)
  local uint = 0
  for i = 1, #str do
    uint = uint + str:byte(i) * 0x100^(i-1)
  end
  return string.format("%.0f", uint)
end

kong.log.set_serialize_value("trace_id", parse_traceid(ngx.ctx.KONG_SPANS[1].trace_id))
kong.log.set_serialize_value("span_id", parse_traceid(ngx.ctx.KONG_SPANS[1].span_id))
`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "environment variable present in file but not set",
			args:    args{[]string{"testdata/file.yaml"}},
			wantErr: true,
		},
		{
			name:    "file with bad environment variable",
			args:    args{[]string{"testdata/bad-env-var/file.yaml"}},
			wantErr: true,
		},
		{
			name:    "invalid file due to leading space",
			args:    args{[]string{"testdata/badyamlwithspace/bar.yml"}},
			wantErr: true,
		},
		{
			name: "multiple files",
			args: args{[]string{"testdata/file.yaml", "testdata/file.json"}},
			envVars: map[string]string{
				"DECK_SVC2_HOST":         "2.example.com",
				"DECK_FILE_LOG_FUNCTION": "kong.log.set_serialize_value('trace_id', 1))",
			},
			want: &Content{
				Services: []FService{
					{
						Service: kong.Service{
							Name: kong.String("svc2"),
							Host: kong.String("2.example.com"),
							Tags: kong.StringSlice("<"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r2"),
									Paths: kong.StringSlice("/r2"),
								},
							},
						},
					},
				},
				Plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("prometheus"),
						},
					},
					{
						Plugin: kong.Plugin{
							Name: kong.String("pre-function"),
							Config: kong.Configuration{
								"log": "kong.log.set_serialize_value('trace_id', 1))\n",
							},
						},
					},
				},
				Consumers: []FConsumer{
					{
						Consumer: kong.Consumer{
							Username: kong.String("foo"),
						},
					},
					{
						Consumer: kong.Consumer{
							Username: kong.String("bar"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid directory",
			args: args{[]string{"testdata/valid"}},
			want: &Content{
				Info: &Info{
					SelectorTags: []string{"tag1"},
				},
				Services: []FService{
					{
						Service: kong.Service{
							Name: kong.String("svc2"),
							Host: kong.String("2.example.com"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r2"),
									Paths: kong.StringSlice("/r2"),
								},
							},
						},
					},
					{
						Service: kong.Service{
							Name: kong.String("svc1"),
							Host: kong.String("1.example.com"),
							Tags: kong.StringSlice("team-svc1"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r1"),
									Paths: kong.StringSlice("/r1"),
								},
							},
						},
					},
				},
				Consumers: []FConsumer{
					{
						Consumer: kong.Consumer{
							Username: kong.String("foo"),
						},
					},
					{
						Consumer: kong.Consumer{
							Username: kong.String("bar"),
						},
					},
					{
						Consumer: kong.Consumer{
							Username: kong.String("harry"),
						},
					},
				},
				Plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("prometheus"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "different workspaces",
			args:    args{[]string{"testdata/differentworkspace"}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "same workspaces",
			args: args{[]string{"testdata/sameworkspace"}},
			want: &Content{
				FormatVersion: *kong.String("1.1"),
				Workspace:     *kong.String("bar"),
				Services: []FService{
					{
						Service: kong.Service{
							Name: kong.String("svc2"),
							Host: kong.String("2.example.com"),
							Tags: kong.StringSlice("team-svc2"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r2"),
									Paths: kong.StringSlice("/r2"),
								},
							},
						},
					},
					{
						Service: kong.Service{
							Name: kong.String("svc1"),
							Host: kong.String("1.example.com"),
							Tags: kong.StringSlice("team-svc1"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r1"),
									Paths: kong.StringSlice("/r1"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "defaults",
			args: args{[]string{"testdata/defaults"}},
			want: &Content{
				FormatVersion: "1.1",
				Upstreams: []FUpstream{
					{
						Upstream: kong.Upstream{
							Name:      kong.String("upstream1"),
							Algorithm: kong.String("round-robin"),
						},
						Targets: []*FTarget{
							{
								Target: kong.Target{
									Target: kong.String("198.51.100.11:80"),
									Weight: kong.Int(0),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "shared workspace",
			args: args{[]string{"testdata/sharedworkspace"}},
			want: &Content{
				FormatVersion: *kong.String("1.1"),
				Workspace:     *kong.String("bar"),
				Services: []FService{
					{
						Service: kong.Service{
							Name: kong.String("svc1"),
							Host: kong.String("1.example.com"),
							Tags: kong.StringSlice("team-svc1"),
						},
						Routes: []*FRoute{
							{
								Route: kong.Route{
									Name:  kong.String("r1"),
									Paths: kong.StringSlice("/r1"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "file with env var and parse bool",
			args: args{[]string{"testdata/parsebool/file.yaml"}},
			envVars: map[string]string{
				"DECK_MOCKBIN_ENABLED": "true",
			},
			want: &Content{
				Services: []FService{
					{
						Service: kong.Service{
							Name:    kong.String("svc1"),
							Host:    kong.String("mockbin.org"),
							Enabled: kong.Bool(true),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "file with env var and parse bool - err on bad value",
			args: args{[]string{"testdata/parsebool/file.yaml"}},
			envVars: map[string]string{
				"DECK_MOCKBIN_ENABLED": "RIP",
			},
			wantErr: true,
		},
		{
			name: "file with env var and parse Int",
			args: args{[]string{"testdata/parseint/file.yaml"}},
			envVars: map[string]string{
				"DECK_WRITE_TIMEOUT": "1337",
			},
			want: &Content{
				Services: []FService{
					{
						Service: kong.Service{
							Name:         kong.String("svc1"),
							Host:         kong.String("mockbin.org"),
							WriteTimeout: kong.Int(1337),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "file with env var and parse Int - err on bad value",
			args: args{[]string{"testdata/parseint/file.yaml"}},
			envVars: map[string]string{
				"DECK_WRITE_TIMEOUT": "RIP",
			},
			wantErr: true,
		},
		{
			name: "file with env var and parse Float64",
			args: args{[]string{"testdata/parsefloat/file.yaml"}},
			envVars: map[string]string{
				"DECK_FOO_FLOAT": "1337",
			},
			want: &Content{
				Plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("foofloat"),
							Config: kong.Configuration{
								"foo": float64(1337),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "file with env var and parse Int - err on bad value",
			args: args{[]string{"testdata/parsefloat/file.yaml"}},
			envVars: map[string]string{
				"DECK_FOO_FLOAT": "RIP",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			got, err := getContent(tt.args.filenames)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opt := []cmp.Option{
				cmpopts.SortSlices(sortSlices),
				cmpopts.SortSlices(func(a, b *string) bool { return *a < *b }),
				cmpopts.EquateEmpty(),
			}
			if diff := cmp.Diff(got, tt.want, opt...); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func Test_yamlUnmarshal(t *testing.T) {
	stringToInterfaceMap := map[string]interface{}{}
	bytes1 := `
versions:
  v1:
    enabled: false
`
	mapOfMap := map[string]interface{}{}
	err := yamlUnmarshal([]byte(bytes1), &mapOfMap)
	if err != nil {
		t.Errorf("yamlUnmarshal() error = %v (should be nil)", err)
	}
	subMap := mapOfMap["versions"]
	if reflect.TypeOf(subMap) != reflect.TypeOf(stringToInterfaceMap) {
		t.Errorf("yamlUnmarshal() expected type: %T, got: %T", stringToInterfaceMap, subMap)
	}

	bytes2 := `
versions:
- enabled: false
  version: 1
`
	mapOfArrayOfMap := map[string]interface{}{}
	err = yamlUnmarshal([]byte(bytes2), &mapOfArrayOfMap)
	if err != nil {
		t.Errorf("yamlUnmarshal() error = %v (should be nil)", err)
	}
	array := mapOfArrayOfMap["versions"].([]interface{})
	element := array[0]
	if reflect.TypeOf(element) != reflect.TypeOf(stringToInterfaceMap) {
		t.Errorf("yamlUnmarshal() expected type: %T, got: %T", stringToInterfaceMap, element)
	}
}
