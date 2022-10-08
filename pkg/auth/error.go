package auth

import (
	"errors"
)

var (
	ErrMissingJWKS    = errors.New("Failed to find matching key for token")
	ErrRequestGrant   = errors.New("Request does not match grant")
	ErrUserIdDataType = errors.New("User ID claim is not a string")
	ErrUserForbidden  = errors.New("Not authorized to perform this action")
)
