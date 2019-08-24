package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

const (
	caCert1 = `-----BEGIN CERTIFICATE-----
MIICtjCCAZ6gAwIBAgIJAPNy+nTdeZTkMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBUhlbGxvMB4XDTE5MDgwMjA2MjA1MFoXDTE5MDkwMTA2MjA1MFowEDEOMAwG
A1UEAwwFSGVsbG8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCsapiV
hFRy7g4+ilFs4Fb3MXcFPoqFXcS1Wyb04O7xZ7zu9SLkauDk5I99Bjd3AH6Ij9iK
K6V82f5Gcu2Z8XNx76yzTCHNeYc7Jq2/oszmgbI+pS4QigBy1wDSPVkbcUpmOrzm
otMoyzGHqsvqJFcRqRPx0OfRKZpUP1k8ctA1tFmVWLvBy/vvrqXQGgysVP7S0ab5
UdP7CTrGT7dxsRwyvFXYaePhL3Gy2QRQTpVHcghNv/viK1mHB7BU4QJZCp6molz1
/FkYl56OIztiQhstnsiCiPoVSZCC5SyVs4rjH4WccUH7MmzTJXwKIJOSfwewG7WZ
qNqGJUWs8KS9j+uFAgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcN
AQELBQADggEBAG6kD1zHpKD9NKhCyvecStleNOPGWnatrYvtXJ/b1E6mShcA70mY
/6ddwTSsHTkNYARQMzBk1lClIlBGQLFXHyvQe+WMzLs617lOTHMdiFJRRPCbFalh
4PYFg7GH9VHNFlnNDVhObgjPHAeBWVDEW0DMnNxtWsqkDWnYPKC/LpUQHMNkU7db
EA/r4GVy9mn66fcbFiNZn2PoDlOJ/UxgYU9jLCIq8iRTTD5BL3BipAIw1m6xgL+R
XTOPho6Y+CY//SRveLbj3qgCChgDX4VBSBftsevNypJSkVVjIpTSsId5CobRyvpA
G5S85YkvWpITa4oa8VwKlvUMP4BtJbfsPbw=
-----END CERTIFICATE-----`

	caCert2 = `-----BEGIN CERTIFICATE-----
MIICtjCCAZ6gAwIBAgIJAITebuKHfaalMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBUhlbGxvMB4XDTE5MDgwMjA2MjMzM1oXDTE5MDkwMTA2MjMzM1owEDEOMAwG
A1UEAwwFSGVsbG8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDnf9F8
HvV4t+zuxYf2cTWq4D8cXNQWZNIW+20EbE0s9+Fih0DT/VQ9glfh0C6E+WosJxLF
TugyB3MC85gUbdsiIhKgOmYxiyqVP7eb59T87UkSHf35n1nPQrPuEOOtU/94J+Lm
ealnVGUs+Cn5cpZiwSNsUJ8PcZc3cysm7EdbHwshZ8/Yj82X0+ugqpgSu+98johU
/k/spORAAlpwcMTfjBxAHJnEn+TQcorbU+4bXaqi5ArA6+9taz4I3ciA911dXVLJ
el+PqsCOGmT2pXTb6JJge7GzfGDzqk9ufluLt5nVahI2DBQNXWCsERE4+lRVfSXo
RHTWRDND8Te9SV3XAgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcN
AQELBQADggEBAD3chIQhVPwrYdICg5fALpHFft8smwrozkZLJ1AYSDFKiqY+A9Rx
KUAtB5JK9x73OLXhZfenOLiVPqWBGR89o3rqp/hDwGcoga3xSmVgYpljEKrHZv0o
Ho1kMb5miys8voiGNFMw0VN7NbVNu0Y8tqaZ2dw9G08/Ci6DdvjyQzbUBPSlVLj/
lgxcjmLAh1PqUIMtnsQvMLbEJbf3Tv4W7pMSb/5vj07txF8ghc0Og9nYuzl5O5Vn
R+KjGWTWN/6jKiW7eOOd/SXbBXTZvKgathHWbrr0bPsI9xtTPvU1D1aLHGbv+XL/
+hKjAZljUDJBg43XfiPEpEyOSD15aSXc3Vw=
-----END CERTIFICATE-----`

	caCert3 = `-----BEGIN CERTIFICATE-----
MIICtjCCAZ6gAwIBAgIJAOJYFWhgoxIjMA0GCSqGSIb3DQEBCwUAMBAxDjAMBgNV
BAMMBUhlbGxvMB4XDTE5MDgwMjA2MjQxOFoXDTE5MDkwMTA2MjQxOFowEDEOMAwG
A1UEAwwFSGVsbG8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQClcwHz
o9cgAP+KuCypLx3b2iPvQP+CYxK6X2IYyP0bD3nkE4cmh2ZPwvIExnd6PAL17xgE
dHdP/CjNx/Kf6E0yQIqeYB0qwHiZpLI1GmGXXMlkGCFEvtFKy1nlEc4TDPAFR0kb
aJPJkvr+DtVUaU+sztOoQkSp2er040IpRkZNxTSA1yM0Z26uCG91iZB/1QEYSWKp
4vEaiJYMcHy6KMaGP/RvX0ooGoyGows1QLaEfQLzAgW+8z/LDePabGyb+xRLTxSi
VYzkg/7JtTa5Hx9rZLSULHZkkqV1ALI9CiF5SrnuzGTuGUOzt0uR794ftz9JFkqK
qinHuiYTXiY5DOZpAgMBAAGjEzARMA8GA1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcN
AQELBQADggEBAHWAzlaT7t6EsTZ59lxOhZYDMM6UQI0wb1DlP4D/41x0PmIakxn3
luMCRIsQmkTDumG2EiCxAIBNYIdab9fE7rf9iNGAbo29N3rPwhXOf9HkoTXDmM3d
OBn+lJ9SxdqrITSf5iGwK30w5TEqoohRrZ/3X6CMbv8P04NAL9wgpuynBnJ7ChyJ
GyjCdC9WA9vyyFD7q2wIJsfOPX71hVjWbjBrQetFEXK2GcOD/HYxx6+dsFnLrwr9
mzxr44k03x8GbDsWkU2amReR6+lf5arOqYEXBBQKgYiItq/db1ce7NKfROvxAL1T
MgtromXLVHfyA84RZSUYh2mh6inNngdIhCM=
-----END CERTIFICATE-----`
)

