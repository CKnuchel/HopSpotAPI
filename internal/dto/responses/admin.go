package responses

import "time"

type PaginatedUsersResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

type InvitationCodeResponse struct {
	ID         uint          `json:"id"`
	Code       string        `json:"code"`
	Comment    string        `json:"comment,omitempty"`
	CreatedBy  UserResponse  `json:"created_by"`
	RedeemedBy *UserResponse `json:"redeemed_by,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	RedeemedAt *time.Time    `json:"redeemed_at,omitempty"`
}

type PaginatedInvitationCodesResponse struct {
	Codes      []InvitationCodeResponse `json:"codes"`
	Pagination PaginationResponse       `json:"pagination"`
}
