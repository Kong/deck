package file

import (
	yaml "github.com/ghodss/yaml"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func validate(content []byte) error {
	var c map[string]interface{}
	err := yaml.Unmarshal(content, &c)
	if err != nil {
		return errors.Wrap(err, "unmarshaling file content")
	}
	c = ensureJSON(c)
	schemaLoader := gojsonschema.NewStringLoader(contentSchema)
	documentLoader := gojsonschema.NewGoLoader(c)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}
	var errs utils.ErrArray
	for _, desc := range result.Errors() {
		err := errors.New(desc.String())
		errs.Errors = append(errs.Errors, err)
	}
	return errs
}
