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
	// if not set otherwise in input, workspace is an empty string
	var workspace string
	type args struct {
		inputFilename          string
		outputFilename         string
		fromFormat             Format
		toFormat               Format
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Convert(tt.args.inputFilename, tt.args.outputFilename, tt.args.fromFormat,
				tt.args.toFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				gotMap, err := file.GetContentFromFiles([]string{tt.args.outputFilename})
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				wantMap, err := file.GetContentFromFiles([]string{tt.args.expectedOutputFilename})
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
				}
				got := wipeServiceID(gotMap[workspace])
				want := wipeServiceID(wantMap[workspace])
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
