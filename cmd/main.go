package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	tgbot "github.com/tian841224/crawler_sportcenter/internal/bot/tg_bot"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/domain/schedule"
	timeslot "github.com/tian841224/crawler_sportcenter/internal/domain/time_slot"
	"github.com/tian841224/crawler_sportcenter/internal/domain/user"
	"github.com/tian841224/crawler_sportcenter/internal/infrastructure/db"
	"github.com/tian841224/crawler_sportcenter/internal/scheduler"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {

	// #region 初始化 Logger
	logger.InitLogger()
	logger.Log.Info("初始化 Logger")
	// #endregion

	// #region 讀取env設定檔
	logger.Log.Info("載入設定檔")
	cfg := config.LoadConfig()
	// #endregion

	//  #region 初始化資料庫
	logger.Log.Info("初始化資料庫")
	dbInstance, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Log.Error("資料庫初始化失敗", zap.Error(err))
		return
	}
	if dbInstance == nil {
		logger.Log.Error("資料庫初始化失敗")
		return
	}
	// Debug用
	// dbInstance.DropDatabase(cfg.DBName)
	// #endregion

	// #region 初始化瀏覽器
	logger.Log.Info("初始化瀏覽器")
	browser := browser.NewBrowserService()
	defer browser.Close()
	nantunSportCenterService := crawler.NewNantunSportCenterService(browser)
	// nantunSportCenterService.CrawlerNantun(cfg)
	// #endregion

	// #region 初始化 Telegram Bot
	logger.Log.Info("初始化Telegram Bot")
	botService := tgbot.NewTGBotService(cfg)
	if botService == nil {
		logger.Log.Error("Failed to initialize Telegram Bot")
		return
	}
	// #endregion

	// #region 初始化Repository
	logger.Log.Info("初始化Repository")
	userRepository := user.NewUserRepository(&dbInstance)
	timeslotRepository := timeslot.NewTimeSlotRepository(&dbInstance)
	scheduleRepository := schedule.NewScheduleRepository(&dbInstance)
	// #endregion

	// #region 初始化Service
	logger.Log.Info("初始化Service")
	userService := user.NewUserService(userRepository)
	timeslotService := timeslot.NewTimeSlotService(timeslotRepository)
	scheduleService := schedule.NewScheduleService(scheduleRepository)
	nantunSportCenterBotService := crawler.NewNantunSportCenterBotService(browser, nantunSportCenterService, cfg)
	// #endregion

	handler := tgbot.NewMessageHandler(botService, userService, timeslotService, scheduleService, &nantunSportCenterBotService)

	// 設定訊息處理
	botService.HandleMessage(handler.HandleUpdate)

	// 使用 context 控制 Bot 的生命週期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 開始接收訊息
	botService.StartReceiveMessage()

	// #region 初始化Scheduler
	logger.Log.Info("初始化Scheduler")
	schedulerService := scheduler.NewSchedulerService(&nantunSportCenterBotService, scheduleService, userService, botService)
	schedulerService.Start(ctx)
	// #endregion

	logger.Log.Info("開始接收訊息")

	// 設定系統信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 等待中斷信號
	<-sigChan
	logger.Log.Info("收到中斷信號，開始關閉程式")

	// 關閉 scheduler
	schedulerService.Stop()

	// 關閉瀏覽器
	if err := browser.Close(); err != nil {
		logger.Log.Error("關閉瀏覽器失敗", zap.Error(err))
	}

	// 關閉資料庫連接
	conn := dbInstance.GetConn().(*gorm.DB)
	sqlDB, err := conn.DB()
	if err != nil {
		logger.Log.Error("取得資料庫連接失敗", zap.Error(err))
	} else {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error("關閉資料庫連接失敗", zap.Error(err))
		}
	}

	logger.Log.Info("程式已關閉")
}
