package schedule

import (
	"time"

	timeslot "github.com/tian841224/crawler_sportcenter/internal/domain/time_slot"
	"github.com/tian841224/crawler_sportcenter/internal/domain/user"
)

// Schedule 使用者設定的排程
type Schedule struct {
	ID         uint               `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UserID     uint               `gorm:"column:user_id" json:"userId"`
	User       *user.User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Weekday    time.Weekday       `gorm:"column:weekday;type:smallint" json:"weekday"`
	TimeSlotID *uint              `gorm:"column:time_slot_id" json:"timeSlotId"`
	TimeSlot   *timeslot.TimeSlot `gorm:"foreignKey:TimeSlotID" json:"timeSlot,omitempty"`
	CreatedAt  time.Time          `gorm:"column:created_at;autoCreateTime" json:"createdAt" swaggerignore:"true"`
	UpdatedAt  time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt" swaggerignore:"true"`
}

// TableName 設定資料表名稱
func (Schedule) TableName() string {
	return "schedule"
}

// ... existing code ...
