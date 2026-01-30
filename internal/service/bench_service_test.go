package service

import (
	"context"
	"testing"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestBenchService_Create_Success(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	description := "A nice bench"
	req := &requests.CreateBenchRequest{
		Name:        "Park Bench",
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
	benchRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Bench")).
		Run(func(ctx context.Context, b *domain.Bench) {
			b.Model = &gorm.Model{ID: 1}
		}).
		Return(nil)

	// Mock FindByID after create (to reload with Creator)
	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(&domain.Bench{
			Model:       &gorm.Model{ID: 1},
			Name:        "Park Bench",
			Latitude:    47.3769,
			Longitude:   8.5417,
			Description: "A nice bench",
			HasToilet:   true,
			HasTrashBin: false,
			CreatedBy:   1,
			Creator:     creator,
		}, nil)

	// NotifyNewBench is called async, but we mock it anyway
	notificationSvc.EXPECT().
		NotifyNewBench(mock.Anything, mock.AnythingOfType("*domain.Bench"), uint(1)).
		Return(nil).
		Maybe() // async call

	// Act
	result, err := svc.Create(context.Background(), req, uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Park Bench", result.Name)
	assert.Equal(t, 47.3769, result.Latitude)
	assert.Equal(t, 8.5417, result.Longitude)
	assert.True(t, result.HasToilet)
	assert.False(t, result.HasTrashBin)
}

func TestBenchService_GetByID_Success(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "creator@example.com",
		DisplayName: "Creator",
	}

	bench := &domain.Bench{
		Model:       &gorm.Model{ID: 1},
		Name:        "Test Bench",
		Latitude:    47.0,
		Longitude:   8.0,
		Description: "Description",
		CreatedBy:   1,
		Creator:     creator,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	// Act
	result, err := svc.GetByID(context.Background(), uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Bench", result.Name)
}

func TestBenchService_GetByID_NotFound(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	result, err := svc.GetByID(context.Background(), uint(999))

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bench not found")
}

func TestBenchService_List_Success(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	benches := []domain.Bench{
		{
			Model:     &gorm.Model{ID: 1},
			Name:      "Bench 1",
			Latitude:  47.0,
			Longitude: 8.0,
		},
		{
			Model:     &gorm.Model{ID: 2},
			Name:      "Bench 2",
			Latitude:  47.1,
			Longitude: 8.1,
		},
	}

	req := &requests.ListBenchesRequest{
		Page:  1,
		Limit: 50,
	}

	benchRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.BenchFilter")).
		Return(benches, int64(2), nil)

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Benches, 2)
	assert.Equal(t, int64(2), result.Pagination.Total)
}

func TestBenchService_List_WithCoordinatesAndRadius(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	// Bench 1: very close (should be included)
	// Bench 2: far away (should be excluded by radius)
	benches := []domain.Bench{
		{
			Model:     &gorm.Model{ID: 1},
			Name:      "Close Bench",
			Latitude:  47.3770, // Very close to search point
			Longitude: 8.5418,
		},
		{
			Model:     &gorm.Model{ID: 2},
			Name:      "Far Bench",
			Latitude:  48.0, // Far from search point
			Longitude: 9.0,
		},
	}

	lat := 47.3769
	lon := 8.5417
	radius := 1000 // 1km radius

	req := &requests.ListBenchesRequest{
		Page:   1,
		Limit:  50,
		Lat:    &lat,
		Lon:    &lon,
		Radius: &radius,
	}

	benchRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.BenchFilter")).
		Return(benches, int64(2), nil)

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Only close bench should be included (far bench is outside 1km)
	assert.Len(t, result.Benches, 1)
	assert.Equal(t, "Close Bench", result.Benches[0].Name)
	assert.NotNil(t, result.Benches[0].Distance)
}

func TestBenchService_Update_Success_AsOwner(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	bench := &domain.Bench{
		Model:       &gorm.Model{ID: 1},
		Name:        "Old Name",
		Description: "Old Description",
		CreatedBy:   1,
		Creator:     creator,
	}

	newName := "New Name"
	req := &requests.UpdateBenchRequest{
		Name: &newName,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	benchRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Bench")).
		Run(func(ctx context.Context, b *domain.Bench) {
			assert.Equal(t, "New Name", b.Name)
		}).
		Return(nil)

	// Act
	result, err := svc.Update(context.Background(), uint(1), req, uint(1), false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.Name)
}

func TestBenchService_Update_Success_AsAdmin(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	bench := &domain.Bench{
		Model:     &gorm.Model{ID: 1},
		Name:      "Old Name",
		CreatedBy: 1, // Created by user 1
		Creator:   creator,
	}

	newName := "Admin Updated Name"
	req := &requests.UpdateBenchRequest{
		Name: &newName,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	benchRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Bench")).
		Return(nil)

	// Act - user 2 is admin
	result, err := svc.Update(context.Background(), uint(1), req, uint(2), true)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Admin Updated Name", result.Name)
}

func TestBenchService_Update_Forbidden(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	creator := domain.User{
		Model:       &gorm.Model{ID: 1},
		DisplayName: "Creator",
	}

	bench := &domain.Bench{
		Model:     &gorm.Model{ID: 1},
		Name:      "Test Bench",
		CreatedBy: 1, // Created by user 1
		Creator:   creator,
	}

	newName := "Hacked Name"
	req := &requests.UpdateBenchRequest{
		Name: &newName,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	// Act - user 2 is NOT admin and NOT owner
	result, err := svc.Update(context.Background(), uint(1), req, uint(2), false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestBenchService_Update_NotFound(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	newName := "Name"
	req := &requests.UpdateBenchRequest{
		Name: &newName,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	result, err := svc.Update(context.Background(), uint(999), req, uint(1), false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bench not found")
}

func TestBenchService_Delete_Success_AsOwner(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	bench := &domain.Bench{
		Model:     &gorm.Model{ID: 1},
		Name:      "Test Bench",
		CreatedBy: 1,
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	benchRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act
	err := svc.Delete(context.Background(), uint(1), uint(1), false)

	// Assert
	assert.NoError(t, err)
}

func TestBenchService_Delete_Success_AsAdmin(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	bench := &domain.Bench{
		Model:     &gorm.Model{ID: 1},
		Name:      "Test Bench",
		CreatedBy: 1, // Owner is user 1
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	benchRepo.EXPECT().
		Delete(mock.Anything, uint(1)).
		Return(nil)

	// Act - user 2 is admin
	err := svc.Delete(context.Background(), uint(1), uint(2), true)

	// Assert
	assert.NoError(t, err)
}

func TestBenchService_Delete_Forbidden(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	bench := &domain.Bench{
		Model:     &gorm.Model{ID: 1},
		Name:      "Test Bench",
		CreatedBy: 1, // Owner is user 1
	}

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(bench, nil)

	// Act - user 2 is NOT admin and NOT owner
	err := svc.Delete(context.Background(), uint(1), uint(2), false)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestBenchService_Delete_NotFound(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	benchRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	err := svc.Delete(context.Background(), uint(999), uint(1), false)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bench not found")
}

func TestBenchService_List_SortByDistance(t *testing.T) {
	// Arrange
	benchRepo := mocks.NewBenchRepository(t)
	notificationSvc := mocks.NewNotificationService(t)
	svc := NewBenchService(benchRepo, notificationSvc)

	// Benches at different distances
	benches := []domain.Bench{
		{
			Model:     &gorm.Model{ID: 1},
			Name:      "Far Bench",
			Latitude:  47.38,
			Longitude: 8.55,
		},
		{
			Model:     &gorm.Model{ID: 2},
			Name:      "Close Bench",
			Latitude:  47.377,
			Longitude: 8.542,
		},
	}

	lat := 47.3769
	lon := 8.5417

	req := &requests.ListBenchesRequest{
		Page:      1,
		Limit:     50,
		Lat:       &lat,
		Lon:       &lon,
		SortBy:    "distance",
		SortOrder: "asc",
	}

	benchRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(f repository.BenchFilter) bool {
			return f.Lat != nil && f.Lon != nil
		})).
		Return(benches, int64(2), nil)

	// Act
	result, err := svc.List(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Benches, 2)
	// Should be sorted by distance ascending - Close Bench first
	assert.Equal(t, "Close Bench", result.Benches[0].Name)
	assert.Equal(t, "Far Bench", result.Benches[1].Name)
}
