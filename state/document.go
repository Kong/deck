package state

import (
	"fmt"

	"github.com/hashicorp/go-memdb"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	documentTableName = "document"
	documentsByParent = "documentsByParent"
)

var (
	errDocumentMissingParent = fmt.Errorf("Document has no Parent")
	errDocumentPathRequired  = fmt.Errorf("Document must have a Path")
)

// DocumentsCollection stores and indexes key-auth credentials.
type DocumentsCollection collection

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
		documentsByParent: {
			Name: documentsByParent,
			Indexer: &indexers.MethodIndexer{
				Method: "ParentKey",
			},
		},
	},
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

	if document.Parent == nil {
		return errDocumentMissingParent
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *document.ID)
	searchBy = append(searchBy, *document.Path)
	_, err := getDocument(txn, document.ParentKey(), searchBy...)
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

func getDocument(txn *memdb.Txn, parentKey string, IDs ...string) (*Document, error) {
	if parentKey == "" {
		return nil, fmt.Errorf("parentKey is required")
	}
	documents, err := getAllDocsByParentKey(txn, parentKey)
	if err != nil {
		return nil, err
	}

	for _, id := range IDs {
		for _, document := range documents {
			if id == *document.ID || id == *document.Path {
				return &Document{Document: *document.ShallowCopy()}, nil
			}
		}
	}
	return nil, ErrNotFound
}

func getAllDocsByParentKey(txn *memdb.Txn, parentKey string) ([]*Document, error) {
	iter, err := txn.Get(documentTableName, documentsByParent, parentKey)
	if err != nil {
		return nil, err
	}

	var documents []*Document
	for el := iter.Next(); el != nil; el = iter.Next() {
		d, ok := el.(*Document)
		if !ok {
			panic(unexpectedType)
		}
		documents = append(documents, &Document{Document: *d.ShallowCopy()})
	}
	return documents, nil
}

// Update updates a Document
func (k *DocumentsCollection) Update(document Document) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(document.ID) {
		return errIDRequired
	}

	if document.Parent == nil {
		return errDocumentMissingParent
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDocument(txn, document.ParentKey(), documentsByParent, *document.ID)
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

func deleteDocument(txn *memdb.Txn, key string, index string, pathOrID string) error {
	document, err := getDocument(txn, key, index, pathOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(documentTableName, document)
	if err != nil {
		return err
	}
	return nil
}

// DeleteByParent deletes a Document by parent and path or ID.
func (k *DocumentsCollection) DeleteByParent(parent konnect.ParentInfoer, pathOrID string) error {
	if pathOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDocument(txn, parent.Key(), documentsByParent, pathOrID)
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
		d, ok := el.(*Document)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Document{Document: *d.ShallowCopy()})
	}
	txn.Commit()
	return res, nil
}

// GetAllByParent returns all documents for a Parent
func (k *DocumentsCollection) GetAllByParent(parent konnect.ParentInfoer) ([]*Document, error) {
	if parent == nil {
		return make([]*Document, 0), errDocumentMissingParent
	}
	txn := k.db.Txn(false)
	return getAllDocsByParentKey(txn, parent.Key())
}

// GetByParent returns a document attached to a Parent with a given path or ID
func (k *DocumentsCollection) GetByParent(parent konnect.ParentInfoer, pathOrID string) (*Document, error) {
	if parent == nil {
		return nil, errDocumentMissingParent
	}
	txn := k.db.Txn(false)
	document, err := getDocument(txn, parent.Key(), documentsByParent, pathOrID)
	if err != nil {
		return nil, err
	}
	return document, nil
}
