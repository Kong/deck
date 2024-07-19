package kong2tf

import "github.com/kong/go-database-reconciler/pkg/file"

type TfConfig struct {
	ControlPlaneId string
}

func Convert(inputContent *file.Content) (string, error) {
	builder := getTerraformBuilder()
	director := newDirector(builder)
	return director.builTerraformResources(inputContent), nil
}
