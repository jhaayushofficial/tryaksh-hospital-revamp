package availability

import (
	"time"
)

// GenerateSlots creates sequential slots of durationMinutes starting at start until end.
// A slot is only created if it entirely fits before the end time.
func GenerateSlots(start, end time.Time, durationMinutes int) []Slot {
	var slots []Slot
	duration := time.Duration(durationMinutes) * time.Minute

	for current := start; current.Add(duration).Before(end) || current.Add(duration).Equal(end); current = current.Add(duration) {
		slots = append(slots, Slot{
			Time:   current,
			Status: StatusAvailable,
		})
	}
	return slots
}

// RemoveBreak filters out slots that overlap with the provided break period.
func RemoveBreak(slots []Slot, breakStart, breakEnd time.Time, durationMinutes int) []Slot {
	var filtered []Slot
	duration := time.Duration(durationMinutes) * time.Minute

	for _, slot := range slots {
		slotEnd := slot.Time.Add(duration)
		
		// If slot is entirely before break OR entirely after break, keep it
		if slotEnd.Before(breakStart) || slotEnd.Equal(breakStart) || slot.Time.After(breakEnd) || slot.Time.Equal(breakEnd) {
			filtered = append(filtered, slot)
		}
	}
	return filtered
}

// MarkPast marks any slots that are before 'now + leadMinutes' as StatusPast.
func MarkPast(slots []Slot, now time.Time, leadMinutes int) []Slot {
	cutoff := now.Add(time.Duration(leadMinutes) * time.Minute)
	
	for i, slot := range slots {
		if slot.Time.Before(cutoff) {
			slots[i].Status = StatusPast
		}
	}
	return slots
}

// MarkBooked marks any slot matching a bookedTime as StatusBooked.
// Assumes slots have the same exact time matching.
func MarkBooked(slots []Slot, bookedTimes []time.Time) []Slot {
	bookedMap := make(map[time.Time]bool)
	for _, bt := range bookedTimes {
		bookedMap[bt] = true
	}

	for i, slot := range slots {
		if bookedMap[slot.Time] {
			slots[i].Status = StatusBooked
		}
	}
	return slots
}

// MergeBlocks takes database availability blocks, generates slots for them,
// applies breaks, marks past slots, and marks booked slots.
// durationMinutes is the length of a single appointment.
// leadMinutes is the minimum notice required (e.g., 60 mins before).
func MergeBlocks(blocks []Block, bookedTimes []time.Time, now time.Time, leadMinutes int, durationMinutes int) []Block {
	var result []Block

	for _, block := range blocks {
		// Generate raw slots
		slots := GenerateSlots(block.Start, block.End, durationMinutes)

		// Remove breaks if they exist
		if block.BreakStart != nil && block.BreakEnd != nil {
			slots = RemoveBreak(slots, *block.BreakStart, *block.BreakEnd, durationMinutes)
		}

		// Mark past (time-travel protection)
		slots = MarkPast(slots, now, leadMinutes)

		// Mark booked slots
		slots = MarkBooked(slots, bookedTimes)

		block.Slots = slots
		result = append(result, block)
	}

	return result
}
