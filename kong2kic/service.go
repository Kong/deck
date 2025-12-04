package kong2kic

import (
	"log"
	"strconv"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Main function to populate KIC services with annotations
func populateKICServicesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	for _, service := range content.Services {
		if service.Name == nil {
			log.Println("Service name is empty. Please provide a name for the service.")
			continue
		}

		// Create Kubernetes service
		k8sService := createK8sService(&service, content.Upstreams)

		// Add annotations from the service object
		addAnnotationsFromService(&service, k8sService.Annotations)

		// Populate upstream policies based on KIC version
		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			populateKICUpstreamPolicy(content, &service, k8sService, kicContent)
		} else {
			populateKICUpstream(content, &service, k8sService, kicContent)
		}

		// Add plugins to the service
		err := addPluginsToService(&service, k8sService, kicContent)
		if err != nil {
			return err
		}

		// Append the Kubernetes service to KIC content
		kicContent.Services = append(kicContent.Services, *k8sService)
	}
	return nil
}

// Helper function to create a Kubernetes Service from a Kong service
func createK8sService(service *file.FService, upstreams []file.FUpstream) *k8scorev1.Service {
	k8sService := &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ServiceAPIVersionv1,
			Kind:       ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(*service.Name),
			Annotations: make(map[string]string),
		},
	}

	// Determine the protocol (default to TCP)
	protocol := k8scorev1.ProtocolTCP
	if service.Protocol != nil && strings.ToUpper(*service.Protocol) == string(k8scorev1.ProtocolUDP) {
		protocol = k8scorev1.ProtocolUDP
	}

	// Set the service port
	if service.Port != nil {
		// check that the port is within the valid range
		if *service.Port > 65535 || *service.Port < 0 {
			log.Fatalf("Port %d is not within the valid range. Please provide a port between 0 and 65535.\n", *service.Port)
		}
		servicePort := k8scorev1.ServicePort{
			Protocol: protocol,
			Port:       int32(*service.Port),
			TargetPort: intstr.FromInt(*service.Port),
		}
		k8sService.Spec.Ports = []k8scorev1.ServicePort{servicePort}
	}

	// Configure the service selector or external name
	if service.Host == nil {
		k8sService.Spec.Selector = map[string]string{"app": *service.Name}
	} else {
		k8sService.Spec.Type = k8scorev1.ServiceTypeExternalName
		k8sService.Spec.ExternalName = *service.Host

		// Check if the host matches any upstream
		if isUpstreamReferenced(service.Host, upstreams) {
			k8sService.Spec.Selector = map[string]string{"app": *service.Name}
			k8sService.Spec.Type = ""
			k8sService.Spec.ExternalName = ""
		}
	}

	return k8sService
}

// Helper function to check if the service host matches any upstream name
func isUpstreamReferenced(host *string, upstreams []file.FUpstream) bool {
	for _, upstream := range upstreams {
		if upstream.Name != nil && strings.EqualFold(*upstream.Name, *host) {
			return true
		}
	}
	return false
}

// Helper function to add annotations from a service to a Kubernetes service
func addAnnotationsFromService(service *file.FService, annotations map[string]string) {
	if service.Protocol != nil {
		annotations[KongHQProtocol] = *service.Protocol
	}
	if service.Path != nil {
		annotations[KongHQPath] = *service.Path
	}
	if service.ClientCertificate != nil && service.ClientCertificate.ID != nil {
		annotations[KongHQClientCert] = *service.ClientCertificate.ID
	}
	if service.ReadTimeout != nil {
		annotations[KongHQReadTimeout] = strconv.Itoa(*service.ReadTimeout)
	}
	if service.WriteTimeout != nil {
		annotations[KongHQWriteTimeout] = strconv.Itoa(*service.WriteTimeout)
	}
	if service.ConnectTimeout != nil {
		annotations[KongHQConnectTimeout] = strconv.Itoa(*service.ConnectTimeout)
	}
	if service.Retries != nil {
		annotations[KongHQRetries] = strconv.Itoa(*service.Retries)
	}
	addTagsToAnnotations(service.Tags, annotations)
}

// processPlugin is a helper function that processes a single plugin for a service
func processPlugin(
	plugin *file.FPlugin,
	ownerName string,
	annotations map[string]string,
	kicContent *KICContent,
) error {
	if plugin.Name == nil {
		log.Println("Plugin name is empty. Please provide a name for the plugin.")
		return nil
	}

	// Create a KongPlugin
	kongPlugin, err := createKongPlugin(plugin, ownerName)
	if err != nil {
		return err
	}
	if kongPlugin == nil {
		return nil
	}

	// Add the plugin name to the service annotations
	addPluginToAnnotations(kongPlugin.Name, annotations)

	// Append the KongPlugin to KIC content
	kicContent.KongPlugins = append(kicContent.KongPlugins, *kongPlugin)
	return nil
}

// addPluginsToService adds plugins from both service-level and top-level plugin configurations
func addPluginsToService(service *file.FService, k8sService *k8scorev1.Service, kicContent *KICContent) error {
	if service.Name == nil {
		log.Println("Service name is empty. Please provide a name for the service.")
		return nil
	}
	ownerName := *service.Name

	// Process service-level plugins
	for _, plugin := range service.Plugins {
		if err := processPlugin(plugin, ownerName, k8sService.Annotations, kicContent); err != nil {
			return err
		}
	}
	return nil
}
