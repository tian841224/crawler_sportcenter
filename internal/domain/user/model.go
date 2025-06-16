package user

import "time"

type User struct {
	ID                  uint      `gorm:"primaryKey;column:id;autoIncrement"`
	AccountID           string    `gorm:"column:account_id;type:varchar(50);unique;not null"`
	Status              bool      `gorm:"column:status;not null"`
	SportCenterAccount  string    `gorm:"column:sport_center_account;type:varchar(50)"`
	SportCenterPassword string    `gorm:"column:sport_center_password;type:varchar(50)"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "user"
}
