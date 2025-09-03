package repository

import (
	"context"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type OAuthRepositoryInterface interface {
	CreateOAuthProvider(ctx context.Context, oauth *entity.OAuthProviderEntity) error
	UpsertOAuthProvider(ctx context.Context, oauth *entity.OAuthProviderEntity) error
	GetOAuthProviderByProviderAndUserID(ctx context.Context, provider string, providerUserID string) (*entity.OAuthProviderEntity, error)
	GetOAuthProvidersByUserID(ctx context.Context, userID int64) ([]*entity.OAuthProviderEntity, error)
	DeleteOAuthProvider(ctx context.Context, id int64) error

	AssignRoleToUser(ctx context.Context, userID int64, roleID int64) error
	LogOAuthActivity(ctx context.Context, activity *entity.OAuthActivityLog) error
}

type OAuthRepository struct {
	db *gorm.DB
}

// LogOAuthActivity implements OAuthRepositoryInterface.
func (o *OAuthRepository) LogOAuthActivity(ctx context.Context, activity *entity.OAuthActivityLog) error {

	modelActivity := model.OAuthActivityLog{
		UserID:    activity.UserID,
		Provider:  activity.Provider,
		Action:    activity.Action,
		IPAddress: activity.IPAddress,
		UserAgent: activity.UserAgent,
		Status:    activity.Status,
		ErrorMsg:  activity.ErrorMsg,
		CreatedAt: time.Now(),
	}

	if err := o.db.WithContext(ctx).Create(&modelActivity).Error; err != nil {
		log.Errorf("[OAuthRepository-13] LogOAuthActivity: %v", err)
		return err
	}

	activity.ID = modelActivity.ID
	activity.CreatedAt = modelActivity.CreatedAt

	return nil

}

// AssignRoleToUser implements OAuthRepositoryInterface.
func (o *OAuthRepository) AssignRoleToUser(ctx context.Context, userID int64, roleID int64) error {
	userRole := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := o.db.WithContext(ctx).Create(userRole).Error; err != nil {
		log.Errorf("[OAuthRepository-1] AssignRoleToUser: %v", err)
		return err
	}
	return nil
}

// DeleteOAuthProvider implements OAuthRepositoryInterface.
func (o *OAuthRepository) DeleteOAuthProvider(ctx context.Context, id int64) error {
	if err := o.db.WithContext(ctx).Delete(&model.OAuthProvider{}, id).Error; err != nil {
		log.Errorf("[OAuthRepository-6] DeleteOAuthProvider: %v", err)
		return err
	}
	return nil
}

// GetOAuthProviderByProviderAndUserID implements OAuthRepositoryInterface.
func (o *OAuthRepository) GetOAuthProviderByProviderAndUserID(ctx context.Context, provider string, providerUserID string) (*entity.OAuthProviderEntity, error) {
	var modelOAuth model.OAuthProvider

	if err := o.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&modelOAuth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Errorf("[OAuthRepository-4] GetOAuthProviderByProviderAndUserID: %v", err)
		return nil, err
	}

	return &entity.OAuthProviderEntity{
		ID:              modelOAuth.ID,
		UserID:          modelOAuth.UserID,
		Provider:        modelOAuth.Provider,
		ProviderUserID:  modelOAuth.ProviderUserID,
		ProviderEmail:   modelOAuth.ProviderEmail,
		ProviderName:    modelOAuth.ProviderName,
		ProviderPicture: modelOAuth.ProviderPicture,
		AccessToken:     modelOAuth.AccessToken,
		RefreshToken:    modelOAuth.RefreshToken,
		TokenExpiresAt:  modelOAuth.TokenExpiresAt,
		CreatedAt:       modelOAuth.CreatedAt,
		UpdatedAt:       modelOAuth.UpdatedAt,
	}, nil
}

