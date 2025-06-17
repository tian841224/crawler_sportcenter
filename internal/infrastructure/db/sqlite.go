package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // 使用純 Go SQLite 驅動
)

// SQLiteDB 包裝 GORM 資料庫連接
type SQLiteDB struct {
	Conn   *gorm.DB
	dbName string
	dbPath string
}

// NewSQLiteDB 建立 SQLite 資料庫連接，從環境變數讀取配置
func NewSQLiteDB() *SQLiteDB {
	// 載入 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("無法載入 .env 文件，使用系統環境變數:", err)
	}

	// 從環境變數讀取資料庫名稱
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "crawler_sportcenter_system" // 預設值
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// 取得執行檔所在目錄
		execPath, err := os.Executable()
		if err != nil {
			// 如果無法取得執行檔路徑，使用當前目錄
			dbPath = filepath.Join(".", dbName+".db")
		} else {
			// 使用執行檔所在目錄
			execDir := filepath.Dir(execPath)
			dbPath = filepath.Join(execDir, dbName+".db")
		}
	}

	db := &SQLiteDB{
		dbName: dbName,
		dbPath: dbPath,
	}

	if err := db.initDatabase(); err != nil {
		logger.Log.Error("SQLite 資料庫初始化失敗", zap.Error(err))
		return nil
	}
	return db
}

// initDatabase 初始化 SQLite 資料庫
func (d *SQLiteDB) initDatabase() error {
	logger.Log.Info("開始初始化 SQLite 資料庫",
		zap.String("database", d.dbName),
		zap.String("path", d.dbPath))

	// 建立資料庫檔案
	if err := d.createDatabaseIfNotExists(d.dbPath); err != nil {
		return fmt.Errorf("建立 SQLite 資料庫失敗: %w", err)
	}

	// 連接到資料庫
	db, err := d.connectToDatabaseWithPath(d.dbPath)
	if err != nil {
		return fmt.Errorf("連接 SQLite 資料庫失敗: %w", err)
	}

	// 設定連接
	d.Conn = db

	// 測試連接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("取得底層資料庫連接失敗: %w", err)
	}

	// 執行簡單查詢測試連接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("SQLite 資料庫連接測試失敗: %w", err)
	}

	logger.Log.Info("SQLite 資料庫連線成功",
		zap.String("database", d.dbName),
		zap.String("path", d.dbPath))
	return nil
}

// connectToDatabaseWithPath 連接到指定路徑的 SQLite 資料庫
func (d *SQLiteDB) connectToDatabaseWithPath(dbPath string) (*gorm.DB, error) {
	// 設定 GORM 配置
	config := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	// 使用 modernc.org/sqlite 驅動連接 SQLite 資料庫
	dsn := "file:" + dbPath + "?cache=shared&mode=rwc"
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        dsn,
	}, config)
	if err != nil {
		return nil, fmt.Errorf("無法連接 SQLite 資料庫: %w", err)
	}

	// 啟用外鍵約束 (SQLite 預設不啟用)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("無法啟用外鍵約束: %w", err)
	}

	return db, nil
}

// createDatabaseIfNotExists 檢查並建立 SQLite 資料庫檔案
func (d *SQLiteDB) createDatabaseIfNotExists(dbPath string) error {
	logger.Log.Info("檢查 SQLite 資料庫檔案是否存在", zap.String("path", dbPath))

	// 確保資料庫目錄存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("無法建立資料庫目錄: %w", err)
	}

	// 檢查檔案是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Log.Info("開始建立 SQLite 資料庫檔案", zap.String("path", dbPath))

		// 建立空檔案
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("建立 SQLite 資料庫檔案失敗: %w", err)
		}
		file.Close()

		logger.Log.Info("SQLite 資料庫檔案建立成功", zap.String("path", dbPath))
	} else {
		logger.Log.Info("SQLite 資料庫檔案已存在", zap.String("path", dbPath))
	}

	return nil
}

// DropDatabase 刪除 SQLite 資料庫檔案
func (d *SQLiteDB) DropDatabase(dbName string) error {
	dbPath := filepath.Join("data", dbName+".db")

	logger.Log.Info("開始刪除 SQLite 資料庫", zap.String("database", dbName), zap.String("path", dbPath))

	// 檢查檔案是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("SQLite 資料庫檔案 %s 不存在", dbPath)
	}

	// 如果目前有連接，先關閉連接
	if d.Conn != nil {
		if err := d.Close(); err != nil {
			logger.Log.Warn("關閉 SQLite 資料庫連接時發生錯誤", zap.Error(err))
		}
	}

	// 刪除資料庫檔案
	if err := os.Remove(dbPath); err != nil {
		return fmt.Errorf("刪除 SQLite 資料庫檔案失敗: %w", err)
	}

	logger.Log.Info("SQLite 資料庫刪除成功", zap.String("database", dbName))
	return nil
}

// Close 關閉資料庫連接
func (d *SQLiteDB) Close() error {
	if d.Conn == nil {
		return nil
	}

	sqlDB, err := d.Conn.DB()
	if err != nil {
		return fmt.Errorf("取得底層資料庫連接失敗: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("關閉 SQLite 資料庫連接失敗: %w", err)
	}

	logger.Log.Info("SQLite 資料庫連接已關閉")
	return nil
}

func (d *SQLiteDB) GetConn() interface{} {
	return d.Conn
}
