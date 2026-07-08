package auth

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrUserDisabled      = errors.New("user is disabled")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("invalid token")
	ErrSessionExpired    = errors.New("session expired")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrQuotaExceeded     = errors.New("resource quota exceeded")
	ErrAPIKeyNotFound    = errors.New("API key not found")
	ErrAPIKeyDisabled    = errors.New("API key is disabled")
	ErrDuplicateUser     = errors.New("username or email already exists")
)
