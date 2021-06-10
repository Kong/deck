package file

import (
	"encoding/json"
	"fmt"

	ghodss "github.com/ghodss/yaml"
	"github.com/kong/deck/utils"
	"github.com/xeipuuv/gojsonschema"
)

func check(item string, t *string) bool {
	if t != nil || len(*t) > 0 {
		fmt.Println(item + "[err] is not allowed to be specified.")
		return false
	}
	return true
}

func checkDefaults(kd KongDefaults) bool {
	if ret := check("Service.ID", kd.Service.ID); !ret {
		return false
	}

	if ret := check("Service.Host", kd.Service.Host); !ret {
		return false
	}

	if ret := check("Service.Name", kd.Service.Name); !ret {
		return false
	}

	if kd.Service.Port != nil {
		return false
	}

	if ret := check("Route.Name", kd.Route.Name); !ret {
		return false
	}

	if ret := check("Route.ID", kd.Route.ID); !ret {
		return false
	}

	if ret := check("Target.Target", kd.Target.Target); !ret {
		return false
	}

	if ret := check("Target.ID", kd.Target.ID); !ret {
		return false
	}

	if ret := check("Upstream.Name", kd.Upstream.Name); !ret {
		return false
	}

	if ret := check("Upstream.ID", kd.Upstream.ID); !ret {
		return false
	}

	return true
}

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

	var kongdefaults *KongDefaults
	if err := json.Unmarshal(content, kongdefaults); err != nil {
		return fmt.Errorf("unmarshaling file into KongDefaults: %w", err)
	}
	if kongdefaults != nil {
		if res := checkDefaults(*kongdefaults); !res {
			return fmt.Errorf("fields are not allowed to specify in defaults")
		}
	}

	return errs
}
