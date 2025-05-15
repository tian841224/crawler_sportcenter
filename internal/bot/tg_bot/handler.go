package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type MessageHandler struct {
	bot TGBotInterface
}

func NewMessageHandler(bot TGBotInterface) *MessageHandler {
	return &MessageHandler{
		bot: bot,
	}
}

// HandleUpdate 處理所有的更新消息
func (h *MessageHandler) HandleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		h.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		h.handleCallback(update.CallbackQuery)
	}
}

// 處理文字消息
func (h *MessageHandler) handleMessage(message *tgbotapi.Message) {
	// 根據消息內容處理
	switch message.Text {
	case "/start":
		h.handleStart(message)
	default:
		h.handleDefault(message)
	}
}

// 處理按鈕回調
func (h *MessageHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	// 處理按鈕回調的邏輯
	logger.Log.Info("收到按鈕回調：" + callback.Data)

	// 這裡可以根據 callback.Data 來處理不同的按鈕操作
}

// 處理 /start 命令
func (h *MessageHandler) handleStart(message *tgbotapi.Message) {
	text := "歡迎使用運動中心查詢機器人！\n請選擇您要查詢的場地。"
	err := h.bot.SendMessage(message.Chat.ID, text)
	if err != nil {
		logger.Log.Error("發送歡迎訊息失敗：" + err.Error())
	}
}

// 處理消息
func (h *MessageHandler) handleDefault(message *tgbotapi.Message) {
	text := "收到您的訊息：" + message.Text
	err := h.bot.SendMessage(message.Chat.ID, text)
	if err != nil {
		logger.Log.Error("發送回覆訊息失敗：" + err.Error())
	}
}
