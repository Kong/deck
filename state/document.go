package state

import (
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

const (
	documentTableName           = "document"
	documentsByServicePackageID = "documentsByServicePackageID"
	documentsByServiceVersionID = "documentsByServiceVersionID"
)

var errDocumentBothForeign = errors.New("Document has both ServicePackage and ServiceVersion set")
var errDocumentInvalidSV = errors.New("ServiceVersion attached to Document has no ID")
var errDocumentInvalidSP = errors.New("ServicePackage attached to Document has no ID")
var errDocumentPathRequired = errors.New("Document must have a Path")

// DocumentsCollection stores and indexes key-auth credentials.
type DocumentsCollection collection

type documentSearcher func(*memdb.Txn, string) ([]*Document, error)

var documentTableSchema = &memdb.TableSchema{
	Name: documentTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		all: allIndex,
		// foreign
		documentsByServicePackageID: {
			Name:         documentsByServicePackageID,
			AllowMissing: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ServicePackage",
						Sub:    "ID",
					},
				},
			},
		},
		// foreign
		documentsByServiceVersionID: {
			Name:         documentsByServiceVersionID,
			AllowMissing: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ServiceVersion",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

func getDocumentForeign(document Document) (string, documentSearcher, error) {
	if document.ServicePackage != nil && document.ServiceVersion != nil {
		return "", nil, errDocumentBothForeign
	}
	if document.ServiceVersion != nil {
		if utils.Empty(document.ServiceVersion.ID) {
			return "", nil, errDocumentInvalidSV
		}
		return *document.ServiceVersion.ID, getAllDocsByVersionID, nil
	}
	if utils.Empty(document.ServicePackage.ID) {
		return "", nil, errDocumentInvalidSP
	}
	return *document.ServicePackage.ID, getAllDocsByPackageID, nil
}

// Add adds a document into DocumentsCollection
// document.ID should not be nil else an error is thrown.
func (k *DocumentsCollection) Add(document Document) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(document.ID) {
		return errIDRequired
	}

	if utils.Empty(document.Path) {
		return errDocumentPathRequired
	}

	foreignID, searcher, err := getDocumentForeign(document)
	if err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *document.ID)
	searchBy = append(searchBy, *document.Path)
	_, err = getDocument(txn, foreignID, searcher, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting document %v: %w", document.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(documentTableName, &document)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getDocument(txn *memdb.Txn, foreignID string, search documentSearcher, IDs ...string) (*Document, error) {
	if foreignID == "" {
		return nil, errors.New("foreignID is required")
	}
	documents, err := search(txn, foreignID)
	if err != nil {
		return nil, err
	}

	for _, id := range IDs {
		for _, document := range documents {
			if id == *document.ID || id == *document.Path {
				return &Document{Document: *document.DeepCopy()}, nil
			}
		}
	}
	return nil, ErrNotFound
}

// GetByPackage gets a Service Package Document by path or ID.
func (k *DocumentsCollection) GetByPackage(packageID, pathOrID string) (*Document, error) {
	if pathOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	document, err := getDocument(txn, packageID, getAllDocsByPackageID, pathOrID)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// GetByVersion gets a Service Version Document by path or ID.
func (k *DocumentsCollection) GetByVersion(versionID, pathOrID string) (*Document, error) {
	if pathOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	document, err := getDocument(txn, versionID, getAllDocsByVersionID, pathOrID)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// Update updates a Service Version.
func (k *DocumentsCollection) Update(document Document) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(document.ID) {
		return errIDRequired
	}
	foreignID, searcher, err := getDocumentForeign(document)
	if err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err = deleteDocument(txn, foreignID, searcher, *document.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(documentTableName, &document)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteDocument(txn *memdb.Txn, foreignID string, searcher documentSearcher, pathOrID string) error {
	document, err := getDocument(txn, foreignID, searcher, pathOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(documentTableName, document)
	if err != nil {
		return err
	}
	return nil
}

// DeleteVersionDocument deletes a Service Version document by name or ID.
func (k *DocumentsCollection) DeleteVersionDocument(versionID, pathOrID string) error {
	if pathOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDocument(txn, versionID, getAllDocsByVersionID, pathOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// DeletePackageDocument deletes a Service Package document by name or ID.
func (k *DocumentsCollection) DeletePackageDocument(packageID, pathOrID string) error {
	if pathOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDocument(txn, packageID, getAllDocsByPackageID, pathOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all Documents.
func (k *DocumentsCollection) GetAll() ([]*Document, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(documentTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Document
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Document)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Document{Document: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

func getAllDocsByPackageID(txn *memdb.Txn, packageID string) ([]*Document, error) {
	iter, err := txn.Get(documentTableName, documentsByServicePackageID, packageID)
	if err != nil {
		return nil, err
	}

	var versions []*Document
	for el := iter.Next(); el != nil; el = iter.Next() {
		v, ok := el.(*Document)
		if !ok {
			panic(unexpectedType)
		}
		versions = append(versions, &Document{Document: *v.DeepCopy()})
	}
	return versions, nil
}

func getAllDocsByVersionID(txn *memdb.Txn, versionID string) ([]*Document, error) {
	iter, err := txn.Get(documentTableName, documentsByServiceVersionID, versionID)
	if err != nil {
		return nil, err
	}

	var versions []*Document
	for el := iter.Next(); el != nil; el = iter.Next() {
		v, ok := el.(*Document)
		if !ok {
			panic(unexpectedType)
		}
		versions = append(versions, &Document{Document: *v.DeepCopy()})
	}
	return versions, nil
}

// GetAllByServicePackageID returns all documents for a ServicePackage id.
func (k *DocumentsCollection) GetAllByServicePackageID(id string) ([]*Document, error) {
	txn := k.db.Txn(false)
	return getAllDocsByPackageID(txn, id)
}

// GetAllByServiceVersionID returns all documents for a ServiceVersion id.
func (k *DocumentsCollection) GetAllByServiceVersionID(id string) ([]*Document, error) {
	txn := k.db.Txn(false)
	return getAllDocsByVersionID(txn, id)
}
