package scheduler

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
)

// 定義通知介面，讓 scheduler 不依賴具體的 bot 實現

type SchedulerService struct {
	nantunSportCenter crawler.NantunSportCenterBotInterface
	subscriptions     map[int64][]Subscription
	mutex             sync.RWMutex
	stopChan          chan struct{}
}

type Subscription struct {
	Weekday  string
	TimeSlot int
	Tag      string
}

type SchedulerInterface interface {
	AddSubscription(chatID int64, weekday string, timeSlot int)
	RemoveSubscription(chatID int64, weekday string, timeSlot int)
	GetUserSubscriptions(chatID int64) []Subscription
	Start()
	Stop()
	CheckAllSubscriptions()
	NotifyUser(chatID int64, sub Subscription, slots []types.CleanTimeSlot) string
}

var _ SchedulerInterface = (*SchedulerService)(nil)

func NewSchedulerService(nantunSportCenter crawler.NantunSportCenterBotInterface) *SchedulerService {
	return &SchedulerService{
		nantunSportCenter: nantunSportCenter,
		subscriptions:     make(map[int64][]Subscription),
		stopChan:          make(chan struct{}),
	}
}

// 增加訂閱時段
func (s *SchedulerService) AddSubscription(chatID int64, weekday string, timeSlot int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tag := fmt.Sprintf("%d", chatID)
	sub := Subscription{
		Weekday:  weekday,
		TimeSlot: timeSlot,
		Tag:      tag,
	}

	// 檢查是否已經存在相同的訂閱
	for _, existingSub := range s.subscriptions[chatID] {
		if existingSub.Weekday == weekday && existingSub.TimeSlot == timeSlot {
			return // 已經存在相同的訂閱
		}
	}

	s.subscriptions[chatID] = append(s.subscriptions[chatID], sub)
	logger.Log.Info(fmt.Sprintf("使用者 %d 訂閱了 %s 的時段 %d", chatID, weekday, timeSlot))
}

// 移除訂閱
func (s *SchedulerService) RemoveSubscription(chatID int64, weekday string, timeSlot int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subs := s.subscriptions[chatID]
	for i, sub := range subs {
		if sub.Weekday == weekday && sub.TimeSlot == timeSlot {
			// 移除這個訂閱
			s.subscriptions[chatID] = append(subs[:i], subs[i+1:]...)
			logger.Log.Info(fmt.Sprintf("用戶 %d 取消訂閱了 %s 的時段 %d", chatID, weekday, timeSlot))
			break
		}
	}
}

// 取得使用者的所有訂閱
func (s *SchedulerService) GetUserSubscriptions(chatID int64) []Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.subscriptions[chatID]
}

// 啟動定時搜尋
func (s *SchedulerService) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 每1分鐘檢查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAllSubscriptions()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// 停止定時搜尋
func (s *SchedulerService) Stop() {
	close(s.stopChan)
}

// 檢查所有訂閱
func (s *SchedulerService) checkAllSubscriptions() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for chatID, subs := range s.subscriptions {
		for _, sub := range subs {
			go s.checkSubscription(chatID, sub)
		}
	}
}

// 檢查單個訂閱
func (s *SchedulerService) checkSubscription(chatID int64, sub Subscription) {
	availableSlots, err := s.nantunSportCenter.GetAvailableTimeSlots(sub.Weekday, sub.TimeSlot, sub.Tag)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("檢查訂閱失敗: %v", err))
		return
	}

	if len(availableSlots) > 0 {
		// 有可用場地，通知用戶
		s.notifyUser(chatID, sub, availableSlots)
	}
}

// 通知使用者 - 使用介面而非直接呼叫 bot
func (s *SchedulerService) notifyUser(chatID int64, sub Subscription, slots []types.CleanTimeSlot) string {
	timeSlotText := types.TimeSlotMap[types.TimeSlotCode(sub.TimeSlot)]
	message := fmt.Sprintf("🎾 場地通知 🎾\n\n星期%s的%s有可用場地了！\n\n可用場地：", sub.Weekday, timeSlotText)

	for _, slot := range slots {
		message += fmt.Sprintf("\n- %s", slot.CourtName)
	}

	message += "\n\n請立即前往預約！"
	
	// 建立預約按鈕
	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	for _, slot := range slots {
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
	s.notifier.SendKeyboardMessage(chatID, "選擇要預約的場地：", keyboard)
	
	return message
}

var _ SchedulerInterface = (*SchedulerService)(nil)

// 同時為了更好的模組化，可以將原有的方法名稱調整為公開方法
func (s *SchedulerService) CheckAllSubscriptions() {
	s.checkAllSubscriptions()
}
 
func (s *SchedulerService) NotifyUser(chatID int64, sub Subscription, slots []types.CleanTimeSlot) string {
	return s.notifyUser(chatID, sub, slots)
}
