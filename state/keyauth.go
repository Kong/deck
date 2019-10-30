package state

// KeyAuthsCollection stores and indexes key-auth credentials.
type KeyAuthsCollection struct {
	credentialsCollection
}

func newKeyAuthsCollection(common collection) *KeyAuthsCollection {
	return &KeyAuthsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "key-auth",
		},
	}
}

// Add adds a key-auth credential to KeyAuthsCollection
func (k *KeyAuthsCollection) Add(keyAuth KeyAuth) error {
	cred := (entity)(&keyAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a key-auth credential by key or ID.
func (k *KeyAuthsCollection) Get(keyOrID string) (*KeyAuth, error) {
	cred, err := k.credentialsCollection.Get(keyOrID)
	if err != nil {
		return nil, err
	}

	keyAuth, ok := cred.(*KeyAuth)
	if !ok {
		panic(unexpectedType)
	}
	return &KeyAuth{KeyAuth: *keyAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all key-auth credentials
// belong to a Consumer with id.
func (k *KeyAuthsCollection) GetAllByConsumerID(id string) ([]*KeyAuth,
	error) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*KeyAuth
	for _, cred := range creds {
		r, ok := cred.(*KeyAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &KeyAuth{KeyAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing key-auth credential.
func (k *KeyAuthsCollection) Update(keyAuth KeyAuth) error {
	cred := (entity)(&keyAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a key-auth credential by key or ID.
func (k *KeyAuthsCollection) Delete(keyOrID string) error {
	return k.credentialsCollection.Delete(keyOrID)
}

// GetAll gets all key-auth credentials.
func (k *KeyAuthsCollection) GetAll() ([]*KeyAuth, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*KeyAuth
	for _, cred := range creds {
		r, ok := cred.(*KeyAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &KeyAuth{KeyAuth: *r.DeepCopy()})
	}
	return res, nil
}
