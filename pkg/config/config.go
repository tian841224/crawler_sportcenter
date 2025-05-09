package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tian841224/crawler_sportcenter/internal/types"
)

type Config struct {
	ChooseWeekday string
	TimeSlotCode  types.TimeSlotCode
	ID            string
	Password      string
}

func LoadConfig() Config {
	// 載入 .env 文件，指定正確的文件路徑
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("無法載入 .env 文件，使用系統環境變數:", err)
	}

	// 從環境變數中獲取 TimeSlotCode 並轉換為整數
	timeSlotCodeStr := os.Getenv("TIME_SLOT_CODE")
	timeSlotCode := types.TimeSlotCode(1) // 默認值
	if code, err := strconv.Atoi(timeSlotCodeStr); err == nil {
		timeSlotCode = types.TimeSlotCode(code)
	} else {
		log.Printf("無法解析 TIME_SLOT_CODE，使用默認值: %v", err)
	}

	// 從環境變數中獲取值
	cfg := Config{
		ChooseWeekday: os.Getenv("CHOOSE_WEEKDAY"),
		TimeSlotCode:  timeSlotCode,
		ID:            os.Getenv("ID"),
		Password:      os.Getenv("Password"),
	}

	return cfg
}
