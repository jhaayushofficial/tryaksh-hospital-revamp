package availability

import (
	"testing"
	"time"

	"github.com/tryaksh/clinic/backend/internal/db"
)

func parseTime(t string) time.Time {
	parsed, _ := time.Parse("15:04", t)
	return parsed
}

func TestCheckOverlap(t *testing.T) {
	existing := []db.Availability{
		{StartTime: parseTime("09:00"), EndTime: parseTime("12:00")},
		{StartTime: parseTime("15:00"), EndTime: parseTime("17:00")},
	}

	tests := []struct {
		name    string
		new     db.Availability
		wantErr bool
	}{
		{
			name:    "completely before",
			new:     db.Availability{StartTime: parseTime("07:00"), EndTime: parseTime("08:00")},
			wantErr: false,
		},
		{
			name:    "completely after",
			new:     db.Availability{StartTime: parseTime("18:00"), EndTime: parseTime("19:00")},
			wantErr: false,
		},
		{
			name:    "between blocks",
			new:     db.Availability{StartTime: parseTime("13:00"), EndTime: parseTime("14:00")},
			wantErr: false,
		},
		{
			name:    "exact overlap",
			new:     db.Availability{StartTime: parseTime("09:00"), EndTime: parseTime("12:00")},
			wantErr: true,
		},
		{
			name:    "partial overlap start",
			new:     db.Availability{StartTime: parseTime("08:30"), EndTime: parseTime("09:30")},
			wantErr: true,
		},
		{
			name:    "partial overlap end",
			new:     db.Availability{StartTime: parseTime("11:30"), EndTime: parseTime("12:30")},
			wantErr: true,
		},
		{
			name:    "contained within",
			new:     db.Availability{StartTime: parseTime("10:00"), EndTime: parseTime("11:00")},
			wantErr: true,
		},
		{
			name:    "contains existing block",
			new:     db.Availability{StartTime: parseTime("08:00"), EndTime: parseTime("13:00")},
			wantErr: true,
		},
		{
			name:    "touching start boundary",
			new:     db.Availability{StartTime: parseTime("08:00"), EndTime: parseTime("09:00")},
			wantErr: false,
		},
		{
			name:    "touching end boundary",
			new:     db.Availability{StartTime: parseTime("12:00"), EndTime: parseTime("13:00")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckOverlap(existing, tt.new)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckOverlap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
