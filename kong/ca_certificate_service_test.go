package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

const (
	caCert1 = `-----BEGIN CERTIFICATE-----
MIIEvjCCAqagAwIBAgIJALabx/Nup200MA0GCSqGSIb3DQEBCwUAMBMxETAPBgNV
BAMMCFlvbG80Mi4xMCAXDTE5MDkxNTE2Mjc1M1oYDzIxMTkwODIyMTYyNzUzWjAT
MREwDwYDVQQDDAhZb2xvNDIuMTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBANIW67Ay0AtTeBY2mORaGet/VPL5jnBRz0zkZ4Jt7fEq3lbxYaJBnFI8wtz3
bHLtLsxkvOFujEMY7HVd+iTqbJ7hLBtK0AdgXDjf+HMmoWM7x0PkZO+3XSqyRBbI
YNoEaQvYBNIXrKKJbXIU6higQaXYszeN8r3+RIbcTIlZxy28msivEGfGTrNujQFc
r/eyf+TLHbRqh0yg4Dy/U/T6fqamGhFrjupRmOMugwF/BHMH2JHhBYkkzuZLgV2u
7Yh1S5FRlh11am5vWuRSbarnx72hkJ99rUb6szOWnJKKew8RSn3CyhXbS5cb0QRc
ugRc33p/fMucJ4mtCJ2Om1QQe83G1iV2IBn6XJuCvYlyWH8XU0gkRxWD7ZQsl0bB
8AFTkVsdzb94OM8Y6tWI5ybS8rwl8b3r3fjyToIWrwK4WDJQuIUx4nUHObDyw+KK
+MmqwpAXQWbNeuAc27FjuJm90yr/163aGuInNY5Wiz6CM8WhFNAi/nkEY2vcxKKx
irSdSTkbnrmLFAYrThaq0BWTbW2mwkOatzv4R2kZzBUOiSjRLPnbyiPhI8dHLeGs
wMxiTXwyPi8iQvaIGyN4DPaSEiZ1GbexyYFdP7sJJD8tG8iccbtJYquq3cDaPTf+
qv5M6R/JuMqtUDheLSpBNK+8vIe5e3MtGFyrKqFXdynJtfHVAgMBAAGjEzARMA8G
A1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQADggIBAK0BmL5B1fPSMbFy8Hbc
/ESEunt4HGaRWmZZSa/aOtTjhKyDXLLJZz3C4McugfOf9BvvmAOZU4uYjfHTnNH2
Z3neBkdTpQuJDvrBPNoCtJns01X/nuqFaTK/Tt9ZjAcVeQmp51RwhyiD7nqOJ/7E
Hp2rC6gH2ABXeexws4BDoZPoJktS8fzGWdFBCHzf4mCJcb4XkI+7GTYpglR818L3
dMNJwXeuUsmxxKScBVH6rgbgcEC/6YwepLMTHB9VcH3X5VCfkDIyPYLWmvE0gKV7
6OU91E2Rs8PzbJ3EuyQpJLxFUQp8ohv5zaNBlnMb76UJOPR6hXfst5V+e7l5Dgwv
Dh4CeO46exmkEsB+6R3pQR8uOFtubH2snA0S3JA1ji6baP5Y9Wh9bJ5McQUgbAPE
sCRBFoDLXOj3EgzibohC5WrxN3KIMxlQnxPl3VdQvp4gF899mn0Z9V5dAsGPbxRd
quE+DwfXkm0Sa6Ylwqrzu2OvSVgbMliF3UnWbNsDD5KcHGIaFxVC1qkwK4cT3pyS
58i/HAB2+P+O+MltQUDiuw0OSUFDC0IIjkDfxLVffbF+27ef9C5NG81QlwTz7TuN
zeigcsBKooMJTszxCl6dtxSyWTj7hJWXhy9pXsm1C1QulG6uT4RwCa3m0QZoO7G+
6Wu6lP/kodPuoNubstIuPdi2
-----END CERTIFICATE-----`
	caCert2 = `-----BEGIN CERTIFICATE-----
MIIEvjCCAqagAwIBAgIJAPf5iqimiR2BMA0GCSqGSIb3DQEBCwUAMBMxETAPBgNV
BAMMCFlvbG80Mi4yMCAXDTE5MDkxNTE2Mjc1OVoYDzIxMTkwODIyMTYyNzU5WjAT
MREwDwYDVQQDDAhZb2xvNDIuMjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBANIW67Ay0AtTeBY2mORaGet/VPL5jnBRz0zkZ4Jt7fEq3lbxYaJBnFI8wtz3
bHLtLsxkvOFujEMY7HVd+iTqbJ7hLBtK0AdgXDjf+HMmoWM7x0PkZO+3XSqyRBbI
YNoEaQvYBNIXrKKJbXIU6higQaXYszeN8r3+RIbcTIlZxy28msivEGfGTrNujQFc
r/eyf+TLHbRqh0yg4Dy/U/T6fqamGhFrjupRmOMugwF/BHMH2JHhBYkkzuZLgV2u
7Yh1S5FRlh11am5vWuRSbarnx72hkJ99rUb6szOWnJKKew8RSn3CyhXbS5cb0QRc
ugRc33p/fMucJ4mtCJ2Om1QQe83G1iV2IBn6XJuCvYlyWH8XU0gkRxWD7ZQsl0bB
8AFTkVsdzb94OM8Y6tWI5ybS8rwl8b3r3fjyToIWrwK4WDJQuIUx4nUHObDyw+KK
+MmqwpAXQWbNeuAc27FjuJm90yr/163aGuInNY5Wiz6CM8WhFNAi/nkEY2vcxKKx
irSdSTkbnrmLFAYrThaq0BWTbW2mwkOatzv4R2kZzBUOiSjRLPnbyiPhI8dHLeGs
wMxiTXwyPi8iQvaIGyN4DPaSEiZ1GbexyYFdP7sJJD8tG8iccbtJYquq3cDaPTf+
qv5M6R/JuMqtUDheLSpBNK+8vIe5e3MtGFyrKqFXdynJtfHVAgMBAAGjEzARMA8G
A1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQADggIBALNx2xaS5nv1QjEqtiCO
EA/ZTXbs+il6cf6ZyUwFXs7d3OKx6Kk2Nr7wGgM1M5WuTyIGKtZspz9ThzYmsuN/
UBCSKLw3X7U2fLiHJDipXboU1txasTErUTPJs/Vq4v7PWh8sMLCQH/ha4FAOXR0M
Uie+VgSJNKoQSj7G1hzU/LZv0KdvJ45mQBCnBXrUrGgeEcRqubbkDKgdBh7dJQzW
Xgy6rPb6H1aXbsSuRuUVv/xFHJoCdZJmqPH4JTMYRbHNS2km9nHVJzmtL6pQFe32
24wfpue9geFndOE9bDU9/cqoRYA4Pce4V5qDL0wL9W4uPmyPDkulKNQtAvZnDA9V
6ccYYthlTBr62UEnw7zZOnSm0q4fB2o82/6bdPwrT7WhbHZQWN7SeqYNWAbYZ1EE
40f5IpTwZ7E5LaG62qPhKLXame7SPAaqaQ9aCTYxaWR7XSYBsvCBRanjRq0r9Tql
T1I8lwssIgbA3XubokI+IMkLDEpCQ27niWXOZL5y2M3xyutd6PPjmEEmoHMkOrZL
etlxzx2CCoUDXKkYW2gZKEozwBZ+eBgUj8WB5g/8jGDAI0qzYnfAgiahjGwlEUtP
hJiPG/YFADw0m5b/8OMCZ6AXNhxjdweHniDxY2HE734Nwm9mG/7UbkdvhR05tqFh
G4KCViLH0cXt/TgW1sYB2o9Z
-----END CERTIFICATE-----`

	caCert3 = `-----BEGIN CERTIFICATE-----
MIIEvjCCAqagAwIBAgIJAPOV4FBF2WzgMA0GCSqGSIb3DQEBCwUAMBMxETAPBgNV
BAMMCFlvbG80Mi4zMCAXDTE5MDkxNTE2MjgwNloYDzIxMTkwODIyMTYyODA2WjAT
MREwDwYDVQQDDAhZb2xvNDIuMzCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBANIW67Ay0AtTeBY2mORaGet/VPL5jnBRz0zkZ4Jt7fEq3lbxYaJBnFI8wtz3
bHLtLsxkvOFujEMY7HVd+iTqbJ7hLBtK0AdgXDjf+HMmoWM7x0PkZO+3XSqyRBbI
YNoEaQvYBNIXrKKJbXIU6higQaXYszeN8r3+RIbcTIlZxy28msivEGfGTrNujQFc
r/eyf+TLHbRqh0yg4Dy/U/T6fqamGhFrjupRmOMugwF/BHMH2JHhBYkkzuZLgV2u
7Yh1S5FRlh11am5vWuRSbarnx72hkJ99rUb6szOWnJKKew8RSn3CyhXbS5cb0QRc
ugRc33p/fMucJ4mtCJ2Om1QQe83G1iV2IBn6XJuCvYlyWH8XU0gkRxWD7ZQsl0bB
8AFTkVsdzb94OM8Y6tWI5ybS8rwl8b3r3fjyToIWrwK4WDJQuIUx4nUHObDyw+KK
+MmqwpAXQWbNeuAc27FjuJm90yr/163aGuInNY5Wiz6CM8WhFNAi/nkEY2vcxKKx
irSdSTkbnrmLFAYrThaq0BWTbW2mwkOatzv4R2kZzBUOiSjRLPnbyiPhI8dHLeGs
wMxiTXwyPi8iQvaIGyN4DPaSEiZ1GbexyYFdP7sJJD8tG8iccbtJYquq3cDaPTf+
qv5M6R/JuMqtUDheLSpBNK+8vIe5e3MtGFyrKqFXdynJtfHVAgMBAAGjEzARMA8G
A1UdEwQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQADggIBAAWkIoAl1g5crjJcdQcN
9gF2+FRDdo84V+srtA5q9bvGDYXt9S8/IDSqjDKz/03nlCdye4bBLacorhhZS97O
jPBdZ30kAVn478z6ZcZQeHHE93uOcO+hZWcQEoFRid8HSVAbFimhC+wW9JQv1CKz
S/3i1uPXAE9V2TlWmKYbqt8jJBJ1qJV1Gt8V9AAB+L5X/I5MQoMk7rmddkUf5fEj
2Lghp4egDnHNE9jqMxx38LY+06TpoAr++ICdJ0P2n/Cmg+Z3jQZklE+8SR9C1XfK
N1W9jGpDFkRgZx6udkGf6NDwyr+W3O//U+P2TLMSlPFHbpqH1jJFWVHaohE4rjPl
MpG0eqM6inuliNewymQPurwxP47wBeQedDw169ehxbwqCmRvNToAfp+GTwrhoIDt
YP97asx55JPxsXniM+L8mSsRvS+k11A8fdq5E3luKJ0Pyfct2AzUoy5OzcUWotv1
M5l19WkET0gOKRTzJzUpZiFQ3S6HIk//kT08VD++7UegLDOn2s0TIPLBpU6uaihm
6VSxYPyCcVX+4FeMGU5T47xnT5ClyhjSgUiY50FVuXuyA2zrsdZj4O5KDktK9hGh
Vlz21lz4fvd3gtPgtVXLmaBKlGQFWNLxARBTQFxJJ7JvYu6FnhzhWt6YO49vAnUG
R+pHRocvtyc8EnkuMw6+jGHr
-----END CERTIFICATE-----`
)

func TestCACertificatesService(T *testing.T) {
	runWhenKong(T, ">=1.3.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
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

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &CACertificate{
		Cert: String(caCert3),
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

	client, err := NewTestClient(nil, nil)
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

	assert.Nil(err)
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
