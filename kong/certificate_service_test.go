package kong

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

const (
	key1 = `-----BEGIN PRIVATE KEY-----
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
x0lR9WSCCs7krwrfYJM0jr7NRhGn
-----END PRIVATE KEY-----`

	cert1 = `-----BEGIN CERTIFICATE-----
MIIEpDCCAowCCQDpoxYguQp6+jANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTgxMjAyMjE0MTI1WhcNMTkxMjAyMjE0MTI1WjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCg
21YOXJB4rjZUvl8dCpLX6oon8qYT6BnpWIfflTU295U72oQGGga3eC2m4YpAWrEE
zA6bGstqcmz7BPzF3ND5kotj5DOSmHSOZ8k+s4z+Sz6+WQWgO2esy6Kxv+KnWYvE
PMgBxgg3v2Kl9v9B9XiHua9FTZVWhY5gF0oBFBlQdhp4FOQa+6CDPQik8++8QcPn
WqlD9CCvheKfSej79ao5Hy1LuSTuHUxgsQ4zYu92S5bYaAPgn09k8axZHIIwhbi8
BaDEdEgySJTvefVpuluVYAKM/bBoClRTCT6i9eqgRln85oMf/yOJ41M+ev2xbLd2
3SdcQ6CsoTT4Xc7Mvi3VC/fNw/6mwmO1lFsRSBE8jUoQRzLtt0w72z7625aMOTwt
njsbtvi5rkF4bf1YW4iqkybIrFr7544cIA4OFdbQ8eWszKVXZnrhyOfsEo4Ir1Kh
JV1eunk6ZG/WhY8MbmKRmjrGabJRYxBL16PEDHTzh01nXsiV5rge86PoXCuzNCTx
oIkhgwq1fSeA8M6Z75u8jxi5N4FygkkjrpsYV6TqhOo20m0BuLiH4OmcgZD4jniH
gGmTJ2h1Ymyv15PCNoPwfEUqJ/cKTfcakwz7WNkSEwTJjjd54zEkz658ggdZkaCh
ek3/pK7itRoLLz/rtaHlOUvAlJc4rphP5t8HNDs+/wIDAQABMA0GCSqGSIb3DQEB
CwUAA4ICAQACRx7PKUGhp0jQLquD0C79086GM4QwCFRlDkewzQiecLE+qz6qYqJK
gSEdL2YHQw2wZOh0GhMMlFk06zDc34gwUdg/aK6oLYJpUZ4jwJKYWQRQY8YWU1gs
Hkq3wKHrPG/YDS07aZBgKvEMHAtlTJeWFcWqKORMxaTpwgQkevUJJaL/Miashz5N
NyUiILKp01kQGBO62BKKVxtxy1EYosdgr8x4TUnW0XuPjLkKuzjJt7v83Ptblu2d
Vhrln5+RLGXldBOnMus8+r2gCbQb5H3fcRizNVnJTTdfq0DoyZSoZx11bKvhZMkx
FiGN/CtLNNBnBJgDSoyesLDs9ZMS6njdCLegxxK5nOL67gKjlbHF+JfAbR4ojyhh
xgsFNDNiVgxssvnR2MOD5rlyqn4UYGQWol90Z52CpQXO1sYRGTA7flf0nSHDaKw9
wuXog4MC3f1dIgvKZYxY0rC/2fMCoop2TK5MqBrIVFcV6IH/T9bEVlUXGk6kvQrf
ZbD+Nn2FKejW43Xfl6Ftd1JGXJr6vzVYG4jEGRNSlZhxnX5gG/fIvor7NZGlWD5b
OsnC3clA3dYwN/mdRAAi6yV/Cdv0ccxcKu1+Ub48zTajwnKliTP59GqjrSFtSoT5
EMP/MtSXGWJ/G5wCCKf/zrmm3J5om8RFywSFLi9ycjmtqi8I1ajJ0w==
-----END CERTIFICATE-----`

	key2 = `-----BEGIN PRIVATE KEY-----
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
N3a/2nCyY29O/S8NtCgE9AI7j+wElcI=
-----END PRIVATE KEY-----`
	cert2 = `-----BEGIN CERTIFICATE-----
MIIEpDCCAowCCQD/wY+0qczvfjANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTgxMjAyMjE0OTI1WhcNMTkxMjAyMjE0OTI1WjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC9
7UR6/nUVRvwWG7ggOCY5p/2FfvFNSuKYeU756rmla5kWndmNeAH4Z3bYrd5g8b1L
WBbfmVCemwPexeBzbDIXfVKRw+AxpJNXxeRs3Bq8hTXo3f53uxLsUYajbzDiLnAw
/nS3v2Qeuy13Sf/V+LmAwJV4PPirQ1ITgnVth/w2oKALBVFwt5t3QyTjiHa8EmxR
iBArfuOGPjtWARbXLQ3IXkyLNMpVuktlwgMqPUbLPdBNkKeTXw5shnA6vAOLeyRC
mdXsH4LDMOGx4HTh6NKtnzG2r7q2bnVChjb2onh+YWeNG7c/oQDlfOIcmcdc0QFn
2xui1dbmKUuOLTxBahhcQhwk2TqxO4ZssnbAgvVwejYWewll9rvumd6wYhpmkjsj
vrR6KNAvEgzBUxOvDAbEepviY2FGgKgMBA7AE6Us92Z9Ie7u0T7wnbhUu+/Ngfr3
mGjdFIYQ5dv6WpClVFsyw5Ynumr3FoBCiuxol//J8zjLcFowDT59ec3DlWdXsSiT
ALuuN83DL1LCZmJ3axmtjxAX/FP3LYOvpHuaBF8FF3m+IIuuMci9JMA/kXV3tfEs
F0/mzZ9o6TW/ze7KpFtnuuOwxLf3DQK9N3BgIIy9RQ73lZg4yfYtU4knec0MVAhE
HYE9pmPQBb8X2/YKegDqNxndVTtl2NFSdyhIajulvwIDAQABMA0GCSqGSIb3DQEB
CwUAA4ICAQC21//GuU+cdj7+dOiPfyODoZVSaHFsTUEOX2kuQ5LnM1chI13Bmzd0
kmw+57Dc5fxzb0mo7uLeU4lXxGhvN3B/2JwVgoVQ+4qqp2w7cFsEpff8gUvTwglI
tkVWMCm+0isRIdFsqsgf4lnPcvMTcymYNR2j8KFbG+vRbDGdG+VSClMxjd/qg3nU
Op5OiyZlzIvoIxcSG5mySeDUimL9REqjD1WCBFgpRVrO5gDlEyDjPOAYoNulXRzR
1PRXHY/lVonO6g1aiOJBZ+BjuE8J81mvZYCOVCliVLEQoeDQ3+qQUGzFzSUQ1+i2
J09JYK3j0MNmI6Uo8x34Ufz+oS1RPRtnPBdTqSx0J6QGdkFe6D4IGwufWeHKe8o5
+OyrctPnx0cnwFC/LE6V/jzFatM6Bpvp54Rt7e3dDLtpQrZrR+4WpjYsNpYwBVWr
lYtlHuS+pOQGOwQyQqh3laQ9CmacR3MMmmtouswhBd81CTh7C2BR965Khl/J/Tmc
K8vpSehmQyvfmJhZgTE21q7M4gGeotd5F6CfoJzqSu2XloDRHXJ7aqUKHNAbR1c7
5+/4rz5abr4xa7usKu2hh5wy7bzYCY3tRFS2mv6Pi/qDmaMSw4JsbB8kisPiP/nL
E7ieP1vQrPWs17eoyI4J84XswdAjyh5OpAX0QVbgQseXrBv8TRECgw==
-----END CERTIFICATE-----`

	key3 = `-----BEGIN PRIVATE KEY-----
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
Kwauzj8I4U9WySIj91DC4rv4ALBfSkU=
-----END PRIVATE KEY-----`
	cert3 = `-----BEGIN CERTIFICATE-----
MIIE+zCCAuOgAwIBAgIJALZU5ftU71m6MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0xODEyMDIyMTU2MjdaFw0xOTEyMDIyMTU2MjdaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBAK9QZkJ1Ihc0488xq1XRoiBtxgwURDRvTWnHnXuN5h7SNxJDsnNWSttQ/xMB
v67a5vznKNjWJeHaQGN40wOkZ3Mb3elkBEVBTl/beqnBuzyL3bhr8SmgN7Dup1PC
tEUZ07AuDg2cqHVfA/C7orsiUJijS969zFu0iJDGgMMxvRmrncqe74uJTzrkW0Bs
5opPLplER99eeGCyuUDaL7WJnx4pnfjBLn9lZMNIWtbdwlvhh38sP7KY+v/k2QVQ
/gLDGr+lFWJwMvdIJVO4eAX6NEZQvm4vAkJl8RC3M3++RtZZKliqdzhtAFxk0wLd
BPmcncmqherf80oJ7lUHTQ9pXL+NXhylNJRb8l8cpDH27w5Vkyv2OGNKtcI0+fTX
TT/VVrPEsLuNjn/7LmYPJqcZrCX5rRsVxmDaObUAdEKJR+EpaPHYZWiwxh+KZwcs
m21pYQacbEWo3MnQU7B0krwYrPzyftPUchvi/eEyy/PE5GRvgzf39vfgT/zGzBhn
SnszvOrWJE3vsErztjJKQLYo9u4Pwq5j4SUyZS65fbT3eQpxGqF6gZnTUdZXVwvO
U026uey+s9WqQWB4p3qGDaThR5fwXz5rLocaAFS3B0IIP0bWCvxADflW9/hRZn7x
P7CeEpeGos4sg4/Ide7nOJl7ub7/4ho54AwL70pniZPmRpoRAgMBAAGjUDBOMB0G
A1UdDgQWBBTOQKJ21L/ciHTkRi1n7rRTInXJ4zAfBgNVHSMEGDAWgBTOQKJ21L/c
iHTkRi1n7rRTInXJ4zAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBCwUAA4ICAQB5
tZqWk0RyeAsCK6sX9tWZFAKFiQrVRGlhcW21nUdZn+jLruup27UontAML0mWHIVi
FUaok3BZ6qEMC0q6DAzfCN7Zmk/K7MeaHc0staCv8qj6XC/CAWgkx3k9WgDp72K1
lyp1hwW8I9tUMoM4C+6LFjp2959v/4mUnLz69atzdomVZiPf2HiUrBAb4eMOXntZ
E4tVyAG3A713QAsOXFMtz8LzlHOTUOPiWcyk92/XfBtsVTmFYpxOKSBrhHIXz+WV
6pKJ557iBpGbu5/CscT5+VN5CYAFxzw0LsRXgJoVqgM5XQS8zztCi8XK9kchpt2u
eULB8qUFnUHqewkBypxDDNQ/mOjY4K5dm9RwM6WUeAlVZGWWn0vHaToUN91f7usr
UfbR/OrU4lizCkznqNqH9IYIB11LSJngr/FMSymRKAQOUSUmqJCUlLvALKEColhW
Ti/feroXva50o6DojtMRBn5G2aTyfIqeiYHBdrBd6NQXKxNSd/qeR3sKRZ4kp6c8
+tCIfUQuN9no4J2cnYULhs2mwInqIny5AgOytXRfDxR61wUezV7OEfUAhhovmuwf
Nez1wdrqpD+3AI7Rv+GU/zCBOCoKl0LlqYchcWWEFgBHcmjgTvGI9yfhoDiezibT
StncqiK5F5CsWRrwQCpoNDkOAQE/l7QZgBzYrXw4vQ==
-----END CERTIFICATE-----`
)

