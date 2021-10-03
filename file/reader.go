package file

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
)

var (
	kong230Version = semver.MustParse("2.3.0")
	kong140Version = semver.MustParse("1.4.0")
)

// RenderConfig contains necessary information to render a correct
// KongConfig from a file.
type RenderConfig struct {
	CurrentState *state.KongState
	KongVersion  semver.Version
}

// GetContentFromFiles reads in a file with a slice of filenames and constructs
// a state. If filename is `-`, then it will read from os.Stdin.
// If filename represents a directory, it will traverse the tree
// rooted at filename, read all the files with .yaml, .yml and .json extensions
// and generate a content after a merge of the content from all the files.
//
// It will return an error if the file representation is invalid
// or if there is any error during processing.
func GetContentFromFiles(filenames []string) (*Content, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	return getContent(filenames)
}

func GetForKonnect(fileContent *Content, opt RenderConfig) (*utils.KongRawState, *utils.KonnectRawState, error) {
	var builder stateBuilder
	// setup
	builder.targetContent = fileContent
	builder.currentState = opt.CurrentState

	kongState, konnectState, err := builder.build()
	if err != nil {
		return nil, nil, fmt.Errorf("building state: %w", err)
	}
	tags := []string{}
	if fileContent.Info != nil {
		tags = fileContent.Info.SelectorTags
	}
	mergeSelectorTags(kongState, opt.KongVersion, tags)

	return kongState, konnectState, nil
}

// Get process the fileContent and renders a RawState.
// IDs of entities are matches based on currentState.
func Get(fileContent *Content, opt RenderConfig) (*utils.KongRawState, error) {
	var builder stateBuilder
	// setup
	builder.targetContent = fileContent
	builder.currentState = opt.CurrentState

	state, _, err := builder.build()
	if err != nil {
		return nil, fmt.Errorf("building state: %w", err)
	}
	tags := []string{}
	if fileContent.Info != nil {
		tags = fileContent.Info.SelectorTags
	}
	mergeSelectorTags(state, opt.KongVersion, tags)

	return state, nil
}

func mergeSelectorTags(state *utils.KongRawState, version semver.Version, tags []string) {
	if version.GTE(kong140Version) {
		utils.MustMergeTags(state.HMACAuths, tags)
		utils.MustMergeTags(state.ACLGroups, tags)
		utils.MustMergeTags(state.KeyAuths, tags)
		utils.MustMergeTags(state.BasicAuths, tags)
		utils.MustMergeTags(state.JWTAuths, tags)
		utils.MustMergeTags(state.Oauth2Creds, tags)
		utils.MustMergeTags(state.HMACAuths, tags)
	}
	if version.GTE(kong230Version) {
		utils.MustMergeTags(state.MTLSAuths, tags)
	}
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
