package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/crossedbot/simplejwt/jwk"
	middleware "github.com/crossedbot/simplemiddleware"
)

func getJwksCert(uri string, token *middleware.Token) (string, error) {
	var jwks jwk.Jwks
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return "", err
	}
	for _, k := range jwks.Keys {
		if token.Header["kid"] == k.KID {
			return fmt.Sprintf(
				"-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----",
				k.X5C[0],
			), nil
		}
	}
	return "", nil
}

func certPemToRsaPublicKey(certPem []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(certPem)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert.PublicKey.(*rsa.PublicKey), nil
}
