package convert

import (
	"testing"

	"github.com/kong/deck/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    Format
		wantErr bool
	}{
		{
			name: "parses valid values",
			args: args{
				key: "kong-gateway",
			},
			want:    FormatKongGateway,
			wantErr: false,
		},
		{
			name: "parses valid values",
			args: args{
				key: "kong-gateway-2.x",
			},
			want:    FormatKongGateway2x,
			wantErr: false,
		},
		{
			name: "parses valid values",
			args: args{
				key: "kong-gateway-3.x",
			},
			want:    FormatKongGateway3x,
			wantErr: false,
		},
		{
			name: "parses values in a case-insensitive manner",
			args: args{
				key: "koNNect",
			},
			want:    FormatKonnect,
			wantErr: false,
		},
		{
			name: "parse fails with invalid values",
			args: args{
				key: "k42",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Convert(t *testing.T) {
	type args struct {
		inputFilename          string
		outputFilename         string
		fromFormat             Format
		toFormat               Format
		runtimeGroupName       string
		expectedOutputFilename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors out due to invalid conversion",
			args: args{
				inputFilename: "testdata/2/input.yaml",
				fromFormat:    FormatKonnect,
				toFormat:      FormatKongGateway,
			},
			wantErr: true,
		},
		{
			name: "errors out due to invalid conversion",
			args: args{
				inputFilename: "testdata/3/input.yaml",
				fromFormat:    FormatKongGateway3x,
				toFormat:      FormatKongGateway2x,
			},
			wantErr: true,
		},
		{
			name: "errors out when input file doesn't exist",
			args: args{
				inputFilename: "testdata/1/input-does-not-exist.yaml",
				fromFormat:    FormatKongGateway,
				toFormat:      FormatKonnect,
			},
			wantErr: true,
		},
		{
			name: "converts from Kong Gateway to Konnect format (no workspace, no RG)",
			args: args{
				inputFilename:          "testdata/2/input.yaml",
				outputFilename:         "testdata/2/output.yaml",
				expectedOutputFilename: "testdata/2/output-expected.yaml",
				fromFormat:             FormatKongGateway,
				toFormat:               FormatKonnect,
			},
			wantErr: false,
		},
		{
			name: "converts from Kong Gateway 2.x to Kong Gateway 3.x format",
			args: args{
				inputFilename:          "testdata/3/input.yaml",
				outputFilename:         "testdata/3/output.yaml",
				expectedOutputFilename: "testdata/3/output-expected.yaml",
				fromFormat:             FormatKongGateway2x,
				toFormat:               FormatKongGateway3x,
			},
			wantErr: false,
		},
		{
			name: "converts from Kong Gateway 2.x to Kong Gateway 3.x format (no _format_version input)",
			args: args{
				inputFilename:          "testdata/4/input.yaml",
				outputFilename:         "testdata/4/output.yaml",
				expectedOutputFilename: "testdata/4/output-expected.yaml",
				fromFormat:             FormatKongGateway2x,
				toFormat:               FormatKongGateway3x,
			},
			wantErr: false,
		},
		{
			name: "converts from Kong Gateway to Konnect format (workspace, no RG)",
			args: args{
				inputFilename:          "testdata/5/input.yaml",
				outputFilename:         "testdata/5/output.yaml",
				expectedOutputFilename: "testdata/5/output-expected.yaml",
				fromFormat:             FormatKongGateway,
				toFormat:               FormatKonnect,
			},
			wantErr: false,
		},
		{
			name: "converts from Kong Gateway to Konnect format (no workspace, RG)",
			args: args{
				inputFilename:          "testdata/6/input.yaml",
				outputFilename:         "testdata/6/output.yaml",
				expectedOutputFilename: "testdata/6/output-expected.yaml",
				fromFormat:             FormatKongGateway,
				toFormat:               FormatKonnect,
				runtimeGroupName:       "foo",
			},
			wantErr: false,
		},
		{
			name: "converts from Kong Gateway to Konnect format (workspace + RG)",
			args: args{
				inputFilename:          "testdata/7/input.yaml",
				outputFilename:         "testdata/7/output.yaml",
				expectedOutputFilename: "testdata/7/output-expected.yaml",
				fromFormat:             FormatKongGateway,
				toFormat:               FormatKonnect,
				runtimeGroupName:       "foo",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otps := Opts{
				InputFilename:    tt.args.inputFilename,
				OutputFilename:   tt.args.outputFilename,
				FromFormat:       tt.args.fromFormat,
				ToFormat:         tt.args.toFormat,
				RuntimeGroupName: tt.args.runtimeGroupName,
			}
			err := Convert(otps)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				got, err := file.GetContentFromFiles([]string{tt.args.outputFilename})
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				want, err := file.GetContentFromFiles([]string{tt.args.expectedOutputFilename})
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				assert.Equal(t, want, got)
			}
		})
	}
}

func Test_convertKongGatewayToKonnect(t *testing.T) {
	type args struct {
		input *file.Content
	}
	tests := []struct {
		name             string
		runtimeGroupName string
		args             args
		want             *file.Content
		wantErr          bool
	}{
		{
			name:    "nil input content fails",
			wantErr: true,
		},
		{
			name: "converts a Kong state file to Konnect state file (no workspace, no RG)",
			args: args{
				input: &file.Content{
					Services: []file.FService{
						{
							Service: kong.Service{
								ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
								Name: kong.String("s1"),
								Host: kong.String("mockbin.org"),
							},
							Routes: []*file.FRoute{
								{
									Route: kong.Route{
										Name:                    kong.String("r1"),
										HTTPSRedirectStatusCode: kong.Int(301),
										Paths:                   []*string{kong.String("/r1")},
									},
								},
							},
						},
					},
				},
			},
			want: &file.Content{
				FormatVersion: "3.0",
				Services: []file.FService{
					{
						Service: kong.Service{
							ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
							Name: kong.String("s1"),
							Host: kong.String("mockbin.org"),
						},
						Routes: []*file.FRoute{
							{
								Route: kong.Route{
									Name:                    kong.String("r1"),
									HTTPSRedirectStatusCode: kong.Int(301),
									Paths:                   []*string{kong.String("/r1")},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "converts a Kong state file to Konnect state file (workspace, no RG)",
			args: args{
				input: &file.Content{
					Workspace: "foo",
					Services: []file.FService{
						{
							Service: kong.Service{
								ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
								Name: kong.String("s1"),
								Host: kong.String("mockbin.org"),
							},
							Routes: []*file.FRoute{
								{
									Route: kong.Route{
										Name:                    kong.String("r1"),
										HTTPSRedirectStatusCode: kong.Int(301),
										Paths:                   []*string{kong.String("/r1")},
									},
								},
							},
						},
					},
				},
			},
			want: &file.Content{
				FormatVersion: "3.0",
				Konnect: &file.Konnect{
					RuntimeGroupName: "foo",
				},
				Services: []file.FService{
					{
						Service: kong.Service{
							ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
							Name: kong.String("s1"),
							Host: kong.String("mockbin.org"),
						},
						Routes: []*file.FRoute{
							{
								Route: kong.Route{
									Name:                    kong.String("r1"),
									HTTPSRedirectStatusCode: kong.Int(301),
									Paths:                   []*string{kong.String("/r1")},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "converts a Kong state file to Konnect state file (no workspace, RG)",
			args: args{
				input: &file.Content{
					Services: []file.FService{
						{
							Service: kong.Service{
								ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
								Name: kong.String("s1"),
								Host: kong.String("mockbin.org"),
							},
							Routes: []*file.FRoute{
								{
									Route: kong.Route{
										Name:                    kong.String("r1"),
										HTTPSRedirectStatusCode: kong.Int(301),
										Paths:                   []*string{kong.String("/r1")},
									},
								},
							},
						},
					},
				},
			},
			runtimeGroupName: "foo",
			want: &file.Content{
				FormatVersion: "3.0",
				Konnect: &file.Konnect{
					RuntimeGroupName: "foo",
				},
				Services: []file.FService{
					{
						Service: kong.Service{
							ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
							Name: kong.String("s1"),
							Host: kong.String("mockbin.org"),
						},
						Routes: []*file.FRoute{
							{
								Route: kong.Route{
									Name:                    kong.String("r1"),
									HTTPSRedirectStatusCode: kong.Int(301),
									Paths:                   []*string{kong.String("/r1")},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "converts a Kong state file to Konnect state file (workspace + RG)",
			args: args{
				input: &file.Content{
					Workspace: "bar",
					Services: []file.FService{
						{
							Service: kong.Service{
								ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
								Name: kong.String("s1"),
								Host: kong.String("mockbin.org"),
							},
							Routes: []*file.FRoute{
								{
									Route: kong.Route{
										Name:                    kong.String("r1"),
										HTTPSRedirectStatusCode: kong.Int(301),
										Paths:                   []*string{kong.String("/r1")},
									},
								},
							},
						},
					},
				},
			},
			runtimeGroupName: "foo",
			want: &file.Content{
				FormatVersion: "3.0",
				Konnect: &file.Konnect{
					RuntimeGroupName: "foo",
				},
				Services: []file.FService{
					{
						Service: kong.Service{
							ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
							Name: kong.String("s1"),
							Host: kong.String("mockbin.org"),
						},
						Routes: []*file.FRoute{
							{
								Route: kong.Route{
									Name:                    kong.String("r1"),
									HTTPSRedirectStatusCode: kong.Int(301),
									Paths:                   []*string{kong.String("/r1")},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertKongGatewayToKonnect(tt.args.input, tt.runtimeGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertKongGatewayToKonnect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}
