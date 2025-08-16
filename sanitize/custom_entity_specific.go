package sanitize

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

var (
	FCertificate  = "FCertificate"
	CACertificate = "CACertificate"
	Key           = "Key"
)

func (s *Sanitizer) handleEntity(entityName string, fieldValue reflect.Value) error {
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return nil
	}

	switch entityName {
	case FCertificate:
		return s.handleFCertificate(fieldValue)
	case CACertificate:
		return s.handleCACertificate(fieldValue)
	case Key:
		return s.handleKey(fieldValue)
	default:
		return fmt.Errorf("no specific handler for entity: %s\n", entityName)
	}
}

func (s *Sanitizer) handleFCertificate(fieldValue reflect.Value) error {
	cert := fieldValue.Interface().(file.FCertificate)
	sanitisedCertificate := cert.DeepCopy()

	certPEM, keyPEM, _, err := generateTestCertAndKey(false)
	if err != nil {
		return fmt.Errorf("error generating test certificate and key: %w", err)
	}

	sanitisedCertificate.Cert = certPEM
	sanitisedCertificate.Key = keyPEM

	s.sanitizedMap[*cert.Cert] = sanitisedCertificate.Cert
	s.sanitizedMap[*cert.Key] = sanitisedCertificate.Key

	return s.setFieldValue(fieldValue, *sanitisedCertificate, FCertificate)
}

func (s *Sanitizer) handleCACertificate(fieldValue reflect.Value) error {
	cert := fieldValue.Interface().(kong.CACertificate)
	sanitisedCertificate := cert.DeepCopy()

	certPEM, _, certDigest, err := generateTestCertAndKey(true)
	if err != nil {
		return fmt.Errorf("error generating test certificate and key: %w", err)
	}

	sanitisedCertificate.Cert = certPEM
	sanitisedCertificate.CertDigest = certDigest

	s.sanitizedMap[*cert.Cert] = sanitisedCertificate.Cert
	s.sanitizedMap[*cert.CertDigest] = sanitisedCertificate.CertDigest

	return s.setFieldValue(fieldValue, *sanitisedCertificate, CACertificate)
}

func (s *Sanitizer) handleKey(fieldValue reflect.Value) error {
	key := fieldValue.Interface().(kong.Key)
	sanitisedKey := key.DeepCopy()

	kid := utils.UUID()
	if key.KID != nil {
		kid = s.sanitizeValue(*key.KID)
	}
	s.sanitizedMap[*key.KID] = kid
	sanitisedKey.KID = &kid

	if key.PEM != nil {
		privateKeyStr, privateKey, err := generateRSAPrivateKeyPEM()
		if err != nil {
			return fmt.Errorf("error generating RSA private key PEM: %w", err)
		}

		publicKey := generateRSAPublicKeyPEM(privateKey)

		sanitisedKey.PEM = &kong.PEM{
			PublicKey:  &publicKey,
			PrivateKey: &privateKeyStr,
		}

		s.sanitizedMap[*key.PEM.PublicKey] = *sanitisedKey.PEM.PublicKey
		s.sanitizedMap[*key.PEM.PrivateKey] = *sanitisedKey.PEM.PrivateKey
	}

	if key.JWK != nil {
		jwk, err := generateRSAJWK(kid)
		if err != nil {
			return fmt.Errorf("error generating RSA JWK: %w", err)
		}
		sanitisedKey.JWK = &jwk
		s.sanitizedMap[*key.JWK] = *sanitisedKey.JWK
	}

	return s.setFieldValue(fieldValue, *sanitisedKey, Key)
}

// generateTestCertAndKey generates a test certificate and key pair.
// It returns the certificate PEM, key PEM, and certificate digest.
// If isCA is true, it generates a CA certificate; otherwise, it generates a regular certificate.
func generateTestCertAndKey(isCA bool) (certPEM, keyPEM, certDigest *string, err error) {
	// Generating a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	subject := "Test Certificate"
	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if isCA {
		subject = "Test CA Certificate"
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	// Creating a certificate template
	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: subject},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(48 * time.Hour),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  isCA,
		BasicConstraintsValid: true,
	}

	// Self-signing the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, err
	}

	// PEM encoding the certificate and key
	certPEMBytes := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))
	keyPEMBytes := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))

	// Computing digest (SHA256 of DER)
	digest := sha256.Sum256(derBytes)
	certDigestStr := hex.EncodeToString(digest[:])

	certPEM = &certPEMBytes
	keyPEM = &keyPEMBytes
	certDigest = &certDigestStr

	return certPEM, keyPEM, certDigest, nil
}

// generateRSAJWK generates a JWK representation of an RSA key with the given KID.
func generateRSAJWK(kid string) (string, error) {
	_, priv, err := generateRSAPrivateKeyPEM()
	if err != nil {
		return "", err
	}

	pub := &priv.PublicKey

	// converting RSA public key components to base64url encoding
	nBytes := pub.N.Bytes()
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}

	jwkJSON, err := json.Marshal(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWK: %w", err)
	}

	return string(jwkJSON), nil
}

// generateRSAPrivateKeyPEM generates RSA private key with PKCS#1 encryption
func generateRSAPrivateKeyPEM() (string, *rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return string(privPEM), priv, nil
}

// generateRSAPublicKeyPEM generates RSA public key with PKCS#1 encryption
func generateRSAPublicKeyPEM(priv *rsa.PrivateKey) string {
	pubBytes := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
	return string(pubPEM)
}

// setFieldValue is a helper function to set the sanitized value on the field
func (s *Sanitizer) setFieldValue(fieldValue reflect.Value, sanitizedValue interface{}, entityName string) error {
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(sanitizedValue))
	} else {
		fmt.Println("Cannot sanitize: ", entityName)
	}
	return nil
}
