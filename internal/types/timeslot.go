package types

// TimeSlot 代表一個可預約的時段
type TimeSlot struct {
	ElementType string `json:"elementType"`
	CourtName   string `json:"courtName,omitempty"`
	Price       string `json:"price,omitempty"`
	Time        string `json:"time,omitempty"`
	BookingID   string `json:"bookingId,omitempty"`
	Date        string `json:"date,omitempty"`
	Period      string `json:"period,omitempty"`
	Fee         string `json:"fee,omitempty"`
	Button      string `json:"button,omitempty"`
}

// CleanTimeSlot 將結果轉換為 JSON，移除 rawText 欄位
type CleanTimeSlot struct {
	CourtName string `json:"courtName"`
	Price     string `json:"price"`
	Time      string `json:"time"`
	BookingID string `json:"bookingId,omitempty"`
	Date      string `json:"date,omitempty"`
	Period    string `json:"period,omitempty"`
	Fee       string `json:"fee,omitempty"`
	Button    string `json:"button,omitempty"`
}

// TimeSlotCode 定義時段代碼
type TimeSlotCode int

const (
	TimeSlot_6_7 TimeSlotCode = iota + 1
	TimeSlot_7_8
	TimeSlot_8_9
	TimeSlot_9_10
	TimeSlot_10_11
	TimeSlot_11_12
	TimeSlot_12_13
	TimeSlot_13_14
	TimeSlot_14_15
	TimeSlot_15_16
	TimeSlot_16_17
	TimeSlot_17_18
	TimeSlot_18_19
	TimeSlot_19_20
	TimeSlot_20_21
	TimeSlot_21_22
)

// TimeSlotMap 時段代碼對應的時間範圍
var TimeSlotMap = map[TimeSlotCode]string{
	TimeSlot_6_7:   "6：00-7：00",
	TimeSlot_7_8:   "7：00-8：00",
	TimeSlot_8_9:   "8：00-9：00",
	TimeSlot_9_10:  "9：00-10：00",
	TimeSlot_10_11: "10：00-11：00",
	TimeSlot_11_12: "11：00-12：00",
	TimeSlot_12_13: "12：00-13：00",
	TimeSlot_13_14: "13：00-14：00",
	TimeSlot_14_15: "14：00-15：00",
	TimeSlot_15_16: "15：00-16：00",
	TimeSlot_16_17: "16：00-17：00",
	TimeSlot_17_18: "17：00-18：00",
	TimeSlot_18_19: "18：00-19：00",
	TimeSlot_19_20: "19：00-20：00",
	TimeSlot_20_21: "20：00-21：00",
	TimeSlot_21_22: "21：00-22：00",
}