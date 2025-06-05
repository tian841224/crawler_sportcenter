package timeslot

import "time"

// TimeSlot 代表可預約的時間區段定義
type TimeSlot struct {
	ID        uint      `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	StartTime time.Time `gorm:"column:start_time;type:time without time zone;not null" json:"startTime" example:"09:00"`
	EndTime   time.Time `gorm:"column:end_time;type:time without time zone;not null" json:"endTime" example:"10:00"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt" swaggerignore:"true"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt" swaggerignore:"true"`
}

// TableName 設定 TimeSlot 的資料表名稱
func (TimeSlot) TableName() string {
	return "time_slot"
}
