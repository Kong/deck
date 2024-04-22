package kong2kic

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sgwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Convert route to Ingress (Ingress API)
func populateKICIngressesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	// For each route under each service create one ingress.
	// If the route has multiple hosts, create the relevant host declarations in the ingress
	// and under each hosts create a declaration for each path in the route.
	// If the route has no hosts, create a declaration in the ingress for each path in the route.
	// Map additional fields in the route to annotations in the ingress.
	// If the route has plugins, create a KongPlugin for each one and add an annotation to the ingress
	// to link the plugin to it.
	for _, service := range content.Services {
		for _, route := range service.Routes {
			// save all ingresses we create for this route so we can then
			// assign them the plugins defined for the route
			var (
				k8sIngress           k8snetv1.Ingress
				pathTypeImplSpecific = k8snetv1.PathTypeImplementationSpecific
			)

			k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
			k8sIngress.TypeMeta.Kind = "Ingress"
			if service.Name != nil && route.Name != nil {
				k8sIngress.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name)
			} else {
				log.Println("Service name or route name is empty. This is not recommended." +
					"Please, provide a name for the service and the route before generating Kong Ingress Controller manifests.")
				continue
			}
			ingressClassName := ClassName
			k8sIngress.Spec.IngressClassName = &ingressClassName
			k8sIngress.ObjectMeta.Annotations = make(map[string]string)

			// add konghq.com/protocols annotation if route.Protocols is not nil
			if route.Protocols != nil {
				var protocols string
				for _, protocol := range route.Protocols {
					if protocols == "" {
						protocols = *protocol
					} else {
						protocols = protocols + "," + *protocol
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/protocols"] = protocols
			}

			// add konghq.com/strip-path annotation if route.StripPath is not nil
			if route.StripPath != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
			}

			// add konghq.com/preserve-host annotation if route.PreserveHost is not nil
			if route.PreserveHost != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
			}

			// add konghq.com/regex-priority annotation if route.RegexPriority is not nil
			if route.RegexPriority != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
			}

			// add konghq.com/https-redirect-status-code annotation if route.HTTPSRedirectStatusCode is not nil
			if route.HTTPSRedirectStatusCode != nil {
				value := strconv.Itoa(*route.HTTPSRedirectStatusCode)
				k8sIngress.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = value
			}

			// add konghq.com/headers.* annotation if route.Headers is not nil
			if route.Headers != nil {
				for key, value := range route.Headers {
					k8sIngress.ObjectMeta.Annotations["konghq.com/headers."+key] = strings.Join(value, ",")
				}
			}

			// add konghq.com/path-handling annotation if route.PathHandling is not nil
			if route.PathHandling != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/path-handling"] = *route.PathHandling
			}

			// add konghq.com/snis annotation if route.SNIs is not nil
			if route.SNIs != nil {
				var snis string
				for _, sni := range route.SNIs {
					if snis == "" {
						snis = *sni
					} else {
						snis = snis + "," + *sni
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/snis"] = snis
			}

			// add konghq.com/request-buffering annotation if route.RequestBuffering is not nil
			if route.RequestBuffering != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
			}

			// add konghq.com/response-buffering annotation if route.ResponseBuffering is not nil
			if route.ResponseBuffering != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
			}

			// add konghq.com/methods annotation if route.Methods is not nil
			if route.Methods != nil {
				var methods string
				for _, method := range route.Methods {
					if methods == "" {
						methods = *method
					} else {
						methods = methods + "," + *method
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/methods"] = methods
			}

			// add konghq.com/tags annotation if route.Tags is not nil
			if route.Tags != nil {
				var tags []string
				for _, tag := range route.Tags {
					if tag != nil {
						tags = append(tags, *tag)
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
			}

			if len(route.Hosts) == 0 {
				// iterate route.Paths and create a k8sIngress.Spec.Rules for each one.
				// If service.Port is not nil, add it to the k8sIngress.Spec.Rules
				ingressRule := k8snetv1.IngressRule{
					IngressRuleValue: k8snetv1.IngressRuleValue{
						HTTP: &k8snetv1.HTTPIngressRuleValue{
							Paths: []k8snetv1.HTTPIngressPath{},
						},
					},
				}
				for _, path := range route.Paths {
					// if path starts with ~ then add / to the beginning of the path
					// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingr
					//                                             ess-regular-expression-paths-for-kong-3x-compatibility
					sCopy := *path
					if strings.HasPrefix(*path, "~") {
						sCopy = "/" + *path
					}
					if service.Port != nil {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
							k8snetv1.HTTPIngressPath{
								Path:     sCopy,
								PathType: &pathTypeImplSpecific,
								Backend: k8snetv1.IngressBackend{
									Service: &k8snetv1.IngressServiceBackend{
										Name: *service.Name,
										Port: k8snetv1.ServiceBackendPort{
											Number: int32(*service.Port),
										},
									},
								},
							})
					} else {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
							k8snetv1.HTTPIngressPath{
								Path:     sCopy,
								PathType: &pathTypeImplSpecific,
								Backend: k8snetv1.IngressBackend{
									Service: &k8snetv1.IngressServiceBackend{
										Name: *service.Name,
									},
								},
							})
					}
				}
				k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
			} else {
				// Iterate route.Hosts and create a k8sIngress.Spec.Rules for each one.
				// For each host, iterate route.Paths and add a k8snetv1.HTTPIngressPath to the k8sIngress.Spec.Rules.
				// If service.Port is not nil, add it to the k8sIngress.Spec.Rules
				for _, host := range route.Hosts {
					ingressRule := k8snetv1.IngressRule{
						Host: *host,
						IngressRuleValue: k8snetv1.IngressRuleValue{
							HTTP: &k8snetv1.HTTPIngressRuleValue{
								Paths: []k8snetv1.HTTPIngressPath{},
							},
						},
					}
					for _, path := range route.Paths {
						// if path starts with ~ then add / to the beginning of the path
						// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingr
						//                                             ess-regular-expression-paths-for-kong-3x-compatibility
						sCopy := *path
						if strings.HasPrefix(*path, "~") {
							sCopy = "/" + *path
						}
						if service.Port != nil {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
								k8snetv1.HTTPIngressPath{
									Path:     sCopy,
									PathType: &pathTypeImplSpecific,
									Backend: k8snetv1.IngressBackend{
										Service: &k8snetv1.IngressServiceBackend{
											Name: *service.Name,
											Port: k8snetv1.ServiceBackendPort{
												Number: int32(*service.Port),
											},
										},
									},
								})
						} else {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
								k8snetv1.HTTPIngressPath{
									Path:     sCopy,
									PathType: &pathTypeImplSpecific,
									Backend: k8snetv1.IngressBackend{
										Service: &k8snetv1.IngressServiceBackend{
											Name: *service.Name,
										},
									},
								})
						}
					}
					k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
				}
			}

			err := addPluginsToRoute(service, route, k8sIngress, kicContent)
			if err != nil {
				return err
			}
			kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
		}
	}
	return nil
}

