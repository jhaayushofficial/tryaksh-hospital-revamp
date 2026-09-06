package availability

import "time"

type SlotStatus string

const (
	StatusAvailable SlotStatus = "available"
	StatusBooked    SlotStatus = "booked"
	StatusPast      SlotStatus = "past"
)

type Slot struct {
	Time   time.Time  `json:"time"`
	Status SlotStatus `json:"status"`
}

type Block struct {
	Start      time.Time  `json:"start"`
	End        time.Time  `json:"end"`
	BreakStart *time.Time `json:"break_start,omitempty"`
	BreakEnd   *time.Time `json:"break_end,omitempty"`
	Slots      []Slot     `json:"slots"`
}
