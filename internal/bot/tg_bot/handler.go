package tgbot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type MessageHandler struct {
	bot                   TGBotInterface
	nantun_sport          crawler.NantunSportCenterBotInterface
	userSelectionDate     string
	userSelectionTimeSlot string
	userList              map[int64]struct{}
}

func NewMessageHandler(bot TGBotInterface, nantun_sport crawler.NantunSportCenterBotInterface) *MessageHandler {
	return &MessageHandler{
		bot:          bot,
		nantun_sport: nantun_sport,
		userList:     make(map[int64]struct{}),
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
		// 儲存使用者ID
		if _, exists := h.userList[id]; !exists {
			h.userList[id] = struct{}{}
		}
		h.handleCallback(update.CallbackQuery)
	}
}

// #region 處理所有格式訊息
// 處理文字訊息
func (h *MessageHandler) handleMessage(message *tgbotapi.Message) {
	switch message.Text {
	case "/start":
		h.handleStart(message)
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
)

// 使用常量
func (h *MessageHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	switch {
	case callback.Data == callbackNantunSport:
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleSportCenterSelection(callback)
	case callback.Data == callbackBackToMain:
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleBackToMain(callback)
	case strings.HasPrefix(callback.Data, prefixDate):
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleDateSelection(callback)
	case strings.HasPrefix(callback.Data, prefixTimeSlot):
		h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		h.handleTimeSlotSelection(callback)
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
	logger.Log.Info("用戶嘗試預約場地：" + selectedCourt)

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

// #endregion

// #region 南屯場地
// 取得南屯所有可預約時間
func (h *MessageHandler) getNantunSportAllAvailableTimeSlots(message *tgbotapi.Message) {
	text := "收到您的訊息：" + message.Text
	h.bot.SendMessage(message.Chat.ID, text)
}

// #endregion
