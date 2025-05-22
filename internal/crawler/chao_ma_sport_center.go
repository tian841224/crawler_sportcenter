package crawler

import (
	"github.com/go-rod/rod/lib/proto"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type ChaoMaSportCenterService struct {
	browserService browser.BrowserService
	Chao_Ma_Url    string // 朝馬運動中心網址
}

func NewChaoMaSportCenterService(browserService browser.BrowserService) *ChaoMaSportCenterService {
	return &ChaoMaSportCenterService{
		browserService: browserService,
		Chao_Ma_Url:    "https://scr.cyc.org.tw/tp11.aspx?module=login_page&files=login",
	}
}

// 爬蟲朝馬運動中心
func (s *ChaoMaSportCenterService) CrawlerChaoMa() error {
	// 建立新頁面
	page, err := s.browserService.GetPage(s.Chao_Ma_Url, "")
	if err != nil {
		logger.Log.Error("無法創建新頁面: " + err.Error())
		return err
	}

	// 讀取網站
	if err := page.Navigate(s.Chao_Ma_Url); err != nil {
		logger.Log.Error("無法讀取網站: " + err.Error())
		return err
	}

	logger.Log.Info("讀取網站")

	// 等待並點擊防詐騙訊息按鈕
	if err := page.MustElement("button.swal2-confirm.swal2-styled").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊確認按鈕: " + err.Error())
	}

	logger.Log.Info("點擊防詐騙訊息按鈕")

	// 等待表單元素載入
	if err := page.MustElement("#ContentPlaceHolder1_loginid").WaitVisible(); err != nil {
		logger.Log.Error("無法找到登入表單: " + err.Error())
		return err
	}

	logger.Log.Info("等待表單元素載入")

	// 填寫身分證字號
	if err := page.MustElement("#ContentPlaceHolder1_loginid").Input("L124035685"); err != nil {
		logger.Log.Error("無法輸入身分證字號: " + err.Error())
		return err
	}

	logger.Log.Info("填寫身分證字號")

	// 填寫密碼
	if err := page.MustElement("#loginpw").Input("j25319456"); err != nil {
		logger.Log.Error("無法輸入密碼: " + err.Error())
		return err
	}

	logger.Log.Info("填寫密碼")

	// 點擊登入按鈕
	if err := page.MustElement("#login_but").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊登入按鈕: " + err.Error())
		return err
	}

	logger.Log.Info("點擊登入按鈕")

	// 等待頁面載入完成
	page.MustWaitStable()

	// 點擊羽球圖片
	if err := page.MustElement("img[src='img/ICON2/ico02-01.png']").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊羽球圖片: " + err.Error())
		return err
	}

	logger.Log.Info("成功點擊羽球圖片")
	return nil
}
