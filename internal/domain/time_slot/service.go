package timeslot

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, timeSlot *TimeSlot) error
	GetByID(ctx context.Context, id uint) (*TimeSlot, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type TimeSlotService struct {
	repo Repository
}

var _ Service = (*TimeSlotService)(nil)

func NewTimeSlotService(repo Repository) Service {
	return &TimeSlotService{repo: repo}
}

func (s *TimeSlotService) Create(ctx context.Context, timeSlot *TimeSlot) error {
	if timeSlot == nil {
		return errors.New("timeSlot 不能為空")
	}
	return s.repo.Create(ctx, timeSlot)
}

func (s *TimeSlotService) GetByID(ctx context.Context, id uint) (*TimeSlot, error) {
	if id == 0 {
		return nil, errors.New("ID 不能為 0")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TimeSlotService) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	if id == 0 {
		return errors.New("ID 不能為 0")
	}
	return s.repo.Update(ctx, id, updates)
}

func (s *TimeSlotService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("ID 不能為 0")
	}
	return s.repo.Delete(ctx, id)
}
