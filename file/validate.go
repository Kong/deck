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
	var invalid []string

	for _, svc := range c.Services {
		if ret := check("Service.ID", svc.Service.ID); len(ret) > 0 {
			invalid = append(invalid, "Service.ID")
		}

		if ret := check("Service.Host", svc.Service.Host); len(ret) > 0 {
			invalid = append(invalid, "Service.Host")
		}

		if ret := check("Service.Name", svc.Service.Name); len(ret) > 0 {
			invalid = append(invalid, "Service.Name")
		}

		if svc.Service.Port != nil {
			invalid = append(invalid, "Service.Port")
		}
	}

	for _, route := range c.Routes {
		if ret := check("Route.Name", route.Route.Name); len(ret) > 0 {
			invalid = append(invalid, "Route.Name")
		}

		if ret := check("Route.ID", route.Route.ID); len(ret) > 0 {
			invalid = append(invalid, "Route.ID")
		}
	}

	for _, upstream := range c.Upstreams {
		if ret := check("Upstream.Name", upstream.Upstream.Name); len(ret) > 0 {
			invalid = append(invalid, "Upstream.Name")
		}

		if ret := check("Upstream.ID", upstream.Upstream.ID); len(ret) > 0 {
			invalid = append(invalid, "Upstream.ID")
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("unacceptable fields in defaults: %s", strings.Join(invalid, ", "))
	}

	return nil
}
