package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
)

// DB 介面定義資料庫的基本操作
type DB interface {
	Close() error
	DropDatabase(dbName string) error
	GetConn() interface{} // 返回 GORM 連接
}

// NewDatabase 根據環境變數建立適當的資料庫連接
func NewDatabase(cfg config.Config) (DB, error) {
	// 載入 .env 文件
	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("無法載入 .env 文件，使用系統環境變數", zap.Error(err))
	}

	// 從環境變數讀取資料庫類型
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "" {
		dbType = "sqlite" // 預設使用 SQLite
	}

	logger.Log.Info("初始化資料庫", zap.String("type", dbType))

	switch dbType {
	case "sqlite":
		db := NewSQLiteDB(cfg)
		if db == nil {
			return nil, fmt.Errorf("無法建立 SQLite 資料庫連接")
		}
		return db, nil
	case "postgres", "postgresql":
		db := NewPostgresDB(cfg)
		if db == nil {
			return nil, fmt.Errorf("無法建立 PostgreSQL 資料庫連接")
		}
		return db, nil
	default:
		return nil, fmt.Errorf("不支援的資料庫類型: %s", dbType)
	}
}
