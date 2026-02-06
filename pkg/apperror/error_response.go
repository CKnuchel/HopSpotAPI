package apperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode represents a unique error identifier
type ErrorCode string

// Error codes - Authentication
const (
	ErrCodeInvalidCredentials  ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrCodeInvalidToken        ErrorCode = "AUTH_INVALID_TOKEN"
	ErrCodeTokenExpired        ErrorCode = "AUTH_TOKEN_EXPIRED"
	ErrCodeInvalidRefreshToken ErrorCode = "AUTH_INVALID_REFRESH_TOKEN"
	ErrCodeAccountDeactivated  ErrorCode = "AUTH_ACCOUNT_DEACTIVATED"
	ErrCodeForbidden           ErrorCode = "AUTH_FORBIDDEN"
	ErrCodeAdminRequired       ErrorCode = "AUTH_ADMIN_REQUIRED"
)

// Error codes - User
const (
	ErrCodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	ErrCodeEmailExists        ErrorCode = "USER_EMAIL_EXISTS"
	ErrCodeCannotDeleteSelf   ErrorCode = "USER_CANNOT_DELETE_SELF"
)

// Error codes - Invitation
const (
	ErrCodeInvitationInvalidCode       ErrorCode = "INVITATION_INVALID_CODE"
	ErrCodeInvitationAlreadyRedeemed   ErrorCode = "INVITATION_ALREADY_REDEEMED"
	ErrCodeInvitationNotFound          ErrorCode = "INVITATION_NOT_FOUND"
	ErrCodeInvitationCannotDeleteRedeemed ErrorCode = "INVITATION_CANNOT_DELETE_REDEEMED"
)

// Error codes - Bench
const (
	ErrCodeBenchNotFound  ErrorCode = "BENCH_NOT_FOUND"
	ErrCodeBenchForbidden ErrorCode = "BENCH_FORBIDDEN"
)

// Error codes - Photo
const (
	ErrCodePhotoNotFound   ErrorCode = "PHOTO_NOT_FOUND"
	ErrCodePhotoMaxReached ErrorCode = "PHOTO_MAX_REACHED"
	ErrCodePhotoTooLarge   ErrorCode = "PHOTO_FILE_TOO_LARGE"
	ErrCodePhotoInvalidType ErrorCode = "PHOTO_INVALID_TYPE"
	ErrCodePhotoForbidden  ErrorCode = "PHOTO_FORBIDDEN"
)

// Error codes - Visit
const (
	ErrCodeVisitNotFound  ErrorCode = "VISIT_NOT_FOUND"
	ErrCodeVisitForbidden ErrorCode = "VISIT_FORBIDDEN"
)

// Error codes - Favorite
const (
	ErrCodeFavoriteNotFound     ErrorCode = "FAVORITE_NOT_FOUND"
	ErrCodeFavoriteAlreadyExists ErrorCode = "FAVORITE_ALREADY_EXISTS"
)

// Error codes - Validation
const (
	ErrCodeValidationInvalidRequest ErrorCode = "VALIDATION_INVALID_REQUEST"
	ErrCodeValidationInvalidID      ErrorCode = "VALIDATION_INVALID_ID"
	ErrCodeValidationInvalidEmail   ErrorCode = "VALIDATION_INVALID_EMAIL"
	ErrCodeValidationPasswordShort  ErrorCode = "VALIDATION_PASSWORD_TOO_SHORT"
	ErrCodeValidationFieldRequired  ErrorCode = "VALIDATION_FIELD_REQUIRED"
)

// Error codes - System
const (
	ErrCodeSystemInternal     ErrorCode = "SYSTEM_INTERNAL_ERROR"
	ErrCodeSystemDatabase     ErrorCode = "SYSTEM_DATABASE_ERROR"
	ErrCodeSystemRateLimited  ErrorCode = "SYSTEM_RATE_LIMITED"
)

