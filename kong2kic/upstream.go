package kong2kic

import (
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function to find the matching upstream
func findMatchingUpstream(serviceHost *string, upstreams []file.FUpstream) *file.FUpstream {
	if serviceHost == nil {
		return nil
	}
	for _, upstream := range upstreams {
		if upstream.Name != nil && strings.EqualFold(*upstream.Name, *serviceHost) {
			return &upstream
		}
	}
	return nil
}

// Helper function to convert HTTP statuses
func convertHTTPStatuses(statuses []int) []kicv1beta1.HTTPStatus {
	if statuses == nil {
		return nil
	}
	result := make([]kicv1beta1.HTTPStatus, len(statuses))
	for i, status := range statuses {
		result[i] = kicv1beta1.HTTPStatus(status)
	}
	return result
}

// Helper function to populate active healthcheck
func populateActiveHealthcheck(active *kong.ActiveHealthcheck) *kicv1beta1.KongUpstreamActiveHealthcheck {
	if active == nil {
		return nil
	}
	return &kicv1beta1.KongUpstreamActiveHealthcheck{
		Type:                   active.Type,
		Concurrency:            active.Concurrency,
		HTTPPath:               active.HTTPPath,
		HTTPSSNI:               active.HTTPSSni,
		HTTPSVerifyCertificate: active.HTTPSVerifyCertificate,
		Timeout:                active.Timeout,
		Headers:                active.Headers,
		Healthy:                populateHealthcheckHealthy(active.Healthy),
		Unhealthy:              populateHealthcheckUnhealthy(active.Unhealthy),
	}
}

// Helper function to populate passive healthcheck
func populatePassiveHealthcheck(passive *kong.PassiveHealthcheck) *kicv1beta1.KongUpstreamPassiveHealthcheck {
	if passive == nil {
		return nil
	}
	return &kicv1beta1.KongUpstreamPassiveHealthcheck{
		Type:      passive.Type,
		Healthy:   populateHealthcheckHealthy(passive.Healthy),
		Unhealthy: populateHealthcheckUnhealthy(passive.Unhealthy),
	}
}

// Helper function to populate healthcheck healthy settings
func populateHealthcheckHealthy(healthy *kong.Healthy) *kicv1beta1.KongUpstreamHealthcheckHealthy {
	if healthy == nil {
		return nil
	}
	return &kicv1beta1.KongUpstreamHealthcheckHealthy{
		Interval:     healthy.Interval,
		Successes:    healthy.Successes,
		HTTPStatuses: convertHTTPStatuses(healthy.HTTPStatuses),
	}
}

// Helper function to populate healthcheck unhealthy settings
func populateHealthcheckUnhealthy(unhealthy *kong.Unhealthy) *kicv1beta1.KongUpstreamHealthcheckUnhealthy {
	if unhealthy == nil {
		return nil
	}
	return &kicv1beta1.KongUpstreamHealthcheckUnhealthy{
		HTTPFailures: unhealthy.HTTPFailures,
		TCPFailures:  unhealthy.TCPFailures,
		Timeouts:     unhealthy.Timeouts,
		Interval:     unhealthy.Interval,
		HTTPStatuses: convertHTTPStatuses(unhealthy.HTTPStatuses),
	}
}

// Function to populate KIC Upstream Policy for KIC v3
func populateKICUpstreamPolicy(
	content *file.Content,
	service *file.FService,
	k8sService *k8scorev1.Service,
	kicContent *KICContent,
) {
	if service.Name == nil {
		log.Println("Service name is empty. Please provide the necessary information.")
		return
	}

	// Find the matching upstream
	upstream := findMatchingUpstream(service.Host, content.Upstreams)
	if upstream == nil {
		return
	}

	// Create KongUpstreamPolicy
	kongUpstreamPolicy := kicv1beta1.KongUpstreamPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: KICAPIVersionV1Beta1,
			Kind:       UpstreamPolicyKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(*service.Name + "-upstream"),
			Annotations: make(map[string]string),
		},
	}

	// Add an annotation to link the upstream policy to the k8s service
	if k8sService.ObjectMeta.Annotations == nil {
		k8sService.ObjectMeta.Annotations = make(map[string]string)
	}
	k8sService.ObjectMeta.Annotations["konghq.com/upstream-policy"] = kongUpstreamPolicy.ObjectMeta.Name

	// Populate the Upstream Policy Spec
	populateKongUpstreamPolicySpec(upstream, &kongUpstreamPolicy)

	// Add tags to annotations
	addTagsToAnnotations(upstream.Tags, kongUpstreamPolicy.ObjectMeta.Annotations)

	// Append the KongUpstreamPolicy to KIC content
	kicContent.KongUpstreamPolicies = append(kicContent.KongUpstreamPolicies, kongUpstreamPolicy)
}

