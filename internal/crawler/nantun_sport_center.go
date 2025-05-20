package crawler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type NantunSportCenterInterface interface {
}

var _ NantunSportCenterInterface = (*NantunSportCenterService)(nil)

type NantunSportCenterService struct {
	browserService browser.BrowserService
	Nantun_Url     string // 南屯運動中心網址
	paymentURL     string // 繳費網址
}

func NewNantunSportCenterService(browserService browser.BrowserService) NantunSportCenterService {
	return NantunSportCenterService{
		browserService: browserService,
		Nantun_Url:     "https://nd01.xuanen.com.tw/BPMember/BPMemberLogin",
		paymentURL:     "https://nd01.xuanen.com.tw/BPMemberOrder/BPMemberOrder",
	}
}

// 快速預定場地
func (s *NantunSportCenterService) QuickCrawlerNantun(cfg config.Config) error {
	page, err := s.browserService.GetPage(s.Nantun_Url)
	if err != nil {
		return err
	}

	if err := s.login(page, cfg); err != nil {
		return err
	}

	for _, buttonIndex := range cfg.ButtonIndex {

		if err := s.clickAgreeButton(page); err != nil {
			return err
		}

		if err := s.selectLocationBooking(page); err != nil {
			return err
		}

		if err := s.selectBadminton(page); err != nil {
			return err
		}

		if err := s.setCheckboxAndProceed(page); err != nil {
			return err
		}

		if err := s.proceedToBooking(page); err != nil {
			return err
		}

		if err := s.selectTimeSlot(page, s.convertDayPeriodToTimeSlot(cfg.DayPeriod)); err != nil {
			return err
		}

		if err := s.fastSelectLastDate(page); err != nil {
			return err
		}

		if err := s.fastBookCourt(page, buttonIndex); err != nil {
			return err
		}
	}
	s.browserService.Close()
	return nil
}

// 爬蟲南屯運動中心
func (s *NantunSportCenterService) CrawlerNantun(cfg config.Config) error {
	page, err := s.browserService.GetPage(s.Nantun_Url)
	if err != nil {
		return err
	}

	if err := s.login(page, cfg); err != nil {
		return err
	}

	bookCount := 0
	for _, timeSlotCode := range cfg.TimeSlotCodes {
		if err := s.clickAgreeButton(page); err != nil {
			return err
		}

		if err := s.selectLocationBooking(page); err != nil {
			continue
		}

		if err := s.selectBadminton(page); err != nil {
			continue
		}

		if err := s.setCheckboxAndProceed(page); err != nil {
			return err
		}

		if err := s.proceedToBooking(page); err != nil {
			return err
		}

		if err := s.selectDate(page, cfg.ChooseWeekday); err != nil {
			return err
		}

		if err := s.selectTimeSlot(page, timeSlotCode); err != nil {
			continue
		}

		cleanSlots, err := s.getAllAvailableTimeSlots(page)
		if err != nil {
			continue
		}

		targetSlot := s.findAvailableCourtsByTimeSlot(cleanSlots, timeSlotCode)

		if err := s.bookCourt(page, targetSlot); err != nil {
			logger.Log.Error(fmt.Sprintf("預約時段 %v 失敗: %s", types.TimeSlotMap[timeSlotCode], err))
			continue
		}

		bookCount++
	}

	if err := s.clickAgreeButton(page); err != nil {
		return err
	}

	if bookCount > 0 {
		if err := s.navigateToPayment(page); err != nil {
			return err
		}
	}

	s.browserService.Close()
	return nil
}

