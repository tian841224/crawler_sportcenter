package schedule

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, schedule *Schedule) error
	GetByID(ctx context.Context, id uint) (*Schedule, error)
	GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error)
	GetAll(ctx context.Context) (*[]Schedule, error)
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, id uint) error
}

type ScheduleService struct {
	repo Repository // 注入 Repository 介面
}

var _ Service = (*ScheduleService)(nil)

func NewScheduleService(db *gorm.DB) Repository {
	return &ScheduleRepository{db: db}
}

func (s *ScheduleService) Create(ctx context.Context, schedule *Schedule) error {
	// 在建立前，先檢查是否已經存在相同使用者在相同星期有排程
	existingSchedules, err := s.repo.GetByUserID(ctx, schedule.UserID)
	if err != nil {
		return err
	}

	for _, existingSchedule := range existingSchedules {
		// 檢查是否已存在相同使用者在相同排程
		if existingSchedule.Weekday == schedule.Weekday && existingSchedule.TimeSlotID == schedule.TimeSlotID {
			return errors.New("已訂閱相同時段")
		}
	}

	// 如果沒有重複，才進行建立操作
	return s.repo.Create(ctx, schedule)
}

func (s *ScheduleService) GetByID(ctx context.Context, id uint) (*Schedule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ScheduleService) GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ScheduleService) GetAll(ctx context.Context) (*[]Schedule, error) {
	return s.repo.GetAll(ctx)
}

func (s *ScheduleService) Update(ctx context.Context, schedule *Schedule) error {
	return s.repo.Update(ctx, schedule)
}

func (s *ScheduleService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
