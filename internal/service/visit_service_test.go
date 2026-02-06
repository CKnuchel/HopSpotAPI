package service

import (
	"context"
	"testing"
	"time"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/mocks"
	"hopSpotAPI/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper function to create a test visit service with nil photo dependencies
// Photo URL will be nil in all test responses, which is acceptable for testing core visit functionality
func newTestVisitService(t *testing.T, visitRepo *mocks.VisitRepository) VisitService {
	photoRepo := mocks.NewPhotoRepository(t)
	// Return nil for all GetMainPhoto calls - no photos in tests
	photoRepo.EXPECT().GetMainPhoto(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	return NewVisitService(visitRepo, photoRepo, (*storage.MinioClient)(nil), nil)
}

func TestVisitService_Create_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	req := &requests.CreateVisitRequest{
		SpotID:  1,
		Comment: "Nice spot!",
	}

	visitRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Visit")).
		Run(func(ctx context.Context, v *domain.Visit) {
			assert.Equal(t, uint(1), v.SpotID)
			assert.Equal(t, uint(5), v.UserID)
			assert.Equal(t, "Nice spot!", v.Comment)
			// Simulate DB setting ID and timestamps
			v.Model = &gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
			}
		}).
		Return(nil)

	// Mock FindByID to return the visit with preloaded spot
	visitRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(&domain.Visit{
			Model:   &gorm.Model{ID: 1, CreatedAt: time.Now()},
			SpotID:  1,
			UserID:  5,
			Comment: "Nice spot!",
			Spot: domain.Spot{
				ID:   1,
				Name: "Test Spot",
			},
		}, nil)

	// Act
	result, err := svc.Create(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Spot.ID)
}

func TestVisitService_Create_WithoutComment(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	req := &requests.CreateVisitRequest{
		SpotID:  2,
		Comment: "",
	}

	visitRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Visit")).
		Run(func(ctx context.Context, v *domain.Visit) {
			assert.Equal(t, uint(2), v.SpotID)
			assert.Equal(t, uint(3), v.UserID)
			assert.Empty(t, v.Comment)
			v.Model = &gorm.Model{
				ID:        2,
				CreatedAt: time.Now(),
			}
		}).
		Return(nil)

	// Mock FindByID to return the visit with preloaded spot
	visitRepo.EXPECT().
		FindByID(mock.Anything, uint(2)).
		Return(&domain.Visit{
			Model:  &gorm.Model{ID: 2, CreatedAt: time.Now()},
			SpotID: 2,
			UserID: 3,
			Spot: domain.Spot{
				ID:   2,
				Name: "Another Spot",
			},
		}, nil)

	// Act
	result, err := svc.Create(context.Background(), req, uint(3))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestVisitService_List_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	visits := []domain.Visit{
		{
			Model:   &gorm.Model{ID: 1, CreatedAt: time.Now()},
			SpotID:  1,
			UserID:  5,
			Comment: "First visit",
			Spot: domain.Spot{
				ID:   1,
				Name: "Spot 1",
			},
		},
		{
			Model:   &gorm.Model{ID: 2, CreatedAt: time.Now()},
			SpotID:  2,
			UserID:  5,
			Comment: "Second visit",
			Spot: domain.Spot{
				ID:   2,
				Name: "Spot 2",
			},
		},
	}

	req := &requests.ListVisitsRequest{
		Page:  1,
		Limit: 50,
	}

	visitRepo.EXPECT().
		FindByUserID(mock.Anything, uint(5), mock.AnythingOfType("repository.VisitFilter")).
		Return(visits, int64(2), nil)

	// Act
	result, err := svc.List(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Visits, 2)
	assert.Equal(t, int64(2), result.Pagination.Total)
	assert.Equal(t, 1, result.Pagination.Page)
}

func TestVisitService_List_WithSpotFilter(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	spotID := uint(1)
	visits := []domain.Visit{
		{
			Model:   &gorm.Model{ID: 1, CreatedAt: time.Now()},
			SpotID:  1,
			UserID:  5,
			Comment: "Visit to spot 1",
			Spot: domain.Spot{
				ID:   1,
				Name: "Spot 1",
			},
		},
	}

	req := &requests.ListVisitsRequest{
		Page:   1,
		Limit:  50,
		SpotID: &spotID,
	}

	visitRepo.EXPECT().
		FindByUserID(mock.Anything, uint(5), mock.MatchedBy(func(f repository.VisitFilter) bool {
			return f.SpotID != nil && *f.SpotID == uint(1)
		})).
		Return(visits, int64(1), nil)

	// Act
	result, err := svc.List(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Visits, 1)
	assert.Equal(t, int64(1), result.Pagination.Total)
}

func TestVisitService_List_Pagination(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	// Page 2 of results (55 total, 50 per page = 5 on page 2)
	visits := make([]domain.Visit, 5)
	for i := 0; i < 5; i++ {
		visits[i] = domain.Visit{
			Model:  &gorm.Model{ID: uint(51 + i), CreatedAt: time.Now()},
			SpotID: uint(i + 1),
			UserID: 5,
			Spot: domain.Spot{
				ID:   uint(i + 1),
				Name: "Spot",
			},
		}
	}

	req := &requests.ListVisitsRequest{
		Page:  2,
		Limit: 50,
	}

	visitRepo.EXPECT().
		FindByUserID(mock.Anything, uint(5), mock.MatchedBy(func(f repository.VisitFilter) bool {
			return f.Page == 2 && f.Limit == 50
		})).
		Return(visits, int64(55), nil)

	// Act
	result, err := svc.List(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Visits, 5)
	assert.Equal(t, int64(55), result.Pagination.Total)
	assert.Equal(t, 2, result.Pagination.TotalPages) // 55/50 = 2 pages
	assert.Equal(t, 2, result.Pagination.Page)
}

func TestVisitService_List_Empty(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	req := &requests.ListVisitsRequest{
		Page:  1,
		Limit: 50,
	}

	visitRepo.EXPECT().
		FindByUserID(mock.Anything, uint(5), mock.AnythingOfType("repository.VisitFilter")).
		Return([]domain.Visit{}, int64(0), nil)

	// Act
	result, err := svc.List(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Visits, 0)
	assert.Equal(t, int64(0), result.Pagination.Total)
}

func TestVisitService_GetCountBySpotID_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	visitRepo.EXPECT().
		CountBySpotID(mock.Anything, uint(1)).
		Return(int64(42), nil)

	// Act
	count, err := svc.GetCountBySpotID(context.Background(), uint(1))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestVisitService_GetCountBySpotID_Zero(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := newTestVisitService(t, visitRepo)

	visitRepo.EXPECT().
		CountBySpotID(mock.Anything, uint(999)).
		Return(int64(0), nil)

	// Act
	count, err := svc.GetCountBySpotID(context.Background(), uint(999))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