// 執行登入
func (s *NantunSportCenterService) login(page *rod.Page, cfg config.Config) error {
	if err := page.MustElement("#txt_Account").Input(cfg.ID); err != nil {
		logger.Log.Error("無法輸入身分證字號: " + err.Error())
		return err
	}
	logger.Log.Info("填寫身分證字號")

	if err := page.MustElement("#txt_Pass").Input(cfg.Password); err != nil {
		logger.Log.Error("無法輸入密碼: " + err.Error())
		return err
	}
	logger.Log.Info("填寫密碼")

	if err := page.MustElement(".CssLoginBtn").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊登入按鈕: " + err.Error())
		return err
	}
	logger.Log.Info("點擊登入按鈕")

	page.MustWaitStable()
	return nil
}

// 點擊預防詐騙確認按鈕
func (s *NantunSportCenterService) clickAgreeButton(page *rod.Page) error {
	if err := page.MustElement("#Msg_Agree").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊確認按鈕: " + err.Error())
		return err
	}
	logger.Log.Info("點擊確認按鈕")
	page.MustWaitStable()
	return nil
}

// 點選場地預約
func (s *NantunSportCenterService) selectLocationBooking(page *rod.Page) error {
	if _, err := page.Eval(`() => {
		const element = document.querySelector('#location');
		if (element) {
			next(3);
			return true;
		}
		return false;
	}`); err != nil {
		logger.Log.Error("無法觸發場地預約按鈕的 onclick 事件: " + err.Error())
		return err
	}
	logger.Log.Info("觸發場地預約按鈕的 onclick 事件")
	page.MustWaitStable()
	return nil
}

// 點擊羽球按鈕
func (s *NantunSportCenterService) selectBadminton(page *rod.Page) error {
	if err := page.MustElement(".CssAdImg[data-slick-index='0']").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊羽球按鈕: " + err.Error())
		return err
	}
	logger.Log.Info("點擊羽球按鈕")
	page.MustWaitStable()
	return nil
}

// 設定勾選框狀態
func (s *NantunSportCenterService) setCheckboxAndProceed(page *rod.Page) error {
	if _, err := page.Eval(`() => {
		const checkbox = document.querySelector('#isRememberAcc');
		if (checkbox) {
			checkbox.checked = true;
			checkclick();
			return true;
		}
		return false;
	}`); err != nil {
		logger.Log.Error("無法設定勾選框狀態和觸發點擊事件: " + err.Error())
		return err
	}

	page.MustWaitStable()
	logger.Log.Info("設定勾選框狀態和觸發點擊事件")
	return nil
}

// 點選預約場地
func (s *NantunSportCenterService) proceedToBooking(page *rod.Page) error {
	if _, err := page.Eval(`() => {
		next();
		return true;
	}`); err != nil {
		logger.Log.Error("無法觸發預約場地按鈕的 onclick 事件: " + err.Error())
		return err
	}

	page.MustWaitStable()
	logger.Log.Info("觸發預約場地按鈕的 onclick 事件")
	return nil
}

// 選擇日期
func (s *NantunSportCenterService) selectDate(page *rod.Page, targetWeekday string) error {
	weekdays := []string{}

	firstDatebox := page.MustElement(".datebox")
	dateElements := firstDatebox.MustElements("div")

	for _, element := range dateElements {
		weekday := element.MustText()
		weekdays = append(weekdays, weekday)
	}

	logger.Log.Info("取得的星期資訊:")
	for i, day := range weekdays {
		logger.Log.Info(fmt.Sprintf("%d: %s", i+1, day))
	}

	secondDatebox := page.MustElements("div.datebox")
	if len(secondDatebox) < 2 {
		logger.Log.Error("找不到第二個日期框")
		return fmt.Errorf("找不到第二個日期框")
	}

	dateButtons := secondDatebox[1].MustElements("div")
	if len(dateButtons) < len(weekdays) {
		logger.Log.Error(fmt.Sprintf("日期按鈕數量不足，只有 %d 個按鈕", len(dateButtons)))
		return fmt.Errorf("日期按鈕數量不足")
	}

	weekdayIndex := -1
	for i, day := range weekdays {
		if day == targetWeekday {
			weekdayIndex = i
			break
		}
	}

	if weekdayIndex == -1 {
		logger.Log.Error(fmt.Sprintf("找不到星期%s", targetWeekday))
		return fmt.Errorf("找不到指定的星期")
	}

	logger.Log.Info(fmt.Sprintf("找到星期%s，索引為 %d", targetWeekday, weekdayIndex))

	dateToClick := dateButtons[weekdayIndex]
	dateText := dateToClick.MustText()
	logger.Log.Info(fmt.Sprintf("選擇的日期是: %s", dateText))

	if err := dateToClick.Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error(fmt.Sprintf("點選日期失敗: %s", err))
		return err
	}

	page.MustWaitStable()
	logger.Log.Info("日期點選成功")
	return nil
}

