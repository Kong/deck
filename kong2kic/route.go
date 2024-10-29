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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	var paths []k8snetv1.HTTPIngressPath
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
			//nolint: gosec
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
		kongPlugin := kicv1.KongPlugin{
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
			kongPlugin.Protocols = kicv1.StringsToKongProtocols(protocols)
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
