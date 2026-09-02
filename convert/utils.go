package convert

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

func randomString(n int) (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzj"
	charsetLength := big.NewInt(int64(len(charset)))

	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("error generating random string: %w", err)
		}
		ret[i] = charset[num.Int64()]
	}

	return string(ret), nil
}

func removeServiceName(service *file.FService) *file.FService {
	serviceCopy := service.DeepCopy()
	serviceCopy.Name = nil
	serviceCopy.ID = kong.String(utils.UUID())
	return serviceCopy
}

// isEmpty checks if a value is considered empty (nil, empty slice/array).
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	}

	return false
}
