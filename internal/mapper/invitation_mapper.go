package mapper

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
)

func InvitationCodesToResponse(codes []domain.InvitationCode) []responses.InvitationCodeResponse {
	result := make([]responses.InvitationCodeResponse, len(codes))
	for i, code := range codes {
		result[i] = InvitationCodeToResponse(&code)
	}
	return result
}

func InvitationCodeToResponse(code *domain.InvitationCode) responses.InvitationCodeResponse {
	response := responses.InvitationCodeResponse{
		ID:        code.ID,
		Code:      code.Code,
		Comment:   code.Comment,
		CreatedAt: code.CreatedAt,
	}

	// Creator (with nil check)
	if code.Creator != nil {
		response.CreatedBy = UserToResponse(code.Creator)
	}

	// Redeemer (if exists)
	if code.Redeemer != nil {
		redeemerResponse := UserToResponse(code.Redeemer)
		response.RedeemedBy = &redeemerResponse
		response.RedeemedAt = &code.UpdatedAt
	}

	return response
}
