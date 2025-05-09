package web

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type NantunSportCenterInterface interface {
}

var _ NantunSportCenterInterface = (*NantunSportCenterService)(nil)

type NantunSportCenterService struct {
	crawlerService crawler.CrawlerService
	Nantun_Url     string // 南屯運動中心網址
}

func NewSportCenterService(crawlerService crawler.CrawlerService) NantunSportCenterService {
	return NantunSportCenterService{
		crawlerService: crawlerService,
		Nantun_Url:     "https://nd01.xuanen.com.tw/BPMember/BPMemberLogin",
	}
}

// 爬蟲南屯運動中心
func (s *NantunSportCenterService) CrawlerNantun(cfg config.Config) error {

	// #region 設定頁面
	// 初始化瀏覽器
	s.crawlerService.InitBrowser()
	// 建立新頁面
	page, err := s.crawlerService.GetPage()
	if err != nil {
		return err
	}

	// 設定網頁模式
	s.crawlerService.SetWebMode(true)

	// 讀取網站
	if err = page.Navigate(s.Nantun_Url); err != nil {
		return err
	}

	// #endregion

	// #region 登入
	// 填寫身分證字號
	if err = page.MustElement("#txt_Account").Input(cfg.ID); err != nil {
		logger.Log.Error("無法輸入身分證字號: " + err.Error())
		return err
	}

	logger.Log.Info("填寫身分證字號")

	// 填寫密碼
	if err = page.MustElement("#txt_Pass").Input(cfg.Password); err != nil {
		logger.Log.Error("無法輸入密碼: " + err.Error())
		return err
	}

	logger.Log.Info("填寫密碼")

	// 點擊登入按鈕
	if err = page.MustElement(".CssLoginBtn").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊登入按鈕: " + err.Error())
		return err
	}

	logger.Log.Info("點擊登入按鈕")

	// 等待頁面載入完成
	page.MustWaitStable()

	// 點擊確認按鈕
	if err = page.MustElement("#Msg_Agree").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊確認按鈕: " + err.Error())
		return err
	}

	logger.Log.Info("點擊確認按鈕")

	// 等待頁面載入完成
	page.MustWaitStable()

	// #endregion

	// #region 點選場地預約
	// 使用 JavaScript 觸發 onclick 事件
	if _, err = page.Eval(`() => {
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

	// 等待頁面載入完成
	page.MustWaitStable()
	// #endregion

	// #region 點擊羽球按鈕
	if err = page.MustElement(".CssAdImg[data-slick-index='0']").Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error("無法點擊羽球按鈕: " + err.Error())
		return err
	}

	logger.Log.Info("點擊羽球按鈕")

	// 等待頁面載入完成
	page.MustWaitStable()
	// #endregion

	// #region 點擊羽球按鈕
	// 使用 JavaScript 設定勾選框狀態並觸發點擊事件
	if _, err = page.Eval(`() => {
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

	logger.Log.Info("設定勾選框狀態和觸發點擊事件")
	// #endregion

	// #region 點選預約場地
	// 使用 JavaScript 觸發 onclick 事件
	if _, err = page.Eval(`() => {
		next();
		return true;
	}`); err != nil {
		logger.Log.Error("無法觸發預約場地按鈕的 onclick 事件: " + err.Error())
		return err
	}

	logger.Log.Info("觸發預約場地按鈕的 onclick 事件")
	// #endregion

	// #region 點選日期對應按鈕
	// 取得日期
	weekdays := []string{}

	// 選擇第一個 .datebox 中的所有 div 元素
	firstDatebox := page.MustElement(".datebox")
	dateElements := firstDatebox.MustElements("div")

	// 遍歷每個元素取出文字，將順序儲存到weekdays
	for _, element := range dateElements {
		weekday := element.MustText()
		weekdays = append(weekdays, weekday)
	}

	// 輸出結果
	logger.Log.Info("取得的星期資訊:")
	for i, day := range weekdays {
		logger.Log.Info(fmt.Sprintf("%d: %s", i+1, day))
	}

	// 選擇對應日期
	secondDatebox := page.MustElements("div.datebox")
	if len(secondDatebox) < 2 {
		logger.Log.Error("找不到第二個日期框")
		return fmt.Errorf("找不到第二個日期框")
	}

	dateButtons := secondDatebox[1].MustElements("div")

	// 檢查找到的按鈕數量是否足夠
	if len(dateButtons) < len(weekdays) {
		logger.Log.Error(fmt.Sprintf("日期按鈕數量不足，只有 %d 個按鈕", len(dateButtons)))
		return fmt.Errorf("日期按鈕數量不足")
	}

	// 設定要選擇的星期幾
	targetWeekday := cfg.ChooseWeekday

	// 尋找對應的星期幾索引
	weekdayIndex := -1
	for i, day := range weekdays {
		if day == targetWeekday {
			weekdayIndex = i
			break
		}
	}

	// 檢查是否找到對應的星期
	if weekdayIndex == -1 {
		logger.Log.Error(fmt.Sprintf("找不到星期%s", targetWeekday))
		return fmt.Errorf("找不到指定的星期")
	}

	logger.Log.Info(fmt.Sprintf("找到星期%s，索引為 %d", targetWeekday, weekdayIndex))

	dateToClick := dateButtons[weekdayIndex]

	// 打印要點選的日期
	dateText := dateToClick.MustText()
	logger.Log.Info(fmt.Sprintf("選擇的日期是: %s", dateText))

	// 執行點選
	if err = dateToClick.Click(proto.InputMouseButtonLeft, 1); err != nil {
		logger.Log.Error(fmt.Sprintf("點選日期失敗: %s", err))
		return err
	}

	// 等待頁面載入完成
	page.MustWaitStable()
	logger.Log.Info("日期點選成功")
	// #endregion

	// #region 點選選擇的時段（1=上午，2=下午，3=晚上）
	if err = s.selectTimeSlot(page, 2); err != nil { // 這裡的 1 表示選擇上午時段
		logger.Log.Error("選擇時段失敗: " + err.Error())
		return err
	}
	// #endregion

	// #region 取得所有可預約時段資訊
	cleanSlots, err := s.getAllAvailableTimeSlots(page)
	if err != nil {
		logger.Log.Error("取得可預約時段資訊失敗: " + err.Error())
	}
	// #endregion

	// #region 篩選出指定要預約的場地
	targetSlot := s.findAvailableCourtsByTimeSlot(cleanSlots, cfg.TimeSlotCode)
	// #endregion

	// #region 預約場地
	s.bookCourt(page, targetSlot)
	// #endregion

	s.crawlerService.Close()
	return nil
}

// SelectTimeSlot 選擇時段（1=上午，2=下午，3=晚上）
func (s *NantunSportCenterService) selectTimeSlot(page *rod.Page, timeSlot int) error {
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
                DoSubmit2(%s,'%s',%s,%s);
                return true;
            } catch (e) {
                console.error(e);
                return false;
            }
        }`, matches[1], matches[2], matches[3], matches[4])

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
			return nil // 完成預約流程後返回
		}
	}

	return fmt.Errorf("所有場地預約嘗試均失敗")
}
