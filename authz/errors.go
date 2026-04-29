package authz

import "errors"

var (
	// ErrInvalidConfig indicates the authz config is incomplete or invalid.
	ErrInvalidConfig = errors.New("authz: invalid config")
	// ErrNotInitialized indicates the package global authorizer has not been set.
	ErrNotInitialized = errors.New("authz: not initialized")
	// ErrUnsupportedAdapter indicates the configured policy adapter is unsupported.
	ErrUnsupportedAdapter = errors.New("authz: unsupported adapter")
	// ErrForbidden indicates an authorization decision denied access.
	ErrForbidden = errors.New("authz: forbidden")
)