// GetOAuthProvidersByUserID implements OAuthRepositoryInterface.
func (o *OAuthRepository) GetOAuthProvidersByUserID(ctx context.Context, userID int64) ([]*entity.OAuthProviderEntity, error) {
	var modelOAuths []model.OAuthProvider

	if err := o.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&modelOAuths).Error; err != nil {
		log.Errorf("[OAuthRepository-5] GetOAuthProvidersByUserID: %v", err)
		return nil, err
	}

	var oauths []*entity.OAuthProviderEntity
	for _, modelOAuth := range modelOAuths {
		oauths = append(oauths, &entity.OAuthProviderEntity{
			ID:              modelOAuth.ID,
			UserID:          modelOAuth.UserID,
			Provider:        modelOAuth.Provider,
			ProviderUserID:  modelOAuth.ProviderUserID,
			ProviderEmail:   modelOAuth.ProviderEmail,
			ProviderName:    modelOAuth.ProviderName,
			ProviderPicture: modelOAuth.ProviderPicture,
			AccessToken:     modelOAuth.AccessToken,
			RefreshToken:    modelOAuth.RefreshToken,
			TokenExpiresAt:  modelOAuth.TokenExpiresAt,
			CreatedAt:       modelOAuth.CreatedAt,
			UpdatedAt:       modelOAuth.UpdatedAt,
		})
	}

	return oauths, nil

}

// UpsertOAuthProvider implements OAuthRepositoryInterface.
func (o *OAuthRepository) UpsertOAuthProvider(ctx context.Context, oauth *entity.OAuthProviderEntity) error {

	modelOAuth := model.OAuthProvider{
		UserID:          oauth.UserID,
		Provider:        oauth.Provider,
		ProviderUserID:  oauth.ProviderUserID,
		ProviderEmail:   oauth.ProviderEmail,
		ProviderName:    oauth.ProviderName,
		ProviderPicture: oauth.ProviderPicture,
		AccessToken:     oauth.AccessToken,
		RefreshToken:    oauth.RefreshToken,
		TokenExpiresAt:  oauth.TokenExpiresAt,
	}

	result := o.db.WithContext(ctx).Model(&model.OAuthProvider{}).
		Where("provider = ? AND provider_user_id = ?", oauth.Provider, oauth.ProviderUserID).
		Updates(map[string]interface{}{
			"user_id":          oauth.UserID,
			"provider_email":   oauth.ProviderEmail,
			"provider_name":    oauth.ProviderName,
			"provider_picture": oauth.ProviderPicture,
			"access_token":     oauth.AccessToken,
			"refresh_token":    oauth.RefreshToken,
			"token_expires_at": oauth.TokenExpiresAt,
			"updated_at":       time.Now(),
		})

	if result.Error != nil {
		log.Errorf("[OAuthRepository-2] UpsertOAuthProvider update: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		if err := o.db.WithContext(ctx).Create(&modelOAuth).Error; err != nil {
			log.Errorf("[OAuthRepository-3] UpsertOAuthProvider create: %v", err)
			return err
		}
		oauth.ID = modelOAuth.ID
		oauth.CreatedAt = modelOAuth.CreatedAt
	}

	return nil

}

// CreateOAuthProvider implements OAuthRepositoryInterface.
func (o *OAuthRepository) CreateOAuthProvider(ctx context.Context, oauth *entity.OAuthProviderEntity) error {
	modelOAuth := model.OAuthProvider{
		UserID:          oauth.UserID,
		Provider:        oauth.Provider,
		ProviderUserID:  oauth.ProviderUserID,
		ProviderEmail:   oauth.ProviderEmail,
		ProviderName:    oauth.ProviderName,
		ProviderPicture: oauth.ProviderPicture,
		AccessToken:     oauth.AccessToken,
		RefreshToken:    oauth.RefreshToken,
		TokenExpiresAt:  oauth.TokenExpiresAt,
	}

	if err := o.db.WithContext(ctx).Create(&modelOAuth).Error; err != nil {
		log.Errorf("[OAuthRepository-1] CreateOAuthProvider: %v", err)
		return err
	}

	oauth.ID = modelOAuth.ID
	oauth.CreatedAt = modelOAuth.CreatedAt
	return nil

}

func NewOAuthRepository(db *gorm.DB) OAuthRepositoryInterface {
	return &OAuthRepository{db: db}
}
