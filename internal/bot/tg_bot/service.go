package tgbot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

type TGBotInterface interface {
	SendMessage(chatID int64, text string) error
	StartReceiveMessage()
	HandleMessage(handler func(update tgbotapi.Update))
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

	// debug 模式
	bot.Debug = true

	return &TGBotService{
		cfg: cfg,
		bot: bot,
	}
}

func (s *TGBotService) SendMessage(chatID int64, text string) error {
	if s.bot == nil {
		return fmt.Errorf("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Log.Error("發送訊息失敗: " + err.Error())
		return err
	}

	return nil
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
