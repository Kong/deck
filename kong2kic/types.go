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

func (k KICContent) marshalKICContentToFormat(format string) ([]byte, error) {
	var output []byte

	const (
		yamlSeparator = "---\n"
	)

	for _, kongIngress := range k.KongIngresses {
		kongIngresses, err := SerializeObjectDroppingFields(kongIngress, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongIngresses...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, kongPlugin := range k.KongPlugins {
		kongPlugins, err := SerializeObjectDroppingFields(kongPlugin, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongPlugins...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, kongClusterPlugin := range k.KongClusterPlugins {
		kongClusterPlugins, err := SerializeObjectDroppingFields(kongClusterPlugin, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongClusterPlugins...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, ingress := range k.Ingresses {
		ingresses, err := SerializeObjectDroppingFields(ingress, format)
		if err != nil {
			return nil, err
		}
		output = append(output, ingresses...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, httpRoute := range k.HTTPRoutes {
		httpRoutes, err := SerializeObjectDroppingFields(httpRoute, format)
		if err != nil {
			return nil, err
		}
		output = append(output, httpRoutes...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, kongUpstreamPolicy := range k.KongUpstreamPolicies {
		kongUpstreamPolicies, err := SerializeObjectDroppingFields(kongUpstreamPolicy, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongUpstreamPolicies...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, service := range k.Services {
		services, err := SerializeObjectDroppingFields(service, format)
		if err != nil {
			return nil, err
		}
		output = append(output, services...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, secret := range k.Secrets {
		secrets, err := SerializeObjectDroppingFields(secret, format)
		if err != nil {
			return nil, err
		}
		output = append(output, secrets...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, kongConsumer := range k.KongConsumers {
		kongConsumers, err := SerializeObjectDroppingFields(kongConsumer, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumers...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
	}

	for _, kongConsumerGroup := range k.KongConsumerGroups {
		kongConsumerGroups, err := SerializeObjectDroppingFields(kongConsumerGroup, format)
		if err != nil {
			return nil, err
		}
		output = append(output, kongConsumerGroups...)
		if format == file.YAML {
			output = append(output, []byte(yamlSeparator)...)
		}
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
