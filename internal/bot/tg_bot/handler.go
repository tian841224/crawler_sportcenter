package tgbot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/domain/schedule"
	timeslot "github.com/tian841224/crawler_sportcenter/internal/domain/time_slot"
	"github.com/tian841224/crawler_sportcenter/internal/domain/user"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MessageHandler struct {
	bot                     TGBotInterface
	nantun_sport            crawler.NantunSportCenterBotInterface
	user                    user.Service
	timeslot                timeslot.Service
	schedule                schedule.Service
	userSelectionDate       string
	userSelectionWeekday    time.Weekday
	userSelectionTimeSlotID uint
	userSelectionTimeSlot   string
	settingState            map[int64]string // 新增：用於追蹤使用者的設定狀態
}

func NewMessageHandler(bot TGBotInterface, user user.Service, timeslot timeslot.Service, schedule schedule.Service, nantun_sport crawler.NantunSportCenterBotInterface) *MessageHandler {
	return &MessageHandler{
		bot:          bot,
		nantun_sport: nantun_sport,
		user:         user,
		timeslot:     timeslot,
		schedule:     schedule,
		settingState: make(map[int64]string), // 初始化 settingState
	}
}

// HandleUpdate 處理所有的更新消息
func (h *MessageHandler) HandleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		h.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		// 取TG ID
		id := update.CallbackQuery.From.ID
		_, err := h.getOrCreateUser(id)
		if err != nil {
			logger.Log.Error("get or create user", zap.Error(err))
			return
		}

		h.handleCallback(update.CallbackQuery)
	}
}

// #region 處理所有格式訊息
// 處理文字訊息
func (h *MessageHandler) handleMessage(message *tgbotapi.Message) {
	// 檢查是否在設定流程中
	if state, exists := h.settingState[message.From.ID]; exists {
		switch state {
		case "waiting_account":
			h.handleAccountInput(message)
			h.settingState[message.From.ID] = "waiting_password"
		case "waiting_password":
			h.handlePasswordInput(message)
			delete(h.settingState, message.From.ID) // 清除設定狀態
		}
		return
	}

	// 處理一般命令
	switch message.Text {
	case "/start":
		h.handleStart(message)
	case "/setting":
		h.handleSetting(message)
	default:
		h.handleDefault(message)
	}
}

// 處理按鈕回饋
const (
	callbackNantunSport = "nantun_sport"
	callbackBackToMain  = "back_to_main"
	prefixDate          = "date_"
	prefixTimeSlot      = "time_slot_"
	prefixBook          = "book_"
	prefixSubWeedDay    = "sub_weed_day_"
	prefixSubTimeSlot   = "sub_time_slot_"
)

func (h *MessageHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	switch {
	// 南屯運動中心
	case callback.Data == callbackNantunSport:
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleSportCenterSelection(callback)
	// 返回主選單
	case callback.Data == callbackBackToMain:
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleBackToMain(callback)
	// 日期選擇
	case strings.HasPrefix(callback.Data, prefixDate):
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleDateSelection(callback)
	// 時段選擇
	case strings.HasPrefix(callback.Data, prefixTimeSlot):
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleTimeSlotSelection(callback)
	// 預約場地
	case strings.HasPrefix(callback.Data, prefixBook):
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleBooking(callback)
	default:
		h.handleUnknownCallback(callback)
	}
}

func (h *MessageHandler) handleUnknownCallback(callback *tgbotapi.CallbackQuery) {
	text := "未知的選項，請重新選擇"
	h.bot.SendMessage(callback.Message.Chat.ID, text)
}

func (h *MessageHandler) handleBackToMain(callback *tgbotapi.CallbackQuery) {
	text := "請選擇您要查詢的場地"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("南屯運動中心", "nantun_sport"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("朝馬運動中心", "chao_ma_sport"),
		),
	)
	h.bot.SendeKeyboardMessage(callback.Message.Chat.ID, text, keyboard)
}

// 處理運動中心選擇
func (h *MessageHandler) handleSportCenterSelection(callback *tgbotapi.CallbackQuery) {
	text := "選擇訂閱時間"
	keyboard := h.createDateSelectionKeyboard()
	h.bot.SendeKeyboardMessage(callback.Message.Chat.ID, text, keyboard)
}

