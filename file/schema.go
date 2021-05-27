package file

import _ "embed" // for embedding only

//go:embed kong_json_schema.json
var kongJSONSchema string
