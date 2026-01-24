package apperror

import "errors"

var (
	ErrEmailAlreadyExists            = errors.New("email already exists")
	ErrInvalidInvitationCode         = errors.New("invalid invitation code")
	ErrInvitationCodeAlreadyRedeemed = errors.New("invitation code already redeemed")
)
