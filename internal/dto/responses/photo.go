package responses

import "time"

type PhotoResponse struct {
	ID           uint      `json:"id"`
	SpotID       uint      `json:"spot_id"`
	IsMain       bool      `json:"is_main"`
	URLOriginal  string    `json:"url_original,omitempty"`
	URLMedium    string    `json:"url_medium,omitempty"`
	URLThumbnail string    `json:"url_thumbnail,omitempty"`
	UploadedBy   uint      `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}
