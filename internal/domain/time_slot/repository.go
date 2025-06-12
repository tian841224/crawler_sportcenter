package timeslot

import (
	"context"
	"fmt"
	"time"

	"github.com/tian841224/crawler_sportcenter/internal/infrastructure/db"
)

type Repository interface {
	Create(ctx context.Context, timeSlot *TimeSlot) error
	GetByID(ctx context.Context, id uint) (*TimeSlot, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type TimeSlotRepository struct {
	db *db.DB
}

var _ Repository = (*TimeSlotRepository)(nil)

func NewTimeSlotRepository(db *db.DB) Repository {
	repo := &TimeSlotRepository{db: db}
	db.Conn.AutoMigrate(&TimeSlot{})
	if err := repo.initData(); err != nil {
		fmt.Printf("初始化時段資料失敗: %v\n", err)
		return nil
	}
	return repo
}

func (r *TimeSlotRepository) Create(ctx context.Context, timeSlot *TimeSlot) error {
	return r.db.Conn.WithContext(ctx).Create(timeSlot).Error
}

func (r *TimeSlotRepository) GetByID(ctx context.Context, id uint) (*TimeSlot, error) {
	var timeSlot TimeSlot
	err := r.db.Conn.WithContext(ctx).First(&timeSlot, id).Error
	return &timeSlot, err
}

func (r *TimeSlotRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.Conn.WithContext(ctx).Model(&TimeSlot{}).Where("id =?", id).Updates(updates).Error
}

func (r *TimeSlotRepository) Delete(ctx context.Context, id uint) error {
	return r.db.Conn.WithContext(ctx).Delete(&TimeSlot{}, id).Error
}

// 建立預設資料
func (r *TimeSlotRepository) initData() error {
	if r.db == nil {
		return fmt.Errorf("資料庫連接未初始化")
	}

	// 檢查是否已經有資料
	var count int64
	if err := r.db.Conn.Model(&TimeSlot{}).Count(&count).Error; err != nil {
		return fmt.Errorf("檢查資料數量失敗: %w", err)
	}

	// 如果已經有資料，就不需要再建立
	if count >= 16 {
		return nil
	}

	// 先建立預設時段
	timeSlots := []TimeSlot{
		{StartTime: parseTime("06:00"), EndTime: parseTime("07:00")},
		{StartTime: parseTime("07:00"), EndTime: parseTime("08:00")},
		{StartTime: parseTime("08:00"), EndTime: parseTime("09:00")},
		{StartTime: parseTime("09:00"), EndTime: parseTime("10:00")},
		{StartTime: parseTime("10:00"), EndTime: parseTime("11:00")},
		{StartTime: parseTime("11:00"), EndTime: parseTime("12:00")},
		{StartTime: parseTime("12:00"), EndTime: parseTime("13:00")},
		{StartTime: parseTime("13:00"), EndTime: parseTime("14:00")},
		{StartTime: parseTime("14:00"), EndTime: parseTime("15:00")},
		{StartTime: parseTime("15:00"), EndTime: parseTime("16:00")},
		{StartTime: parseTime("16:00"), EndTime: parseTime("17:00")},
		{StartTime: parseTime("17:00"), EndTime: parseTime("18:00")},
		{StartTime: parseTime("18:00"), EndTime: parseTime("19:00")},
		{StartTime: parseTime("19:00"), EndTime: parseTime("20:00")},
		{StartTime: parseTime("20:00"), EndTime: parseTime("21:00")},
		{StartTime: parseTime("21:00"), EndTime: parseTime("22:00")},
	}

	// 建立或取得時段資料
	for i := range timeSlots {
		result := r.db.Conn.Where("start_time = ? AND end_time = ?",
			timeSlots[i].StartTime, timeSlots[i].EndTime).FirstOrCreate(&timeSlots[i])
		if result.Error != nil {
			return fmt.Errorf("建立時段失敗 (%s-%s): %w",
				timeSlots[i].StartTime, timeSlots[i].EndTime, result.Error)
		}
	}
	return nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("15:04", s)
	return t
}