// 前往繳費頁面
func (s *NantunSportCenterService) navigateToPayment(page *rod.Page) error {
	if err := page.Navigate(s.paymentURL); err != nil {
		return err
	}
	page.MustWaitStable()
	return nil
}

// SelectTimeSlot 選擇時段（1=上午，2=下午，3=晚上）
func (s *NantunSportCenterService) selectTimeSlot(page *rod.Page, timeSlotCode types.TimeSlotCode) error {

	// 判斷時段，1-12 為上午，13-18 為下午，19-24 為晚上
	var timeSlot int
	if timeSlotCode <= types.TimeSlot_11_12 {
		timeSlot = 1
	} else if timeSlotCode <= types.TimeSlot_17_18 {
		timeSlot = 2
	} else {
		timeSlot = 3
	}

	// 檢查時段參數是否有效
	if timeSlot < 1 || timeSlot > 3 {
		return fmt.Errorf("無效的時段參數：%d，必須是 1-3 之間", timeSlot)
	}

	// 使用 JavaScript 找到對應的時段按鈕並點擊
	script := fmt.Sprintf(`() => {
        const timeSlots = document.querySelectorAll('.selectweek');
        if (timeSlots.length === 0) {
            return false;
        }
        Selecttime(%d);
        return true;
    }`, timeSlot)

	result, err := page.Eval(script)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("執行時段選擇腳本失敗: %s", err))
		return err
	}

	if !result.Value.Bool() {
		logger.Log.Error("找不到時段選擇按鈕")
		return fmt.Errorf("找不到時段選擇按鈕")
	}

	// 等待頁面載入完成
	page.MustWaitStable()

	// 記錄選擇的時段
	timeSlotNames := map[int]string{1: "上午", 2: "下午", 3: "晚上"}
	logger.Log.Info(fmt.Sprintf("已選擇%s時段", timeSlotNames[timeSlot]))
	return nil
}

