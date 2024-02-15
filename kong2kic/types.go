package kong2kic

import (
	"encoding/json"

	"github.com/kong/go-database-reconciler/pkg/file"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	k8sgwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"
)

// KICContent represents a serialized Kong state for KIC.
// +k8s:deepcopy-gen=true
type KICContent struct {
	KongIngresses        []kicv1.KongIngress             `json:"kongIngresses,omitempty" yaml:",omitempty"`
	KongPlugins          []kicv1.KongPlugin              `json:"kongPlugins,omitempty" yaml:",omitempty"`
	KongClusterPlugins   []kicv1.KongClusterPlugin       `json:"clusterPlugins,omitempty" yaml:",omitempty"`
	Ingresses            []k8snetv1.Ingress              `json:"ingresses,omitempty" yaml:",omitempty"`
	Services             []k8scorev1.Service             `json:"services,omitempty" yaml:",omitempty"`
	Secrets              []k8scorev1.Secret              `json:"secrets,omitempty" yaml:",omitempty"`
	KongConsumers        []kicv1.KongConsumer            `json:"consumers,omitempty" yaml:",omitempty"`
	KongConsumerGroups   []kicv1beta1.KongConsumerGroup  `json:"consumerGroups,omitempty" yaml:",omitempty"`
	HTTPRoutes           []k8sgwapiv1.HTTPRoute          `json:"httpRoutes,omitempty" yaml:",omitempty"`
	KongUpstreamPolicies []kicv1beta1.KongUpstreamPolicy `json:"upstreamPolicies,omitempty" yaml:",omitempty"`
}

func (k KICContent) marshalKICContentToYaml() ([]byte, error) {
	var output []byte

	const (
		yamlSeparator = "---\n"
	)

	// iterate over the slices of kongIngresses, kongPlugins,
	// kongClusterPlugins, ingresses, services, secrets, kongConsumers
	// and marshal each one in yaml format
	// and append it to the output slice
	// then return the output slice
	for _, kongIngress := range k.KongIngresses {
		kongIngresses, err := SerializeObjectDroppingFields(kongIngress, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongIngresses...)
		output = append(output, []byte(yamlSeparator)...)
	}

	for _, kongPlugin := range k.KongPlugins {
		kongPlugins, err := SerializeObjectDroppingFields(kongPlugin, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongPlugins...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, kongClusterPlugin := range k.KongClusterPlugins {
		kongClusterPlugins, err := SerializeObjectDroppingFields(kongClusterPlugin, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongClusterPlugins...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, ingress := range k.Ingresses {
		ingresses, err := SerializeObjectDroppingFields(ingress, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, ingresses...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, httpRoute := range k.HTTPRoutes {
		httpRoutes, err := SerializeObjectDroppingFields(httpRoute, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, httpRoutes...)
		output = append(output, []byte(yamlSeparator)...)
	}

	for _, kongUpstreamPolicy := range k.KongUpstreamPolicies {
		kongUpstreamPolicies, err := SerializeObjectDroppingFields(kongUpstreamPolicy, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongUpstreamPolicies...)
		output = append(output, []byte(yamlSeparator)...)
	}

	for _, service := range k.Services {
		services, err := SerializeObjectDroppingFields(service, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, services...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, secret := range k.Secrets {
		secrets, err := SerializeObjectDroppingFields(secret, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, secrets...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, kongConsumer := range k.KongConsumers {
		kongConsumers, err := SerializeObjectDroppingFields(kongConsumer, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumers...)
		output = append(output, []byte(yamlSeparator)...)

	}

	for _, kongConsumerGroup := range k.KongConsumerGroups {
		kongConsumerGroups, err := SerializeObjectDroppingFields(kongConsumerGroup, file.YAML)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumerGroups...)
		output = append(output, []byte(yamlSeparator)...)
	}

	return output, nil
}

func (k KICContent) marshalKICContentToJSON() ([]byte, error) {
	var output []byte

	// iterate over the slices of kongIngresses, kongPlugins,
	// kongClusterPlugins, ingresses, services, secrets, kongConsumers
	// and marshal each one in json format
	// and append it to the output slice
	// then return the output slice
	for _, kongIngress := range k.KongIngresses {
		kongIngresses, err := SerializeObjectDroppingFields(kongIngress, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongIngresses...)
	}

	for _, kongPlugin := range k.KongPlugins {
		kongPlugins, err := SerializeObjectDroppingFields(kongPlugin, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongPlugins...)
	}

	for _, kongClusterPlugin := range k.KongClusterPlugins {
		kongClusterPlugins, err := SerializeObjectDroppingFields(kongClusterPlugin, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongClusterPlugins...)
	}

	for _, ingress := range k.Ingresses {
		ingresses, err := SerializeObjectDroppingFields(ingress, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, ingresses...)
	}

	for _, httpRoute := range k.HTTPRoutes {
		httpRoutes, err := SerializeObjectDroppingFields(httpRoute, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, httpRoutes...)
	}

	for _, kongUpstreamPolicy := range k.KongUpstreamPolicies {
		kongUpstreamPolicies, err := SerializeObjectDroppingFields(kongUpstreamPolicy, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongUpstreamPolicies...)
	}

	for _, service := range k.Services {
		services, err := SerializeObjectDroppingFields(service, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, services...)
	}

	for _, secret := range k.Secrets {
		secrets, err := SerializeObjectDroppingFields(secret, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, secrets...)
	}

	for _, kongConsumer := range k.KongConsumers {
		kongConsumers, err := SerializeObjectDroppingFields(kongConsumer, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumers...)
	}

	for _, kongConsumerGroup := range k.KongConsumerGroups {
		kongConsumerGroups, err := SerializeObjectDroppingFields(kongConsumerGroup, file.JSON)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumerGroups...)
	}

	return output, nil
}

func SerializeObjectDroppingFields(obj interface{}, format string) ([]byte, error) {
	objBytes, err := json.Marshal(obj)
	result := []byte{}
	if err != nil {
		return nil, err
	}
	genericObj := map[string]interface{}{}
	if err := json.Unmarshal(objBytes, &genericObj); err != nil {
		return nil, err
	}

	// We're deleting fields that are not meant to be supplied by users.
	delete(genericObj, "status")
	delete(genericObj["metadata"].(map[string]interface{}), "creationTimestamp")

	if format == file.JSON {
		result, err = json.MarshalIndent(genericObj, "", "    ")
		if err != nil {
			return nil, err
		}
	} else if format == file.YAML {
		result, err = yaml.Marshal(genericObj)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}
