package kong2kic

import (
	"encoding/json"

	kicv1 "github.com/kong/kubernetes-ingress-controller/v2/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v2/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	k8sgwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"
)

// KICContent represents a serialized Kong state for KIC.
// +k8s:deepcopy-gen=true
type KICContent struct {
	KongIngresses      []kicv1.KongIngress            `json:"kongIngresses,omitempty" yaml:",omitempty"`
	KongPlugins        []kicv1.KongPlugin             `json:"kongPlugins,omitempty" yaml:",omitempty"`
	KongClusterPlugins []kicv1.KongClusterPlugin      `json:"clusterPlugins,omitempty" yaml:",omitempty"`
	Ingresses          []k8snetv1.Ingress             `json:"ingresses,omitempty" yaml:",omitempty"`
	Services           []k8scorev1.Service            `json:"services,omitempty" yaml:",omitempty"`
	Secrets            []k8scorev1.Secret             `json:"secrets,omitempty" yaml:",omitempty"`
	KongConsumers      []kicv1.KongConsumer           `json:"consumers,omitempty" yaml:",omitempty"`
	KongConsumerGroups []kicv1beta1.KongConsumerGroup `json:"consumerGroups,omitempty" yaml:",omitempty"`
	HTTPRoutes         []k8sgwapiv1.HTTPRoute         `json:"httpRoutes,omitempty" yaml:",omitempty"`
}

func (k KICContent) marshalKICContentToYaml() ([]byte, error) {
	var kongIngresses []byte
	var kongPlugins []byte
	var kongClusterPlugins []byte
	var ingresses []byte
	var services []byte
	var secrets []byte
	var kongConsumers []byte
	var kongConsumerGroups []byte
	var err error
	var output []byte

	// iterate over the slices of kongIngresses, kongPlugins,
	// kongClusterPlugins, ingresses, services, secrets, kongConsumers
	// and marshal each one in yaml format
	// and append it to the output slice
	// then return the output slice
	for _, kongIngress := range k.KongIngresses {
		kongIngresses, err = yaml.Marshal(kongIngress)
		if err != nil {
			return nil, err
		}
		output = append(output, kongIngresses...)
		output = append(output, []byte("---\n")...)
	}

	for _, kongPlugin := range k.KongPlugins {
		kongPlugins, err = yaml.Marshal(kongPlugin)
		if err != nil {
			return nil, err
		}
		output = append(output, kongPlugins...)
		output = append(output, []byte("---\n")...)

	}

	for _, kongClusterPlugin := range k.KongClusterPlugins {
		kongClusterPlugins, err = yaml.Marshal(kongClusterPlugin)
		if err != nil {
			return nil, err
		}
		output = append(output, kongClusterPlugins...)
		output = append(output, []byte("---\n")...)

	}

	for _, ingress := range k.Ingresses {
		ingresses, err = yaml.Marshal(ingress)
		if err != nil {
			return nil, err
		}
		output = append(output, ingresses...)
		output = append(output, []byte("---\n")...)

	}

	for _, httpRoute := range k.HTTPRoutes {
		httpRoutes, err := yaml.Marshal(httpRoute)
		if err != nil {
			return nil, err
		}
		output = append(output, httpRoutes...)
		output = append(output, []byte("---\n")...)
	}

	for _, service := range k.Services {
		services, err = yaml.Marshal(service)
		if err != nil {
			return nil, err
		}
		output = append(output, services...)
		output = append(output, []byte("---\n")...)

	}

	for _, secret := range k.Secrets {
		secrets, err = yaml.Marshal(secret)
		if err != nil {
			return nil, err
		}
		output = append(output, secrets...)
		output = append(output, []byte("---\n")...)

	}

	for _, kongConsumer := range k.KongConsumers {
		kongConsumers, err = yaml.Marshal(kongConsumer)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumers...)
		output = append(output, []byte("---\n")...)

	}

	for _, kongConsumerGroup := range k.KongConsumerGroups {
		kongConsumerGroups, err = yaml.Marshal(kongConsumerGroup)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumerGroups...)
		output = append(output, []byte("---\n")...)
	}

	return output, nil
}

func (k KICContent) marshalKICContentToJSON() ([]byte, error) {
	var kongIngresses []byte
	var kongPlugins []byte
	var kongClusterPlugins []byte
	var ingresses []byte
	var services []byte
	var secrets []byte
	var kongConsumers []byte
	var kongConsumerGroups []byte
	var err error
	var output []byte

	// iterate over the slices of kongIngresses, kongPlugins,
	// kongClusterPlugins, ingresses, services, secrets, kongConsumers
	// and marshal each one in json format
	// and append it to the output slice
	// then return the output slice
	for _, kongIngress := range k.KongIngresses {
		kongIngresses, err = json.MarshalIndent(kongIngress, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, kongIngresses...)
	}

	for _, kongPlugin := range k.KongPlugins {
		kongPlugins, err = json.MarshalIndent(kongPlugin, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, kongPlugins...)
	}

	for _, kongClusterPlugin := range k.KongClusterPlugins {
		kongClusterPlugins, err = json.MarshalIndent(kongClusterPlugin, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, kongClusterPlugins...)
	}

	for _, ingress := range k.Ingresses {
		ingresses, err = json.MarshalIndent(ingress, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, ingresses...)
	}

	for _, httpRoute := range k.HTTPRoutes {
		httpRoutes, err := json.MarshalIndent(httpRoute, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, httpRoutes...)
	}

	for _, service := range k.Services {
		services, err = json.MarshalIndent(service, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, services...)
	}

	for _, secret := range k.Secrets {
		secrets, err = json.MarshalIndent(secret, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, secrets...)
	}

	for _, kongConsumer := range k.KongConsumers {
		kongConsumers, err = json.MarshalIndent(kongConsumer, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumers...)
	}

	for _, kongConsumerGroup := range k.KongConsumerGroups {
		kongConsumerGroups, err = json.MarshalIndent(kongConsumerGroup, "", "    ")
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumerGroups...)
	}

	return output, nil
}
