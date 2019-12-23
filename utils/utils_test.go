package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	notEmpty := "not-empty"
	emptyString := ""
	var nilPointer *string
	assert.False(Empty(&notEmpty))
	assert.True(Empty(nilPointer))
	assert.True(Empty(&emptyString))
}

func TestUUID(t *testing.T) {
	assert := assert.New(t)
	uuid := UUID()
	assert.NotEmpty(uuid)
	assert.Regexp(regexp.MustCompile(
		"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"),
		uuid)
}

func Test_cleanKongVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{
				version: "1.0.1",
			},
			want:    "1.0",
			wantErr: false,
		},
		{
			args: args{
				version: "1.3.0.1",
			},
			want:    "1.3",
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.1",
			},
			want:    "0.14",
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.2rc",
			},
			want:    "0.14",
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.2rc1",
			},
			want:    "0.14",
			wantErr: false,
		},
		{
			args: args{
				version: "0.33-enterprise-edition",
			},
			want:    "0.33",
			wantErr: false,
		},
		{
			args: args{
				version: "1.3.0-0-enterprise-edition",
			},
			want:    "1.3",
			wantErr: false,
		},
		{
			args: args{
				version: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			args: args{
				version: "0-1.1",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanKongVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanKongVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleanKongVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
