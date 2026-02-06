package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/logger"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/utils"
)

const (
	MaxPhotosPerSpot = 10
	MaxFileSize      = 10 * 1024 * 1024 // 10 MB
)

type PhotoService interface {
	Upload(ctx context.Context, spotID uint, userID uint, file *multipart.FileHeader, isMain bool) (*responses.PhotoResponse, error)
	Delete(ctx context.Context, photoID uint, userID uint, isAdmin bool) error
	SetMainPhoto(ctx context.Context, photoID uint, userID uint, isAdmin bool) error
	GetBySpotID(ctx context.Context, spotID uint) ([]responses.PhotoResponse, error)
	GetPresignedURL(ctx context.Context, photoID uint, size string) (string, error)
}

type photoService struct {
	photoRepo   repository.PhotoRepository
	spotRepo    repository.SpotRepository
	minioClient *storage.MinioClient
}

func NewPhotoService(photoRepo repository.PhotoRepository, spotRepo repository.SpotRepository, minioClient *storage.MinioClient) PhotoService {
	return &photoService{
		photoRepo:   photoRepo,
		spotRepo:    spotRepo,
		minioClient: minioClient,
	}
}

func (s *photoService) Upload(ctx context.Context, spotID uint, userID uint, file *multipart.FileHeader, isMain bool) (*responses.PhotoResponse, error) {
	// Check if the referenced spot exists
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, err
	}
	if spot == nil {
		return nil, apperror.ErrSpotNotFound
	}

	// Check the number of existing photos for the spot
	count, err := s.photoRepo.CountBySpotID(ctx, spotID)
	if err != nil {
		return nil, err
	}
	if count >= MaxPhotosPerSpot {
		return nil, apperror.ErrMaxPhotosReached
	}

	// Checking file size
	if file.Size > MaxFileSize {
		return nil, apperror.ErrFileTooLarge
	}

	// Checking MIME type
	if !utils.ValidateImageType(file.Header.Get("Content-Type")) {
		return nil, apperror.ErrInvalidFileType
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Creating the photo versions
	processed, err := utils.ProcessImage(src)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Creating the photo record to get the ID
	photo := &domain.Photo{
		SpotID:     spotID,
		UploadedBy: userID,
		IsMain:     isMain,
		MimeType:   "image/jpeg",
		FileSize:   len(processed.Original),
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, err
	}

	// Generating storage paths
	pathOriginal := utils.GeneratePhotoPath(spotID, photo.ID, "original")
	pathMedium := utils.GeneratePhotoPath(spotID, photo.ID, "medium")
	pathThumbnail := utils.GeneratePhotoPath(spotID, photo.ID, "thumbnail")

	// Uploading the images to MinIO
	if err := s.minioClient.Upload(ctx, pathOriginal, bytes.NewReader(processed.Original), int64(len(processed.Original)), "image/jpeg"); err != nil {
		// Cleanup: delete the photo record if upload fails
		if cleanupErr := s.photoRepo.Delete(ctx, photo.ID); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Uint("photoID", photo.ID).Msg("cleanup: failed to delete photo record")
		}
		return nil, fmt.Errorf("failed to upload original: %w", err)
	}

	if err := s.minioClient.Upload(ctx, pathMedium, bytes.NewReader(processed.Medium), int64(len(processed.Medium)), "image/jpeg"); err != nil {
		if cleanupErr := s.photoRepo.Delete(ctx, photo.ID); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Uint("photoID", photo.ID).Msg("cleanup: failed to delete photo record")
		}
		if cleanupErr := s.minioClient.Delete(ctx, pathOriginal); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Str("path", pathOriginal).Msg("cleanup: failed to delete original from storage")
		}
		return nil, fmt.Errorf("failed to upload medium: %w", err)
	}

	if err := s.minioClient.Upload(ctx, pathThumbnail, bytes.NewReader(processed.Thumbnail), int64(len(processed.Thumbnail)), "image/jpeg"); err != nil {
		if cleanupErr := s.photoRepo.Delete(ctx, photo.ID); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Uint("photoID", photo.ID).Msg("cleanup: failed to delete photo record")
		}
		if cleanupErr := s.minioClient.Delete(ctx, pathOriginal); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Str("path", pathOriginal).Msg("cleanup: failed to delete original from storage")
		}
		if cleanupErr := s.minioClient.Delete(ctx, pathMedium); cleanupErr != nil {
			logger.Warn().Err(cleanupErr).Str("path", pathMedium).Msg("cleanup: failed to delete medium from storage")
		}
		return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Saving paths to the photo record
	photo.FilePathOriginal = pathOriginal
	photo.FilePathMedium = pathMedium
	photo.FilePathThumbnail = pathThumbnail

	if err := s.photoRepo.Update(ctx, photo); err != nil {
		return nil, err
	}

	// If it's the first photo, set it as main
	if isMain || count == 0 {
		if err := s.photoRepo.SetMainPhoto(ctx, photo.ID, spotID); err != nil {
			logger.Warn().Err(err).Uint("photoID", photo.ID).Msg("failed to set main photo")
		}
		photo.IsMain = true
	}

	return mapper.PhotoToResponse(photo), nil
}

