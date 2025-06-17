package scheduler

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	tgbot "github.com/tian841224/crawler_sportcenter/internal/bot/tg_bot"
	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/domain/schedule"
	"github.com/tian841224/crawler_sportcenter/internal/domain/user"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
)

type SchedulerService struct {
	nantunSportCenter crawler.NantunSportCenterBotInterface
	schedule          schedule.Service
	user              user.Service
	tgBot             tgbot.TGBotInterface
	mutex             sync.RWMutex
	stopChan          chan struct{}
}

type SchedulerInterface interface {
	Start(ctx context.Context)
	Stop()
}

var _ SchedulerInterface = (*SchedulerService)(nil)

func NewSchedulerService(nantunSportCenter crawler.NantunSportCenterBotInterface, schedule schedule.Service, user user.Service, tgBot tgbot.TGBotInterface) *SchedulerService {
	return &SchedulerService{
		nantunSportCenter: nantunSportCenter,
		tgBot:             tgBot,
		schedule:          schedule,
		user:              user,
		stopChan:          make(chan struct{}),
	}
}

// 啟動定時搜尋
func (s *SchedulerService) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 每1分鐘檢查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAllSubscriptions(ctx)
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
func (s *SchedulerService) checkAllSubscriptions(ctx context.Context) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	scheduleList, err := s.schedule.GetAll(ctx)
	if err != nil {
		return err
	}

	sort.Slice(*scheduleList, func(i, j int) bool {
		// 先比星期
		if (*scheduleList)[i].Weekday != (*scheduleList)[j].Weekday {
			return (*scheduleList)[i].Weekday < (*scheduleList)[j].Weekday
		}
		// 星期一樣時，比開始時間
		if (*scheduleList)[i].TimeSlot == nil || (*scheduleList)[j].TimeSlot == nil {
			return false // 或依你需求決定
		}
		return (*scheduleList)[i].TimeSlot.StartTime.Before((*scheduleList)[j].TimeSlot.StartTime)
	})

	currentWeekday := time.Now().Weekday()
	currentTime := time.Now().Hour()
	availableTimeSlotsLength := 0

	for _, subs := range *scheduleList {
		// 檢查 TimeSlot 是否為空值
		if subs.TimeSlot == nil {
			logger.Log.Warn("TimeSlot is nil", zap.Uint("scheduleID", subs.ID))
			continue
		}

		// 查詢的時間一樣直接通知使用者
		if subs.Weekday != currentWeekday || subs.TimeSlot.StartTime.Hour() != currentTime {
			// 檢查是否有可用場地
			var err error
			dayMap := map[string]string{
				"0": "日", "1": "一", "2": "二",
				"3": "三", "4": "四", "5": "五", "6": "六",
			}

			weekday := dayMap[strconv.Itoa(int(subs.Weekday))]
			availableTimeSlots, err := s.nantunSportCenter.GetAvailableTimeSlotsForSchedule(weekday, int(*subs.TimeSlotID), strconv.Itoa(int(subs.UserID)))
			if err != nil {
				logger.Log.Error("checkAllSubscriptions", zap.Error(err))
				continue
			}
			availableTimeSlotsLength = len(availableTimeSlots)
		}

		// 沒有可用場地直接跳過
		if availableTimeSlotsLength == 0 {
			continue
		}

		// 如果有可用場地，通知使用者
		message := fmt.Sprintf("星期 %s 時段 %d:00-%d:00 有可用場地",
			subs.Weekday,
			subs.TimeSlot.StartTime.Hour(),
			subs.TimeSlot.EndTime.Hour())

		user, err := s.user.GetByID(ctx, subs.UserID)
		if err != nil {
			logger.Log.Error("checkAllSubscriptions", zap.Error(err))
			continue
		}
		// TODO: 修改成TG帳號
		accountID, err := strconv.ParseInt(user.AccountID, 10, 64)
		if err != nil {
			logger.Log.Error("invalid AccountID", zap.String("AccountID", user.AccountID), zap.Error(err))
			continue
		}
		s.tgBot.SendMessage(accountID, message)

		currentWeekday = subs.Weekday
		currentTime = subs.TimeSlot.StartTime.Hour()
		logger.Log.Debug("checkAllSubscriptions", zap.Uint("scheduleID", subs.ID), zap.Any("subs", subs))
	}

	return nil
}
