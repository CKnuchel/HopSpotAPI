package service

import (
	"context"
	"testing"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/mocks"
	"hopSpotAPI/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestSpotService_Create_Success(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	description := "A nice spot"
	req := &requests.CreateSpotRequest{
		Name:        "Park Spot",
		Latitude:    47.3769,
		Longitude:   8.5417,
		Description: &description,
		HasToilet:   true,
		HasTrashBin: false,
	}

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "creator@example.com",
		DisplayName: "Creator",
		Role:        domain.RoleUser,
	}

	// Mock Create - set ID
	spotRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Spot")).
		Run(func(ctx context.Context, s *domain.Spot) {
			s.ID = 1
		}).
		Return(nil)

	// Mock FindByID after create (to reload with Creator)
	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(&domain.Spot{
			ID:          1,
			Name:        "Park Spot",
			Latitude:    47.3769,
			Longitude:   8.5417,
			Description: "A nice spot",
			HasToilet:   true,
			HasTrashBin: false,
			CreatedBy:   1,
			Creator:     creator,
		}, nil)

	// NotifyNewSpot is called async, but we mock it anyway
	notificationSvc.EXPECT().
		NotifyNewSpot(mock.Anything, mock.AnythingOfType("*domain.Spot"), uint(1)).
		Return(nil).
		Maybe() // async call

	// Act
	result, err := svc.Create(context.Background(), req, uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Park Spot", result.Name)
	assert.Equal(t, 47.3769, result.Latitude)
	assert.Equal(t, 8.5417, result.Longitude)
	assert.True(t, result.HasToilet)
	assert.False(t, result.HasTrashBin)
}

func TestSpotService_GetByID_Success(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "creator@example.com",
		DisplayName: "Creator",
	}

	spot := &domain.Spot{
		ID:          1,
		Name:        "Test Spot",
		Latitude:    47.0,
		Longitude:   8.0,
		Description: "Description",
		CreatedBy:   1,
		Creator:     creator,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(&domain.Photo{
			FilePathThumbnail: "spots/1/photos/1_thumbnail.jpg",
		}, nil)

	// Act
	result, err := svc.GetByID(context.Background(), uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Spot", result.Name)
}

func TestSpotService_GetByID_NotFound(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	result, err := svc.GetByID(context.Background(), uint(999))

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "spot not found")
}

func TestSpotService_List_Success(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spots := []domain.Spot{
		{
			ID:        1,
			Name:      "Spot 1",
			Latitude:  47.0,
			Longitude: 8.0,
		},
		{
			ID:        2,
			Name:      "Spot 2",
			Latitude:  47.1,
			Longitude: 8.1,
		},
	}

	req := &requests.ListSpotsRequest{
		Page:  1,
		Limit: 50,
	}

	spotRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.SpotFilter")).
		Return(spots, int64(2), nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(nil, nil).Maybe()

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(2)).
		Return(nil, nil).Maybe()

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Spots, 2)
	assert.Equal(t, int64(2), result.Pagination.Total)
}

func TestSpotService_List_WithCoordinatesAndRadius(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	// Spot 1: very close (should be included)
	// Spot 2: far away (should be excluded by radius)
	spots := []domain.Spot{
		{
			ID:        1,
			Name:      "Close Spot",
			Latitude:  47.3770, // Very close to search point
			Longitude: 8.5418,
		},
		{
			ID:        2,
			Name:      "Far Spot",
			Latitude:  48.0, // Far from search point
			Longitude: 9.0,
		},
	}

	lat := 47.3769
	lon := 8.5417
	radius := 1000 // 1km radius

	req := &requests.ListSpotsRequest{
		Page:   1,
		Limit:  50,
		Lat:    &lat,
		Lon:    &lon,
		Radius: &radius,
	}

	spotRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.SpotFilter")).
		Return(spots, int64(2), nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(nil, nil).Maybe()

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(2)).
		Return(nil, nil).Maybe()

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Only close spot should be included (far spot is outside 1km)
	assert.Len(t, result.Spots, 1)
	assert.Equal(t, "Close Spot", result.Spots[0].Name)
	assert.NotNil(t, result.Spots[0].Distance)
}

