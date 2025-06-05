package timeslot

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, timeSlot *TimeSlot) error
	GetByID(ctx context.Context, id uint) (*TimeSlot, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type TimeSlotRepository struct {
	db *gorm.DB
}

var _ Repository = (*TimeSlotRepository)(nil)

func NewTimeSlotRepository(db *gorm.DB) Repository {
	return &TimeSlotRepository{db: db}
}

func (r *TimeSlotRepository) Create(ctx context.Context, timeSlot *TimeSlot) error {
	return r.db.WithContext(ctx).Create(timeSlot).Error
}

func (r *TimeSlotRepository) GetByID(ctx context.Context, id uint) (*TimeSlot, error) {
	var timeSlot TimeSlot
	err := r.db.WithContext(ctx).First(&timeSlot, id).Error
	return &timeSlot, err
}

func (r *TimeSlotRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&TimeSlot{}).Where("id =?", id).Updates(updates).Error
}

func (r *TimeSlotRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&TimeSlot{}, id).Error
}
