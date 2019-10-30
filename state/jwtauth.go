package state

// JWTAuthsCollection stores and indexes jwt-auth credentials.
type JWTAuthsCollection struct {
	credentialsCollection
}

func newJWTAuthsCollection(common collection) *JWTAuthsCollection {
	return &JWTAuthsCollection{
		credentialsCollection: credentialsCollection{
			collection: common,
			CredType:   "jwt-auth",
		},
	}
}

// Add adds a jwt-auth credential to JWTAuthsCollection
func (k *JWTAuthsCollection) Add(jwtAuth JWTAuth) error {
	cred := (entity)(&jwtAuth)
	return k.credentialsCollection.Add(cred)
}

// Get gets a jwt-auth credential by key or ID.
func (k *JWTAuthsCollection) Get(keyOrID string) (*JWTAuth, error) {
	cred, err := k.credentialsCollection.Get(keyOrID)
	if err != nil {
		return nil, err
	}

	jwtAuth, ok := cred.(*JWTAuth)
	if !ok {
		panic(unexpectedType)
	}
	return &JWTAuth{JWTAuth: *jwtAuth.DeepCopy()}, nil
}

// GetAllByConsumerID returns all jwt-auth credentials
// belong to a Consumer with id.
func (k *JWTAuthsCollection) GetAllByConsumerID(id string) ([]*JWTAuth,
	error) {
	creds, err := k.credentialsCollection.GetAllByConsumerID(id)
	if err != nil {
		return nil, err
	}

	var res []*JWTAuth
	for _, cred := range creds {
		r, ok := cred.(*JWTAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &JWTAuth{JWTAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing jwt-auth credential.
func (k *JWTAuthsCollection) Update(jwtAuth JWTAuth) error {
	cred := (entity)(&jwtAuth)
	return k.credentialsCollection.Update(cred)
}

// Delete deletes a jwt-auth credential by key or ID.
func (k *JWTAuthsCollection) Delete(keyOrID string) error {
	return k.credentialsCollection.Delete(keyOrID)
}

// GetAll gets all jwt-auth credentials.
func (k *JWTAuthsCollection) GetAll() ([]*JWTAuth, error) {
	creds, err := k.credentialsCollection.GetAll()
	if err != nil {
		return nil, err
	}

	var res []*JWTAuth
	for _, cred := range creds {
		r, ok := cred.(*JWTAuth)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &JWTAuth{JWTAuth: *r.DeepCopy()})
	}
	return res, nil
}
