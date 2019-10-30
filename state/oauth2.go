package state

// Oauth2CredsCollection stores and indexes oauth2 credentials.
type Oauth2CredsCollection struct {
	credentialsCollection
}

func newOauth2CredsCollection(common collection) *Oauth2CredsCollection {
	return &Oauth2CredsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "oauth2",
		},
	}
}

// Add adds a oauth2 credential to Oauth2CredsCollection
func (k *Oauth2CredsCollection) Add(keyAuth Oauth2Credential) error {
	cred := (entity)(&keyAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a oauth2 credential by key or ID.
func (k *Oauth2CredsCollection) Get(keyOrID string) (*Oauth2Credential, error) {
	cred, err := k.credentialsCollection.Get(keyOrID)
	if err != nil {
		return nil, err
	}

	keyAuth, ok := cred.(*Oauth2Credential)
	if !ok {
		panic(unexpectedType)
	}
	return &Oauth2Credential{Oauth2Credential: *keyAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all oauth2 credentials
// belong to a Consumer with id.
func (k *Oauth2CredsCollection) GetAllByConsumerID(id string) ([]*Oauth2Credential,
	error) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*Oauth2Credential
	for _, cred := range creds {
		r, ok := cred.(*Oauth2Credential)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Oauth2Credential{Oauth2Credential: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing oauth2 credential.
func (k *Oauth2CredsCollection) Update(keyAuth Oauth2Credential) error {
	cred := (entity)(&keyAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a oauth2 credential by key or ID.
func (k *Oauth2CredsCollection) Delete(keyOrID string) error {
	return k.credentialsCollection.Delete(keyOrID)
}

// GetAll gets all oauth2 credentials.
func (k *Oauth2CredsCollection) GetAll() ([]*Oauth2Credential, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*Oauth2Credential
	for _, cred := range creds {
		r, ok := cred.(*Oauth2Credential)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Oauth2Credential{Oauth2Credential: *r.DeepCopy()})
	}
	return res, nil
}
