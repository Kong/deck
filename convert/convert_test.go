package convert

import (
	"reflect"
	"testing"

	"github.com/kong/deck/file"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
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

func Test_kongServiceToKonnectServicePackage(t *testing.T) {
	type args struct {
		service file.FService
	}
	tests := []struct {
		name    string
		args    args
		want    file.FServicePackage
		wantErr bool
	}{
		{
			name: "converts a kong service to service package",
			args: args{
				service: file.FService{
					Service: kong.Service{
						Name: kong.String("foo"),
						Host: kong.String("foo.example.com"),
					},
				},
			},
			want: file.FServicePackage{
				Name:        kong.String("foo"),
				Description: kong.String("placeholder description for foo service package"),
				Versions: []file.FServiceVersion{
					{
						Version: kong.String("v1"),
						Implementation: &file.Implementation{
							Type: utils.ImplementationTypeKongGateway,
							Kong: &file.Kong{
								Service: &file.FService{
									Service: kong.Service{
										Host: kong.String("foo.example.com"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "converts fails for kong services without a name",
			args: args{
				service: file.FService{
					Service: kong.Service{
						ID:   kong.String("service-id"),
						Host: kong.String("foo.example.com"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kongServiceToKonnectServicePackage(tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("kongServiceToKonnectServicePackage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got = zeroOutID(got)
			if !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func zeroOutID(sp file.FServicePackage) file.FServicePackage {
	res := sp.DeepCopy()
	for _, v := range res.Versions {
		if v.Implementation != nil && v.Implementation.Kong != nil &&
			v.Implementation.Kong.Service != nil {
			v.Implementation.Kong.Service.ID = nil
		}
	}
	return *res
}

func Test_Convert(t *testing.T) {
	type args struct {
		inputFilename          string
		inputFilenames         []string
		outputFilename         string
		fromFormat             Format
		toFormat               Format
		disableMocks           bool
		envVars                map[string]string
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
			name: "errors out when a nameless service is present in the input",
			args: args{
				inputFilename: "testdata/1/input.yaml",
				fromFormat:    FormatKongGateway,
				toFormat:      FormatKonnect,
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
			name: "converts from Kong Gateway to Konnect format",
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
			name: "converts from distributed to kong gateway (no deck specific fields)",
			args: args{
				inputFilename:          "testdata/5/input.yaml",
				outputFilename:         "testdata/5/output.yaml",
				expectedOutputFilename: "testdata/5/output-expected.yaml",
				fromFormat:             FormatDistributed,
				toFormat:               FormatKongGateway,
			},
			wantErr: false,
		},
		{
			name: "converts from distributed to kong gateway with defaults",
			args: args{
				inputFilename:          "testdata/6/input.yaml",
				outputFilename:         "testdata/6/output.yaml",
				expectedOutputFilename: "testdata/6/output-expected.yaml",
				fromFormat:             FormatDistributed,
				toFormat:               FormatKongGateway,
			},
			wantErr: false,
		},
		{
			name: "converts from distributed to kong gateway with multiple files",
			args: args{
				inputFilenames:         []string{"testdata/7/input-1.yaml", "testdata/7/input-2.yaml"},
				outputFilename:         "testdata/7/output.yaml",
				expectedOutputFilename: "testdata/7/output-expected.yaml",
				fromFormat:             FormatDistributed,
				toFormat:               FormatKongGateway,
			},
			wantErr: false,
		},
		{
			name: "converts from distributed to kong gateway with env variables",
			args: args{
				inputFilenames:         []string{"testdata/8/input.yaml"},
				outputFilename:         "testdata/8/output.yaml",
				expectedOutputFilename: "testdata/8/output-expected.yaml",
				fromFormat:             FormatDistributed,
				toFormat:               FormatKongGateway,
				disableMocks:           true,
				envVars: map[string]string{
					"DECK_MOCKBIN_HOST":    "mockbin.org",
					"DECK_MOCKBIN_ENABLED": "true",
					"DECK_WRITE_TIMEOUT":   "777",
					"DECK_FOO_FLOAT":       "666",
				},
			},
			wantErr: false,
		},
		{
			name: "converts from distributed to kong gateway with env variables (mocked)",
			args: args{
				inputFilenames:         []string{"testdata/9/input.yaml"},
				outputFilename:         "testdata/9/output.yaml",
				expectedOutputFilename: "testdata/9/output-expected.yaml",
				fromFormat:             FormatDistributed,
				toFormat:               FormatKongGateway,
				disableMocks:           false,
			},
			wantErr: false,
		},
		{
			name: "errors from distributed to kong gateway with env variables not set",
			args: args{
				inputFilenames: []string{"testdata/9/input.yaml"},
				fromFormat:     FormatDistributed,
				toFormat:       FormatKongGateway,
				disableMocks:   true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFiles := tt.args.inputFilenames
			if tt.args.inputFilename != "" {
				inputFiles = []string{tt.args.inputFilename}
			}
			for k, v := range tt.args.envVars {
				t.Setenv(k, v)
			}
			err := Convert(inputFiles, tt.args.outputFilename, file.YAML, tt.args.fromFormat,
				tt.args.toFormat, !tt.args.disableMocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				got, err := file.GetContentFromFiles([]string{tt.args.outputFilename}, !tt.args.disableMocks)
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				want, err := file.GetContentFromFiles([]string{tt.args.expectedOutputFilename}, !tt.args.disableMocks)
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				got = wipeServiceID(got)
				want = wipeServiceID(want)
				assert.Equal(t, want, got)
			}
		})
	}
}

func wipeServiceID(content *file.Content) *file.Content {
	result := content.DeepCopy()
	result.ServicePackages = nil
	for _, sp := range content.ServicePackages {
		sp := sp
		sp = zeroOutID(sp)
		result.ServicePackages = append(result.ServicePackages, sp)
	}
	return result
}

func Test_convertKongGatewayToKonnect(t *testing.T) {
	type args struct {
		input *file.Content
	}
	tests := []struct {
		name    string
		args    args
		want    *file.Content
		wantErr bool
	}{
		{
			name:    "nil input content fails",
			wantErr: true,
		},
		{
			name: "errors out when a nameless service is present",
			args: args{
				input: &file.Content{
					Services: []file.FService{
						{
							Service: kong.Service{
								ID:   kong.String("1404df16-48c4-42e6-beab-b7f8792587dc"),
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
			wantErr: true,
		},
		{
			name: "converts a Kong state file to Konnect state file",
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
				ServicePackages: []file.FServicePackage{
					{
						Name:        kong.String("s1"),
						Description: kong.String("s2"),
						Versions: []file.FServiceVersion{
							{
								Version: kong.String("v1"),
								Implementation: &file.Implementation{
									Type: utils.ImplementationTypeKongGateway,
									Kong: &file.Kong{
										Service: &file.FService{
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
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertKongGatewayToKonnect(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertKongGatewayToKonnect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}