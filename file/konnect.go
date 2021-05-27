package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// PopulateDocumentContent updates the Documents contained within a Content with the
// contents of their files on disk. Document files are stored at
// <root>/ServicePackage.Name/Document.Path and <root>/ServicePackage.Name/ServiceVersion.Version/Document.Path,
// where <root> is the directory containing the first state file.
func (c Content) PopulateDocumentContent(filenames []string) error {
	if len(filenames) == 0 {
		return fmt.Errorf("cannot populate documents without a location")
	}
	// TODO decK actually allows you to use _multiple_ state files
	// We currently choose the first arbitrarily and assume document content is under its directory
	// Future plans are to rework the multiple state file functionality to require all state files
	// be in the same directory.
	root := filepath.Dir(filenames[0])
	for _, sp := range c.ServicePackages {
		if sp.Document != nil {
			path := filepath.Join(root, utils.FilenameToName(*sp.Document.Path))
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading document file: %w", err)
			}
			sp.Document.Content = kong.String(string(content))
		}
		for _, sv := range sp.Versions {
			if sv.Document != nil {
				path := filepath.Join(root, utils.FilenameToName(*sv.Document.Path))
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("error reading document file: %w", err)
				}
				sv.Document.Content = kong.String(string(content))
			}
		}
	}
	return nil
}

// StripLocalDocumentPath removes local path information from a target state document, returning the base path with a
// prepended slash. These path values match typical path values for documents created in the Konnect GUI, whereas path
// values in decK state files are local relative paths with service package and service version directories.
func (c Content) StripLocalDocumentPath() {
	for _, sp := range c.ServicePackages {
		if sp.Document != nil {
			trunc := "/" + filepath.Base(utils.FilenameToName(*sp.Document.Path))
			sp.Document.Path = &trunc
		}
		for _, sv := range sp.Versions {
			if sv.Document != nil {
				trunc := "/" + filepath.Base(utils.FilenameToName(*sv.Document.Path))
				sv.Document.Path = &trunc
			}
		}
	}
}