func addPluginsToRoute(
	service file.FService,
	route *file.FRoute,
	ingress k8snetv1.Ingress,
	kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name, route name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service, route and the plugin before generating Kong Ingress Controller manifests.")
			continue
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}

		// populate enabled, runon, ordering and protocols
		if plugin.Enabled != nil {
			kongPlugin.Disabled = !*plugin.Enabled
		}
		if plugin.RunOn != nil {
			kongPlugin.RunOn = *plugin.RunOn
		}
		if plugin.Ordering != nil {
			kongPlugin.Ordering = &kong.PluginOrdering{
				Before: plugin.Ordering.Before,
				After:  plugin.Ordering.After,
			}
		}
		if plugin.Protocols != nil {
			protocols := make([]string, len(plugin.Protocols))
			for i, protocol := range plugin.Protocols {
				if protocol != nil {
					protocols[i] = *protocol
				}
			}
			kongPlugin.Protocols = kicv1.StringsToKongProtocols(protocols)
		}

		// add konghq.com/tags annotation if plugin.Tags is not nil
		if plugin.Tags != nil {
			var tags []string
			for _, tag := range plugin.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongPlugin.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		if ingress.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
		} else {
			annotations := ingress.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

// Convert route to HTTPRoute (Gateway API)
func populateKICIngressesWithGatewayAPI(content *file.Content, kicContent *KICContent) error {
	for _, service := range content.Services {
		for _, route := range service.Routes {
			var httpRoute k8sgwapiv1.HTTPRoute
			httpRoute.Kind = "HTTPRoute"
			if targetKICVersionAPI == KICV3GATEWAY {
				httpRoute.APIVersion = GatewayAPIVersionV1
			} else {
				httpRoute.APIVersion = GatewayAPIVersionV1Beta1
			}
			if service.Name != nil && route.Name != nil {
				httpRoute.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name)
			} else {
				log.Println("Service name or route name is empty. This is not recommended." +
					"Please, provide a name for the service and the route before generating HTTPRoute manifests.")
				continue
			}
			httpRoute.ObjectMeta.Annotations = make(map[string]string)

			// add konghq.com/preserve-host annotation if route.PreserveHost is not nil
			if route.PreserveHost != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
			}

			// add konghq.com/strip-path annotation if route.StripPath is not nil
			if route.StripPath != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
			}

			// add konghq.com/https-redirect-status-code annotation if route.HTTPSRedirectStatusCode is not nil
			if route.HTTPSRedirectStatusCode != nil {
				value := strconv.Itoa(*route.HTTPSRedirectStatusCode)
				httpRoute.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = value
			}

			// add konghq.com/regex-priority annotation if route.RegexPriority is not nil
			if route.RegexPriority != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
			}

			// add konghq.com/path-handling annotation if route.PathHandling is not nil
			if route.PathHandling != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/path-handling"] = *route.PathHandling
			}

			// add konghq.com/snis annotation if route.SNIs is not nil
			if route.SNIs != nil {
				var snis string
				for _, sni := range route.SNIs {
					if snis == "" {
						snis = *sni
					} else {
						snis = snis + "," + *sni
					}
				}
				httpRoute.ObjectMeta.Annotations["konghq.com/snis"] = snis
			}

			// add konghq.com/request-buffering annotation if route.RequestBuffering is not nil
			if route.RequestBuffering != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
			}

			// add konghq.com/response-buffering annotation if route.ResponseBuffering is not nil
			if route.ResponseBuffering != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
			}

			// add konghq.com/tags annotation if route.Tags is not nil
			if route.Tags != nil {
				var tags []string
				for _, tag := range route.Tags {
					if tag != nil {
						tags = append(tags, *tag)
					}
				}
				httpRoute.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
			}

			// if route.hosts is not nil, add them to the httpRoute
			if route.Hosts != nil {
				for _, host := range route.Hosts {
					httpRoute.Spec.Hostnames = append(httpRoute.Spec.Hostnames, k8sgwapiv1.Hostname(*host))
				}
			}

			// add kong as the spec.parentRef.name
			httpRoute.Spec.ParentRefs = append(httpRoute.Spec.ParentRefs, k8sgwapiv1.ParentReference{
				Name: k8sgwapiv1.ObjectName(ClassName),
			})

			// add service details to HTTPBackendRef

			backendRef := k8sgwapiv1.BackendRef{
				BackendObjectReference: k8sgwapiv1.BackendObjectReference{
					Name: k8sgwapiv1.ObjectName(*service.Name),
				},
			}
			if service.Port != nil {
				portNumber := k8sgwapiv1.PortNumber(*service.Port)
				backendRef.Port = &portNumber
			}

			var httpHeaderMatch []k8sgwapiv1.HTTPHeaderMatch
			headerMatchExact := k8sgwapiv1.HeaderMatchExact
			headerMatchRegex := k8sgwapiv1.HeaderMatchRegularExpression
			// if route.Headers is not nil, add them to the httpHeaderMatch
			if route.Headers != nil {
				for key, values := range route.Headers {
					// if values has only one value and that value starts with
					// the special prefix ~*, the value is interpreted as a regular expression.
					if len(values) == 1 && strings.HasPrefix(values[0], "~*") {
						httpHeaderMatch = append(httpHeaderMatch, k8sgwapiv1.HTTPHeaderMatch{
							Name:  k8sgwapiv1.HTTPHeaderName(key),
							Value: values[0][2:],
							Type:  &headerMatchRegex,
						})
					} else {
						// if multiple values are present, add them as comma separated values
						// if only one value is present, add it as a single value
						var value string
						if len(values) > 1 {
							value = strings.Join(values, ",")
						} else {
							value = values[0]
						}
						httpHeaderMatch = append(httpHeaderMatch, k8sgwapiv1.HTTPHeaderMatch{
							Name:  k8sgwapiv1.HTTPHeaderName(key),
							Value: value,
							Type:  &headerMatchExact,
						})
					}
				}
			}

			// If path is not nil, then for each path, for each method, add a httpRouteRule
			// to the httpRoute
			if route.Paths != nil {
				for _, path := range route.Paths {
					var httpPathMatch k8sgwapiv1.HTTPPathMatch
					pathMatchRegex := k8sgwapiv1.PathMatchRegularExpression
					pathMatchPrefix := k8sgwapiv1.PathMatchPathPrefix

					if strings.HasPrefix(*path, "~") {
						httpPathMatch.Type = &pathMatchRegex
						regexPath := (*path)[1:]
						httpPathMatch.Value = &regexPath
					} else {
						httpPathMatch.Type = &pathMatchPrefix
						httpPathMatch.Value = path
					}

					// if no method is specified, add a httpRouteRule to the httpRoute with headers and path
					if route.Methods == nil {
						httpRoute.Spec.Rules = append(httpRoute.Spec.Rules, k8sgwapiv1.HTTPRouteRule{
							Matches: []k8sgwapiv1.HTTPRouteMatch{
								{
									Path:    &httpPathMatch,
									Headers: httpHeaderMatch,
								},
							},
							BackendRefs: []k8sgwapiv1.HTTPBackendRef{
								{
									BackendRef: backendRef,
								},
							},
						})
					}

					for _, method := range route.Methods {
						httpMethod := k8sgwapiv1.HTTPMethod(*method)
						httpRoute.Spec.Rules = append(httpRoute.Spec.Rules, k8sgwapiv1.HTTPRouteRule{
							Matches: []k8sgwapiv1.HTTPRouteMatch{
								{
									Path:    &httpPathMatch,
									Method:  &httpMethod,
									Headers: httpHeaderMatch,
								},
							},
							BackendRefs: []k8sgwapiv1.HTTPBackendRef{
								{
									BackendRef: backendRef,
								},
							},
						})
					}
				}
			} else {
				// If path is nil, then for each method, add a httpRouteRule
				// to the httpRoute with headers and no path
				for _, method := range route.Methods {
					httpMethod := k8sgwapiv1.HTTPMethod(*method)
					httpRoute.Spec.Rules = append(httpRoute.Spec.Rules, k8sgwapiv1.HTTPRouteRule{
						Matches: []k8sgwapiv1.HTTPRouteMatch{
							{
								Method:  &httpMethod,
								Headers: httpHeaderMatch,
							},
						},
						BackendRefs: []k8sgwapiv1.HTTPBackendRef{
							{
								BackendRef: backendRef,
							},
						},
					})
				}
			}
			err := addPluginsToGatewayAPIRoute(service, route, httpRoute, kicContent)
			if err != nil {
				return err
			}
			kicContent.HTTPRoutes = append(kicContent.HTTPRoutes, httpRoute)
		}
	}
	return nil
}