// GetAvailableTimeSlots 取得所有可預約的時段資訊
func (s *NantunSportCenterService) getAllAvailableTimeSlots(page *rod.Page) ([]types.CleanTimeSlot, error) {
	// 設定超時和等待參數
	timeout := 10 * time.Second
	page.Timeout(timeout)

	// 等待頁面加載完成
	err := page.WaitStable(2 * time.Second)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("等待頁面穩定失敗: %s", err))
		return nil, err
	}

	// 搜尋所有時段元素
	listItems, err := page.Elements("div.listbackground > div.imformation1, div.listbackground > div.imformation2")
	if err != nil {
		logger.Log.Error(fmt.Sprintf("找不到時段元素: %s", err))
		return nil, err
	}

	logger.Log.Info(fmt.Sprintf("找到 %d 個時段元素", len(listItems)))

	var slots []types.TimeSlot

	// 處理每個時段元素
	for _, item := range listItems {
		// 確保元素有效
		if item == nil {
			continue
		}

		// 檢查是否有 listbtn（可預約按鈕）
		var bookBtn *rod.Element
		bookBtn, err = item.Element("div.courseintro div.listbtn")
		if err != nil || bookBtn == nil {
			// 沒有 listbtn，表示不可預約
			continue
		}

		slot := types.TimeSlot{}

		// 取得場地名稱和價格
		listTexts, err := item.Elements("div.textcss div.listtext")
		if err == nil && len(listTexts) >= 2 {
			// 場地名稱（第一個 listtext）
			courtName, err := listTexts[0].Text()
			if err == nil {
				slot.CourtName = strings.TrimSpace(courtName)
			}

			// 價格（第二個 listtext）
			price, err := listTexts[1].Text()
			if err == nil {
				slot.Price = strings.TrimSpace(price)
			}

			// 時間（第三個 listtext）
			if len(listTexts) >= 3 {
				time, err := listTexts[2].Text()
				if err == nil {
					slot.Time = strings.TrimSpace(time)
				}
			}
		}

		// 取得預約參數
		onclick, err := bookBtn.Attribute("onclick")
		if err == nil && onclick != nil && *onclick != "" {
			// 解析預約參數 (例如 DoSubmit2(84,'2025-05-13',6,250))
			re := regexp.MustCompile(`DoSubmit2\((\d+),['"](\S+)['"],(\d+),(\d+)\)`)
			matches := re.FindStringSubmatch(*onclick)

			if len(matches) >= 5 {
				slot.Button = *onclick
				slot.Date = matches[2]
				slot.Period = matches[3]
			}
		}

		slots = append(slots, slot)
	}

	cleanSlots := make([]types.CleanTimeSlot, 0, len(slots))
	for _, slot := range slots {
		cleanSlots = append(cleanSlots, types.CleanTimeSlot{
			CourtName: slot.CourtName,
			Price:     slot.Price,
			Time:      slot.Time,
			BookingID: slot.BookingID,
			Date:      slot.Date,
			Period:    slot.Period,
			Fee:       slot.Fee,
			Button:    slot.Button,
		})
	}
	logger.Log.Info(fmt.Sprintf("找到 %d 個可預約時段:", len(cleanSlots)))
	return cleanSlots, nil
}

// 根據時段代碼查找可用場地
func (s *NantunSportCenterService) findAvailableCourtsByTimeSlot(slots []types.CleanTimeSlot, code types.TimeSlotCode) []types.CleanTimeSlot {
	targetTime := types.TimeSlotMap[code]
	var availableCourts []types.CleanTimeSlot

	for _, slot := range slots {
		if slot.Time == targetTime {
			availableCourts = append(availableCourts, slot)
		}
	}

	logger.Log.Info(fmt.Sprintf("找到 %d 個 %v 點可預約時段:", len(availableCourts), targetTime))
	return availableCourts
}

