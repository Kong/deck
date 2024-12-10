package kong2kic

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sgwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Helper function to add annotations from a route to an ingress
func addAnnotationsFromRoute(route *file.FRoute, annotations map[string]string) {
	if route.Protocols != nil {
		var protocols []string
		for _, protocol := range route.Protocols {
			if protocol != nil {
				protocols = append(protocols, *protocol)
			}
		}
		annotations["konghq.com/protocols"] = strings.Join(protocols, ",")
	}
	if route.StripPath != nil {
		annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
	}
	if route.PreserveHost != nil {
		annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
	}
	if route.RegexPriority != nil {
		annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
	}
	if route.HTTPSRedirectStatusCode != nil {
		annotations["konghq.com/https-redirect-status-code"] = strconv.Itoa(*route.HTTPSRedirectStatusCode)
	}
	if route.Headers != nil {
		for key, value := range route.Headers {
			annotations["konghq.com/headers."+key] = strings.Join(value, ",")
		}
	}
	if route.PathHandling != nil {
		annotations["konghq.com/path-handling"] = *route.PathHandling
	}
	if route.SNIs != nil {
		var snis []string
		for _, sni := range route.SNIs {
			if sni != nil {
				snis = append(snis, *sni)
			}
		}
		annotations["konghq.com/snis"] = strings.Join(snis, ",")
	}
	if route.RequestBuffering != nil {
		annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
	}
	if route.ResponseBuffering != nil {
		annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
	}
	if route.Methods != nil {
		var methods []string
		for _, method := range route.Methods {
			if method != nil {
				methods = append(methods, *method)
			}
		}
		annotations["konghq.com/methods"] = strings.Join(methods, ",")
	}
	if route.Tags != nil {
		var tags []string
		for _, tag := range route.Tags {
			if tag != nil {
				tags = append(tags, *tag)
			}
		}
		annotations["konghq.com/tags"] = strings.Join(tags, ",")
	}
}

// Helper function to create ingress paths
func createIngressPaths(
	route *file.FRoute,
	serviceName string,
	servicePort *int,
	pathType k8snetv1.PathType,
) []k8snetv1.HTTPIngressPath {
	paths := make([]k8snetv1.HTTPIngressPath, 0, len(route.Paths))
	for _, path := range route.Paths {
		sCopy := *path
		if strings.HasPrefix(sCopy, "~") {
			sCopy = "/" + sCopy
		}
		backend := k8snetv1.IngressBackend{
			Service: &k8snetv1.IngressServiceBackend{
				Name: serviceName,
			},
		}
		if servicePort != nil {
			// check that the port is within the valid range
			if *servicePort > 65535 || *servicePort < 0 {
				log.Fatalf("Port %d is not within the valid range. Please provide a port between 0 and 65535.\n", *servicePort)
			}
			backend.Service.Port.Number = int32(*servicePort)
		}
		paths = append(paths, k8snetv1.HTTPIngressPath{
			Path:     sCopy,
			PathType: &pathType,
			Backend:  backend,
		})
	}
	return paths
}

// Convert route to Ingress (Ingress API)
func populateKICIngressesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	for _, service := range content.Services {
		if service.Name == nil {
			log.Println("Service name is empty. Please provide a name for the service.")
			continue
		}
		serviceName := *service.Name
		for _, route := range service.Routes {
			if route.Name == nil {
				log.Println("Route name is empty. Please provide a name for the route.")
				continue
			}
			routeName := *route.Name

			k8sIngress := k8snetv1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking.k8s.io/v1",
					Kind:       "Ingress",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        calculateSlug(serviceName + "-" + routeName),
					Annotations: make(map[string]string),
				},
			}
			ingressClassName := ClassName
			k8sIngress.Spec.IngressClassName = &ingressClassName

			addAnnotationsFromRoute(route, k8sIngress.ObjectMeta.Annotations)

			pathType := k8snetv1.PathTypeImplementationSpecific

			if len(route.Hosts) == 0 {
				ingressRule := k8snetv1.IngressRule{
					IngressRuleValue: k8snetv1.IngressRuleValue{
						HTTP: &k8snetv1.HTTPIngressRuleValue{
							Paths: createIngressPaths(route, serviceName, service.Port, pathType),
						},
					},
				}
				k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
			} else {
				for _, host := range route.Hosts {
					ingressRule := k8snetv1.IngressRule{
						Host: *host,
						IngressRuleValue: k8snetv1.IngressRuleValue{
							HTTP: &k8snetv1.HTTPIngressRuleValue{
								Paths: createIngressPaths(route, serviceName, service.Port, pathType),
							},
						},
					}
					k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
				}
			}

			err := addPluginsToRoute(service, route, &k8sIngress, kicContent)
			if err != nil {
				return err
			}
			kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
		}
	}
	return nil
}

