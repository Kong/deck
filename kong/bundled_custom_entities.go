package kong

import (
	"github.com/kong/go-kong/kong/custom"
)

var defaultCustomEntities = []custom.EntityCRUDDefinition{
	{
		Name:       "key-auth",
		CRUDPath:   "/consumers/${consumer_id}/key-auth",
		PrimaryKey: "id",
	},
	{
		Name:       "basic-auth",
		CRUDPath:   "/consumers/${consumer_id}/basic-auth",
		PrimaryKey: "id",
	},
	{
		Name:       "acl",
		CRUDPath:   "/consumers/${consumer_id}/acls",
		PrimaryKey: "id",
	},
	{
		Name:       "hmac-auth",
		CRUDPath:   "/consumers/${consumer_id}/hmac-auth",
		PrimaryKey: "id",
	},
	{
		Name:       "jwt",
		CRUDPath:   "/consumers/${consumer_id}/jwt",
		PrimaryKey: "id",
	},
	{
		Name:       "oauth2",
		CRUDPath:   "/consumers/${consumer_id}/oauth2",
		PrimaryKey: "id",
	},
	{
		Name:       "mtls-auth",
		CRUDPath:   "/consumers/${consumer_id}/mtls-auth",
		PrimaryKey: "id",
	},
}
