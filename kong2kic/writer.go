package kong2kic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
)

func WriteContentToFile(content *file.Content, filename string, format file.Format, yamlOrJSON string) error {
	var c []byte
	var err error

	switch format {
	// case YAML:
	// 	c, err = yaml.Marshal(content)
	// 	if err != nil {
	// 		return err
	// 	}
	// case JSON:
	// 	c, err = json.MarshalIndent(content, "", "  ")
	// 	if err != nil {
	// 		return err
	// 	}
	case KICV3GATEWAY:
		if yamlOrJSON == file.YAML {
			c, err = MarshalKongToKICYaml(content, KICV3GATEWAY)
		} else {
			c, err = MarshalKongToKICJson(content, KICV3GATEWAY)
		}
		if err != nil {
			return err
		}
	case KICV3INGRESS:
		if yamlOrJSON == file.YAML {
			c, err = MarshalKongToKICYaml(content, KICV3INGRESS)
		} else {
			c, err = MarshalKongToKICJson(content, KICV3INGRESS)
		}
		if err != nil {
			return err
		}
	case KICV2GATEWAY:
		if yamlOrJSON == file.YAML {
			c, err = MarshalKongToKICYaml(content, KICV2GATEWAY)
		} else {
			c, err = MarshalKongToKICJson(content, KICV2GATEWAY)
		}
		if err != nil {
			return err
		}
	case KICV2INGRESS:
		if yamlOrJSON == file.YAML {
			c, err = MarshalKongToKICYaml(content, KICV2INGRESS)
		} else {
			c, err = MarshalKongToKICJson(content, KICV2INGRESS)
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown file format: " + string(format))
	}

	if filename == "-" {
		if _, err := fmt.Print(string(c)); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
	} else {
		filename = utils.AddExtToFilename(filename, strings.ToLower(string(format)))
		prefix, _ := filepath.Split(filename)
		if err := os.WriteFile(filename, c, 0o600); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		for _, sp := range content.ServicePackages {
			if sp.Document != nil {
				if err := os.MkdirAll(filepath.Join(prefix, filepath.Dir(*sp.Document.Path)), 0o700); err != nil {
					return fmt.Errorf("creating document directory: %w", err)
				}
				if err := os.WriteFile(filepath.Join(prefix, *sp.Document.Path),
					[]byte(*sp.Document.Content), 0o600); err != nil {
					return fmt.Errorf("writing document file: %w", err)
				}
			}
			for _, v := range sp.Versions {
				if v.Document != nil {
					if err := os.MkdirAll(filepath.Join(prefix, filepath.Dir(*v.Document.Path)), 0o700); err != nil {
						return fmt.Errorf("creating document directory: %w", err)
					}
					if err := os.WriteFile(filepath.Join(prefix, *v.Document.Path),
						[]byte(*v.Document.Content), 0o600); err != nil {
						return fmt.Errorf("writing document file: %w", err)
					}
				}
			}
		}
	}
	return nil
}
