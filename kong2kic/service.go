package kong2kic

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	k8scorev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func populateKICServicesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	// Iterate Kong Services and create k8s Services,
	// then create KongIngress resources for Kong Service Upstream data.
	// Finally, create KongPlugin resources for each plugin
	// associated with the service.
	for i := range content.Services {
		service := content.Services[i]

		var k8sService k8scorev1.Service
		var protocol k8scorev1.Protocol

		k8sService.TypeMeta.APIVersion = "v1"
		k8sService.TypeMeta.Kind = "Service"
		if service.Name != nil {
			k8sService.ObjectMeta.Name = calculateSlug(*service.Name)
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
			continue
		}
		k8sService.ObjectMeta.Annotations = make(map[string]string)

		// default TCP unless service.Protocol is equal to k8scorev1.ProtocolUDP
		if service.Protocol != nil && k8scorev1.Protocol(strings.ToUpper(*service.Protocol)) == k8scorev1.ProtocolUDP {
			protocol = k8scorev1.ProtocolUDP
		} else {
			protocol = k8scorev1.ProtocolTCP
		}

		if service.Port != nil {
			sPort := k8scorev1.ServicePort{
				Protocol:   protocol,
				Port:       int32(*service.Port),
				TargetPort: intstr.IntOrString{IntVal: int32(*service.Port)},
			}
			k8sService.Spec.Ports = append(k8sService.Spec.Ports, sPort)
		}

		if service.Host == nil {
			// the service host is not informed
			// then the service name is the selector
			k8sService.Spec.Selector = map[string]string{"app": *service.Name}
		} else {
			// Assume the service host is not referencing an upstream by default
			// we need to create an external service
			k8sService.Spec.Type = k8scorev1.ServiceTypeExternalName
			k8sService.Spec.ExternalName = *service.Host

			for _, upstream := range content.Upstreams {
				if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
					// the service host is referencing an upstream
					// we do nothing here as upstreams are handled in populateKICUpstream
					// interpret service name as the selector
					k8sService.Spec.Selector = map[string]string{"app": *service.Name}
					// Reset the external service settings as they are not needed
					k8sService.Spec.Type = ""
					k8sService.Spec.ExternalName = ""
					break
				}
			}
		}

		// add konghq.com/protocol annotation if service.Protocol is not nil
		if service.Protocol != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/protocol"] = *service.Protocol
		}

		// add konghq.com/path annotation if service.Path is not nil
		if service.Path != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/path"] = *service.Path
		}

		// add konghq.com/client-cert annotation if service.ClientCertificate is not nil
		if service.ClientCertificate != nil && service.ClientCertificate.ID != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/client-cert"] = *service.ClientCertificate.ID
		}

		// add konghq.com/read-timeout annotation if service.ReadTimeout is not nil
		if service.ReadTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/read-timeout"] = strconv.Itoa(*service.ReadTimeout)
		}

		// add konghq.com/write-timeout annotation if service.WriteTimeout is not nil
		if service.WriteTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/write-timeout"] = strconv.Itoa(*service.WriteTimeout)
		}

		// add konghq.com/connect-timeout annotation if service.ConnectTimeout is not nil
		if service.ConnectTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/connect-timeout"] = strconv.Itoa(*service.ConnectTimeout)
		}

		// add konghq.com/retries annotation if service.Retries is not nil
		if service.Retries != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/retries"] = strconv.Itoa(*service.Retries)
		}

		// add konghq.com/tags annotation if service.Tags is not nil
		if service.Tags != nil {
			var tags []string
			for _, tag := range service.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			k8sService.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// Use KongUpstreamPolicy for KICv3
			populateKICUpstreamPolicy(content, &service, &k8sService, kicContent)
		} else {
			// Use KongIngress for KICv2
			populateKICUpstream(content, &service, &k8sService, kicContent)
		}

		// iterate over the plugins for this service, create a KongPlugin for each one and add an annotation to the service
		// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
		// create a plugins annotation in the k8sservice to link the plugin to it
		err := addPluginsToService(service, k8sService, kicContent)
		if err != nil {
			return err
		}

		kicContent.Services = append(kicContent.Services, k8sService)

	}
	return nil
}

func addPluginsToService(service file.FService, k8sService k8scorev1.Service, kicContent *KICContent) error {
	for _, plugin := range service.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && service.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service and the plugin before generating Kong Ingress Controller manifests.")
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

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

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

		if k8sService.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
		} else {
			annotations := k8sService.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}
