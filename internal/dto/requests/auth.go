package requests

type RegisterRequest struct {
	Email          string `json:"email" binding:"required,email,max=255"`
	Password       string `json:"password" binding:"required,min=8,max=100"`
	DisplayName    string `json:"display_name" binding:"required,min=1,max=100"`
	InvitationCode string `json:"invitation_code" binding:"required,len=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshFCMTokenRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