// 建立日期選擇鍵盤
func (h *MessageHandler) createDateSelectionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("日", "date_0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("一", "date_1"),
			tgbotapi.NewInlineKeyboardButtonData("二", "date_2"),
			tgbotapi.NewInlineKeyboardButtonData("三", "date_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("四", "date_4"),
			tgbotapi.NewInlineKeyboardButtonData("五", "date_5"),
			tgbotapi.NewInlineKeyboardButtonData("六", "date_6"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("返回主選單", "back_to_main"),
		),
	)
}

// 處理日期選擇
func (h *MessageHandler) handleDateSelection(callback *tgbotapi.CallbackQuery) {
	dayMap := map[string]string{
		"0": "日", "1": "一", "2": "二",
		"3": "三", "4": "四", "5": "五", "6": "六",
	}
	h.userSelectionDate = dayMap[callback.Data[5:]]

	weekdayInt, err := strconv.Atoi(callback.Data[5:])
	if err != nil {
		logger.Log.Error("invalid weekday", zap.String("weekday", callback.Data[5:]), zap.Error(err))
		return
	}
	h.userSelectionWeekday = time.Weekday(weekdayInt)
	logger.Log.Info("收到按鈕回調：" + h.userSelectionDate)

	text := "選擇訂閱時間"
	keyboard := h.createTimeSlotKeyboard()
	h.bot.SendeKeyboardMessage(callback.Message.Chat.ID, text, keyboard)
}

