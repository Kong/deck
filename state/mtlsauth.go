package state

// MTLSAuthsCollection stores and indexes mtls-auth credentials.
type MTLSAuthsCollection struct {
	credentialsCollection
}

func newMTLSAuthsCollection(common collection) *MTLSAuthsCollection {
	return &MTLSAuthsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "mtls-auth",
		},
	}
}

// Add adds a mtls-auth credential to MTLSAuthsCollection
func (k *MTLSAuthsCollection) Add(mtlsAuth MTLSAuth) error {
	cred := (entity)(&mtlsAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a mtls-auth credential by ID.
func (k *MTLSAuthsCollection) Get(ID string) (*MTLSAuth, error) {
	cred, err := k.credentialsCollection.Get(ID)
	if err != nil {
		return nil, err
	}

	mtlsAuth, ok := cred.(*MTLSAuth)
	if !ok {
		panic(unexpectedType)
	}
	return &MTLSAuth{MTLSAuth: *mtlsAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all mtls-auth credentials
// belong to a Consumer with id.
func (k *MTLSAuthsCollection) GetAllByConsumerID(id string) ([]*MTLSAuth,
	error,
) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*MTLSAuth
	for _, cred := range creds {
		r, ok := cred.(*MTLSAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &MTLSAuth{MTLSAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing mtls-auth credential.
func (k *MTLSAuthsCollection) Update(mtlsAuth MTLSAuth) error {
	cred := (entity)(&mtlsAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a mtls-auth credential by ID.
func (k *MTLSAuthsCollection) Delete(ID string) error {
	return k.credentialsCollection.Delete(ID)
}

// GetAll gets all mtls-auth credentials.
func (k *MTLSAuthsCollection) GetAll() ([]*MTLSAuth, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*MTLSAuth
	for _, cred := range creds {
		r, ok := cred.(*MTLSAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &MTLSAuth{MTLSAuth: *r.DeepCopy()})
	}
	return res, nil
}
