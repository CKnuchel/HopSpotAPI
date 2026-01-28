package mapper

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
)

func PhotoToResponse(photo *domain.Photo) *responses.PhotoResponse {
	return &responses.PhotoResponse{
		ID:           photo.ID,
		BenchID:      photo.BenchID,
		IsMain:       photo.IsMain,
		URLOriginal:  photo.FilePathOriginal,
		URLMedium:    photo.FilePathMedium,
		URLThumbnail: photo.FilePathThumbnail,
		UploadedBy:   photo.UploadedBy,
		CreatedAt:    photo.CreatedAt,
	}
}

func PhotosToResponse(photos []domain.Photo) []responses.PhotoResponse {
	result := make([]responses.PhotoResponse, len(photos))
	for i, photo := range photos {
		result[i] = *PhotoToResponse(&photo)
	}
	return result
}
