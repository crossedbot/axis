package auth

import (
	"errors"
)

var (
	ErrMissingJWKS    = errors.New("failed to find matching key for token")
	ErrUserIdDataType = errors.New("user ID claim is not a string")
	ErrUserForbidden  = errors.New("Not authorized to perform this action")
)
