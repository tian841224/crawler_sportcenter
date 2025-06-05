package user

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID string) (*User, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type UserRepository struct {
	db *gorm.DB
}

var _ Repository = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB) Repository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*User, error) {
	var user []*User
	err := r.db.WithContext(ctx).Find(&user).Error
	return user, err
}

func (r *UserRepository) GetByAccountID(ctx context.Context, accountID string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id =?", id).Updates(updates).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}
