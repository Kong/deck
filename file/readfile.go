package file

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// getContent reads reads all the YAML files in the directory or the
// file, depending on what fileOrDir represents, merges the content of
// these files and renders a Content.
func getContent(fileOrDir string) (*Content, error) {
	readers, err := getReaders(fileOrDir)
	if err != nil {
		return nil, err
	}
	var res Content
	for _, r := range readers {
		content, err := readContent(r)
		if err != nil {
			return nil, errors.Wrap(err, "reading file")
		}
		err = mergo.Merge(&res, content, mergo.WithAppendSlice)
	}
	return &res, nil
}

// getReaders returns back io.Readers representing all the YAML
// files in a directory. If fileOrDir is a single file, then it
// returns back the reader for the file.
// If fileOrDir is equal to "-" string, then it returns back a io.Reader
// for the os.Stdin file descriptor.
func getReaders(fileOrDir string) ([]io.Reader, error) {
	// special case where `-` means stdin
	if fileOrDir == "-" {
		return []io.Reader{os.Stdin}, nil
	}

	finfo, err := os.Stat(fileOrDir)
	if err != nil {
		return nil, errors.Wrap(err, "reading state file")
	}

	var files []string
	if finfo.IsDir() {
		files, err = yamlFilesInDir(fileOrDir)
		if err != nil {
			return nil,
				errors.Wrap(err, "getting YAML files from directory")
		}
	} else {
		files = append(files, fileOrDir)
	}

	var res []io.Reader
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, errors.Wrap(err, "opening file")
		}
		res = append(res, bufio.NewReader(f))
	}
	return res, nil
}

// readContent reads all the byes untill io.EOF and unmarshals the read
// bytes into Content.
func readContent(reader io.Reader) (*Content, error) {
	var content Content
	var bytes []byte
	var err error
	bytes, err = ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// yamlFilesInDir traverses the directory rooted at dir and
// returns all the files with a case-insensitive extension of `yml` or `yaml`.
func yamlFilesInDir(dir string) ([]string, error) {
	var res []string
	yamlExt := regexp.MustCompile("[Yy]([Aa])?[Mm][Ll]")
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if yamlExt.MatchString(path) {
				res = append(res, path)
			}
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "reading state directory")
	}
	return res, nil
}
