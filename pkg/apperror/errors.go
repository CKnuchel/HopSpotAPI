package apperror

import "errors"

// User-related errors
var (
	ErrEmailAlreadyExists            = errors.New("email already exists")
	ErrUserNotFound                  = errors.New("user not found")
	ErrInvalidInvitationCode         = errors.New("invalid invitation code")
	ErrInvitationCodeAlreadyRedeemed = errors.New("invitation code already redeemed")
	ErrCannotDeleteSelf              = errors.New("admin cannot delete themselves")
)

// Authentication-related errors
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDeactivated = errors.New("account is deactivated")
	ErrForbidden          = errors.New("forbidden")
)

// Refresh Token Errors
var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

// Bench-related errors
var (
	ErrBenchNotFound = errors.New("bench not found")
)

// Photo Errors
var (
	ErrPhotoNotFound    = errors.New("photo not found")
	ErrMaxPhotosReached = errors.New("maximum photos per bench reached")
	ErrFileTooLarge     = errors.New("file size exceeds limit")
	ErrInvalidFileType  = errors.New("invalid file type")
)
