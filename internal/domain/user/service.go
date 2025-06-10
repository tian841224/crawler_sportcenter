package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID string) (*User, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type UserService struct {
	repo Repository
}

var _ Service = (*UserService)(nil)

func NewUserService(repo Repository) Service {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, user *User) error {
	if user == nil {
		return errors.New("user 不能為空")
	}

	// 檢查用戶是否已存在
	existingUser, err := s.repo.GetByAccountID(ctx, user.AccountID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else if existingUser != nil {
		return errors.New("使用者已存在")
	}

	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*User, error) {
	if id == 0 {
		return nil, errors.New("ID 不能為 0")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByAccountID(ctx context.Context, accountID string) (*User, error) {
	if accountID == "" {
		return nil, errors.New("accountID 不能為空")
	}
	return s.repo.GetByAccountID(ctx, accountID)
}

func (s *UserService) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	if id == 0 {
		return errors.New("ID 不能為 0")
	}
	return s.repo.Update(ctx, id, updates)
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("不能為 0")
	}
	return s.repo.Delete(ctx, id)
}
