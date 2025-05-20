package main

import (
	"context"

	tgbot "github.com/tian841224/crawler_sportcenter/internal/bot/tg_bot"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

func main() {

	// #region 初始化 Logger
	logger.InitLogger()
	logger.Log.Info("初始化 Logger")
	// #endregion

	// #region 讀取設定檔
	cfg := config.LoadConfig()
	logger.Log.Info("載入設定檔")
	// #endregion

	// #region 初始化瀏覽器
	browser := browser.NewBrowserService()
	// logger.Log.Info("初始化瀏覽器")
	nantunSportCenterService := crawler.NewNantunSportCenterService(browser)
	// nantunSportCenterService.CrawlerNantun(cfg)
	// #endregion

	// #region 初始化 Telegram Bot
	botService := tgbot.NewTGBotService(cfg)
	if botService == nil {
		logger.Log.Error("Failed to initialize Telegram Bot")
		return
	}

	nantunSportCenterBotService := crawler.NewNantunSportCenterBotService(browser, nantunSportCenterService, cfg)
	handler := tgbot.NewMessageHandler(botService, &nantunSportCenterBotService)

	// 設定訊息處理
	botService.HandleMessage(handler.HandleUpdate)

	// 使用 context 控制 Bot 的生命週期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 開始接收訊息
	botService.StartReceiveMessage()

	// 阻塞主程式，直到收到取消信號（例如 Ctrl+C）
	<-ctx.Done()

	logger.Log.Info("Shutting down Telegram Bot...")
	// #endregion
}
