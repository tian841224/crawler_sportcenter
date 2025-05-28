package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type TGBotInterface interface {
	SendMessage(chatID int64, text string)
	SendeKeyboardMessage(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup)
	StartReceiveMessage()
	HandleMessage(handler func(update tgbotapi.Update))
	Request(request tgbotapi.CallbackConfig)
}

var _ TGBotInterface = (*TGBotService)(nil)

type TGBotService struct {
	cfg            config.Config
	bot            *tgbotapi.BotAPI
	messageHandler func(update tgbotapi.Update)
}

func NewTGBotService(cfg config.Config) *TGBotService {
	bot, err := tgbotapi.NewBotAPI(cfg.TG_Bot_Token)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}

	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true,
	})

	if cfg.TG_Bot_Webhook_Domain != "" {
		// 設定 webhook
		webhookURL := cfg.TG_Bot_Webhook_Domain
		webhookConfig, err := tgbotapi.NewWebhook(webhookURL)
		if err != nil {
			logger.Log.Error("設定 webhook 失敗: " + err.Error())
			return nil
		}

		_, err = bot.Request(webhookConfig)
		if err != nil {
			logger.Log.Error("設定 webhook 失敗: " + err.Error())
			return nil
		}
		logger.Log.Info("成功設定 webhook: " + webhookURL)
	} else {
		logger.Log.Info("使用輪詢模式")
	}

	// debug 模式
	// bot.Debug = true
	return &TGBotService{
		cfg: cfg,
		bot: bot,
	}
}

// StartReceiveMessage 開始接收消息
func (s *TGBotService) StartReceiveMessage() {
	if s.bot == nil {
		logger.Log.Error("bot not initialized")
		return
	}

	// 參數設定
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// 取得更新通道
	updates := s.bot.GetUpdatesChan(u)

	// 在 goroutine 中處理消息
	go func() {
		for update := range updates {
			if s.messageHandler != nil {
				s.messageHandler(update)
			}
		}
	}()
}

// HandleMessage 設置消息處理函數
func (s *TGBotService) HandleMessage(handler func(update tgbotapi.Update)) {
	s.messageHandler = handler
}

func (s *TGBotService) SendMessage(chatID int64, text string) {
	if s.bot == nil {
		logger.Log.Error("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Log.Error("發送訊息失敗: " + err.Error())
	}

}

func (s *TGBotService) SendeKeyboardMessage(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	if s.bot == nil {
		logger.Log.Error("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Log.Error("發送訊息失敗: " + err.Error())
	}
}

func (s *TGBotService) Request(request tgbotapi.CallbackConfig) {
	if _, err := s.bot.Request(request); err != nil {
		logger.Log.Error("回覆 callback 失敗：" + err.Error())
	}
}