// Simplify the plugin addition function
func addPluginsToRoute(
	service file.FService,
	route *file.FRoute,
	ingress *k8snetv1.Ingress,
	kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		if plugin.Name == nil {
			log.Println("Plugin name is empty. Please provide a name for the plugin.")
			continue
		}
		pluginName := *plugin.Name
		kongPlugin := configurationv1.KongPlugin{
			TypeMeta: metav1.TypeMeta{
				APIVersion: KICAPIVersion,
				Kind:       KongPluginKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        calculateSlug(*service.Name + "-" + *route.Name + "-" + pluginName),
				Annotations: map[string]string{IngressClass: ClassName},
			},
			PluginName: pluginName,
		}

		// Populate plugin fields
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
			var protocols []string
			for _, protocol := range plugin.Protocols {
				if protocol != nil {
					protocols = append(protocols, *protocol)
				}
			}
			kongPlugin.Protocols = configurationv1.StringsToKongProtocols(protocols)
		}
		if plugin.Tags != nil {
			var tags []string
			for _, tag := range plugin.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongPlugin.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		configJSON, err := json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = apiextensionsv1.JSON{
			Raw: configJSON,
		}

		// Add plugin reference to ingress annotations
		pluginsAnnotation := ingress.ObjectMeta.Annotations["konghq.com/plugins"]
		if pluginsAnnotation == "" {
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
		} else {
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = pluginsAnnotation + "," + kongPlugin.ObjectMeta.Name
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

// HeadersByNameTypeAndValue is a type to sort headers by name, type and value.
// This is needed to ensure that the order of headers is consistent across runs.
type HeadersByNameTypeAndValue []k8sgwapiv1.HTTPHeaderMatch

func (a HeadersByNameTypeAndValue) Len() int      { return len(a) }
func (a HeadersByNameTypeAndValue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HeadersByNameTypeAndValue) Less(i, j int) bool {
	if a[i].Name < a[j].Name {
		return true
	}
	if a[i].Name > a[j].Name {
		return false
	}

	if a[i].Type != nil && a[j].Type == nil {
		return true
	}
	if a[i].Type == nil && a[j].Type != nil {
		return false
	}

	if *a[i].Type < *a[j].Type {
		return true
	}
	if *a[i].Type > *a[j].Type {
		return false
	}
	return a[i].Value < a[j].Value
}

// Convert route to HTTPRoute (Gateway API)
func populateKICIngressesWithGatewayAPI(content *file.Content, kicContent *KICContent) error {
	for _, service := range content.Services {
		for _, route := range service.Routes {
			httpRoute, err := createHTTPRoute(service, route)
			if err != nil {
				log.Println(err)
				continue
			}

			err = addPluginsToGatewayAPIRoute(service, route, httpRoute, kicContent)
			if err != nil {
				return err
			}
			kicContent.HTTPRoutes = append(kicContent.HTTPRoutes, httpRoute)
		}
	}
	return nil
}

func createHTTPRoute(service file.FService, route *file.FRoute) (k8sgwapiv1.HTTPRoute, error) {
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
		return httpRoute, errors.New(
			"service name or route name is empty. Please provide a name for the service" +
				"and the route before generating HTTPRoute manifests",
		)
	}
	httpRoute.ObjectMeta.Annotations = make(map[string]string)

	addAnnotations(&httpRoute, route)
	addHosts(&httpRoute, route)
	addParentRefs(&httpRoute)
	addBackendRefs(&httpRoute, service, route)

	return httpRoute, nil
}

func addAnnotations(httpRoute *k8sgwapiv1.HTTPRoute, route *file.FRoute) {
	if route.PreserveHost != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
	}
	if route.StripPath != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
	}
	if route.HTTPSRedirectStatusCode != nil {
		value := strconv.Itoa(*route.HTTPSRedirectStatusCode)
		httpRoute.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = value
	}
	if route.RegexPriority != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
	}
	if route.PathHandling != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/path-handling"] = *route.PathHandling
	}
	if route.Tags != nil {
		var tags []string
		for _, tag := range route.Tags {
			if tag != nil {
				tags = append(tags, *tag)
			}
		}
		httpRoute.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
	}
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
	if route.RequestBuffering != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
	}
	if route.ResponseBuffering != nil {
		httpRoute.ObjectMeta.Annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
	}
}

func addHosts(httpRoute *k8sgwapiv1.HTTPRoute, route *file.FRoute) {
	if route.Hosts != nil {
		for _, host := range route.Hosts {
			httpRoute.Spec.Hostnames = append(httpRoute.Spec.Hostnames, k8sgwapiv1.Hostname(*host))
		}
	}
}

func addParentRefs(httpRoute *k8sgwapiv1.HTTPRoute) {
	httpRoute.Spec.ParentRefs = append(httpRoute.Spec.ParentRefs, k8sgwapiv1.ParentReference{
		Name: k8sgwapiv1.ObjectName(ClassName),
	})
}

func addBackendRefs(httpRoute *k8sgwapiv1.HTTPRoute, service file.FService, route *file.FRoute) {
	backendRef := k8sgwapiv1.BackendRef{
		BackendObjectReference: k8sgwapiv1.BackendObjectReference{
			Name: k8sgwapiv1.ObjectName(*service.Name),
		},
	}
	if service.Port != nil {
		portNumber := k8sgwapiv1.PortNumber(*service.Port) //nolint:gosec
		backendRef.Port = &portNumber
	}

	var httpHeaderMatch []k8sgwapiv1.HTTPHeaderMatch
	headerMatchExact := k8sgwapiv1.HeaderMatchExact
	headerMatchRegex := k8sgwapiv1.HeaderMatchRegularExpression
	if route.Headers != nil {
		for key, values := range route.Headers {
			if len(values) == 1 && strings.HasPrefix(values[0], "~*") {
				httpHeaderMatch = append(httpHeaderMatch, k8sgwapiv1.HTTPHeaderMatch{
					Name:  k8sgwapiv1.HTTPHeaderName(key),
					Value: values[0][2:],
					Type:  &headerMatchRegex,
				})
			} else {
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
		sort.Sort(HeadersByNameTypeAndValue(httpHeaderMatch))
	}

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
}

func addPluginsToGatewayAPIRoute(
	service file.FService, route *file.FRoute, httpRoute k8sgwapiv1.HTTPRoute, kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name, route name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service, route and the plugin before generating Kong Ingress Controller manifests.")
			continue
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongPlugin.PluginName = *plugin.Name

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
