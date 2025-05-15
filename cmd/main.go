package main

import (
	tgbot "github.com/tian841224/crawler_sportcenter/internal/bot/tg_bot"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

func main() {

	// 初始化 Logger
	logger.InitLogger()
	logger.Log.Info("初始化 Logger")

	// 載入設定檔
	cfg := config.LoadConfig()
	logger.Log.Info("載入設定檔")

	// 初始化瀏覽器
	browser := browser.NewBrowserService()
	logger.Log.Info("初始化瀏覽器")
	nantunSportCenterService := crawler.NewNantunSportCenterService(browser)
	nantunSportCenterService.CrawlerNantun(cfg)

	botService := tgbot.NewTGBotService(cfg)
    handler := tgbot.NewMessageHandler(botService)
    
    // 設置消息處理
    botService.HandleMessage(handler.HandleUpdate)
    
    // 開始接收消息
    botService.StartReceiveMessage()
}
