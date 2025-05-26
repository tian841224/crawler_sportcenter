package browser

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"

	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type BrowserInterface interface {
	initBrowser() error
	initPage(tag string) (*rod.Page, error)
	setWebMode(isMobileMode bool) error
	GetPage(url string, tag string) (*rod.Page, error)
	SwitchToPageByTag(tag string) (*rod.Page, error)
	Close() error
}

var _ BrowserInterface = (*BrowserService)(nil)

type BrowserService struct {
	browser *rod.Browser
	page    *rod.Page
	pages   []PageInfo
}

type PageInfo struct {
	Page  *rod.Page
	Tag   string
	pages []PageInfo
}

func NewBrowserService() BrowserService {
	return BrowserService{
		pages: make([]PageInfo, 0),
	}
}

// 取得頁面
func (s *BrowserService) GetPage(url string, tag string) (*rod.Page, error) {
	if s.browser == nil {
		if err := s.initBrowser(); err != nil {
			return nil, err
		}
	}

	if s.hasTag(tag) {
		return s.SwitchToPageByTag(tag)
	}

	page, err := s.initPage(tag)
	if err != nil {
		return nil, err
	}
	s.setWebMode(true)

	if err = page.Navigate(url); err != nil {
		return nil, err
	}

	return page, nil
}

// 根據標籤切換頁面
func (s *BrowserService) SwitchToPageByTag(tag string) (*rod.Page, error) {
	for _, p := range s.pages {
		if p.Tag == tag {
			s.page = p.Page
			s.page.MustActivate()
			return p.Page, nil
		}
	}
	return nil, fmt.Errorf("找不到標籤為 %s 的頁面", tag)
}

// 關閉瀏覽器
func (s *BrowserService) Close() error {
	for _, pageInfo := range s.pages {
		if pageInfo.Page != nil {
			pageInfo.Page.Close()
		}
	}

	if s.browser != nil {
		return s.browser.Close()
	}
	return nil
}

// 初始化爬蟲
func (s *BrowserService) initBrowser() error {
	// 設定瀏覽器啟動選項
	l := launcher.New().
		Headless(false).
		Leakless(false). // Disable leakless mode
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-features", "IsolateOrigins,site-per-process").
		Devtools(false).
		NoSandbox(true)

	path := l.MustLaunch()

	// 初始化瀏覽器
	browser := rod.New().ControlURL(path).MustConnect()

	s.browser = browser

	return nil
}

// 取得頁面
func (s *BrowserService) initPage(tag string) (*rod.Page, error) {
	// 建立新頁面
	s.page = stealth.MustPage(s.browser)

	err := s.page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		AcceptLanguage: "zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	})

	if err != nil {
		logger.Log.Error("設定 User Agent 失敗:" + err.Error())
		return nil, err
	}

	// 注入反檢測腳本
	_, err = s.page.Eval(`() => {
		Object.defineProperty(navigator, 'webdriver', { get: () => false });
		Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
		Object.defineProperty(navigator, 'languages', { get: () => ['zh-TW', 'zh', 'en-US', 'en'] });
	}`)

	if err != nil {
		logger.Log.Error("注入反檢測腳本失敗:" + err.Error())
		return nil, err
	}

	// 注入反彈窗腳本
	_, err = s.page.Eval(`() => {
			window.alert = () => {};
			window.confirm = () => true;
			window.prompt = () => null;
		}`)

	if err != nil {
		logger.Log.Error("注入反彈窗腳本失敗:" + err.Error())
		return nil, err
	}

	s.pages = append(s.pages, PageInfo{
		Page: s.page,
		Tag:  tag,
	})
	s.pages = append(s.pages, PageInfo{
		Page: s.page,
		Tag:  tag,
	})
	return s.page, nil
}

// 設定網頁模式
func (s *BrowserService) setWebMode(isMobileMode bool) error {
	if s.page == nil {
		return nil
	}

	// 使用正確的 proto.NetworkSetUserAgentOverride 結構
	ua := &proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}

	if isMobileMode {
		ua.UserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Mobile/15E148 Safari/604.1"
		// 設置移動設備參數
		err := s.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:             375, // iPhone 的寬度
			Height:            812, // iPhone 的高度
			DeviceScaleFactor: 3,   // 設備像素比
			Mobile:            true,
		})
		if err != nil {
			return err
		}
	}

	return s.page.SetUserAgent(ua)
}

// 新增：檢查標籤是否存在
func (s *BrowserService) hasTag(tag string) bool {
	for _, pageInfo := range s.pages {
		if pageInfo.Tag == tag {
			return true
		}
	}
	return false
}
