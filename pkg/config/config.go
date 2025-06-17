package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tian841224/crawler_sportcenter/internal/types"
)

type Config struct {
	DBType                string
	DBPath                string
	DBHost                string
	DBPort                string
	DBName                string
	DBUser                string
	DBPassword            string
	ChooseWeekday         string
	TimeSlotCodes         []types.TimeSlotCode // 改為切片以支援多個時段
	DayPeriod             int
	ButtonIndex           []int
	ID                    string
	Password              string
	TG_Bot_Token          string
	TG_Bot_Webhook_Domain string
	// TG_Bot_Webhook_Port   string
	// TG_Bot_Secret_Token string
}

func LoadConfig() Config {
	var envPath string
	var err error

	// 先讀取當前工作目錄的 .env（For部署環境）
	currentDirEnv := ".env"
	if _, err := os.Stat(currentDirEnv); err == nil {
		envPath = currentDirEnv
		err = godotenv.Load(envPath)
		if err == nil {
			log.Printf("成功載入 .env 檔案: %s", envPath)
		}
	}

	// 如果當前目錄沒有或載入失敗，嘗試從程式碼位置推算的專案根目錄 (For開發環境)
	if envPath == "" || err != nil {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		projectRoot := filepath.Join(dir, "..", "..")
		fallbackEnvPath := filepath.Join(projectRoot, ".env")

		if _, statErr := os.Stat(fallbackEnvPath); statErr == nil {
			envPath = fallbackEnvPath
			err = godotenv.Load(envPath)
			if err == nil {
				log.Printf("成功載入 .env 檔案 (備援路徑): %s", envPath)
			}
		}
	}

	// 如果兩種方式都失敗，記錄錯誤但繼續執行（使用系統環境變數）
	if err != nil {
		log.Printf("無法載入 .env 檔案，使用系統環境變數。嘗試的路徑: 當前目錄/.env, %s", envPath)
	}

	timeSlotCodesStr := strings.Split(os.Getenv("TIME_SLOT_CODE"), ",")
	var timeSlotCodes []types.TimeSlotCode

	for _, codeStr := range timeSlotCodesStr {
		if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
			timeSlotCodes = append(timeSlotCodes, types.TimeSlotCode(code))
		}
	}

	if len(timeSlotCodes) == 0 {
		timeSlotCodes = append(timeSlotCodes, types.TimeSlotCode(1))
	}

	// 從環境變數中獲取值
	cfg := Config{
		DBType:                os.Getenv("DB_TYPE"),
		DBPath:                os.Getenv("DB_PATH"),
		DBHost:                os.Getenv("DB_HOST"),
		DBPort:                os.Getenv("DB_PORT"),
		DBName:                os.Getenv("DB_NAME"),
		DBUser:                os.Getenv("DB_USER"),
		DBPassword:            os.Getenv("DB_PASSWORD"),
		ChooseWeekday:         os.Getenv("CHOOSE_WEEKDAY"),
		TimeSlotCodes:         timeSlotCodes,
		ID:                    os.Getenv("ID"),
		Password:              os.Getenv("Password"),
		TG_Bot_Token:          os.Getenv("TELEGRAM_BOT_TOKEN"),
		TG_Bot_Webhook_Domain: os.Getenv("TELEGRAM_BOT_WEBHOOK_DOMAIN"),
		// TG_Bot_Webhook_Port:   os.Getenv("TG_Bot_Webhook_Port"),
		// TG_Bot_Secret_Token: os.Getenv("TELEGRAM_BOT_SECRET_TOKEN"),
		DayPeriod: func() int {
			period, err := strconv.Atoi(os.Getenv("DAY_PERIOD"))
			if err != nil {
				return 1
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
