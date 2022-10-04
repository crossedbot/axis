package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	middleware "github.com/crossedbot/simplemiddleware"
	"github.com/stretchr/testify/require"
)

func TestContainsGrant(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "hello.world/test", nil)
	require.Nil(t, err)
	grants := []string{"that"}
	SetAuthGrants(grants)

	// Auth grant is contained in request grants (1 contains *)
	reqGrants := fmt.Sprintf("this,%s,those",
		strings.Join(grants, GrantDelimiter))
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.Nil(t, ContainsGrant(req))

	// Auth grant matches request grants (1 == 1)
	reqGrants = "that"
	ctx = req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.Nil(t, ContainsGrant(req))

	// Auth grant is not contained in request grants (1 !contains *)
	reqGrants = "this,those"
	ctx = req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.NotNil(t, ContainsGrant(req))

	// Auth grant does not match request grants (1 != 1)
	reqGrants = "this"
	ctx = req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.NotNil(t, ContainsGrant(req))

	grants = []string{"that,this"}
	SetAuthGrants(grants)

	// Auth grants is contained in the request grants (* contains *)
	reqGrants = fmt.Sprintf("%s,those",
		strings.Join(grants, GrantDelimiter))
	ctx = req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.Nil(t, ContainsGrant(req))

	// Auth grants is not contained in request grant (* !contains 1)
	reqGrants = "those"
	ctx = req.Context()
	ctx = context.WithValue(ctx, middleware.ClaimGrant, reqGrants)
	req = req.WithContext(ctx)
	require.NotNil(t, ContainsGrant(req))
}
