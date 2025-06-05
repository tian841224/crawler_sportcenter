package schedule

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, schedule *Schedule) error
	GetByID(ctx context.Context, id uint) (*Schedule, error)
	GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error)
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, id uint) error
}

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) Repository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id uint) (*Schedule, error) {
	res := &Schedule{}
	if err := r.db.WithContext(ctx).First(res, id).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ScheduleRepository) GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error) {
	var schedules []*Schedule
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}


func (r *ScheduleRepository) Update(ctx context.Context, schedule *Schedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

func (r *ScheduleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Schedule{}, id).Error
}

