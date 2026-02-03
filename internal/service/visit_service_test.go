package service

import (
	"context"
	"testing"
	"time"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestVisitService_Create_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	req := &requests.CreateVisitRequest{
		BenchID: 1,
		Comment: "Nice bench!",
	}

	visitRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Visit")).
		Run(func(ctx context.Context, v *domain.Visit) {
			assert.Equal(t, uint(1), v.BenchID)
			assert.Equal(t, uint(5), v.UserID)
			assert.Equal(t, "Nice bench!", v.Comment)
			// Simulate DB setting ID and timestamps
			v.Model = &gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
			}
			// Simulate DB preloading Bench
			v.Bench = domain.Bench{
				ID:   1,
				Name: "Test Bench",
			}
		}).
		Return(nil)

	// Act
	result, err := svc.Create(context.Background(), req, uint(5))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Bench.ID)
}

func TestVisitService_Create_WithoutComment(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	req := &requests.CreateVisitRequest{
		BenchID: 2,
		Comment: "",
	}

	visitRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Visit")).
		Run(func(ctx context.Context, v *domain.Visit) {
			assert.Equal(t, uint(2), v.BenchID)
			assert.Equal(t, uint(3), v.UserID)
			assert.Empty(t, v.Comment)
			v.Model = &gorm.Model{
				ID:        2,
				CreatedAt: time.Now(),
			}
			// Simulate DB preloading Bench
			v.Bench = domain.Bench{
				ID:   2,
				Name: "Another Bench",
			}
		}).
		Return(nil)

	// Act
	result, err := svc.Create(context.Background(), req, uint(3))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestVisitService_List_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	visits := []domain.Visit{
		{
			Model:   &gorm.Model{ID: 1, CreatedAt: time.Now()},
			BenchID: 1,
			UserID:  5,
			Comment: "First visit",
			Bench: domain.Bench{
				ID:   1,
				Name: "Bench 1",
			},
		},
		{
			Model:   &gorm.Model{ID: 2, CreatedAt: time.Now()},
			BenchID: 2,
			UserID:  5,
			Comment: "Second visit",
			Bench: domain.Bench{
				ID:   2,
				Name: "Bench 2",
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

func TestVisitService_List_WithBenchFilter(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	benchID := uint(1)
	visits := []domain.Visit{
		{
			Model:   &gorm.Model{ID: 1, CreatedAt: time.Now()},
			BenchID: 1,
			UserID:  5,
			Comment: "Visit to bench 1",
			Bench: domain.Bench{
				ID:   1,
				Name: "Bench 1",
			},
		},
	}

	req := &requests.ListVisitsRequest{
		Page:    1,
		Limit:   50,
		BenchID: &benchID,
	}

	visitRepo.EXPECT().
		FindByUserID(mock.Anything, uint(5), mock.MatchedBy(func(f repository.VisitFilter) bool {
			return f.BenchID != nil && *f.BenchID == uint(1)
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
	svc := NewVisitService(visitRepo)

	// Page 2 of results (55 total, 50 per page = 5 on page 2)
	visits := make([]domain.Visit, 5)
	for i := 0; i < 5; i++ {
		visits[i] = domain.Visit{
			Model:   &gorm.Model{ID: uint(51 + i), CreatedAt: time.Now()},
			BenchID: uint(i + 1),
			UserID:  5,
			Bench: domain.Bench{
				ID:   uint(i + 1),
				Name: "Bench",
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
	svc := NewVisitService(visitRepo)

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

func TestVisitService_GetCountByBenchID_Success(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	visitRepo.EXPECT().
		CountByBenchID(mock.Anything, uint(1)).
		Return(int64(42), nil)

	// Act
	count, err := svc.GetCountByBenchID(context.Background(), uint(1))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestVisitService_GetCountByBenchID_Zero(t *testing.T) {
	// Arrange
	visitRepo := mocks.NewVisitRepository(t)
	svc := NewVisitService(visitRepo)

	visitRepo.EXPECT().
		CountByBenchID(mock.Anything, uint(999)).
		Return(int64(0), nil)

	// Act
	count, err := svc.GetCountByBenchID(context.Background(), uint(999))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
