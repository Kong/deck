package file

import (
	"bytes"
	"io"

	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	yaml "gopkg.in/yaml.v2"
)

func validate(content []byte) error {
	var errs utils.ErrArray
	var err error

	decoder := yaml.NewDecoder(bytes.NewReader(content))
	for err == nil {
		var c map[string]interface{}
		err = decoder.Decode(&c)
		if err != nil && err != io.EOF {
			return err
		}
		c = ensureJSON(c)
		schemaLoader := gojsonschema.NewStringLoader(contentSchema)
		documentLoader := gojsonschema.NewGoLoader(c)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return err
		}

		for _, desc := range result.Errors() {
			err := errors.New(desc.String())
			errs.Errors = append(errs.Errors, err)
		}
	}

	if err != io.EOF {
		return err
	}
	if errs.Errors == nil {
		return nil
	}
	return errs
}
