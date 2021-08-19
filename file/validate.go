package file

import (
	"encoding/json"
	"fmt"
	"strings"

	ghodss "github.com/ghodss/yaml"
	"github.com/kong/deck/utils"
	"github.com/xeipuuv/gojsonschema"
)

func validate(content []byte) error {
	var c map[string]interface{}
	err := ghodss.Unmarshal(content, &c)
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

	kongContent := Content{}
	if err := json.Unmarshal(content, &kongContent); err != nil {
		return fmt.Errorf("unmarshaling file into KongDefaults: %w", err)
	}

	if err := checkDefaults(kongContent); err != nil {
		return fmt.Errorf("default values are not allowed for these fields: %w", err)
	}

	return errs
}

func check(item string, t *string) string {
	if t != nil && len(*t) > 0 {
		return item
	}
	return ""
}

func checkDefaults(c Content) error {
	if c.Info == nil {
		return nil
	}
	var invalid []string

	defaults := c.Info.Defaults
	svc := defaults.Service
	if svc != nil {
		if ret := check("Service.ID", svc.ID); len(ret) > 0 {
			invalid = append(invalid, "Service.ID")
		}

		if ret := check("Service.Host", svc.Host); len(ret) > 0 {
			invalid = append(invalid, "Service.Host")
		}

		if ret := check("Service.Name", svc.Name); len(ret) > 0 {
			invalid = append(invalid, "Service.Name")
		}

		if svc.Port != nil {
			invalid = append(invalid, "Service.Port")
		}
	}

	route := defaults.Route
	if route != nil {
		if ret := check("Route.Name", route.Name); len(ret) > 0 {
			invalid = append(invalid, "Route.Name")
		}

		if ret := check("Route.ID", route.ID); len(ret) > 0 {
			invalid = append(invalid, "Route.ID")
		}
	}

	upstream := defaults.Upstream
	if upstream != nil {
		if ret := check("Upstream.Name", upstream.Name); len(ret) > 0 {
			invalid = append(invalid, "Upstream.Name")
		}

		if ret := check("Upstream.ID", upstream.ID); len(ret) > 0 {
			invalid = append(invalid, "Upstream.ID")
		}
	}

	target := defaults.Target
	if target != nil {
		if ret := check("target.ID", target.ID); len(ret) > 0 {
			invalid = append(invalid, "target.ID")
		}

		if ret := check("target.Target", target.Target); len(ret) > 0 {
			invalid = append(invalid, "target.Target")
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("unacceptable fields in defaults: %s", strings.Join(invalid, ", "))
	}

	return nil
}
