package utils

import (
	"encoding/base64"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_AddExtToFilename(t *testing.T) {
	type args struct {
		filename string
		ext      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				filename: "foo",
				ext:      "yolo",
			},
			want: "foo.yolo",
		},
		{
			args: args{
				filename: "foo.json",
				ext:      "yolo",
			},
			want: "foo.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddExtToFilename(tt.args.filename, tt.args.ext); got != tt.want {
				t.Errorf("AddExtToFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NameToFilename(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "leading separator",
			args: args{
				name: string(os.PathSeparator) + "foo.md",
			},
			want: "foo.md",
		},
		{
			name: "inner separator",
			args: args{
				name: "bar" + string(os.PathSeparator) + "foo.md",
			},
			want: "bar" + url.PathEscape(string(os.PathSeparator)) + "foo.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NameToFilename(tt.args.name); got != tt.want {
				t.Errorf("NameToFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FilenameToName(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "inner separator",
			args: args{
				filename: "bar" + url.PathEscape(string(os.PathSeparator)) + "foo.md",
			},
			want: "bar" + string(os.PathSeparator) + "foo.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilenameToName(tt.args.filename); got != tt.want {
				t.Errorf("FilenameToName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_BasicAuthFormat(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "symmetric test",
			args: args{
				username: "mickey@mouse.com",
				password: "showMeTheCheese$",
			},
			want: "mickey@mouse.com:showMeTheCheese$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded := BasicAuthFormat(tt.args.username, tt.args.password)
			// Decode
			decoded := make([]byte, len(tt.want))
			_, err := base64.StdEncoding.Decode(decoded, []byte(encoded))
			require.NoError(t, err)
			require.Equal(t, string(decoded), tt.want)
		})
	}
}
