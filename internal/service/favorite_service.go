package service

import (
	"context"
	"math"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/logger"
	"hopSpotAPI/pkg/storage"
)

type FavoriteService interface {
	Add(ctx context.Context, userID, spotID uint) error
	Remove(ctx context.Context, userID, spotID uint) error
	IsFavorite(ctx context.Context, userID, spotID uint) (bool, error)
	List(ctx context.Context, userID uint, page, limit int) (*responses.PaginatedFavoritesResponse, error)
	GetFavoriteSpotIDs(ctx context.Context, userID uint) ([]uint, error)
}

type favoriteService struct {
	favoriteRepo    repository.FavoriteRepository
	spotRepo        repository.SpotRepository
	photoRepo       repository.PhotoRepository
	minioClient     *storage.MinioClient
	activityService ActivityService
}

func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	spotRepo repository.SpotRepository,
	photoRepo repository.PhotoRepository,
	minioClient *storage.MinioClient,
	activityService ActivityService,
) FavoriteService {
	return &favoriteService{
		favoriteRepo:    favoriteRepo,
		spotRepo:        spotRepo,
		photoRepo:       photoRepo,
		minioClient:     minioClient,
		activityService: activityService,
	}
}

func (s *favoriteService) Add(ctx context.Context, userID, spotID uint) error {
	// Check if already favorited
	exists, err := s.favoriteRepo.Exists(ctx, userID, spotID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already favorited, no error
	}

	favorite := &domain.Favorite{
		UserID: userID,
		SpotID: spotID,
	}

	if err := s.favoriteRepo.Create(ctx, favorite); err != nil {
		return err
	}

	// Create activity for favorite (async)
	go func() {
		sid := spotID
		if err := s.activityService.Create(context.Background(), userID, domain.ActionFavoriteAdded, &sid); err != nil {
			logger.Warn().Err(err).Uint("spotID", spotID).Msg("failed to create favorite_added activity")
		}
	}()

	return nil
}

func (s *favoriteService) Remove(ctx context.Context, userID, spotID uint) error {
	return s.favoriteRepo.Delete(ctx, userID, spotID)
}

func (s *favoriteService) IsFavorite(ctx context.Context, userID, spotID uint) (bool, error) {
	return s.favoriteRepo.Exists(ctx, userID, spotID)
}

func (s *favoriteService) List(ctx context.Context, userID uint, page, limit int) (*responses.PaginatedFavoritesResponse, error) {
	filter := repository.FavoriteFilter{
		Page:  page,
		Limit: limit,
	}

	favorites, total, err := s.favoriteRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	// Map to responses
	favoriteResponses := make([]responses.FavoriteResponse, len(favorites))
	for i, fav := range favorites {
		favoriteResponses[i] = responses.FavoriteResponse{
			ID:        fav.ID,
			CreatedAt: fav.CreatedAt,
			Spot: responses.FavoriteSpotResponse{
				ID:          fav.Spot.ID,
				Name:        fav.Spot.Name,
				Latitude:    fav.Spot.Latitude,
				Longitude:   fav.Spot.Longitude,
				Rating:      fav.Spot.Rating,
				HasToilet:   fav.Spot.HasToilet,
				HasTrashBin: fav.Spot.HasTrashBin,
			},
		}
		// Get main photo URL
		favoriteResponses[i].Spot.MainPhotoURL = s.getMainPhotoURL(ctx, fav.SpotID)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &responses.PaginatedFavoritesResponse{
		Favorites: favoriteResponses,
		Pagination: responses.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *favoriteService) GetFavoriteSpotIDs(ctx context.Context, userID uint) ([]uint, error) {
	return s.favoriteRepo.GetSpotIDsByUserID(ctx, userID)
}

func (s *favoriteService) getMainPhotoURL(ctx context.Context, spotID uint) *string {
	mainPhoto, err := s.photoRepo.GetMainPhoto(ctx, spotID)
	if err != nil || mainPhoto == nil {
		return nil
	}

	url := s.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)
	return &url
}
