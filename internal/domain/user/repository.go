package user

import (
	"context"

	"github.com/tian841224/crawler_sportcenter/internal/infrastructure/db"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID string) (*User, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type UserRepository struct {
	db *db.DB
}

var _ Repository = (*UserRepository)(nil)

func NewUserRepository(db *db.DB) Repository {
	if err := db.Conn.AutoMigrate(&User{}); err != nil {
		logger.Log.Error("資料庫遷移失敗", zap.Error(err))
		return nil
	}
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	return r.db.Conn.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.Conn.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*User, error) {
	var user []*User
	err := r.db.Conn.WithContext(ctx).Find(&user).Error
	return user, err
}

func (r *UserRepository) GetByAccountID(ctx context.Context, accountID string) (*User, error) {
	var user User
	err := r.db.Conn.WithContext(ctx).Where("account_id = ?", accountID).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.Conn.WithContext(ctx).Model(&User{}).Where("id =?", id).Updates(updates).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Conn.WithContext(ctx).Delete(&User{}, id).Error
}
