package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
	cfg  config.Config
}

func NewDB(cfg config.Config) *DB {
	db := &DB{
		cfg: cfg,
	}
	if err := db.initDatabase(); err != nil {
		logger.Log.Error("資料庫初始化失敗", zap.Error(err))
		return nil
	}
	return db
}

// InitDatabase 初始化資料庫
func (d *DB) initDatabase() error {
	// 建立基礎連接字串 - 連接到預設的 postgres 資料庫
	baseDSN := fmt.Sprintf("host=localhost user=%s password=%s dbname=postgres sslmode=disable",
		d.cfg.DBUser, d.cfg.DBPassword)

	logger.Log.Info("開始初始化資料庫",
		zap.String("database", d.cfg.DBName),
		zap.String("user", d.cfg.DBUser))

	// 檢查並建立資料庫
	if err := d.createDatabaseIfNotExists(baseDSN, d.cfg.DBName); err != nil {
		return fmt.Errorf("建立資料庫失敗: %w", err)
	}

	// 連接到指定的資料庫
	db, err := d.connectToDatabase()
	if err != nil {
		return fmt.Errorf("連接資料庫失敗: %w", err)
	}

	// 設定連接
	d.Conn = db

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("層資料庫連接失敗: %w", err)
	}

	// 執行簡單查詢測試連接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("資料庫連接測試失敗: %w", err)
	}

	logger.Log.Info("資料庫連線成功",
		zap.String("database", d.cfg.DBName),
		zap.String("user", d.cfg.DBUser))
	return nil
}

// connectToDatabase 連接到指定的資料庫
func (d *DB) connectToDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s sslmode=disable",
		d.cfg.DBUser, d.cfg.DBPassword, d.cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("連接資料庫失敗: %w", err)
	}

	return db, nil
}

// createDatabaseIfNotExists 檢查並建立資料庫
func (d *DB) createDatabaseIfNotExists(baseDSN, dbName string) error {
	logger.Log.Info("檢查資料庫是否存在", zap.String("database", dbName))

	sqlDB, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return fmt.Errorf("連接到 postgres 資料庫失敗: %w", err)
	}
	defer sqlDB.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = sqlDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("檢查資料庫是否存在失敗: %w", err)
	}

	if !exists {
		logger.Log.Info("開始建立資料庫", zap.String("database", dbName))
		_, err = sqlDB.Exec("CREATE DATABASE " + dbName)
		if err != nil {
			return fmt.Errorf("建立資料庫失敗: %w", err)
		}
		logger.Log.Info("資料庫建立成功", zap.String("database", dbName))

		// 驗證資料庫是否確實建立
		err = sqlDB.QueryRow(query, dbName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("驗證資料庫建立失敗: %w", err)
		}
		if !exists {
			return fmt.Errorf("資料庫 %s 建立失敗", dbName)
		}
	} else {
		logger.Log.Info("資料庫已存在", zap.String("database", dbName))
	}

	return nil
}

// DropDatabase 刪除指定的資料庫
func (d *DB) DropDatabase(dbName string) error {
	// 建立基礎連接字串連接到 postgres 資料庫
	baseDSN := fmt.Sprintf("host=localhost user=%s password=%s dbname=postgres sslmode=disable",
		d.cfg.DBUser, d.cfg.DBPassword)

	// 連接到 postgres 資料庫
	sqlDB, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return fmt.Errorf("連接到 postgres 資料庫失敗: %w", err)
	}
	defer sqlDB.Close()

	// 檢查資料庫是否存在
	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = sqlDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("檢查資料庫是否存在失敗: %w", err)
	}

	if !exists {
		return fmt.Errorf("資料庫 %s 不存在", dbName)
	}

	// 關閉所有與該資料庫的連接
	_, err = sqlDB.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()", dbName))
	if err != nil {
		return fmt.Errorf("關閉資料庫連接失敗: %w", err)
	}

	// 刪除資料庫
	_, err = sqlDB.Exec("DROP DATABASE " + dbName)
	if err != nil {
		return fmt.Errorf("刪除資料庫失敗: %w", err)
	}

	logger.Log.Info("資料庫刪除成功: " + dbName)
	return nil
}