func TestCertificatesService(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &Certificate{
		Key:  String("foo"),
		Cert: String("bar"),
		SNIs: StringSlice("host1.com", "host2.com"),
	}

	createdCertificate, err := client.Certificates.Create(defaultCtx, certificate)
	assert.NotNil(err) // invalid cert and key
	assert.Nil(createdCertificate)

	certificate.Key = String(key1)
	certificate.Cert = String(cert1)
	createdCertificate, err = client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)

	certificate, err = client.Certificates.Get(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
	assert.NotNil(certificate)
	assert.Equal(2, len(createdCertificate.SNIs))

	certificate.Key = String(key2)
	certificate.Cert = String(cert2)
	certificate, err = client.Certificates.Update(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(certificate)
	assert.Equal(key2, *certificate.Key)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)

	// ID can be specified
	id := uuid.NewV4().String()
	certificate = &Certificate{
		Key:  String(key3),
		Cert: String(cert3),
		ID:   String(id),
	}

	createdCertificate, err = client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)
	assert.Equal(id, *createdCertificate.ID)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}

func TestCertificateWithTags(T *testing.T) {
	runWhenKong(T, ">=1.1.0")
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	certificate := &Certificate{
		Key:  String(key3),
		Cert: String(cert3),
		Tags: StringSlice("tag1", "tag2"),
	}

	createdCertificate, err := client.Certificates.Create(defaultCtx, certificate)
	assert.Nil(err)
	assert.NotNil(createdCertificate)
	assert.Equal(StringSlice("tag1", "tag2"), createdCertificate.Tags)

	err = client.Certificates.Delete(defaultCtx, createdCertificate.ID)
	assert.Nil(err)
}

func TestCertificateListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewTestClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	// fixtures
	certificates := []*Certificate{
		{
			Cert: String(cert1),
			Key:  String(key1),
		},
		{
			Cert: String(cert2),
			Key:  String(key2),
		},
		{
			Cert: String(cert3),
			Key:  String(key3),
		},
	}

	// create fixturs
	for i := 0; i < len(certificates); i++ {
		certificate, err := client.Certificates.Create(defaultCtx, certificates[i])
		assert.Nil(err)
		assert.NotNil(certificate)
		certificates[i] = certificate
	}

	certificatesFromKong, next, err := client.Certificates.List(defaultCtx, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(certificatesFromKong)
	assert.Equal(3, len(certificatesFromKong))

	// check if we see all certificates
	assert.True(compareCertificates(certificates, certificatesFromKong))

	// Test pagination
	certificatesFromKong = []*Certificate{}

	// first page
	page1, next, err := client.Certificates.List(defaultCtx, &ListOpt{Size: 1})
	assert.Nil(err)
	assert.NotNil(next)
	assert.NotNil(page1)
	assert.Equal(1, len(page1))
	certificatesFromKong = append(certificatesFromKong, page1...)

	// last page
	next.Size = 2
	page2, next, err := client.Certificates.List(defaultCtx, next)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(page2)
	assert.Equal(2, len(page2))
	certificatesFromKong = append(certificatesFromKong, page2...)

	assert.True(compareCertificates(certificates, certificatesFromKong))

	certificates, err = client.Certificates.ListAll(defaultCtx)
	assert.Nil(err)
	assert.NotNil(certificates)
	assert.Equal(3, len(certificates))

	for i := 0; i < len(certificates); i++ {
		assert.Nil(client.Certificates.Delete(defaultCtx, certificates[i].ID))
	}
}

func compareCertificates(expected, actual []*Certificate) bool {
	var expectedUsernames, actualUsernames []string
	for _, certificate := range expected {
		expectedUsernames = append(expectedUsernames, *certificate.Cert)
	}

	for _, certificate := range actual {
		actualUsernames = append(actualUsernames, *certificate.Cert)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}