// ErrorResponse is the JSON response structure for errors
type ErrorResponse struct {
	ErrorCode ErrorCode `json:"error_code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
}

// AppError is a structured error with HTTP status and error code
type AppError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error // underlying error for wrapping
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
	}
}

// WithDetails creates a new AppError with additional details
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    details,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// Predefined AppErrors - Authentication
var (
	AppErrInvalidCredentials  = NewAppError(ErrCodeInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
	AppErrInvalidToken        = NewAppError(ErrCodeInvalidToken, "Invalid token", http.StatusUnauthorized)
	AppErrTokenExpired        = NewAppError(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
	AppErrInvalidRefreshToken = NewAppError(ErrCodeInvalidRefreshToken, "Invalid or expired refresh token", http.StatusUnauthorized)
	AppErrAccountDeactivated  = NewAppError(ErrCodeAccountDeactivated, "Account is deactivated", http.StatusForbidden)
	AppErrForbidden           = NewAppError(ErrCodeForbidden, "Access forbidden", http.StatusForbidden)
	AppErrAdminRequired       = NewAppError(ErrCodeAdminRequired, "Admin access required", http.StatusForbidden)
)

// Predefined AppErrors - User
var (
	AppErrUserNotFound     = NewAppError(ErrCodeUserNotFound, "User not found", http.StatusNotFound)
	AppErrEmailExists      = NewAppError(ErrCodeEmailExists, "Email already exists", http.StatusConflict)
	AppErrCannotDeleteSelf = NewAppError(ErrCodeCannotDeleteSelf, "Admin cannot delete themselves", http.StatusBadRequest)
)

// Predefined AppErrors - Invitation
var (
	AppErrInvitationInvalidCode       = NewAppError(ErrCodeInvitationInvalidCode, "Invalid invitation code", http.StatusBadRequest)
	AppErrInvitationAlreadyRedeemed   = NewAppError(ErrCodeInvitationAlreadyRedeemed, "Invitation code already redeemed", http.StatusBadRequest)
	AppErrInvitationNotFound          = NewAppError(ErrCodeInvitationNotFound, "Invitation code not found", http.StatusNotFound)
	AppErrInvitationCannotDeleteRedeemed = NewAppError(ErrCodeInvitationCannotDeleteRedeemed, "Cannot delete redeemed invitation code", http.StatusBadRequest)
)

// Predefined AppErrors - Bench
var (
	AppErrBenchNotFound  = NewAppError(ErrCodeBenchNotFound, "Bench not found", http.StatusNotFound)
	AppErrBenchForbidden = NewAppError(ErrCodeBenchForbidden, "No permission for this bench", http.StatusForbidden)
)

// Predefined AppErrors - Photo
var (
	AppErrPhotoNotFound    = NewAppError(ErrCodePhotoNotFound, "Photo not found", http.StatusNotFound)
	AppErrPhotoMaxReached  = NewAppError(ErrCodePhotoMaxReached, "Maximum 10 photos per bench reached", http.StatusBadRequest)
	AppErrPhotoTooLarge    = NewAppError(ErrCodePhotoTooLarge, "File size exceeds 10 MB limit", http.StatusBadRequest)
	AppErrPhotoInvalidType = NewAppError(ErrCodePhotoInvalidType, "Only JPEG, PNG, and WebP files allowed", http.StatusBadRequest)
	AppErrPhotoForbidden   = NewAppError(ErrCodePhotoForbidden, "No permission for this photo", http.StatusForbidden)
)

// Predefined AppErrors - Visit
var (
	AppErrVisitNotFound  = NewAppError(ErrCodeVisitNotFound, "Visit not found", http.StatusNotFound)
	AppErrVisitForbidden = NewAppError(ErrCodeVisitForbidden, "Can only delete own visits", http.StatusForbidden)
)

// Predefined AppErrors - Favorite
var (
	AppErrFavoriteNotFound     = NewAppError(ErrCodeFavoriteNotFound, "Favorite not found", http.StatusNotFound)
	AppErrFavoriteAlreadyExists = NewAppError(ErrCodeFavoriteAlreadyExists, "Already in favorites", http.StatusConflict)
)

// Predefined AppErrors - Validation
var (
	AppErrValidationInvalidRequest = NewAppError(ErrCodeValidationInvalidRequest, "Invalid request", http.StatusBadRequest)
	AppErrValidationInvalidID      = NewAppError(ErrCodeValidationInvalidID, "Invalid ID", http.StatusBadRequest)
	AppErrValidationInvalidEmail   = NewAppError(ErrCodeValidationInvalidEmail, "Invalid email address", http.StatusBadRequest)
	AppErrValidationPasswordShort  = NewAppError(ErrCodeValidationPasswordShort, "Password must be at least 8 characters", http.StatusBadRequest)
	AppErrValidationFieldRequired  = NewAppError(ErrCodeValidationFieldRequired, "Required field missing", http.StatusBadRequest)
)

// Predefined AppErrors - System
var (
	AppErrSystemInternal    = NewAppError(ErrCodeSystemInternal, "An unexpected error occurred", http.StatusInternalServerError)
	AppErrSystemDatabase    = NewAppError(ErrCodeSystemDatabase, "Database error", http.StatusInternalServerError)
	AppErrSystemRateLimited = NewAppError(ErrCodeSystemRateLimited, "Too many requests", http.StatusTooManyRequests)
)

// RespondWithError sends a structured error response
func RespondWithError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		response := ErrorResponse{
			ErrorCode: appErr.Code,
			Message:   appErr.Message,
		}
		c.JSON(appErr.HTTPStatus, response)
		return
	}

	// Fallback for unstructured errors
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		ErrorCode: ErrCodeSystemInternal,
		Message:   "An unexpected error occurred",
	})
}

// RespondWithErrorAndDetails sends a structured error response with additional details
func RespondWithErrorAndDetails(c *gin.Context, err error, details string) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		response := ErrorResponse{
			ErrorCode: appErr.Code,
			Message:   appErr.Message,
			Details:   details,
		}
		c.JSON(appErr.HTTPStatus, response)
		return
	}

	// Fallback for unstructured errors
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		ErrorCode: ErrCodeSystemInternal,
		Message:   "An unexpected error occurred",
		Details:   details,
	})
}

// AbortWithError aborts the request and sends a structured error response
func AbortWithError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.HTTPStatus, ErrorResponse{
			ErrorCode: appErr.Code,
			Message:   appErr.Message,
		})
		return
	}

	// Fallback for unstructured errors
	c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
		ErrorCode: ErrCodeSystemInternal,
		Message:   "An unexpected error occurred",
	})
}
