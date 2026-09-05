package availability

import (
	"testing"
	"time"
)

func parse(t string) time.Time {
	parsed, _ := time.Parse("2006-01-02 15:04", "2026-09-09 "+t)
	return parsed
}

func ptr(t time.Time) *time.Time {
	return &t
}

func TestGenerateSlots(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		duration int
		wantLen  int
	}{
		{
			name:     "Normal: 10:00-13:00, 15 min",
			start:    parse("10:00"),
			end:      parse("13:00"),
			duration: 15,
			wantLen:  12,
		},
		{
			name:     "End time not on 15-min boundary: 10:00-13:10",
			start:    parse("10:00"),
			end:      parse("13:10"),
			duration: 15,
			wantLen:  12, // last 10 mins dropped
		},
		{
			name:     "Start == end",
			start:    parse("10:00"),
			end:      parse("10:00"),
			duration: 15,
			wantLen:  0,
		},
		{
			name:     "Exactly one slot",
			start:    parse("10:00"),
			end:      parse("10:15"),
			duration: 15,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlots(tt.start, tt.end, tt.duration)
			if len(got) != tt.wantLen {
				t.Errorf("GenerateSlots() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestRemoveBreak(t *testing.T) {
	slots := GenerateSlots(parse("10:00"), parse("13:00"), 15) // 12 slots

	tests := []struct {
		name       string
		breakStart time.Time
		breakEnd   time.Time
		wantLen    int
	}{
		{
			name:       "Break aligned: 11:30-11:45",
			breakStart: parse("11:30"),
			breakEnd:   parse("11:45"),
			wantLen:    11, // removes 1 slot (11:30)
		},
		{
			name:       "Break NOT aligned: 11:20-11:50",
			breakStart: parse("11:20"),
			breakEnd:   parse("11:50"),
			wantLen:    9, // removes 3 slots (11:15, 11:30, 11:45)
		},
		{
			name:       "Break covers entire block",
			breakStart: parse("10:00"),
			breakEnd:   parse("13:00"),
			wantLen:    0,
		},
		{
			name:       "Break outside block (before)",
			breakStart: parse("08:00"),
			breakEnd:   parse("09:00"),
			wantLen:    12,
		},
		{
			name:       "Break touching start",
			breakStart: parse("09:00"),
			breakEnd:   parse("10:00"),
			wantLen:    12,
		},
		{
			name:       "Break touching end",
			breakStart: parse("13:00"),
			breakEnd:   parse("14:00"),
			wantLen:    12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveBreak(slots, tt.breakStart, tt.breakEnd, 15)
			if len(got) != tt.wantLen {
				t.Errorf("RemoveBreak() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestMarkPast(t *testing.T) {
	slots := GenerateSlots(parse("10:00"), parse("12:00"), 15) // 8 slots

	tests := []struct {
		name        string
		now         time.Time
		leadMinutes int
		wantPast    int
	}{
		{
			name:        "Entire day past",
			now:         parse("15:00"),
			leadMinutes: 0,
			wantPast:    8,
		},
		{
			name:        "Entire day future",
			now:         parse("08:00"),
			leadMinutes: 0,
			wantPast:    0,
		},
		{
			name:        "Partial day (now is 10:30, no lead)",
			now:         parse("10:30"),
			leadMinutes: 0,
			wantPast:    2, // 10:00, 10:15
		},
		{
			name:        "Partial day (now is 10:00, 60 min lead)",
			now:         parse("10:00"),
			leadMinutes: 60,
			wantPast:    4, // cutoff is 11:00, so 10:00, 10:15, 10:30, 10:45 are past
		},
		{
			name:        "Lead exactly on slot boundary",
			now:         parse("09:01"),
			leadMinutes: 60, // cutoff is 10:01
			wantPast:    1,  // 10:00 is past, 10:15 is future
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy slots
			slotsCopy := make([]Slot, len(slots))
			copy(slotsCopy, slots)

			got := MarkPast(slotsCopy, tt.now, tt.leadMinutes)
			pastCount := 0
			for _, s := range got {
				if s.Status == StatusPast {
					pastCount++
				}
			}
			if pastCount != tt.wantPast {
				t.Errorf("MarkPast() pastCount = %v, want %v", pastCount, tt.wantPast)
			}
		})
	}
}

func TestMarkBooked(t *testing.T) {
	slots := GenerateSlots(parse("10:00"), parse("11:00"), 15) // 4 slots
	booked := []time.Time{parse("10:15"), parse("10:45")}

	got := MarkBooked(slots, booked)
	
	if got[0].Status != StatusAvailable {
		t.Errorf("Expected 10:00 to be available")
	}
	if got[1].Status != StatusBooked {
		t.Errorf("Expected 10:15 to be booked")
	}
	if got[2].Status != StatusAvailable {
		t.Errorf("Expected 10:30 to be available")
	}
	if got[3].Status != StatusBooked {
		t.Errorf("Expected 10:45 to be booked")
	}
}

func TestMergeBlocks(t *testing.T) {
	blocks := []Block{
		{
			Start:      parse("09:00"),
			End:        parse("12:00"),
			BreakStart: ptr(parse("11:30")),
			BreakEnd:   ptr(parse("11:45")),
		},
		{
			Start: parse("15:00"),
			End:   parse("17:00"),
		},
	}
	
	booked := []time.Time{parse("10:00"), parse("15:30")}
	now := parse("09:30")
	lead := 60 // cutoff is 10:30

	got := MergeBlocks(blocks, booked, now, lead, 15)

	if len(got) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(got))
	}

	b1 := got[0].Slots
	if len(b1) != 11 { // 12 slots - 1 break slot
		t.Fatalf("Expected 11 slots in block 1, got %d", len(b1))
	}

	// 09:00 is past
	if b1[0].Status != StatusPast {
		t.Errorf("Expected 09:00 to be past")
	}
	// 10:00 is past AND booked, but MarkPast overrides or MarkBooked overrides?
	// The order in MergeBlocks is MarkPast -> MarkBooked. So it becomes booked.
	// But logically, a booked slot that is in the past is both. For display, booked is fine.
	if b1[4].Time.Equal(parse("10:00")) && b1[4].Status != StatusBooked {
		t.Errorf("Expected 10:00 to be booked, got %s", b1[4].Status)
	}
	// 10:30 is available (now + 60 = 10:30. Before(10:30) is false)
	if b1[6].Time.Equal(parse("10:30")) && b1[6].Status != StatusAvailable {
		t.Errorf("Expected 10:30 to be available, got %s", b1[6].Status)
	}

	b2 := got[1].Slots
	if len(b2) != 8 {
		t.Fatalf("Expected 8 slots in block 2, got %d", len(b2))
	}
	if b2[2].Time.Equal(parse("15:30")) && b2[2].Status != StatusBooked {
		t.Errorf("Expected 15:30 to be booked")
	}
}
