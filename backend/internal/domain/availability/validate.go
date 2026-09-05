package availability

import (
	"fmt"

	"github.com/tryaksh/clinic/backend/internal/db"
)

// CheckOverlap ensures that the newBlock does not overlap with any of the existing blocks.
// Both start and end times are expected to be on the same date (usually represented as just time of day).
func CheckOverlap(existing []db.Availability, newBlock db.Availability) error {
	newStart := newBlock.StartTime
	newEnd := newBlock.EndTime

	for _, e := range existing {
		// If editing an existing block, ignore itself
		if newBlock.ID != (db.Availability{}.ID) && newBlock.ID == e.ID {
			continue
		}

		if newStart.Before(e.EndTime) && newEnd.After(e.StartTime) {
			return fmt.Errorf("overlaps with existing block %s–%s", e.StartTime.Format("15:04"), e.EndTime.Format("15:04"))
		}
	}
	return nil
}