func TestSpotService_Update_Success_AsOwner(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	spot := &domain.Spot{
		ID:          1,
		Name:        "Old Name",
		Description: "Old Description",
		CreatedBy:   1,
		Creator:     creator,
	}

	newName := "New Name"
	req := &requests.UpdateSpotRequest{
		Name: &newName,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	spotRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Spot")).
		Run(func(ctx context.Context, s *domain.Spot) {
			assert.Equal(t, "New Name", s.Name)
		}).
		Return(nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(nil, nil)

	// Act
	result, err := svc.Update(context.Background(), uint(1), req, uint(1), false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.Name)
}

func TestSpotService_Update_Success_AsAdmin(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	spot := &domain.Spot{
		ID:        1,
		Name:      "Old Name",
		CreatedBy: 1, // Created by user 1
		Creator:   creator,
	}

	newName := "Admin Updated Name"
	req := &requests.UpdateSpotRequest{
		Name: &newName,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	spotRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Spot")).
		Return(nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(nil, nil)

	// Act - user 2 is admin
	result, err := svc.Update(context.Background(), uint(1), req, uint(2), true)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Admin Updated Name", result.Name)
}

func TestSpotService_Update_Forbidden(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1, // Created by user 1
		Creator:   creator,
	}

	newName := "Hacked Name"
	req := &requests.UpdateSpotRequest{
		Name: &newName,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	// Act - user 2 is NOT admin and NOT owner
	result, err := svc.Update(context.Background(), uint(1), req, uint(2), false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestSpotService_Update_NotFound(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	newName := "Name"
	req := &requests.UpdateSpotRequest{
		Name: &newName,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	result, err := svc.Update(context.Background(), uint(999), req, uint(1), false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "spot not found")
}

func TestSpotService_Delete_Success_AsOwner(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1,
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	photoRepo.EXPECT().
		FindBySpotIDUnscoped(mock.Anything, uint(1)).
		Return([]domain.Photo{}, nil)

	spotRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act
	err := svc.Delete(context.Background(), uint(1), uint(1), false)

	// Assert
	assert.NoError(t, err)
}

func TestSpotService_Delete_Success_AsAdmin(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1, // Owner is user 1
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	photoRepo.EXPECT().
		FindBySpotIDUnscoped(mock.Anything, uint(1)).
		Return([]domain.Photo{}, nil)

	spotRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act - user 2 is admin
	err := svc.Delete(context.Background(), uint(1), uint(2), true)

	// Assert
	assert.NoError(t, err)
}

func TestSpotService_Delete_WithPhotos_UsesHardDelete(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1,
	}

	photos := []domain.Photo{
		{
			Model:             &gorm.Model{ID: 1},
			SpotID:            1,
			FilePathOriginal:  "photos/1/original.jpg",
			FilePathMedium:    "photos/1/medium.jpg",
			FilePathThumbnail: "photos/1/thumb.jpg",
		},
		{
			Model:             &gorm.Model{ID: 2},
			SpotID:            1,
			FilePathOriginal:  "photos/2/original.jpg",
			FilePathMedium:    "photos/2/medium.jpg",
			FilePathThumbnail: "photos/2/thumb.jpg",
		},
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	photoRepo.EXPECT().
		FindBySpotIDUnscoped(mock.Anything, uint(1)).
		Return(photos, nil)

	// Verify HardDelete is called for each photo
	photoRepo.EXPECT().
		HardDelete(mock.Anything, uint(1)).
		Return(nil).Once()

	photoRepo.EXPECT().
		HardDelete(mock.Anything, uint(2)).
		Return(nil).Once()

	spotRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act
	err := svc.Delete(context.Background(), uint(1), uint(1), false)

	// Assert
	assert.NoError(t, err)
}

func TestSpotService_Delete_WithSoftDeletedPhotos_IncludesAll(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1,
	}

	// Photos include both active and soft-deleted (DeletedAt is set)
	deletedAt := gorm.DeletedAt{Valid: true}
	photos := []domain.Photo{
		{
			Model:             &gorm.Model{ID: 1},
			SpotID:            1,
			FilePathOriginal:  "photos/1/original.jpg",
			FilePathMedium:    "photos/1/medium.jpg",
			FilePathThumbnail: "photos/1/thumb.jpg",
		},
		{
			Model:             &gorm.Model{ID: 2, DeletedAt: deletedAt},
			SpotID:            1,
			FilePathOriginal:  "photos/2/original.jpg",
			FilePathMedium:    "photos/2/medium.jpg",
			FilePathThumbnail: "photos/2/thumb.jpg",
		},
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	// FindBySpotIDUnscoped returns all photos including soft-deleted ones
	photoRepo.EXPECT().
		FindBySpotIDUnscoped(mock.Anything, uint(1)).
		Return(photos, nil)

	// Verify HardDelete is called for both photos (including soft-deleted)
	photoRepo.EXPECT().
		HardDelete(mock.Anything, uint(1)).
		Return(nil).Once()

	photoRepo.EXPECT().
		HardDelete(mock.Anything, uint(2)).
		Return(nil).Once()

	spotRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act
	err := svc.Delete(context.Background(), uint(1), uint(1), false)

	// Assert
	assert.NoError(t, err)
}

func TestSpotService_Delete_Forbidden(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spot := &domain.Spot{
		ID:        1,
		Name:      "Test Spot",
		CreatedBy: 1, // Owner is user 1
	}

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(spot, nil)

	// Act - user 2 is NOT admin and NOT owner
	err := svc.Delete(context.Background(), uint(1), uint(2), false)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestSpotService_Delete_NotFound(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	spotRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	err := svc.Delete(context.Background(), uint(999), uint(1), false)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spot not found")
}

func TestSpotService_List_SortByDistance(t *testing.T) {
	// Arrange
	spotRepo := mocks.NewSpotRepository(t)
	photoRepo := mocks.NewPhotoRepository(t)
	minioClient := &storage.MinioClient{}
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewSpotService(spotRepo, photoRepo, nil, nil, nil, nil, minioClient, notificationSvc, nil)

	// Spots at different distances
	spots := []domain.Spot{
		{
			ID:        1,
			Name:      "Far Spot",
			Latitude:  47.38,
			Longitude: 8.55,
		},
		{
			ID:        2,
			Name:      "Close Spot",
			Latitude:  47.377,
			Longitude: 8.542,
		},
	}

	lat := 47.3769
	lon := 8.5417

	req := &requests.ListSpotsRequest{
		Page:      1,
		Limit:     50,
		Lat:       &lat,
		Lon:       &lon,
		SortBy:    "distance",
		SortOrder: "asc",
	}

	spotRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(f repository.SpotFilter) bool {
			return f.Lat != nil && f.Lon != nil
		})).
		Return(spots, int64(2), nil)

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(1)).
		Return(nil, nil).Maybe()

	photoRepo.EXPECT().
		GetMainPhoto(mock.Anything, uint(2)).
		Return(nil, nil).Maybe()

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Spots, 2)
	// Should be sorted by distance ascending - Close Spot first
	assert.Equal(t, "Close Spot", result.Spots[0].Name)
	assert.Equal(t, "Far Spot", result.Spots[1].Name)
}
