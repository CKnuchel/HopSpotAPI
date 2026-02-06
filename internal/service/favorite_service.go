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
	Add(ctx context.Context, userID, benchID uint) error
	Remove(ctx context.Context, userID, benchID uint) error
	IsFavorite(ctx context.Context, userID, benchID uint) (bool, error)
	List(ctx context.Context, userID uint, page, limit int) (*responses.PaginatedFavoritesResponse, error)
	GetFavoriteBenchIDs(ctx context.Context, userID uint) ([]uint, error)
}

type favoriteService struct {
	favoriteRepo    repository.FavoriteRepository
	benchRepo       repository.BenchRepository
	photoRepo       repository.PhotoRepository
	minioClient     *storage.MinioClient
	activityService ActivityService
}

func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	benchRepo repository.BenchRepository,
	photoRepo repository.PhotoRepository,
	minioClient *storage.MinioClient,
	activityService ActivityService,
) FavoriteService {
	return &favoriteService{
		favoriteRepo:    favoriteRepo,
		benchRepo:       benchRepo,
		photoRepo:       photoRepo,
		minioClient:     minioClient,
		activityService: activityService,
	}
}

func (s *favoriteService) Add(ctx context.Context, userID, benchID uint) error {
	// Check if already favorited
	exists, err := s.favoriteRepo.Exists(ctx, userID, benchID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already favorited, no error
	}

	favorite := &domain.Favorite{
		UserID:  userID,
		BenchID: benchID,
	}

	if err := s.favoriteRepo.Create(ctx, favorite); err != nil {
		return err
	}

	// Create activity for favorite (async)
	go func() {
		bid := benchID
		if err := s.activityService.Create(context.Background(), userID, domain.ActionFavoriteAdded, &bid); err != nil {
			logger.Warn().Err(err).Uint("benchID", benchID).Msg("failed to create favorite_added activity")
		}
	}()

	return nil
}

func (s *favoriteService) Remove(ctx context.Context, userID, benchID uint) error {
	return s.favoriteRepo.Delete(ctx, userID, benchID)
}

func (s *favoriteService) IsFavorite(ctx context.Context, userID, benchID uint) (bool, error) {
	return s.favoriteRepo.Exists(ctx, userID, benchID)
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
			Bench: responses.FavoriteBenchResponse{
				ID:          fav.Bench.ID,
				Name:        fav.Bench.Name,
				Latitude:    fav.Bench.Latitude,
				Longitude:   fav.Bench.Longitude,
				Rating:      fav.Bench.Rating,
				HasToilet:   fav.Bench.HasToilet,
				HasTrashBin: fav.Bench.HasTrashBin,
			},
		}
		// Get main photo URL
		favoriteResponses[i].Bench.MainPhotoURL = s.getMainPhotoURL(ctx, fav.BenchID)
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

func (s *favoriteService) GetFavoriteBenchIDs(ctx context.Context, userID uint) ([]uint, error) {
	return s.favoriteRepo.GetBenchIDsByUserID(ctx, userID)
}

func (s *favoriteService) getMainPhotoURL(ctx context.Context, benchID uint) *string {
	mainPhoto, err := s.photoRepo.GetMainPhoto(ctx, benchID)
	if err != nil || mainPhoto == nil {
		return nil
	}

	url := s.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)
	return &url
}
