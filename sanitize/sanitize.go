package sanitize

import (
	"context"
	"reflect"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

type Sanitizer struct {
	ctx          context.Context
	client       *kong.Client
	content      *file.Content
	isKonnect    bool
	salt         string
	sanitizedMap map[string]interface{}
}

type SanitizerOptions struct {
	Ctx       context.Context
	Client    *kong.Client
	Content   *file.Content
	Salt      string
	IsKonnect bool
}

func NewSanitizer(opts *SanitizerOptions) *Sanitizer {
	saltToUse := opts.Salt
	if saltToUse == "" {
		saltToUse = utils.UUID()
	}

	return &Sanitizer{
		ctx:          opts.Ctx,
		client:       opts.Client,
		content:      opts.Content,
		isKonnect:    opts.IsKonnect,
		salt:         saltToUse,
		sanitizedMap: make(map[string]interface{}),
	}
}

func (s *Sanitizer) Sanitize() (*file.Content, error) {
	content := reflect.ValueOf(s.content)
	if content.Kind() == reflect.Ptr {
		content = content.Elem()
	}
	contentType := content.Type()

	for i := 0; i < content.NumField(); i++ {
		fieldName := contentType.Field(i).Name
		if _, exempt := topLevelExemptedFields[fieldName]; exempt {
			continue
		}
		fieldValueSet := content.Field(i)

		if fieldValueSet.IsValid() && fieldValueSet.CanInterface() && !fieldValueSet.IsZero() {
			s.sanitizeField(fieldValueSet)
		}
	}

	return s.content, nil
}

func (s *Sanitizer) sanitizeField(_ reflect.Value) {
	// sanitize the field based on type
	// to be added in next commit
}
