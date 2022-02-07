package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	ghodss "github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

// getContent reads all the YAML and JSON files in the directory or the
// file, depending on the type of each item in filenames, merges the content of
// these files and renders a Content.
func getContent(filenames []string) (*Content, error) {
	var allReaders []io.Reader
	var workspaces []string
	for _, fileOrDir := range filenames {
		readers, err := getReaders(fileOrDir)
		if err != nil {
			return nil, err
		}
		allReaders = append(allReaders, readers...)
	}
	var res Content
	for _, r := range allReaders {
		content, err := readContent(r)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		workspaces = append(workspaces, content.Workspace)
		err = mergo.Merge(&res, content, mergo.WithAppendSlice)
		if err != nil {
			return nil, fmt.Errorf("merging file contents: %w", err)
		}
	}
	if err := validateWorkspaces(workspaces); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var files []string
	if finfo.IsDir() {
		files, err = configFilesInDir(fileOrDir)
		if err != nil {
			return nil, fmt.Errorf("getting files from directory: %w", err)
		}
	} else {
		files = append(files, fileOrDir)
	}

	var res []io.Reader
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		res = append(res, bufio.NewReader(f))
	}
	return res, nil
}

// readContent reads all the byes until io.EOF and unmarshals the read
// bytes into Content.
func readContent(reader io.Reader) (*Content, error) {
	var err error
	contentBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	renderedContent, err := renderTemplate(string(contentBytes))
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}
	renderedContentBytes := []byte(renderedContent)
	err = validate(renderedContentBytes)
	if err != nil {
		return nil, fmt.Errorf("validating file content: %w", err)
	}
	var result Content
	err = yamlUnmarshal(renderedContentBytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// yamlUnmarshal is a wrapper around yaml.Unmarshal to ensure that the right
// yaml package is in use. Using ghodss/yaml ensures that no
// `map[interface{}]interface{}` is present in go-kong.Plugin.Configuration.
// If it is present, then it leads to a silent error. See Github Issue #144.
// The verification for this is done using a test.
func yamlUnmarshal(bytes []byte, v interface{}) error {
	return ghodss.Unmarshal(bytes, v)
}

// configFilesInDir traverses the directory rooted at dir and
// returns all the files with a case-insensitive extension of `yml` or `yaml`.
func configFilesInDir(dir string) ([]string, error) {
	var res []string
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".yaml", ".yml", ".json":
				res = append(res, path)
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("reading state directory: %w", err)
	}
	return res, nil
}

func getPrefixedEnvVar(key string) (string, error) {
	const envVarPrefix = "DECK_"
	if !strings.HasPrefix(key, envVarPrefix) {
		return "", fmt.Errorf("environment variables in the state file must "+
			"be prefixed with 'DECK_', found: '%s'", key)
	}
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("environment variable '%s' present in state file but not set", key)
	}
	return value, nil
}

func renderTemplate(content string) (string, error) {
	t := template.New("state").Funcs(template.FuncMap{
		"env": getPrefixedEnvVar,
	}).Delims("${{", "}}")
	t, err := t.Parse(content)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, nil)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
