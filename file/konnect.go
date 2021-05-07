package file

import (
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

// PopulateDocumentContent updates the Documents contained within a Content with the
// contents of their files on disk. Document files are stored at
// <root>/ServicePackage.Name/Document.Path and <root>/ServicePackage.Name/ServiceVersion.Version/Document.Path,
// where <root> is the directory containing the first state file.
func (c Content) PopulateDocumentContent(filenames []string) error {
	if len(filenames) == 0 {
		return errors.New("cannot populate documents without a location")
	}
	// TODO decK actually allows you to use _multiple_ state files
	// How should we choose which to use as the search path for documents?
	root := filepath.Dir(filenames[0])
	for _, sp := range c.ServicePackages {
		spPath := utils.NameToFilename(*sp.Name)
		path := filepath.Join(root, spPath, utils.NameToFilename(*sp.Document.Path))
		content, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "error reading document file")
		}
		sp.Document.Content = kong.String(string(content))
		for _, sv := range sp.Versions {
			path := filepath.Join(root, spPath, utils.NameToFilename(*sv.Version),
				utils.NameToFilename(*sv.Document.Path))
			content, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, "error reading document file")
			}
			sv.Document.Content = kong.String(string(content))
		}
	}
	return nil
}