// Delete implements PhotoService.
func (s *photoService) Delete(ctx context.Context, photoID uint, userID uint, isAdmin bool) error {
	// Fetch the photo
	photo, err := s.photoRepo.FindByID(ctx, photoID)
	if err != nil {
		return err
	}
	if photo == nil {
		return apperror.ErrPhotoNotFound
	}

	// Authorization check: only admin or uploader can delete
	if photo.UploadedBy != userID && !isAdmin {
		return apperror.ErrForbidden
	}

	// Delete from storage
	if photo.FilePathOriginal != "" {
		if err := s.minioClient.Delete(ctx, photo.FilePathOriginal); err != nil {
			logger.Warn().Err(err).Str("path", photo.FilePathOriginal).Msg("failed to delete original from storage")
		}
	}
	if photo.FilePathMedium != "" {
		if err := s.minioClient.Delete(ctx, photo.FilePathMedium); err != nil {
			logger.Warn().Err(err).Str("path", photo.FilePathMedium).Msg("failed to delete medium from storage")
		}
	}
	if photo.FilePathThumbnail != "" {
		if err := s.minioClient.Delete(ctx, photo.FilePathThumbnail); err != nil {
			logger.Warn().Err(err).Str("path", photo.FilePathThumbnail).Msg("failed to delete thumbnail from storage")
		}
	}

	// Delete from database
	if err := s.photoRepo.Delete(ctx, photoID); err != nil {
		return err
	}

	// If it was the main photo, set another as main
	if photo.IsMain {
		photos, err := s.photoRepo.FindBySpotID(ctx, photo.SpotID)
		if err == nil && len(photos) > 0 {
			if err := s.photoRepo.SetMainPhoto(ctx, photos[0].ID, photo.SpotID); err != nil {
				logger.Warn().Err(err).Uint("photoID", photos[0].ID).Msg("failed to set new main photo")
			}
		}
	}

	return nil
}

// SetMainPhoto implements PhotoService.
func (s *photoService) SetMainPhoto(ctx context.Context, photoID uint, userID uint, isAdmin bool) error {
	// Fetch the photo
	photo, err := s.photoRepo.FindByID(ctx, photoID)
	if err != nil {
		return err
	}
	if photo == nil {
		return apperror.ErrPhotoNotFound
	}

	// Fetch the spot
	spot, err := s.spotRepo.FindByID(ctx, photo.SpotID)
	if err != nil {
		return err
	}
	if spot == nil {
		return apperror.ErrSpotNotFound
	}

	// Authorization check: only admin or spot owner can set main photo
	if spot.CreatedBy != userID && !isAdmin {
		return apperror.ErrForbidden
	}

	return s.photoRepo.SetMainPhoto(ctx, photoID, photo.SpotID)
}

// GetBySpotID implements PhotoService.
func (s *photoService) GetBySpotID(ctx context.Context, spotID uint) ([]responses.PhotoResponse, error) {
	photos, err := s.photoRepo.FindBySpotID(ctx, spotID)
	if err != nil {
		return nil, err
	}

	// Generate public URLs for each photo
	result := make([]responses.PhotoResponse, len(photos))
	for i, photo := range photos {
		result[i] = responses.PhotoResponse{
			ID:           photo.ID,
			SpotID:       photo.SpotID,
			IsMain:       photo.IsMain,
			URLOriginal:  s.minioClient.GetPublicURL(photo.FilePathOriginal),
			URLMedium:    s.minioClient.GetPublicURL(photo.FilePathMedium),
			URLThumbnail: s.minioClient.GetPublicURL(photo.FilePathThumbnail),
			UploadedBy:   photo.UploadedBy,
			CreatedAt:    photo.CreatedAt,
		}
	}

	return result, nil
}

// GetPresignedURL implements PhotoService.
func (s *photoService) GetPresignedURL(ctx context.Context, photoID uint, size string) (string, error) {
	// Fetch the photo
	photo, err := s.photoRepo.FindByID(ctx, photoID)
	if err != nil {
		return "", err
	}
	if photo == nil {
		return "", apperror.ErrPhotoNotFound
	}

	// Determine the file path based on size
	var path string
	switch size {
	case "original":
		path = photo.FilePathOriginal
	case "medium":
		path = photo.FilePathMedium
	case "thumbnail":
		path = photo.FilePathThumbnail
	default:
		path = photo.FilePathMedium
	}

	// Generate presigned URL, Expires in 1 hour
	return s.minioClient.GetPresignedURL(ctx, path, time.Hour)
}
