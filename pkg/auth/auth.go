package auth

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/crossedbot/common/golang/server"
	middleware "github.com/crossedbot/simplemiddleware"

	"github.com/crossedbot/axis/pkg/pins/models"
)

const (
	GrantDelimiter = ","
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

var authGrants string
var SetAuthGrants = func(grants []string) {
	authGrants = strings.Join(grants, GrantDelimiter)
}

func Authorize(handler server.Handler) server.Handler {
	h := func(w http.ResponseWriter, r *http.Request, p server.Parameters) {
		if len(authGrants) > 0 {
			if err := ContainsGrant(r); err != nil {
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

func ContainsGrant(r *http.Request) error {
	reqGrantStr, ok := r.Context().Value(middleware.ClaimGrant).(string)
	if !ok {
		return middleware.ErrGrantDataType
	}
	reqGrantStr = strings.ToLower(reqGrantStr)
	authGrants = strings.ToLower(authGrants)
	a := strings.Split(reqGrantStr, GrantDelimiter)
	b := strings.Split(authGrants, GrantDelimiter)
	contains, ok := Contains(a, b)
	if !ok || !contains {
		return ErrRequestGrant
	}
	return nil
}
