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
	// We currently choose the first arbitrarily and assume document content is under its directory
	// Future plans are to rework the multiple state file functionality to require all state files
	// be in the same directory.
	root := filepath.Dir(filenames[0])
	for _, sp := range c.ServicePackages {
		if sp.Document != nil {
			path := filepath.Join(root, utils.FilenameToName(*sp.Document.Path))
			content, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, "error reading document file")
			}
			sp.Document.Content = kong.String(string(content))
		}
		for _, sv := range sp.Versions {
			if sv.Document != nil {
				path := filepath.Join(root, utils.FilenameToName(*sv.Document.Path))
				content, err := os.ReadFile(path)
				if err != nil {
					return errors.Wrap(err, "error reading document file")
				}
				sv.Document.Content = kong.String(string(content))
			}
		}
	}
	return nil
}

// StripLocalDocum
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
