package state

import (
	"reflect"
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func documentCollection() *DocumentsCollection {
	return state().Documents
}

func TestDocumentCollection_Add(t *testing.T) {
	type args struct {
		document Document
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				document: Document{
					Document: konnect.Document{
						Path: kong.String("foo"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors without a path",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID: kong.String("id1"),
						Parent: &konnect.ServiceVersion{
							ID: kong.String("abc"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors without a parent",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID:   kong.String("id1"),
						Path: kong.String("foo"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "works with ServiceVersion parent",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID:   kong.String("id2"),
						Path: kong.String("bar"),
						Parent: &konnect.ServiceVersion{
							ID:      kong.String("whatever"),
							Version: kong.String("abc"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "works with ServicePackage parent",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID:   kong.String("id3"),
						Path: kong.String("bar"),
						Parent: &konnect.ServicePackage{
							ID:   kong.String("whatever"),
							Name: kong.String("abc"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID:   kong.String("id4"),
						Path: kong.String("abc"),
						Parent: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: true,
		},
	}
	k := documentCollection()
	d1 := Document{
		Document: konnect.Document{
			ID:   kong.String("id4"),
			Path: kong.String("abc"),
			Parent: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	k.Add(d1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.document); (err != nil) != tt.wantErr {
				t.Errorf("DocumentCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocumentCollection_GetByParent(t *testing.T) {
	type args struct {
		pathOrID string
		parent   konnect.ParentInfoer
	}
	d1 := Document{
		Document: konnect.Document{
			ID:   kong.String("foo-id"),
			Path: kong.String("path"),
			Parent: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	d2 := Document{
		Document: konnect.Document{
			ID:   kong.String("bar-id"),
			Path: kong.String("path"),
			Parent: &konnect.ServiceVersion{
				ID: kong.String("id2"),
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *Document
		wantErr bool
	}{
		{
			name: "gets a document by parent and ID",
			args: args{
				pathOrID: "foo-id",
				parent: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
			want:    &d1,
			wantErr: false,
		},
		{
			name: "gets a document by parent and path",
			args: args{
				pathOrID: "path",
				parent: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
			want:    &d1,
			wantErr: false,
		},
		{
			name: "returns an error when parent missing",
			args: args{
				pathOrID: "bar-name",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an ErrNotFound when no document found",
			args: args{
				pathOrID: "baz-id",
				parent: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an error when ID is empty",
			args: args{
				pathOrID: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	k := documentCollection()
	k.Add(d1)
	k.Add(d2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.GetByParent(tt.args.parent, tt.args.pathOrID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocumentCollection.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocumentCollection.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocumentCollection_Update(t *testing.T) {
	d1 := Document{
		Document: konnect.Document{
			ID:   kong.String("foo-id"),
			Path: kong.String("foo-path"),
			Parent: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	d2 := Document{
		Document: konnect.Document{
			ID:   kong.String("bar-id"),
			Path: kong.String("bar-path"),
			Parent: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	d3 := Document{
		Document: konnect.Document{
			ID:   kong.String("foo-id"),
			Path: kong.String("new-foo-path"),
			Parent: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	type args struct {
		document Document
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		updatedDocument *Document
	}{
		{
			name: "update errors if document.ID is nil",
			args: args{
				document: Document{
					Document: konnect.Document{
						Path: kong.String("name"),
						Parent: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if document does not exist",
			args: args{
				document: Document{
					Document: konnect.Document{
						ID: kong.String("does-not-exist"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				document: d3,
			},
			wantErr:         false,
			updatedDocument: &d3,
		},
	}
	k := documentCollection()
	k.Add(d1)
	k.Add(d2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.document); (err != nil) != tt.wantErr {
				t.Errorf("DocumentCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.GetByParent(tt.updatedDocument.Parent, *tt.updatedDocument.ID)

				if !reflect.DeepEqual(got, tt.updatedDocument) {
					t.Errorf("update document, got = %#v, want %#v", got, tt.updatedDocument)
				}
			}
		})
	}
}

func TestDocumentDeleteByParent(t *testing.T) {
	assert := assert.New(t)
	collection := documentCollection()

	var document Document
	document.Path = kong.String("my-document")
	document.ID = kong.String("first")
	document.Parent = &konnect.ServicePackage{
		ID: kong.String("package-id1"),
	}
	err := collection.Add(document)
	assert.Nil(err)

	re, err := collection.GetByParent(document.Parent, "my-document")
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.DeleteByParent(document.Parent, *re.ID)
	assert.Nil(err)

	err = collection.DeleteByParent(document.Parent, *re.ID)
	assert.NotNil(err)
}

func TestDocumentGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := documentCollection()

	var d1 Document
	d1.Path = kong.String("my-d1")
	d1.ID = kong.String("first")
	d1.Parent = &konnect.ServicePackage{
		ID: kong.String("id1"),
	}
	err := collection.Add(d1)
	assert.Nil(err)

	var d2 Document
	d2.Path = kong.String("my-d2")
	d2.ID = kong.String("second")
	d2.Parent = &konnect.ServicePackage{
		ID: kong.String("id1"),
	}
	err = collection.Add(d2)
	assert.Nil(err)

	documents, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(documents))
}

func TestDocumentGetAllByParent(t *testing.T) {
	assert := assert.New(t)
	collection := documentCollection()

	documents := []*Document{
		{
			Document: konnect.Document{
				ID:   kong.String("d1-id"),
				Path: kong.String("d1-path"),
				Parent: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
		},
		{
			Document: konnect.Document{
				ID:   kong.String("d2-id"),
				Path: kong.String("d2-path"),
				Parent: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
		},
		{
			Document: konnect.Document{
				ID:   kong.String("d3-id"),
				Path: kong.String("d3-path"),
				Parent: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
		{
			Document: konnect.Document{
				ID:   kong.String("d4-id"),
				Path: kong.String("d4-path"),
				Parent: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
		{
			Document: konnect.Document{
				ID:   kong.String("d5-id"),
				Path: kong.String("d5-path"),
				Parent: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
	}

	for _, document := range documents {
		err := collection.Add(*document)
		assert.Nil(err)
	}

	documents, err := collection.GetAllByParent(&konnect.ServicePackage{ID: kong.String("id1")})
	assert.Nil(err)
	assert.Equal(2, len(documents))

	documents, err = collection.GetAllByParent(&konnect.ServicePackage{ID: kong.String("id2")})
	assert.Nil(err)
	assert.Equal(3, len(documents))
}
