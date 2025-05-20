package crawler

import (
	"github.com/go-rod/rod"
	"github.com/tian841224/crawler_sportcenter/internal/browser"
	"github.com/tian841224/crawler_sportcenter/internal/types"
	"github.com/tian841224/crawler_sportcenter/pkg/config"
)

type NantunSportCenterBotInterface interface {
	GetAvailableTimeSlots(weedday string, time_slot int) ([]types.CleanTimeSlot, error)
	BookCourt(targetSlot []types.CleanTimeSlot) error
	GetPaymentURL() string
}

var _ NantunSportCenterBotInterface = (*NantunSportCenterBotService)(nil)

type NantunSportCenterBotService struct {
	browserService           browser.BrowserService
	nantunSportCenterService NantunSportCenterService
	Nantun_Url               string // 南屯運動中心網址
	paymentURL               string // 付款網址
	cfg                      config.Config
	page                     *rod.Page ``
}

func NewNantunSportCenterBotService(browserService browser.BrowserService, nantunSportCenterService NantunSportCenterService, cfg config.Config) NantunSportCenterBotService {
	return NantunSportCenterBotService{
		browserService:           browserService,
		nantunSportCenterService: nantunSportCenterService,
		Nantun_Url:               "https://nd01.xuanen.com.tw/BPMember/BPMemberLogin",
		paymentURL:               "https://nd01.xuanen.com.tw/BPMemberOrder/BPMemberOrder",
		cfg:                      cfg,
	}
}

func (s *NantunSportCenterBotService) GetPaymentURL() string {
	return s.paymentURL
}

func (s *NantunSportCenterBotService) GetAvailableTimeSlots(weekday string, time_slot int) ([]types.CleanTimeSlot, error) {

	timeSlotCode := types.TimeSlotCode(time_slot) // 將 int 轉換為 TimeSlotCode

	var err error
	s.page, err = s.browserService.GetPage(s.Nantun_Url)
	if err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.login(s.page, s.cfg); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.clickAgreeButton(s.page); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.selectLocationBooking(s.page); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.selectBadminton(s.page); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.setCheckboxAndProceed(s.page); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.proceedToBooking(s.page); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.selectDate(s.page, weekday); err != nil {
		return nil, err
	}

	if err = s.nantunSportCenterService.selectTimeSlot(s.page, timeSlotCode); err != nil {
		return nil, err
	}

	cleanSlots, err := s.nantunSportCenterService.getAllAvailableTimeSlots(s.page)
	if err != nil {
		return nil, err
	}

	targetSlot := s.nantunSportCenterService.findAvailableCourtsByTimeSlot(cleanSlots, timeSlotCode)

	return targetSlot, nil
}

func (s *NantunSportCenterBotService) BookCourt(targetSlot []types.CleanTimeSlot) error {
	if err := s.nantunSportCenterService.bookCourt(s.page, targetSlot); err != nil {
		return err
	}
	return nil
}
