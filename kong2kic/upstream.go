package kong2kic

import (
	"log"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
)

// UpstreamPolicy for KIC v3
func populateKICUpstreamPolicy(
	content *file.Content,
	service *file.FService,
	k8sservice *k8scorev1.Service,
	kicContent *KICContent,
) {
	if content.Upstreams != nil {
		var kongUpstreamPolicy kicv1beta1.KongUpstreamPolicy
		kongUpstreamPolicy.APIVersion = KICAPIVersionV1Beta1
		kongUpstreamPolicy.Kind = UpstreamPolicyKind
		if service.Name != nil {
			kongUpstreamPolicy.ObjectMeta.Name = calculateSlug(*service.Name + "-upstream")
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
			return
		}

		// Find the upstream (if any) whose name matches the service host and copy the upstream
		// into kongUpstreamPolicy. Append the kongUpstreamPolicy to kicContent.KongUpstreamPolicies.
		for _, upstream := range content.Upstreams {
			if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
				// add an annotation to the k8sservice to link this kongUpstreamPolicy to it
				k8sservice.ObjectMeta.Annotations["konghq.com/upstream-policy"] = kongUpstreamPolicy.ObjectMeta.Name
				var threshold int
				if upstream.Healthchecks != nil && upstream.Healthchecks.Threshold != nil {
					threshold = int(*upstream.Healthchecks.Threshold)
				}
				var activeHealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var activeUnhealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var passiveHealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var passiveUnhealthyHTTPStatuses []kicv1beta1.HTTPStatus

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Active != nil &&
					upstream.Healthchecks.Active.Healthy != nil {
					activeHealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus, len(upstream.Healthchecks.Active.Healthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Active.Healthy.HTTPStatuses {
						activeHealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Active != nil &&
					upstream.Healthchecks.Active.Unhealthy != nil {
					activeUnhealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus,
						len(upstream.Healthchecks.Active.Unhealthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Active.Unhealthy.HTTPStatuses {
						activeUnhealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Passive != nil &&
					upstream.Healthchecks.Passive.Healthy != nil {
					passiveHealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus, len(upstream.Healthchecks.Passive.Healthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Passive.Healthy.HTTPStatuses {
						passiveHealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Passive != nil &&
					upstream.Healthchecks.Passive.Unhealthy != nil {
					passiveUnhealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus,
						len(upstream.Healthchecks.Passive.Unhealthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Passive.Unhealthy.HTTPStatuses {
						passiveUnhealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				// populeate kongUpstreamPolicy.Spec with the
				// non-nil attributes in upstream.
				if upstream.Algorithm != nil {
					kongUpstreamPolicy.Spec.Algorithm = upstream.Algorithm
				}
				if upstream.Slots != nil {
					kongUpstreamPolicy.Spec.Slots = upstream.Slots
				}

				if upstream.HashOn != nil && upstream.Algorithm != nil && *upstream.Algorithm == "consistent-hashing" {
					kongUpstreamPolicy.Spec.HashOn = &kicv1beta1.KongUpstreamHash{
						Input:      (*kicv1beta1.HashInput)(upstream.HashOn),
						Header:     upstream.HashOnHeader,
						Cookie:     upstream.HashOnCookie,
						CookiePath: upstream.HashOnCookiePath,
						QueryArg:   upstream.HashOnQueryArg,
						URICapture: upstream.HashOnURICapture,
					}
				}
				if upstream.HashFallback != nil && upstream.Algorithm != nil && *upstream.Algorithm == "consistent-hashing" {
					kongUpstreamPolicy.Spec.HashOnFallback = &kicv1beta1.KongUpstreamHash{
						Input:      (*kicv1beta1.HashInput)(upstream.HashFallback),
						Header:     upstream.HashFallbackHeader,
						QueryArg:   upstream.HashFallbackQueryArg,
						URICapture: upstream.HashFallbackURICapture,
					}
				}
				if upstream.Healthchecks != nil {
					kongUpstreamPolicy.Spec.Healthchecks = &kicv1beta1.KongUpstreamHealthcheck{
						Threshold: &threshold,
						Active: &kicv1beta1.KongUpstreamActiveHealthcheck{
							Type:                   upstream.Healthchecks.Active.Type,
							Concurrency:            upstream.Healthchecks.Active.Concurrency,
							HTTPPath:               upstream.Healthchecks.Active.HTTPPath,
							HTTPSSNI:               upstream.Healthchecks.Active.HTTPSSni,
							HTTPSVerifyCertificate: upstream.Healthchecks.Active.HTTPSVerifyCertificate,
							Timeout:                upstream.Healthchecks.Active.Timeout,
							Headers:                upstream.Healthchecks.Active.Headers,
							Healthy: &kicv1beta1.KongUpstreamHealthcheckHealthy{
								Interval:     upstream.Healthchecks.Active.Healthy.Interval,
								Successes:    upstream.Healthchecks.Active.Healthy.Successes,
								HTTPStatuses: activeHealthyHTTPStatuses,
							},
							Unhealthy: &kicv1beta1.KongUpstreamHealthcheckUnhealthy{
								HTTPFailures: upstream.Healthchecks.Active.Unhealthy.HTTPFailures,
								TCPFailures:  upstream.Healthchecks.Active.Unhealthy.TCPFailures,
								Timeouts:     upstream.Healthchecks.Active.Unhealthy.Timeouts,
								Interval:     upstream.Healthchecks.Active.Unhealthy.Interval,
								HTTPStatuses: activeUnhealthyHTTPStatuses,
							},
						},
						Passive: &kicv1beta1.KongUpstreamPassiveHealthcheck{
							Type: upstream.Healthchecks.Passive.Type,
							Healthy: &kicv1beta1.KongUpstreamHealthcheckHealthy{
								HTTPStatuses: passiveHealthyHTTPStatuses,
								Interval:     upstream.Healthchecks.Passive.Healthy.Interval,
								Successes:    upstream.Healthchecks.Passive.Healthy.Successes,
							},
							Unhealthy: &kicv1beta1.KongUpstreamHealthcheckUnhealthy{
								HTTPFailures: upstream.Healthchecks.Passive.Unhealthy.HTTPFailures,
								HTTPStatuses: passiveUnhealthyHTTPStatuses,
								TCPFailures:  upstream.Healthchecks.Passive.Unhealthy.TCPFailures,
								Timeouts:     upstream.Healthchecks.Passive.Unhealthy.Timeouts,
								Interval:     upstream.Healthchecks.Passive.Unhealthy.Interval,
							},
						},
					}
				}
				// add konghq.com/tags annotation if upstream.Tags is not nil
				if upstream.Tags != nil {
					var tags []string
					for _, tag := range upstream.Tags {
						if tag != nil {
							tags = append(tags, *tag)
						}
					}
					// initialize the annotations map if it is nil
					if kongUpstreamPolicy.ObjectMeta.Annotations == nil {
						kongUpstreamPolicy.ObjectMeta.Annotations = make(map[string]string)
					}
					kongUpstreamPolicy.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
				}
				kicContent.KongUpstreamPolicies = append(kicContent.KongUpstreamPolicies, kongUpstreamPolicy)
				break
			}
		}
	}
}

// KongIngress with upstream section for KIC v2
func populateKICUpstream(
	content *file.Content,
	service *file.FService,
	k8sservice *k8scorev1.Service,
	kicContent *KICContent,
) {
	// add Kong specific configuration to the k8s service via a KongIngress resource

	if content.Upstreams != nil {
		var kongIngress kicv1.KongIngress
		kongIngress.APIVersion = KICAPIVersion
		kongIngress.Kind = IngressKind
		if service.Name != nil {
			kongIngress.ObjectMeta.Name = calculateSlug(*service.Name + "-upstream")
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
			return
		}
		kongIngress.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}

		// add an annotation to the k8sservice to link this kongIngress to it
		k8sservice.ObjectMeta.Annotations["konghq.com/override"] = kongIngress.ObjectMeta.Name

		// Find the upstream (if any) whose name matches the service host and copy the upstream
		// into a kicv1.KongIngress resource. Append the kicv1.KongIngress to kicContent.KongIngresses.
		for _, upstream := range content.Upstreams {
			if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
				kongIngress.Upstream = &kicv1.KongIngressUpstream{
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
				}
				// add konghq.com/tags annotation if upstream.Tags is not nil
				if upstream.Tags != nil {
					var tags []string
					for _, tag := range upstream.Tags {
						if tag != nil {
							tags = append(tags, *tag)
						}
					}
					kongIngress.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
				}
				kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
				break
			}
		}
	}
}
