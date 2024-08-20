package kong2tf

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"regexp"
	"text/template"

	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/mitchellh/hashstructure"
)

// cleanField removes all characters from the input string that are not letters, digits, underscores, and dashes.
func cleanField(input string) string {
	// Regular expression to match disallowed characters and replace them
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return re.ReplaceAllString(input, "")
}

// dashToUnderscore replaces all dashes in the input string with underscores.
func dashToUnderscore(input string) string {
	// Regular expression to match dashes and replace them with underscores
	re := regexp.MustCompile(`-`)
	return re.ReplaceAllString(input, "_")
}

var funcs = template.FuncMap{
	"hash":             hashstructure.Hash,
	"jsonmarshal":      json.Marshal,
	"cleanField":       cleanField,
	"dashToUnderscore": dashToUnderscore,
}

type DefaultTerraformBuider struct {
	content string
}

type TemplateObjectWrapper struct {
	Content                          interface{}
	GenerateImportsForControlPlaneID *string
	IgnoreCredentialChanges          bool
}

func newDefaultTerraformBuilder() *DefaultTerraformBuider {
	return &DefaultTerraformBuider{}
}

//go:embed templates/service.go.tmpl
var terraformServiceTemplate string

func (b *DefaultTerraformBuider) buildServices(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	tmpl, err := template.New(terraformServiceTemplate).Funcs(funcs).Parse(terraformServiceTemplate)
	if err != nil {
		log.Fatal(err, "Failed to parse template")
		return // Changed from log.Fatalf to return after logging the error
	}

	for index, service := range content.Services {

		var buffer bytes.Buffer
		err = tmpl.Execute(&buffer, service)
		if err != nil {
			log.Fatal(err, "Failed to execute template for service", "serviceIndex", index+1)
		}

		b.content += buffer.String()
	}
}

//go:embed templates/route.go.tmpl
var terraformRouteTemplate string

func (b *DefaultTerraformBuider) buildRoutes(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	logbasics.Info("Starting to build routes")
	logbasics.Info("Template content before parsing", "template", terraformRouteTemplate)

	tmpl, err := template.New(terraformRouteTemplate).Funcs(funcs).Parse(terraformRouteTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range content.Services {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, service)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/global_plugin.go.tmpl
var terraformGlobalPluginTemplate string

func (b *DefaultTerraformBuider) buildGlobalPlugins(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	logbasics.Info("Starting to build global plugins")
	logbasics.Info("Template content before parsing", "template", terraformGlobalPluginTemplate)

	tmpl, err := template.New(terraformGlobalPluginTemplate).Funcs(funcs).Parse(terraformGlobalPluginTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, globalPlugin := range content.Plugins {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, globalPlugin)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/consumer.go.tmpl
var terraformConsumerTemplate string

func (b *DefaultTerraformBuider) buildConsumers(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	logbasics.Info("Starting to build consumers")
	logbasics.Info("Template content before parsing", "template", terraformConsumerTemplate)

	tmpl, err := template.New(terraformConsumerTemplate).Funcs(funcs).Parse(terraformConsumerTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, consumer := range content.Consumers {
		wrapper := TemplateObjectWrapper{
			Content:                          consumer,
			GenerateImportsForControlPlaneID: generateImportsForControlPlaneID,
			IgnoreCredentialChanges:          ignoreCredentialChanges,
		}
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, wrapper)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/consumer_group.go.tmpl
var terraformConsumerGroupTemplate string

func (b *DefaultTerraformBuider) buildConsumerGroups(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	logbasics.Info("Starting to build consumer groups")
	logbasics.Info("Template content before parsing", "template", terraformConsumerGroupTemplate)

	tmpl, err := template.New(terraformConsumerGroupTemplate).Funcs(funcs).Parse(terraformConsumerGroupTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, consumerGroup := range content.ConsumerGroups {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, consumerGroup)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/upstream.go.tmpl
var terraformUpstreamTemplate string

func (b *DefaultTerraformBuider) buildUpstreams(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	tmpl, err := template.New(terraformUpstreamTemplate).Funcs(funcs).Parse(terraformUpstreamTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, upstream := range content.Upstreams {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, upstream)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/ca_certificate.go.tmpl
var terraformCACertificateTemplate string

func (b *DefaultTerraformBuider) buildCACertificates(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	tmpl, err := template.New(terraformCACertificateTemplate).Funcs(funcs).Parse(terraformCACertificateTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, caCertificate := range content.CACertificates {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, caCertificate)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/certificate.go.tmpl
var terraformCertificateTemplate string

func (b *DefaultTerraformBuider) buildCertificates(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	tmpl, err := template.New(terraformCertificateTemplate).Funcs(funcs).Parse(terraformCertificateTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, certificate := range content.Certificates {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, certificate)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

//go:embed templates/vault.go.tmpl
var terraformVaultTemplate string

func (b *DefaultTerraformBuider) buildVaults(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	tmpl, err := template.New(terraformVaultTemplate).Funcs(funcs).Parse(terraformVaultTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for _, vault := range content.Vaults {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, vault)
		if err != nil {
			log.Fatal(err)
		}
		b.content += buffer.String()
	}
}

func (b *DefaultTerraformBuider) getContent() string {
	return b.content
}
