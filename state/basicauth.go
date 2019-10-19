package state

// BasicAuthsCollection stores and indexes basic-auth credentials.
type BasicAuthsCollection struct {
	credentialsCollection
}

func newBasicAuthsCollection(common collection) *BasicAuthsCollection {
	return &BasicAuthsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "basic-auth",
		},
	}
}

// Add adds a basic-auth credential to BasicAuthsCollection
func (k *BasicAuthsCollection) Add(basicAuth BasicAuth) error {
	cred := (entity)(&basicAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a basic-auth credential by key or ID.
func (k *BasicAuthsCollection) Get(keyOrID string) (*BasicAuth, error) {
	cred, err := k.credentialsCollection.Get(keyOrID)
	if err != nil {
		return nil, err
	}

	basicAuth, ok := cred.(*BasicAuth)
	if !ok {
		panic(unexpectedType)
	}
	return &BasicAuth{BasicAuth: *basicAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all basic-auth credentials
// belong to a Consumer with id.
func (k *BasicAuthsCollection) GetAllByConsumerID(id string) ([]*BasicAuth,
	error) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*BasicAuth
	for _, cred := range creds {
		r, ok := cred.(*BasicAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &BasicAuth{BasicAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing basic-auth credential.
func (k *BasicAuthsCollection) Update(basicAuth BasicAuth) error {
	cred := (entity)(&basicAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a basic-auth credential by key or ID.
func (k *BasicAuthsCollection) Delete(keyOrID string) error {
	return k.credentialsCollection.Delete(keyOrID)
}

// GetAll gets all basic-auth credentials.
func (k *BasicAuthsCollection) GetAll() ([]*BasicAuth, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*BasicAuth
	for _, cred := range creds {
		r, ok := cred.(*BasicAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &BasicAuth{BasicAuth: *r.DeepCopy()})
	}
	return res, nil
}