// 預約指定場地
func (s *NantunSportCenterService) bookCourt(page *rod.Page, targetSlot []types.CleanTimeSlot) error {
	for _, slot := range targetSlot {
		// 從 Button 字符串中提取參數
		re := regexp.MustCompile(`DoSubmit2\((\d+),['"](\S+)['"],(\d+),(\d+)\)`)
		matches := re.FindStringSubmatch(slot.Button)
		if len(matches) < 5 {
			logger.Log.Error("無法解析預約參數")
			continue
		}

		// 構建 JavaScript 函數調用
		script := fmt.Sprintf(`() => {
            try {
                %s;
                return true;
            } catch (e) {
                console.error(e);
                return false;
            }
        }`, slot.Button)

		// 執行 JavaScript
		result, err := page.Eval(script)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("執行預約腳本失敗: %s", err))
			continue
		}

		// 檢查是否成功執行
		if !result.Value.Bool() {
			logger.Log.Error("預約腳本執行失敗")
			continue
		}

		// 等待頁面跳轉或更新
		page.MustWaitStable()

		// 檢查是否跳轉到預約確認頁面
		currentURL := page.MustInfo().URL
		if strings.Contains(currentURL, "tFlag=2") {

			// 點擊確認按鈕
			confirmScript := fmt.Sprintf(`() => {
                try {
                    DoSubmit3('%s');
                    return true;
                } catch (e) {
                    console.error(e);
                    return false;
                }
            }`, matches[2])
			logger.Log.Info(fmt.Sprintf("執行確認腳本: %s", confirmScript))

			// 執行確認按鈕點擊
			confirmResult, err := page.Eval(confirmScript)
			if err != nil {
				logger.Log.Error(fmt.Sprintf("執行確認按鈕點擊失敗: %s", err))
				continue
			}

			// 檢查確認按鈕是否點擊成功
			if !confirmResult.Value.Bool() {
				logger.Log.Error("確認按鈕點擊失敗")
				continue
			}

			// 等待最終確認頁面載入
			page.MustWaitStable()
			logger.Log.Info(fmt.Sprintf("成功預約場地：%s，時間：%s", slot.CourtName, slot.Time))

			// 點擊首頁按鈕返回
			script := `() => {
				try {
					window.location = '/BPHome/BPHome';
					return true;
				} catch (e) {
					console.error(e);
					return false;
				}
			}`

			// 執行返回首頁腳本
			result, err := page.Eval(script)
			if err != nil {
				logger.Log.Error(fmt.Sprintf("執行返回首頁腳本失敗: %s", err))
				return err
			}

			// 檢查是否成功執行
			if !result.Value.Bool() {
				logger.Log.Error("返回首頁失敗")
				return fmt.Errorf("返回首頁失敗")
			}

			// 等待頁面載入完成
			page.MustWaitStable()
			logger.Log.Info("成功返回首頁")

			return nil // 完成預約流程後返回
		}
	}

	return fmt.Errorf("所有場地預約嘗試均失敗")
}

// 快速點選最新日期
func (s *NantunSportCenterService) fastSelectLastDate(page *rod.Page) error {

	secondDatebox := page.MustElements("div.datebox")
	if len(secondDatebox) < 2 {
		logger.Log.Error("找不到第二個日期框")
		return fmt.Errorf("找不到第二個日期框")
	}

	dateButtons := secondDatebox[1].MustElements("div")
	dateToClick := dateButtons[6]

	if err := dateToClick.Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error(fmt.Sprintf("點選日期失敗: %s", err))
		return err
	}

	// 使用 JavaScript 查找並點擊最後一個可用日期
	script := `() => {
	    // 先嘗試找帶有 selectweek class 的日期按鈕
	    let dateButton = document.querySelector('div.selectweek[onclick*="SelectDate"]');
	    if (!dateButton) {
	        // 如果沒有找到，則查找普通的日期按鈕
	        const allDateButtons = Array.from(document.querySelectorAll('div[onclick*="SelectDate"]'));
	        if (allDateButtons.length > 0) {
	            dateButton = allDateButtons[allDateButtons.length - 1];
	        }
	    }

	    if (dateButton) {
	        // 從 onclick 屬性中提取日期
	        const onclickAttr = dateButton.getAttribute('onclick');
	        const dateMatch = onclickAttr.match(/'(\d{4}-\d{2}-\d{2})'/);
	        if (dateMatch) {
	            SelectDate(dateMatch[1]);
	            return true;
	        }
	    }
	    return false;
	}`

	for {
		currentTime := time.Now()
		if currentTime.Hour() == 12 { // 12 點後停止點擊
			break
		}
		if currentTime.Hour() == 12 && currentTime.Minute() == 59 { // 12:59分後開始點
			// 執行腳本
			result, err := page.Eval(script)
			if err != nil {
				logger.Log.Error(fmt.Sprintf("執行日期選擇腳本失敗: %s", err))
				return err
			}

			if !result.Value.Bool() {
				logger.Log.Error("找不到可點擊的日期按鈕")
				return fmt.Errorf("找不到可點擊的日期按鈕")
			}

			// 等待頁面穩定
			page.MustWaitStable()
		}
	}

	logger.Log.Info("日期點選成功")
	return nil
}

