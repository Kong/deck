package file

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"

	"github.com/kong/deck/utils"
)

func validate(content []byte) error {
	var c map[string]interface{}
	err := yaml.Unmarshal(content, &c)
	if err != nil {
		return fmt.Errorf("unmarshaling file content: %w", err)
	}
	c = ensureJSON(c)
	schemaLoader := gojsonschema.NewStringLoader(kongJSONSchema)
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
		err := fmt.Errorf(desc.String())
		errs.Errors = append(errs.Errors, err)
	}
	return errs
}

func validateWorkspaces(workspaces []string) error {
	utils.RemoveDuplicates(&workspaces)
	if len(workspaces) > 1 {
		return fmt.Errorf("it seems like you are trying to sync multiple workspaces "+
			"at the same time (%v).\ndecK doesn't support syncing multiple workspaces at the same time, "+
			"please sync one workspace at a time", workspaces)
	}
	return nil
}
