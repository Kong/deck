package kong2tf

import (
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

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

func newDefaultTerraformBuilder() *DefaultTerraformBuider {
	return &DefaultTerraformBuider{}
}

// Generic function that takes type T and returns map[string]any using JSON marshalling
func toMapAny(resource any) map[string]any {
	resourceMap := make(map[string]interface{})
	resourceJSON, err := json.Marshal(resource)
	if err != nil {
		log.Fatal(err, "Failed to marshal resource")
		return resourceMap
	}
	err = json.Unmarshal(resourceJSON, &resourceMap)
	if err != nil {
		log.Fatal(err, "Failed to unmarshal resource")
		return resourceMap
	}
	return resourceMap
}

func (b *DefaultTerraformBuider) buildControlPlaneVar(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	cpId := "YOUR_CONTROL_PLANE_ID"
	if generateImportsForControlPlaneID != nil {
		cpId = *generateImportsForControlPlaneID
	}
	b.content += fmt.Sprintf(`variable "control_plane_id" {
  type = "string"
  default = "%s"
}`, cpId) + "\n\n"
}

func (b *DefaultTerraformBuider) buildServices(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, service := range content.Services {
		parentResourceName := strings.ReplaceAll(*service.Name, "-", "_")
		b.content += generateResource("gateway_service", parentResourceName, toMapAny(service), map[string]string{}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": service.ID,
			},
		}, []string{})

		for _, route := range service.Routes {
			resourceName := strings.ReplaceAll(*route.Name, "-", "_")
			b.content += generateResource("gateway_route", resourceName, toMapAny(route), map[string]string{
				"service": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": route.ID,
				},
			}, []string{})

			for _, plugin := range route.Plugins {
				pluginName := strings.ReplaceAll(*plugin.Name, "-", "_")
				b.content += generateResource("gateway_plugin", pluginName, toMapAny(plugin), map[string]string{
					"route": resourceName,
				}, importConfig{
					controlPlaneId: generateImportsForControlPlaneID,
					importValues: map[string]*string{
						"id": plugin.ID,
					},
				}, []string{})
			}
		}

		for _, plugin := range service.Plugins {
			resourceName := strings.ReplaceAll(*plugin.Name, "-", "_")
			b.content += generateResource("gateway_plugin", resourceName, toMapAny(plugin), map[string]string{
				"service": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": plugin.ID,
				},
			}, []string{})
		}
	}
}

func (b *DefaultTerraformBuider) buildRoutes(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, route := range content.Routes {
		parentResourceName := strings.ReplaceAll(*route.Name, "-", "_")
		parents := map[string]string{}
		if route.Service != nil {
			parents["service"] = strings.ReplaceAll(*route.Service.Name, "-", "_")
		}
		b.content += generateResource("gateway_route", parentResourceName, toMapAny(route), parents, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": route.ID,
			},
		}, []string{})

		for _, plugin := range route.Plugins {
			resourceName := strings.ReplaceAll(*plugin.Name, "-", "_")
			b.content += generateResource("gateway_plugin", resourceName, toMapAny(plugin), map[string]string{
				"route": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": plugin.ID,
				},
			}, []string{})
		}
	}
}

func (b *DefaultTerraformBuider) buildGlobalPlugins(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, globalPlugin := range content.Plugins {
		resourceName := strings.ReplaceAll(*globalPlugin.Name, "-", "_")
		b.content += generateResource("gateway_plugin", resourceName, toMapAny(globalPlugin), map[string]string{}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": globalPlugin.ID,
			},
		}, []string{})
	}
}

func (b *DefaultTerraformBuider) buildConsumers(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, consumer := range content.Consumers {
		parentResourceName := strings.ReplaceAll(*consumer.Username, "-", "_")
		b.content += generateResource("gateway_consumer", parentResourceName, toMapAny(consumer), map[string]string{}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": consumer.ID,
			},
		}, []string{})

		for _, cg := range consumer.Groups {
			resourceName := strings.ReplaceAll(*cg.Name, "-", "_")

			b.content += generateRelationship("gateway_consumer_group_member", resourceName+"_"+parentResourceName, map[string]string{
				"consumer":       parentResourceName,
				"consumer_group": resourceName,
			}, toMapAny(consumer), importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"consumer_id":       consumer.ID,
					"consumer_group_id": cg.ID,
				},
			})
		}

		for _, acl := range consumer.ACLGroups {
			resourceName := "acl_" + strings.ReplaceAll(*acl.Group, "-", "_")
			b.content += generateResource("gateway_acl", resourceName, toMapAny(acl), map[string]string{
				"consumer_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          acl.ID,
					"consumer_id": consumer.ID,
				},
			}, []string{})
		}

		for _, basicauth := range consumer.BasicAuths {
			lifecycle := []string{}

			if ignoreCredentialChanges {
				lifecycle = []string{
					"password",
				}
			}

			resourceName := "basic_auth_" + strings.ReplaceAll(*basicauth.Username, "-", "_")
			b.content += generateResource("gateway_basic_auth", resourceName, toMapAny(basicauth), map[string]string{
				"consumer_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          basicauth.ID,
					"consumer_id": consumer.ID,
				},
			}, lifecycle)
		}

		for _, keyauth := range consumer.KeyAuths {
			resourceName := "key_auth_" + strings.ReplaceAll(*keyauth.Key, "-", "_")
			b.content += generateResource("gateway_key_auth", resourceName, toMapAny(keyauth), map[string]string{
				"consumer_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          keyauth.ID,
					"consumer_id": consumer.ID,
				},
			}, []string{})
		}

		for _, jwt := range consumer.JWTAuths {
			lifecycle := []string{}

			if ignoreCredentialChanges {
				lifecycle = []string{
					"secret", "key",
				}
			}
			resourceName := "jwt_" + strings.ReplaceAll(*jwt.Key, "-", "_")
			b.content += generateResource("gateway_jwt", resourceName, toMapAny(jwt), map[string]string{
				"consumer_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          jwt.ID,
					"consumer_id": consumer.ID,
				},
			}, lifecycle)
		}

		for _, hmacauth := range consumer.HMACAuths {
			resourceName := "hmac_auth_" + strings.ReplaceAll(*hmacauth.Username, "-", "_")
			b.content += generateResource("gateway_hmac_auth", resourceName, toMapAny(hmacauth), map[string]string{
				"consumer_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          hmacauth.ID,
					"consumer_id": consumer.ID,
				},
			}, []string{})
		}

		for _, plugin := range consumer.Plugins {
			pluginName := strings.ReplaceAll(*plugin.Name, "-", "_")
			b.content += generateResource("gateway_plugin", pluginName, toMapAny(plugin), map[string]string{
				"consumer": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": plugin.ID,
				},
			}, []string{})
		}

	}
}

