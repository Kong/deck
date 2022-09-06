package utils

import (
	"net/url"
	"os"
	"testing"

	"github.com/blang/semver/v4"
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

func Test_ParseKongVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    semver.Version
		wantErr bool
	}{
		{
			args: args{
				version: "1.0.1",
			},
			want:    semver.Version{Major: 1, Minor: 0},
			wantErr: false,
		},
		{
			args: args{
				version: "1.3.0.1",
			},
			want:    semver.Version{Major: 1, Minor: 3},
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.1",
			},
			want:    semver.Version{Major: 0, Minor: 14},
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.2rc",
			},
			want:    semver.Version{Major: 0, Minor: 14},
			wantErr: false,
		},
		{
			args: args{
				version: "0.14.2rc1",
			},
			want:    semver.Version{Major: 0, Minor: 14},
			wantErr: false,
		},
		{
			args: args{
				version: "0.33-enterprise-edition",
			},
			want:    semver.Version{Major: 0, Minor: 33},
			wantErr: false,
		},
		{
			args: args{
				version: "1.3.0-0-enterprise-edition",
			},
			want:    semver.Version{Major: 1, Minor: 3},
			wantErr: false,
		},
		{
			args: args{
				version: "",
			},
			want:    semver.Version{},
			wantErr: true,
		},
		{
			args: args{
				version: "0-1.1",
			},
			want:    semver.Version{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKongVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKongVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equals(tt.want) {
				t.Errorf("ParseKongVersion() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}
