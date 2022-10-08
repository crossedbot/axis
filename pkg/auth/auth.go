package auth

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/crossedbot/common/golang/server"
	"github.com/crossedbot/simpleauth/pkg/grants"
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

var authGrants grants.Grant
var SetAuthGrants = func(grantStrs []string) error {
	err := grants.SetCustomGrants(grantStrs)
	if err != nil {
		return err
	}
	s := strings.Join(grantStrs, grants.GrantDelimiter)
	authGrants, err = grants.ToGrant(s)
	authGrants = authGrants.Clean()
	return err
}

func Authorize(handler server.Handler) server.Handler {
	h := func(w http.ResponseWriter, r *http.Request, p server.Parameters) {
		if authGrants != grants.GrantUnknown {
			err := grants.ContainsGrant(authGrants, r)
			if err != nil {
				server.JsonResponse(w, models.NewFailure(
					server.ErrUnauthorizedCode,
					ErrUserForbidden.Error(),
				), http.StatusForbidden)
				return
			}
		}
		handler(w, r, p)
	}
	return middleware.Authorize(h)
}
