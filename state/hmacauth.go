package state

// HMACAuthsCollection stores and indexes hmac-auth credentials.
type HMACAuthsCollection struct {
	credentialsCollection
}

func newHMACAuthsCollection(common collection) *HMACAuthsCollection {
	return &HMACAuthsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "hmac-auth",
		},
	}
}

// Add adds a hmac-auth credential to HMACAuthsCollection
func (k *HMACAuthsCollection) Add(hmacAuth HMACAuth) error {
	cred := (entity)(&hmacAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a hmac-auth credential by key or ID.
func (k *HMACAuthsCollection) Get(keyOrID string) (*HMACAuth, error) {
	cred, err := k.credentialsCollection.Get(keyOrID)
	if err != nil {
		return nil, err
	}

	hmacAuth, ok := cred.(*HMACAuth)
	if !ok {
		panic(unexpectedType)
	}
	return &HMACAuth{HMACAuth: *hmacAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all hmac-auth credentials
// belong to a Consumer with id.
func (k *HMACAuthsCollection) GetAllByConsumerID(id string) ([]*HMACAuth,
	error) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*HMACAuth
	for _, cred := range creds {
		r, ok := cred.(*HMACAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &HMACAuth{HMACAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing hmac-auth credential.
func (k *HMACAuthsCollection) Update(hmacAuth HMACAuth) error {
	cred := (entity)(&hmacAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a hmac-auth credential by key or ID.
func (k *HMACAuthsCollection) Delete(keyOrID string) error {
	return k.credentialsCollection.Delete(keyOrID)
}

// GetAll gets all hmac-auth credentials.
func (k *HMACAuthsCollection) GetAll() ([]*HMACAuth, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*HMACAuth
	for _, cred := range creds {
		r, ok := cred.(*HMACAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &HMACAuth{HMACAuth: *r.DeepCopy()})
	}
	return res, nil
}
