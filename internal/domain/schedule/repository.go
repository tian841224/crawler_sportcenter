package schedule

import (
	"context"

	"github.com/tian841224/crawler_sportcenter/internal/infrastructure/db"
)

type Repository interface {
	Create(ctx context.Context, schedule *Schedule) error
	GetByID(ctx context.Context, id uint) (*Schedule, error)
	GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error)
	GetAll(ctx context.Context) (*[]Schedule, error)
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, id uint) error
}

type ScheduleRepository struct {
	db *db.DB
}

var _ Repository = (*ScheduleRepository)(nil)

func NewScheduleRepository(db *db.DB) Repository {
	db.Conn.AutoMigrate(&Schedule{})
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *Schedule) error {
	return r.db.Conn.WithContext(ctx).Create(schedule).Error
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id uint) (*Schedule, error) {
	res := &Schedule{}
	if err := r.db.Conn.WithContext(ctx).First(res, id).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ScheduleRepository) GetByUserID(ctx context.Context, userID uint) ([]*Schedule, error) {
	var schedules []*Schedule
	if err := r.db.Conn.WithContext(ctx).Where("user_id = ?", userID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepository) GetAll(ctx context.Context) (*[]Schedule, error) {
	var schedules []Schedule
	if err := r.db.Conn.WithContext(ctx).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return &schedules, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *Schedule) error {
	return r.db.Conn.WithContext(ctx).Save(schedule).Error
}

func (r *ScheduleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Conn.WithContext(ctx).Delete(&Schedule{}, id).Error
}
