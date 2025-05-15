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
	ChooseWeekday         string
	TimeSlotCodes         []types.TimeSlotCode // 改為切片以支援多個時段
	DayPeriod    int
	ButtonIndex  []int
	ID                    string
	Password              string
	TG_Bot_Token          string
	TG_Bot_Webhook_Domain string
	// TG_Bot_Webhook_Port   string
	TG_Bot_Secret_Token   string
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
		TG_Bot_Token:          os.Getenv("TG_Bot_Token"),
		TG_Bot_Webhook_Domain: os.Getenv("TG_Bot_Webhook_Domain"),
		// TG_Bot_Webhook_Port:   os.Getenv("TG_Bot_Webhook_Port"),
		TG_Bot_Secret_Token:   os.Getenv("TG_Bot_Secret_Token"),
		DayPeriod: func() int {
			period, err := strconv.Atoi(os.Getenv("DAY_PERIOD"))
			if err != nil {
				return 1 // Default value if conversion fails
			}
			return period
		}(),
		ButtonIndex: func() []int {
			indexStr := os.Getenv("BUTTON_INDEX")
			if indexStr == "" {
				return []int{} 
			}
			
			// 用逗號分隔字串
			indexStrArray := strings.Split(indexStr, ",")
			indexArray := make([]int, 0, len(indexStrArray))
			
			// 轉換每個字串為整數
			for _, str := range indexStrArray {
				num, err := strconv.Atoi(strings.TrimSpace(str))
				if err == nil {
					indexArray = append(indexArray, num)
				}
			}
			return indexArray
		}(),
	}

	return cfg
}
