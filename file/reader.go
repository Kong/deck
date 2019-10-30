package file

import (
	"fmt"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
)

// GetContentFromFile reads in a file with filename and constructs
// a state. If filename is `-`, then it will read from os.Stdin.
// If filename represents a directory, it will traverse the tree
// rooted at filename, read all the files with .yaml, .yml and .json extensions
// and generate a content after a merge of the content from all the files.
//
// It will return an error if the file representation is invalid
// or if there is any error during processing.
func GetContentFromFile(filename string) (*Content, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	return getContent(filename)
}

// Get process the fileContent and renders a RawState.
// IDs of entities are matches based on currentState.
func Get(fileContent *Content,
	currentState *state.KongState) (*utils.KongRawState, error) {

	var builder stateBuilder
	// setup
	builder.targetContent = fileContent
	builder.currentState = currentState
	d, err := utils.GetKongDefaulter()
	if err != nil {
		return nil, errors.Wrap(err, "creating defaulter")
	}
	builder.defaulter = d

	return builder.build()
}

func ensureJSON(m map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = yamlToJSON(v2)
		case []interface{}:
			var array []interface{}
			for _, element := range v2 {
				switch el := element.(type) {
				case map[interface{}]interface{}:
					array = append(array, yamlToJSON(el))
				default:
					array = append(array, el)
				}
			}
			if array != nil {
				res[fmt.Sprint(k)] = array
			} else {
				res[fmt.Sprint(k)] = v
			}
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}

func yamlToJSON(m map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = yamlToJSON(v2)
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}
