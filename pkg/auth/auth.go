package auth

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/crossedbot/common/golang/server"
	middleware "github.com/crossedbot/simplemiddleware"

	"github.com/crossedbot/axis/pkg/pins/models"
)

var KeyFunc = func(authAddr string) middleware.KeyFunc {
	return func(token *middleware.Token) ([]byte, error) {
		cert, err := getJwksCert(authAddr, token)
		if err != nil {
			return nil, err
		}
		if cert == "" {
			return nil, ErrMissingJWKS
		}
		key, err := certPemToRsaPublicKey([]byte(cert))
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(key),
			},
		), nil
	}
}

var ErrFunc = func() middleware.ErrFunc {
	return func(w http.ResponseWriter, err error) {
		server.JsonResponse(w, models.NewFailure(
			models.ErrUnauthorizedCode,
			err.Error(),
		), http.StatusUnauthorized)
	}
}
