package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/tian841224/crawler_sportcenter/internal/crawler"
	"github.com/tian841224/crawler_sportcenter/internal/domain/schedule"
	"github.com/tian841224/crawler_sportcenter/pkg/logger"
	"go.uber.org/zap"
)

type SchedulerService struct {
	nantunSportCenter crawler.NantunSportCenterBotInterface
	schedule          schedule.Service
	mutex             sync.RWMutex
	stopChan          chan struct{}
}

type SchedulerInterface interface {
	Start(ctx context.Context)
	Stop()
}

var _ SchedulerInterface = (*SchedulerService)(nil)

func NewSchedulerService(nantunSportCenter crawler.NantunSportCenterBotInterface, schedule schedule.Service) *SchedulerService {
	return &SchedulerService{
		nantunSportCenter: nantunSportCenter,
		schedule:          schedule,
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

	for chatID, subs := range *scheduleList {
		// TODO:
		// 排序時段
		// 檢查是否有可用場地
		// 如果有可用場地，通知使用者

		logger.Log.Debug("checkAllSubscriptions", zap.Int("chatID", chatID), zap.Any("subs", subs))
	}

	return nil
}

// TODO: 通知使用者 - 使用介面而非直接呼叫 bot