// 快速預約場地
// 預約指定場地
func (s *NantunSportCenterService) fastBookCourt(page *rod.Page, buttonIndex int) error {
	// 使用 JavaScript 找到所有預約按鈕
	script := `() => {
        const buttons = document.querySelectorAll('.listbtn[onclick*="DoSubmit2"]');
        return Array.from(buttons).map(btn => btn.getAttribute('onclick'));
    }`

	// 執行腳本獲取所有按鈕的 onclick 屬性
	result, err := page.Eval(script)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("獲取預約按鈕失敗: %s", err))
		return err
	}

	// 將結果轉換為字符串切片
	var buttons []string
	if err := result.Value.Unmarshal(&buttons); err != nil {
		logger.Log.Error(fmt.Sprintf("解析按鈕資訊失敗: %s", err))
		return err
	}

	// 您可以指定要點擊第幾個按鈕（例如第一個按鈕索引為 0）
	if buttonIndex >= len(buttons) {
		return fmt.Errorf("指定的按鈕索引 %d 超出範圍，總共有 %d 個按鈕", buttonIndex, len(buttons))
	}

	// 從選定按鈕的 onclick 屬性中提取參數
	selectedButton := buttons[buttonIndex]
	re := regexp.MustCompile(`DoSubmit2\((\d+),['"](\S+)['"],(\d+),(\d+)\)`)
	matches := re.FindStringSubmatch(selectedButton)
	if len(matches) < 5 {
		return fmt.Errorf("無法解析選定按鈕的預約參數")
	}

	// 執行預約
	bookScript := fmt.Sprintf(`() => {
        try {
            DoSubmit2(%s,'%s',%s,%s);
            return true;
        } catch (e) {
            console.error(e);
            return false;
        }
    }`, matches[1], matches[2], matches[3], matches[4])

	// 執行預約腳本
	bookResult, err := page.Eval(bookScript)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("執行預約腳本失敗: %s", err))
		return err
	}

	if !bookResult.Value.Bool() {
		return fmt.Errorf("預約失敗")
	}

	// 等待頁面跳轉或更新
	page.MustWaitStable()

	// 檢查是否跳轉到預約確認頁面
	currentURL := page.MustInfo().URL
	if strings.Contains(currentURL, "tFlag=2") {
		// 點擊確認按鈕
		confirmScript := fmt.Sprintf(`() => {
            try {
                DoSubmit3('%s');
                return true;
            } catch (e) {
                console.error(e);
                return false;
            }
        }`, matches[2])

		// 執行確認按鈕點擊
		confirmResult, err := page.Eval(confirmScript)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("執行確認按鈕點擊失敗: %s", err))
			return err
		}

		if !confirmResult.Value.Bool() {
			return fmt.Errorf("確認按鈕點擊失敗")
		}

		// 等待最終確認頁面載入
		page.MustWaitStable()
		logger.Log.Info("成功預約場地")

		// 返回首頁
		script := `() => {
            try {
                window.location = '/BPHome/BPHome';
                return true;
            } catch (e) {
                console.error(e);
                return false;
            }
        }`

		result, err := page.Eval(script)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("執行返回首頁腳本失敗: %s", err))
			return err
		}

		if !result.Value.Bool() {
			return fmt.Errorf("返回首頁失敗")
		}

		page.MustWaitStable()
		logger.Log.Info("成功返回首頁")

		return nil
	}

	return fmt.Errorf("預約流程未完成")
}

func (s *NantunSportCenterService) convertDayPeriodToTimeSlot(dayPeriod int) types.TimeSlotCode {
	switch dayPeriod {
	case 1: // 上午
		return types.TimeSlot_8_9 // 或其他合適的上午時段
	case 2: // 下午
		return types.TimeSlot_14_15 // 或其他合適的下午時段
	case 3: // 晚上
		return types.TimeSlot_19_20 // 或其他合適的晚上時段
	default:
		return types.TimeSlot_8_9 // 默認返回上午時段
	}
}
