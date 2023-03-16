package state

import (
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func snisCollection() *SNIsCollection {
	return state().SNIs
}

func TestSNIsCollection_Add(t *testing.T) {
	type args struct {
		sni SNI
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						Name: kong.String("foo"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts without a name",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID: kong.String("id1"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "inserts with a name and ID",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID:   kong.String("id2"),
						Name: kong.String("bar-name"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert when name is present",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID:   kong.String("id4"),
						Name: kong.String("foo-name"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID:   kong.String("id3"),
						Name: kong.String("foobar-name"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID:   kong.String("id3"),
						Name: kong.String("foobar-name"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
	}
	k := snisCollection()
	sni1 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("id3"),
			Name: kong.String("foo-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	k.Add(sni1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.sni); (err != nil) != tt.wantErr {
				t.Errorf("SNIsCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSNIsCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	sni1 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	sni2 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *SNI
		wantErr bool
	}{
		{
			name: "gets a sni by ID",
			args: args{
				nameOrID: "foo-id",
			},
			want:    &sni1,
			wantErr: false,
		},
		{
			name: "gets a sni by Name",
			args: args{
				nameOrID: "bar-name",
			},
			want:    &sni2,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no sni found",
			args: args{
				nameOrID: "baz-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an error when ID is empty",
			args: args{
				nameOrID: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	k := snisCollection()
	k.Add(sni1)
	k.Add(sni2)
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tc.args.nameOrID)
			if (err != nil) != tc.wantErr {
				t.Errorf("SNIsCollection.Get() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SNIsCollection.Get() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSNIsInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection := snisCollection()

	type derivedSNI struct {
		SNI
	}

	var sni derivedSNI
	sni.SNI = SNI{
		SNI: kong.SNI{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	txn := collection.db.Txn(true)
	txn.Insert(sniTableName, &sni)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("foo-id")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestSNIsCollection_Update(t *testing.T) {
	sni1 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	sni2 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	sni3 := SNI{
		SNI: kong.SNI{
			ID:   kong.String("foo-id"),
			Name: kong.String("name"),
			Certificate: &kong.Certificate{
				ID: kong.String("cert1-id"),
			},
		},
	}
	type args struct {
		sni SNI
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		updatedSNI *SNI
	}{
		{
			name: "update errors if sni.ID is nil",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						Name: kong.String("name"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if sni does not exist",
			args: args{
				sni: SNI{
					SNI: kong.SNI{
						ID: kong.String("does-not-exist"),
						Certificate: &kong.Certificate{
							ID: kong.String("cert1-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				sni: sni3,
			},
			wantErr:    false,
			updatedSNI: &sni3,
		},
	}
	k := snisCollection()
	k.Add(sni1)
	k.Add(sni2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.sni); (err != nil) != tt.wantErr {
				t.Errorf("SNIsCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedSNI.ID)

				if !reflect.DeepEqual(got, tt.updatedSNI) {
					t.Errorf("update sni, got = %#v, want %#v", got, tt.updatedSNI)
				}
			}
		})
	}
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestSNIGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := snisCollection()

	var sni SNI
	sni.Name = kong.String("my-sni")
	sni.ID = kong.String("first")
	sni.Certificate = &kong.Certificate{
		ID: kong.String("cert1-id"),
	}
	err := collection.Add(sni)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-sni", *re.Name)

	re, err = collection.Get("my-sni")
	assert.Nil(err)
	assert.NotNil(re)
}

func TestSNIDelete(t *testing.T) {
	assert := assert.New(t)
	collection := snisCollection()

	var sni SNI
	sni.Name = kong.String("my-sni")
	sni.ID = kong.String("first")
	sni.Certificate = &kong.Certificate{
		ID: kong.String("cert1-id"),
	}
	err := collection.Add(sni)
	assert.Nil(err)

	re, err := collection.Get("my-sni")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("first", *re.ID)

	err = collection.Delete(*re.ID)
	assert.Nil(err)

	err = collection.Delete(*re.ID)
	assert.NotNil(err)
}

func TestSNIGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := snisCollection()

	var sni SNI
	sni.Name = kong.String("my-sni1")
	sni.ID = kong.String("first")
	sni.Certificate = &kong.Certificate{
		ID: kong.String("cert1-id"),
	}
	err := collection.Add(sni)
	assert.Nil(err)

	var sni2 SNI
	sni2.Name = kong.String("my-sni2")
	sni2.ID = kong.String("second")
	sni2.Certificate = &kong.Certificate{
		ID: kong.String("cert1-id"),
	}
	err = collection.Add(sni2)
	assert.Nil(err)

	snis, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(snis))
}

func TestSNIGetAllByServiceID(t *testing.T) {
	assert := assert.New(t)
	collection := snisCollection()

	snis := []*SNI{
		{
			SNI: kong.SNI{
				ID:   kong.String("sni1-id"),
				Name: kong.String("sni1-name"),
				Certificate: &kong.Certificate{
					ID: kong.String("cert1-id"),
				},
			},
		},
		{
			SNI: kong.SNI{
				ID: kong.String("sni2-id"),
				Certificate: &kong.Certificate{
					ID: kong.String("cert1-id"),
				},
			},
		},
		{
			SNI: kong.SNI{
				ID:   kong.String("sni3-id"),
				Name: kong.String("sni3-name"),
				Certificate: &kong.Certificate{
					ID: kong.String("cert2-id"),
				},
			},
		},
		{
			SNI: kong.SNI{
				ID:   kong.String("sni4-id"),
				Name: kong.String("sni4-name"),
				Certificate: &kong.Certificate{
					ID: kong.String("cert2-id"),
				},
			},
		},
		{
			SNI: kong.SNI{
				ID: kong.String("sni5-id"),
				Certificate: &kong.Certificate{
					ID: kong.String("cert2-id"),
				},
			},
		},
	}

	for _, sni := range snis {
		err := collection.Add(*sni)
		assert.Nil(err)
	}

	snis, err := collection.GetAllByCertID("cert1-id")
	assert.Nil(err)
	assert.Equal(2, len(snis))

	snis, err = collection.GetAllByCertID("cert2-id")
	assert.Nil(err)
	assert.Equal(3, len(snis))
}
