package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crossedbot/common/golang/server"
	"github.com/crossedbot/simplejwt"
	"github.com/crossedbot/simplejwt/algorithms"
	middleware "github.com/crossedbot/simplemiddleware"
	"github.com/stretchr/testify/require"

	"github.com/crossedbot/axis/pkg/pins/models"
)

func TestAuthenticate(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.Nil(t, err)
	claims := simplejwt.CustomClaims{
		middleware.ClaimUserId: "myuserid",
		"exp": time.Now().Local().Add(
			time.Hour * time.Duration(1),
		).Unix(),
	}
	tkn, err := simplejwt.New(claims, algorithms.AlgorithmRS256).
		Sign([]byte(testPrivateKey))
	require.Nil(t, err)
	bearer := fmt.Sprintf("Bearer %s", tkn)
	r.Header.Set(middleware.AuthHeader, bearer)
	authenticator = middleware.New(
		middleware.AuthHeader,
		func(token *middleware.Token) ([]byte, error) {
			return []byte(testPublicKey), nil
		},
		func(w http.ResponseWriter, err error) {
			server.JsonResponse(w, models.NewFailure(
				models.ErrUnauthorizedCode,
				err.Error(),
			), http.StatusUnauthorized)
		},
	)
	handler := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		inner := Authenticate(
			server.Handler(func(
				w http.ResponseWriter,
				r *http.Request,
				p server.Parameters,
			) {
				w.WriteHeader(http.StatusOK)
			}),
		)
		p := server.GetParameters(r.Context())
		inner(w, r, p)
	})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, r)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "", rr.Body.String())
}
