package main

import (
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
	nantunSportCenterService.QuickCrawlerNantun(cfg)
}
