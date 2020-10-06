package file

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// getContent reads all the YAML and JSON files in the directory or the
// file, depending on the type of each item in filenames, merges the content of
// these files and renders a Content.
func getContent(filenames []string) (*Content, error) {
	var allReaders []io.Reader
	for _, fileOrDir := range filenames {
		readers, err := getReaders(fileOrDir)
		if err != nil {
			return nil, err
		}
		allReaders = append(allReaders, readers...)
	}
	var res Content
	for _, r := range allReaders {
		rawContents, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		err = validate(rawContents)
		if err != nil {
			return nil, errors.Wrap(err, "validating file content")
		}

		contents, err := readContents(bytes.NewReader(rawContents))
		if err != nil {
			return nil, errors.Wrap(err, "reading file")
		}
		for _, content := range contents {
			err = mergo.Merge(&res, content, mergo.WithAppendSlice)
			if err != nil {
				return nil, errors.Wrap(err, "merging file contents")
			}
		}
	}
	return &res, nil
}

// getReaders returns back io.Readers representing all the YAML and JSON
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
		files, err = configFilesInDir(fileOrDir)
		if err != nil {
			return nil,
				errors.Wrap(err, "getting files from directory")
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

// readContents reads all the bytes until io.EOF and unmarshals the read
// bytes into Content.
func readContents(reader io.Reader) ([]*Content, error) {
	var contents []*Content
	var err error

	decoder := yaml.NewDecoder(reader)

	for err == nil {
		var content Content
		err = decoder.Decode(&content)
		contents = append(contents, &content)
	}

	if err != io.EOF {
		return nil, err
	}
	return contents, nil
}

// configFilesInDir traverses the directory rooted at dir and
// returns all the files with a case-insensitive extension of `yml` or `yaml`.
func configFilesInDir(dir string) ([]string, error) {
	var res []string
	exts := regexp.MustCompile("[Yy]([Aa])?[Mm][Ll]|[Jj][Ss][Oo][Nn]")
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if exts.MatchString(path) {
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
