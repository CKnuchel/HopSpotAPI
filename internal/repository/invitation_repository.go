package repository

import (
	"context"
	"hopSpotAPI/internal/domain"

	"gorm.io/gorm"
)

type invitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository Constructor for InvitationRepository
func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r invitationRepository) Create(ctx context.Context, code *domain.InvitationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r invitationRepository) FindByID(ctx context.Context, id uint) (*domain.InvitationCode, error) {
	var code domain.InvitationCode
	err := r.db.WithContext(ctx).First(&code, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}

func (r invitationRepository) Update(ctx context.Context, code *domain.InvitationCode) error {
	return r.db.WithContext(ctx).Save(code).Error
}

func (r invitationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.InvitationCode{}, id).Error
}

func (r invitationRepository) FindByCode(ctx context.Context, code string) (*domain.InvitationCode, error) {
	var invitationCode domain.InvitationCode
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&invitationCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invitationCode, nil
}

func (r invitationRepository) FindAll(ctx context.Context, filter InvitationFilter) ([]domain.InvitationCode, int64, error) {
	var codes []domain.InvitationCode
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.InvitationCode{})

	if filter.IsRedeemed != nil {
		if *filter.IsRedeemed {
			query = query.Where("redeemed_by IS NOT NULL")
		} else {
			query = query.Where("redeemed_by IS NULL")
		}
	}

	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Execute query
	if err := query.Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

func (r invitationRepository) MarkAsRedeemed(ctx context.Context, codeID uint, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.InvitationCode{}).
		Where("id = ?", codeID).
		Update("redeemed_by", userID).Error
}
