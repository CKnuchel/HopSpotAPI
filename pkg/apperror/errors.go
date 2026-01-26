package apperror

import "errors"

// User-related errors
var (
	ErrEmailAlreadyExists            = errors.New("email already exists")
	ErrUserNotFound                  = errors.New("user not found")
	ErrInvalidInvitationCode         = errors.New("invalid invitation code")
	ErrInvitationCodeAlreadyRedeemed = errors.New("invitation code already redeemed")
)

// Authentication-related errors
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account is deactivated")
	ErrForbidden          = errors.New("forbidden")
)

// Bench-related errors
var (
	ErrBenchNotFound = errors.New("bench not found")
)
