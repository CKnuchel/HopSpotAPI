package service

import (
	"bytes"
	"context"
	"fmt"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/utils"
	"mime/multipart"
	"time"
)

const (
	MaxPhotosPerBench = 10
	MaxFileSize       = 10 * 1024 * 1024 // 10 MB
)

type PhotoService interface {
	Upload(ctx context.Context, benchID uint, userID uint, file *multipart.FileHeader, isMain bool) (*responses.PhotoResponse, error)
	Delete(ctx context.Context, photoID uint, userID uint, isAdmin bool) error
	SetMainPhoto(ctx context.Context, photoID uint, userID uint, isAdmin bool) error
	GetByBenchID(ctx context.Context, benchID uint) ([]responses.PhotoResponse, error)
	GetPresignedURL(ctx context.Context, photoID uint, size string) (string, error)
}

type photoService struct {
	photoRepo   repository.PhotoRepository
	benchRepo   repository.BenchRepository
	minioClient *storage.MinioClient
}

func NewPhotoService(photoRepo repository.PhotoRepository, benchRepo repository.BenchRepository, minioClient *storage.MinioClient) PhotoService {
	return &photoService{
		photoRepo:   photoRepo,
		benchRepo:   benchRepo,
		minioClient: minioClient,
	}
}

func (s *photoService) Upload(ctx context.Context, benchID uint, userID uint, file *multipart.FileHeader, isMain bool) (*responses.PhotoResponse, error) {
	// Check if the referenced bench exists
	bench, err := s.benchRepo.FindByID(ctx, benchID)
	if err != nil {
		return nil, err
	}
	if bench == nil {
		return nil, apperror.ErrBenchNotFound
	}

	// Check the number of existing photos for the bench
	count, err := s.photoRepo.CountByBenchID(ctx, benchID)
	if err != nil {
		return nil, err
	}
	if count >= MaxPhotosPerBench {
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
		BenchID:    benchID,
		UploadedBy: userID,
		IsMain:     isMain,
		MimeType:   "image/jpeg",
		FileSize:   len(processed.Original),
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, err
	}

	// Generating storage paths
	pathOriginal := utils.GeneratePhotoPath(benchID, photo.ID, "original")
	pathMedium := utils.GeneratePhotoPath(benchID, photo.ID, "medium")
	pathThumbnail := utils.GeneratePhotoPath(benchID, photo.ID, "thumbnail")

	// Uploading the images to MinIO
	if err := s.minioClient.Upload(ctx, pathOriginal, bytes.NewReader(processed.Original), int64(len(processed.Original)), "image/jpeg"); err != nil {
		// Cleanup: delete the photo record if upload fails
		s.photoRepo.Delete(ctx, photo.ID)
		return nil, fmt.Errorf("failed to upload original: %w", err)
	}

	if err := s.minioClient.Upload(ctx, pathMedium, bytes.NewReader(processed.Medium), int64(len(processed.Medium)), "image/jpeg"); err != nil {
		s.photoRepo.Delete(ctx, photo.ID)
		s.minioClient.Delete(ctx, pathOriginal)
		return nil, fmt.Errorf("failed to upload medium: %w", err)
	}

	if err := s.minioClient.Upload(ctx, pathThumbnail, bytes.NewReader(processed.Thumbnail), int64(len(processed.Thumbnail)), "image/jpeg"); err != nil {
		s.photoRepo.Delete(ctx, photo.ID)
		s.minioClient.Delete(ctx, pathOriginal)
		s.minioClient.Delete(ctx, pathMedium)
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
		s.photoRepo.SetMainPhoto(ctx, photo.ID, benchID)
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
		s.minioClient.Delete(ctx, photo.FilePathOriginal)
	}
	if photo.FilePathMedium != "" {
		s.minioClient.Delete(ctx, photo.FilePathMedium)
	}
	if photo.FilePathThumbnail != "" {
		s.minioClient.Delete(ctx, photo.FilePathThumbnail)
	}

	// Delete from database
	if err := s.photoRepo.Delete(ctx, photoID); err != nil {
		return err
	}

	// If it was the main photo, set another as main
	if photo.IsMain {
		photos, err := s.photoRepo.FindByBenchID(ctx, photo.BenchID)
		if err == nil && len(photos) > 0 {
			s.photoRepo.SetMainPhoto(ctx, photos[0].ID, photo.BenchID)
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

	// Fetch the bench
	bench, err := s.benchRepo.FindByID(ctx, photo.BenchID)
	if err != nil {
		return err
	}
	if bench == nil {
		return apperror.ErrBenchNotFound
	}

	// Authorization check: only admin or bench owner can set main photo
	if bench.CreatedBy != userID && !isAdmin {
		return apperror.ErrForbidden
	}

	return s.photoRepo.SetMainPhoto(ctx, photoID, photo.BenchID)
}

// GetByBenchID implements PhotoService.
func (s *photoService) GetByBenchID(ctx context.Context, benchID uint) ([]responses.PhotoResponse, error) {
	photos, err := s.photoRepo.FindByBenchID(ctx, benchID)
	if err != nil {
		return nil, err
	}

	return mapper.PhotosToResponse(photos), nil
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