func (b *DefaultTerraformBuider) buildConsumerGroups(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, cg := range content.ConsumerGroups {
		parentResourceName := strings.ReplaceAll(*cg.Name, "-", "_")
		parents := map[string]string{}
		b.content += generateResource("gateway_consumer_group", parentResourceName, toMapAny(cg), parents, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": cg.ID,
			},
		}, []string{})

		// We intentionally don't generate consumers here. Consumers is a FK reference, not a definition.
		for _, consumer := range cg.Consumers {
			resourceName := strings.ReplaceAll(*consumer.Username, "-", "_")

			b.content += generateRelationship("gateway_consumer_group_member", parentResourceName+"_"+resourceName, map[string]string{
				"consumer":       resourceName,
				"consumer_group": parentResourceName,
			}, toMapAny(consumer), importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"consumer_id":       consumer.ID,
					"consumer_group_id": cg.ID,
				},
			})
		}

		for _, plugin := range cg.Plugins {
			resourceName := strings.ReplaceAll(*plugin.Name, "-", "_")
			b.content += generateResource("gateway_plugin", resourceName, toMapAny(plugin), map[string]string{
				"consumer_group": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": plugin.ID,
				},
			}, []string{})
		}
	}
}

func (b *DefaultTerraformBuider) buildUpstreams(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, upstream := range content.Upstreams {
		parentResourceName := strings.ReplaceAll(*upstream.Name, "-", "_")
		parentResourceName = "upstream_" + strings.ReplaceAll(parentResourceName, ".", "_")
		parents := map[string]string{}
		b.content += generateResource("gateway_upstream", parentResourceName, toMapAny(upstream), parents, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": upstream.ID,
			},
		}, []string{})

		for _, target := range upstream.Targets {
			resourceName := strings.ReplaceAll(*target.Target.Target, ".", "_")
			resourceName = "target_" + strings.ReplaceAll(resourceName, ":", "_")
			b.content += generateResource("gateway_target", resourceName, toMapAny(target), map[string]string{
				"upstream_id": parentResourceName,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id":          target.ID,
					"upstream_id": upstream.ID,
				},
			}, []string{})
		}
	}
}

func (b *DefaultTerraformBuider) buildCACertificates(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	idx := 0
	for _, caCertificate := range content.CACertificates {
		hashedCert := fmt.Sprintf("%x", md5.Sum([]byte(*caCertificate.Cert)))
		resourceName := "ca_cert_" + hashedCert
		idx += 1
		b.content += generateResource("gateway_ca_certificate", resourceName, toMapAny(caCertificate), map[string]string{}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": caCertificate.ID,
			},
		}, []string{})
	}
}

func (b *DefaultTerraformBuider) buildCertificates(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, certificate := range content.Certificates {
		hashedCert := fmt.Sprintf("%x", md5.Sum([]byte(*certificate.Cert)))
		resourceName := "cert_" + hashedCert
		b.content += generateResource("gateway_certificate", resourceName, toMapAny(certificate), map[string]string{}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": certificate.ID,
			},
		}, []string{})

		for _, sni := range certificate.SNIs {
			resourceName := "sni_" + strings.ReplaceAll(*sni.Name, ".", "_")
			b.content += generateResource("gateway_sni", resourceName, toMapAny(sni), map[string]string{
				"certificate": "cert_" + hashedCert,
			}, importConfig{
				controlPlaneId: generateImportsForControlPlaneID,
				importValues: map[string]*string{
					"id": sni.ID,
				},
			}, []string{})
		}
	}
}

func (b *DefaultTerraformBuider) buildVaults(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) {
	for _, vault := range content.Vaults {
		parentResourceName := strings.ReplaceAll(*vault.Name, "-", "_")
		parents := map[string]string{}
		b.content += generateResourceWithCustomizations("gateway_vault", parentResourceName, toMapAny(vault), parents, map[string]string{
			"config": "jsonencode",
		}, importConfig{
			controlPlaneId: generateImportsForControlPlaneID,
			importValues: map[string]*string{
				"id": vault.ID,
			},
		}, []string{})
	}
}

func (b *DefaultTerraformBuider) getContent() string {
	return b.content
}
