package convert

import (
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

// secureServiceProtocols are the protocols for which tls_verify is relevant.
var secureServiceProtocols = map[string]bool{
	"https": true,
	"tls":   true,
	"grpcs": true,
	"wss":   true,
}

// updateServicesFor314 updates all services in the content to explicitly set
// tls_verify=false when the service uses a secure protocol and tls_verify is
// not already set. In Kong Gateway 3.14, the global tls_certificate_verify
// option changed from off to on, which means services without an explicit
// tls_verify setting will now have TLS verification enforced.
func updateServicesFor314(content *file.Content) {
	for idx := range content.Services {
		setServiceTLSVerifyDefaultFor314(&content.Services[idx])
	}
}

// setServiceTLSVerifyDefaultFor314 sets tls_verify=false on a service if the
// service uses a secure protocol and tls_verify is not explicitly set.
func setServiceTLSVerifyDefaultFor314(service *file.FService) {
	if service == nil {
		return
	}
	// Only set tls_verify for services with secure protocols
	if service.Protocol != nil && secureServiceProtocols[*service.Protocol] && service.TLSVerify == nil {
		service.TLSVerify = kong.Bool(false)
		serviceName := "<unnamed>"
		if service.Name != nil {
			serviceName = *service.Name
		} else if service.ID != nil {
			serviceName = *service.ID
		}
		cprint.UpdatePrintf(
			"Service '%s': setting tls_verify to false "+
				"(old default, 3.14 enforces TLS certificate verification by default)\n", serviceName)
	}
}
