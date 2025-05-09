package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tian841224/crawler_sportcenter/internal/types"
)

type Config struct {
	ChooseWeekday string
	TimeSlotCodes []types.TimeSlotCode // 改為切片以支援多個時段
	ID            string
	Password      string
}

func LoadConfig() Config {
	// 載入 .env 文件，指定正確的文件路徑
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("無法載入 .env 文件，使用系統環境變數:", err)
	}

	// 從環境變數中獲取 TIME_SLOT_CODE 並轉換為整數切片
	timeSlotCodesStr := strings.Split(os.Getenv("TIME_SLOT_CODE"), ",")
	var timeSlotCodes []types.TimeSlotCode

	for _, codeStr := range timeSlotCodesStr {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			timeSlotCodes = append(timeSlotCodes, types.TimeSlotCode(code))
		}
	}

	if len(timeSlotCodes) == 0 {
		timeSlotCodes = append(timeSlotCodes, types.TimeSlotCode(1)) // 默認值
	}

	// 從環境變數中獲取值
	cfg := Config{
		ChooseWeekday: os.Getenv("CHOOSE_WEEKDAY"),
		TimeSlotCodes: timeSlotCodes,
		ID:            os.Getenv("ID"),
		Password:      os.Getenv("Password"),
	}

	return cfg
}