func (h *MessageHandler) createTimeSlotKeyboard() tgbotapi.InlineKeyboardMarkup {
	timeSlots := []struct {
		Text string
		Data string
	}{
		{Text: "6:00-7:00", Data: "time_slot_1"},
		{"7:00-8:00", "time_slot_2"},
		{"8:00-9:00", "time_slot_3"},
		{"9:00-10:00", "time_slot_4"},
		{"10:00-11:00", "time_slot_5"},
		{"11:00-12:00", "time_slot_6"},
		{"12:00-13:00", "time_slot_7"},
		{"13:00-14:00", "time_slot_8"},
		{"14:00-15:00", "time_slot_9"},
		{"15:00-16:00", "time_slot_10"},
		{"16:00-17:00", "time_slot_11"},
		{"17:00-18:00", "time_slot_12"},
		{"18:00-19:00", "time_slot_13"},
		{"19:00-20:00", "time_slot_14"},
		{"20:00-21:00", "time_slot_15"},
		{"21:00-22:00", "time_slot_16"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	// 每行放置3個按鈕
	for i := 0; i < len(timeSlots); i += 3 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 3 && i+j < len(timeSlots); j++ {
			slot := timeSlots[i+j]
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(slot.Text, slot.Data))
		}
		rows = append(rows, row)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("返回主選單", "back_to_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// 處理時段選擇
func (h *MessageHandler) handleTimeSlotSelection(callback *tgbotapi.CallbackQuery) {
	h.userSelectionTimeSlot = callback.Data[10:]
	num, _ := strconv.Atoi(h.userSelectionTimeSlot)
	timeSlotIDInt, err := strconv.Atoi(h.userSelectionTimeSlot)
	if err != nil {
		logger.Log.Error("invalid time slot", zap.String("time slot", h.userSelectionTimeSlot), zap.Error(err))
		return
	}
	timeSlotID := uint(timeSlotIDInt)

	userObj, err := h.user.GetByAccountID(context.Background(), fmt.Sprintf("%d", callback.Message.Chat.ID))
	if err != nil {
		logger.Log.Error("get user by id", zap.Error(err))
		return
	}
	if userObj == nil {
		err = h.user.Create(context.Background(), &user.User{
			AccountID: fmt.Sprintf("%d", callback.Message.Chat.ID),
			Status:    true,
		})
		if err != nil {
			logger.Log.Error("create user", zap.Error(err))
			return
		}
	}

	err = h.schedule.Create(context.Background(), &schedule.Schedule{
		UserID:     userObj.ID,
		Weekday:    h.userSelectionWeekday,
		TimeSlotID: &timeSlotID,
	})

	if err != nil {
		logger.Log.Error("create schedule", zap.Error(err))
		return
	}

	logger.Log.Info("User selected time slot: " + h.userSelectionTimeSlot)

	availableSlots, err := h.nantun_sport.GetAvailableTimeSlots(h.userSelectionDate, num, fmt.Sprint(callback.Message.Chat.ID))
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if len(availableSlots) == 0 {
		text := "目前無場地可預約，請重新選擇"
		h.bot.SendMessage(callback.Message.Chat.ID, text)
		return
	}

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	for _, slot := range availableSlots {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(slot.CourtName, "book_"+slot.Button),
		)
		keyboardRows = append(keyboardRows, row)
	}

	backRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("返回主選單", "back_to_main"),
	)
	keyboardRows = append(keyboardRows, backRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
	text := "以下是可預約的場地："
	h.bot.SendeKeyboardMessage(callback.Message.Chat.ID, text, keyboard)
}

func (h *MessageHandler) handleBooking(callback *tgbotapi.CallbackQuery) error {
	selectedCourt := callback.Data[5:]
	logger.Log.Info("使用者嘗試預約場地：" + selectedCourt)

	targetSlot := []types.CleanTimeSlot{{Button: selectedCourt}}
	if err := h.nantun_sport.BookCourt(targetSlot); err != nil {
		logger.Log.Error("預約失敗，原因：" + err.Error())
		text := fmt.Sprintf("預約失敗：%v，請重新選擇", err)
		h.bot.SendMessage(callback.Message.Chat.ID, text)
		return err
	}

	h.bot.SendMessage(callback.Message.Chat.ID, "成功預約場地，請前往以下網址完成付款：")
	h.bot.SendMessage(callback.Message.Chat.ID, h.nantun_sport.GetPaymentURL())

	return nil
}

// #endregion

// #region 預設指令
// 處理 /start 命令
func (h *MessageHandler) handleStart(message *tgbotapi.Message) {
	text := "歡迎使用運動中心查詢機器人！\n請選擇您要查詢的場地。"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("南屯運動中心", "nantun_sport"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("朝馬運動中心", "chao_ma_sport"),
		),
	)

	// 發送帶有按鈕的消息
	h.bot.SendeKeyboardMessage(message.Chat.ID, text, keyboard)
}

// 處理訊息
func (h *MessageHandler) handleDefault(message *tgbotapi.Message) {
	text := "收到您的訊息：" + message.Text
	h.bot.SendMessage(message.Chat.ID, text)
}

// 處理 /setting 命令
func (h *MessageHandler) handleSetting(message *tgbotapi.Message) {
	text := "請輸入您的運動中心帳號："
	h.bot.SendMessage(message.Chat.ID, text)
	h.settingState[message.From.ID] = "waiting_account"
}

// 處理帳號輸入
func (h *MessageHandler) handleAccountInput(message *tgbotapi.Message) {
	userObj, err := h.getOrCreateUser(message.From.ID)
	if err != nil {
		logger.Log.Error("get or create user", zap.Error(err))
		return
	}

	// 更新使用者帳號
	err = h.user.Update(context.Background(), userObj.ID, map[string]interface{}{
		"sport_center_account": message.Text,
	})
	if err != nil {
		logger.Log.Error("update user account", zap.Error(err))
		text := "設定帳號失敗，請重試"
		h.bot.SendMessage(message.Chat.ID, text)
		return
	}

	text := "請輸入您的運動中心密碼："
	h.bot.SendMessage(message.Chat.ID, text)
}

// 處理密碼輸入
func (h *MessageHandler) handlePasswordInput(message *tgbotapi.Message) {
	userObj, err := h.getOrCreateUser(message.From.ID)
	if err != nil {
		logger.Log.Error("get or create user", zap.Error(err))
		return
	}

	// 更新使用者密碼
	err = h.user.Update(context.Background(), userObj.ID, map[string]interface{}{
		"sport_center_password": message.Text,
	})
	if err != nil {
		logger.Log.Error("update user password", zap.Error(err))
		text := "設定密碼失敗，請重試"
		h.bot.SendMessage(message.Chat.ID, text)
		return
	}

	text := "帳號密碼設定完成！"
	h.bot.SendMessage(message.Chat.ID, text)
}

// #endregion

// #region 南屯場地
// 取得南屯所有可預約時間
func (h *MessageHandler) getNantunSportAllAvailableTimeSlots(message *tgbotapi.Message) {
	text := "收到您的訊息：" + message.Text
	h.bot.SendMessage(message.Chat.ID, text)
}

// #endregion

func (h *MessageHandler) getOrCreateUser(id int64) (*user.User, error) {
	userObj, err := h.user.GetByAccountID(context.Background(), fmt.Sprintf("%d", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 使用者不存在，建立新使用者
			newUser := &user.User{
				AccountID: fmt.Sprintf("%d", id),
				Status:    true,
			}
			if err := h.user.Create(context.Background(), newUser); err != nil {
				return nil, fmt.Errorf("create user: %w", err)
			}
			return newUser, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return userObj, nil
}
