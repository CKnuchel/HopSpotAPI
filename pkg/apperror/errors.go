package apperror

import "errors"

// Legacy sentinel errors - kept for backward compatibility during migration
// These are used in service layer and mapped to AppError in handlers

// User-related errors
var (
	ErrEmailAlreadyExists            = errors.New("email already exists")
	ErrUserNotFound                  = errors.New("user not found")
	ErrInvalidInvitationCode         = errors.New("invalid invitation code")
	ErrInvitationCodeAlreadyRedeemed = errors.New("invitation code already redeemed")
	ErrCannotDeleteSelf              = errors.New("admin cannot delete themselves")
	ErrInvitationCodeNotFound        = errors.New("invitation code not found")
	ErrCannotDeleteRedeemedCode      = errors.New("cannot delete redeemed invitation code")
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
	ErrBenchNotFound  = errors.New("bench not found")
	ErrBenchForbidden = errors.New("no permission for this bench")
)

// Photo Errors
var (
	ErrPhotoNotFound    = errors.New("photo not found")
	ErrMaxPhotosReached = errors.New("maximum photos per bench reached")
	ErrFileTooLarge     = errors.New("file size exceeds limit")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrPhotoForbidden   = errors.New("no permission for this photo")
)

// Visit Errors
var (
	ErrVisitNotFound  = errors.New("visit not found")
	ErrVisitForbidden = errors.New("can only delete own visits")
)

// Favorite Errors
var (
	ErrFavoriteNotFound     = errors.New("favorite not found")
	ErrFavoriteAlreadyExists = errors.New("already in favorites")
)

// MapToAppError converts a legacy sentinel error to an AppError
// This function is used to bridge the transition from simple errors to structured errors
func MapToAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	switch {
	// Authentication errors
	case errors.Is(err, ErrInvalidCredentials):
		return AppErrInvalidCredentials
	case errors.Is(err, ErrInvalidToken):
		return AppErrInvalidToken
	case errors.Is(err, ErrInvalidRefreshToken):
		return AppErrInvalidRefreshToken
	case errors.Is(err, ErrAccountDeactivated):
		return AppErrAccountDeactivated
	case errors.Is(err, ErrForbidden):
		return AppErrForbidden

	// User errors
	case errors.Is(err, ErrUserNotFound):
		return AppErrUserNotFound
	case errors.Is(err, ErrEmailAlreadyExists):
		return AppErrEmailExists
	case errors.Is(err, ErrCannotDeleteSelf):
		return AppErrCannotDeleteSelf

	// Invitation errors
	case errors.Is(err, ErrInvalidInvitationCode):
		return AppErrInvitationInvalidCode
	case errors.Is(err, ErrInvitationCodeAlreadyRedeemed):
		return AppErrInvitationAlreadyRedeemed
	case errors.Is(err, ErrInvitationCodeNotFound):
		return AppErrInvitationNotFound
	case errors.Is(err, ErrCannotDeleteRedeemedCode):
		return AppErrInvitationCannotDeleteRedeemed

	// Bench errors
	case errors.Is(err, ErrBenchNotFound):
		return AppErrBenchNotFound
	case errors.Is(err, ErrBenchForbidden):
		return AppErrBenchForbidden

	// Photo errors
	case errors.Is(err, ErrPhotoNotFound):
		return AppErrPhotoNotFound
	case errors.Is(err, ErrMaxPhotosReached):
		return AppErrPhotoMaxReached
	case errors.Is(err, ErrFileTooLarge):
		return AppErrPhotoTooLarge
	case errors.Is(err, ErrInvalidFileType):
		return AppErrPhotoInvalidType
	case errors.Is(err, ErrPhotoForbidden):
		return AppErrPhotoForbidden

	// Visit errors
	case errors.Is(err, ErrVisitNotFound):
		return AppErrVisitNotFound
	case errors.Is(err, ErrVisitForbidden):
		return AppErrVisitForbidden

	// Favorite errors
	case errors.Is(err, ErrFavoriteNotFound):
		return AppErrFavoriteNotFound
	case errors.Is(err, ErrFavoriteAlreadyExists):
		return AppErrFavoriteAlreadyExists

	default:
		return nil
	}
}

// RespondWithMappedError maps a legacy error to AppError and sends the response
// Falls back to internal server error for unmapped errors
func RespondWithMappedError(c interface{ JSON(int, any) }, err error) {
	appErr := MapToAppError(err)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, ErrorResponse{
			ErrorCode: appErr.Code,
			Message:   appErr.Message,
		})
		return
	}

	// Unmapped error - return internal server error
	c.JSON(500, ErrorResponse{
		ErrorCode: ErrCodeSystemInternal,
		Message:   "An unexpected error occurred",
	})
}
