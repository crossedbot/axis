package auth

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"sync"

	"github.com/crossedbot/axis/pkg/pins/models"
	"github.com/crossedbot/common/golang/server"
	middleware "github.com/crossedbot/simplemiddleware"
)

var authenticatorAddr string
var SetAuthenticatorAddr = func(addr string) {
	authenticatorAddr = addr
}

var once sync.Once
var authenticator = func() (mw middleware.Middleware) {
	once.Do(func() {
		mw = middleware.New(
			middleware.AuthHeader,
			func(token *middleware.Token) ([]byte, error) {
				cert, err := getJwksCert(authenticatorAddr, token)
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
			},
			func(w http.ResponseWriter, err error) {
				server.JsonResponse(w, models.NewFailure(
					models.ErrUnauthorizedCode,
					err.Error(),
				), http.StatusUnauthorized)
			},
		)
	})
	return
}()

func Authenticate(handler server.Handler) server.Handler {
	h := authenticator.Handle(func(w http.ResponseWriter, r *http.Request) {
		p := server.GetParameters(r.Context())
		userID, err := getUserIdFromRequest(r)
		if err != nil || userID == "" {
			server.JsonResponse(w, models.NewFailure(
				models.ErrUnauthorizedCode,
				"user identifier is missing or invalid",
			), http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		r = r.WithContext(context.WithValue(
			ctx,
			middleware.ClaimUserId,
			userID,
		))
		handler(w, r, p)
	})
	return server.NewHandler(h)
}