func TestCACertificatesService(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &CACertificate{
		Cert: String("bar"),
	}

	createdCertificate, err := client.CACertificates.Create(defaultCtx,
		certificate)
	assert.NotNil(err) // invalid cert and key
	assert.Nil(createdCertificate)

	certificate.Cert = String(caCert1)
	createdCertificate, err = client.CACertificates.Create(defaultCtx,
		certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)

	certificate, err = client.CACertificates.Get(defaultCtx,
		createdCertificate.ID)
	assert.Nil(err)
	assert.NotNil(certificate)

	certificate.Cert = String(caCert2)
	certificate, err = client.CACertificates.Update(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(certificate)

	err = client.CACertificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	certificate = &CACertificate{
		Cert: String(caCert3),
		ID:   String(id),
	}

	createdCertificate, err = client.CACertificates.Create(defaultCtx,
		certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)
	assert.Equal(id, *createdCertificate.ID)

	err = client.CACertificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}

func TestCACertificateWithTags(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &CACertificate{
		Cert: String(cert3),
		Tags: StringSlice("tag1", "tag2"),
	}

	createdCertificate, err := client.CACertificates.Create(defaultCtx,
		certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)
	assert.Equal(StringSlice("tag1", "tag2"), createdCertificate.Tags)

	err = client.CACertificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}

func TestCACertificateListEndpoint(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	certificates := []*CACertificate{
		{
			Cert: String(caCert1),
		},
		{
			Cert: String(caCert2),
		},
		{
			Cert: String(caCert3),
		},
	}

	// create fixturs
	for i := 0; i < len(certificates); i++ {
		certificate, err := client.CACertificates.Create(defaultCtx,
			certificates[i])
		assert.Nil(err)
		assert.NotNil(certificate)
		certificates[i] = certificate
	}

	certificatesFromKong, next, err :=
		client.CACertificates.List(defaultCtx, nil)

	assert.Nil(next)
	assert.NotNil(certificatesFromKong)
	assert.Equal(3, len(certificatesFromKong))

	// check if we see all certificates
	assert.True(compareCACertificates(certificates, certificatesFromKong))

	// Test pagination
	certificatesFromKong = []*CACertificate{}

	// first page
	page1, next, err := client.CACertificates.List(defaultCtx, &ListOpt{
		Size: 1,
	})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	certificatesFromKong = append(certificatesFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.CACertificates.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	certificatesFromKong = append(certificatesFromKong, page2...)

	assert.True(compareCACertificates(certificates, certificatesFromKong))

	certificates, err = client.CACertificates.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(certificates)
	assert.Equal(3, len(certificates))

	for i := 0; i < len(certificates); i++ {
		assert.Nil(client.CACertificates.Delete(defaultCtx, certificates[i].ID))
	}
}

func compareCACertificates(expected, actual []*CACertificate) bool {
	var expectedUsernames, actualUsernames []string
	for _, certificate := range expected {
		expectedUsernames = append(expectedUsernames, *certificate.Cert)
	}

	for _, certificate := range actual {
		actualUsernames = append(actualUsernames, *certificate.Cert)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}