func addPluginsToGatewayAPIRoute(
	service file.FService, route *file.FRoute, httpRoute k8sgwapiv1.HTTPRoute, kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name, route name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service, route and the plugin before generating Kong Ingress Controller manifests.")
			continue
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}

		// populate enabled, runon, ordering and protocols
		if plugin.Enabled != nil {
			kongPlugin.Disabled = !*plugin.Enabled
		}
		if plugin.RunOn != nil {
			kongPlugin.RunOn = *plugin.RunOn
		}
		if plugin.Ordering != nil {
			kongPlugin.Ordering = &kong.PluginOrdering{
				Before: plugin.Ordering.Before,
				After:  plugin.Ordering.After,
			}
		}
		if plugin.Protocols != nil {
			protocols := make([]string, len(plugin.Protocols))
			for i, protocol := range plugin.Protocols {
				if protocol != nil {
					protocols[i] = *protocol
				}
			}
			kongPlugin.Protocols = kicv1.StringsToKongProtocols(protocols)
		}

		// add konghq.com/tags annotation if plugin.Tags is not nil
		if plugin.Tags != nil {
			var tags []string
			for _, tag := range plugin.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongPlugin.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		// add plugins as extensionRef under filters for every rule
		for i := range httpRoute.Spec.Rules {
			httpRoute.Spec.Rules[i].Filters = append(httpRoute.Spec.Rules[i].Filters, k8sgwapiv1.HTTPRouteFilter{
				ExtensionRef: &k8sgwapiv1.LocalObjectReference{
					Name:  k8sgwapiv1.ObjectName(kongPlugin.ObjectMeta.Name),
					Kind:  KongPluginKind,
					Group: "configuration.konghq.com",
				},
				Type: k8sgwapiv1.HTTPRouteFilterExtensionRef,
			})
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}
