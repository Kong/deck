
# Service
A Kong Service maps to a Kubernetes Service.

The generated k8s service is independent of KIC version or GW/INGRESS API.

For each Kong service, it creates a new Kubernetes service. 
If the service has a name, it is converted into a slug format 
using the calculateSlug function and assigned to the Kubernetes service's name. 
If the service doesn't have a name, a warning message is logged and it is skipped.

The protocol defaults to TCP unless the service's protocol is explicitly set to UDP.

If the service has a specified port, a ServicePort object is created with the protocol 
and port information, and this object is appended to the Kubernetes service's ports. 
The service name, is used as a selector for the Kubernetes service.

Various properties of the service (like ReadTimeout, WriteTimeout, ConnectTimeout, 
Protocol, Path, and Retries) are added as annotations to the Kubernetes service.

If the service host is informed then search for an Upstream which name matches.
If there is no upstream, then configure an external service.

# Route
A Kong Route maps to a HTTPRoute (Gateway API) or Ingress (Ingress API)

## Ingress API
The generated k8s Ingress is independent of KIC version.

If the service and route names are not nil, they are used to generate a slug 
that is used as the name of the Ingress resource. If either the service or 
route name is nil, a log message is printed and the loop continues to the next iteration.

It then checks various properties of the route (like Protocols, StripPath, PreserveHost, etc.) 
and if they are not nil, it adds corresponding annotations to the k8sIngress.

If the route doesn't have hosts defined, it creates an IngressRule 
for each path in the route. If the service has a port, it is included in the 
IngressRule. If the route does have hosts, it creates an IngressRule for each 
host and for each path in the route. Again, if the service has a port, 
it is included in the IngressRule.

## Gateway API
The generated k8s HTTPRoute is different for KIC v2 vs v3. The only difference
is apiVersion: gateway.networking.k8s.io/v1beta1 vs apiVersion: gateway.networking.k8s.io/v1.


If the service and route names are not nil, they are used to generate a slug 
that is used as the name of the Ingress resource. If either the service or 
route name is nil, a log message is printed and the loop continues to the next iteration.

It then checks various properties of the route (like Protocols, StripPath, PreserveHost, etc.) 
and if they are not nil, it adds corresponding annotations.

The function then checks if route.Hosts is not nil and if it's not, 
it appends each host to httpRoute.Spec.Hostnames. 
It also appends a ParentReference with the name "kong" to httpRoute.Spec.ParentRefs.

The function then creates a BackendRef with the service name and port and assigns it to backendRef. 
It also creates an HTTPHeaderMatch slice and populates it with header matches if route.Headers is not nil.

Next, the function checks if route.Paths is not nil and if it's not, 
it creates an HTTPRouteRule for each path and each method in the route and 
appends it to httpRoute.Spec.Rules. If route.Paths is nil, it creates an 
HTTPRouteRule for each method in the route with no path and appends it to httpRoute.Spec.Rules

# Plugin
## Service and Route plugins
A Kong Plugin maps to a KongPlugin in K8s.

Plugins do not depend on KIC v2 vs v3.

The plugin configuration is copied verbatim and an annotation
is added to the entity to which the plugin is applied (Service, Ingress, Consumer, ConsumerGroup)

HTTPRoute is a special case in which the plugin is referenced 
as an extensionRef and not an annotation.

## Global Plugins
A Kong Global Plugin maps to a KongClusterPlugin.
The plugin configuration is copied verbatim.

## Consumer Group Plugin
A Kong Consumer Group plugin maps to a KongPlugin.
It's treated differently because the Kong Consumer Group plugin fields are different.

# Consumer
A Consumer maps to a KongConsumer object.

Its credentials are mapped to Secrets in K8s.

# Consumer Group
A Consumer Group is mapped to a KongConsumerGroup.

# Certificate
A certificate is mapped to a Secret of type "kubernetes.io/tls".

# CA Certificate
A CA Certificate maps to an Opaque Secret.

# Upstream
An Upstream maps to a KongIngress for KIC v2 and to a KongUpstreamPolicy for KIC v3.