// Helper function to populate KongUpstreamPolicy Spec
func populateKongUpstreamPolicySpec(upstream *file.FUpstream, policy *kicv1beta1.KongUpstreamPolicy) {
	if upstream.Algorithm != nil {
		policy.Spec.Algorithm = upstream.Algorithm
	}
	if upstream.Slots != nil {
		policy.Spec.Slots = upstream.Slots
	}

	if upstream.Algorithm != nil && *upstream.Algorithm == "consistent-hashing" {
		if upstream.HashOn != nil {
			policy.Spec.HashOn = &kicv1beta1.KongUpstreamHash{
				Input:      (*kicv1beta1.HashInput)(upstream.HashOn),
				Header:     upstream.HashOnHeader,
				Cookie:     upstream.HashOnCookie,
				CookiePath: upstream.HashOnCookiePath,
				QueryArg:   upstream.HashOnQueryArg,
				URICapture: upstream.HashOnURICapture,
			}
		}
		if upstream.HashFallback != nil {
			policy.Spec.HashOnFallback = &kicv1beta1.KongUpstreamHash{
				Input:      (*kicv1beta1.HashInput)(upstream.HashFallback),
				Header:     upstream.HashFallbackHeader,
				QueryArg:   upstream.HashFallbackQueryArg,
				URICapture: upstream.HashFallbackURICapture,
			}
		}
	}

	// Handle healthchecks
	if upstream.Healthchecks != nil {
		var threshold int
		if upstream.Healthchecks.Threshold != nil {
			threshold = int(*upstream.Healthchecks.Threshold)
		}
		policy.Spec.Healthchecks = &kicv1beta1.KongUpstreamHealthcheck{
			Threshold: &threshold,
			Active:    populateActiveHealthcheck(upstream.Healthchecks.Active),
			Passive:   populatePassiveHealthcheck(upstream.Healthchecks.Passive),
		}
	}
}

// Function to populate KIC Upstream for KIC v2
func populateKICUpstream(
	content *file.Content,
	service *file.FService,
	k8sService *k8scorev1.Service,
	kicContent *KICContent,
) {
	if service.Name == nil {
		log.Println("Service name is empty. Please provide the necessary information.")
		return
	}

	// Find the matching upstream
	upstream := findMatchingUpstream(service.Host, content.Upstreams)
	if upstream == nil {
		return
	}

	// Create KongIngress
	kongIngress := kicv1.KongIngress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: KICAPIVersion,
			Kind:       IngressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(*service.Name + "-upstream"),
			Annotations: map[string]string{IngressClass: ClassName},
		},
		Upstream: &kicv1.KongIngressUpstream{
			HostHeader:             upstream.HostHeader,
			Algorithm:              upstream.Algorithm,
			Slots:                  upstream.Slots,
			Healthchecks:           upstream.Healthchecks,
			HashOn:                 upstream.HashOn,
			HashFallback:           upstream.HashFallback,
			HashOnHeader:           upstream.HashOnHeader,
			HashFallbackHeader:     upstream.HashFallbackHeader,
			HashOnCookie:           upstream.HashOnCookie,
			HashOnCookiePath:       upstream.HashOnCookiePath,
			HashOnQueryArg:         upstream.HashOnQueryArg,
			HashFallbackQueryArg:   upstream.HashFallbackQueryArg,
			HashOnURICapture:       upstream.HashOnURICapture,
			HashFallbackURICapture: upstream.HashFallbackURICapture,
		},
	}

	// Add an annotation to link the KongIngress to the k8s service
	if k8sService.ObjectMeta.Annotations == nil {
		k8sService.ObjectMeta.Annotations = make(map[string]string)
	}
	k8sService.ObjectMeta.Annotations["konghq.com/override"] = kongIngress.ObjectMeta.Name

	// Add tags to annotations
	addTagsToAnnotations(upstream.Tags, kongIngress.ObjectMeta.Annotations)

	// Append the KongIngress to KIC content
	kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
}
