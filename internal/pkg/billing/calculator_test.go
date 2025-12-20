package billing

import (
	"testing"
	"time"
)

func TestCalculateMonths(t *testing.T) {
	tests := []struct {
		name       string
		subStart   string // MM-YYYY
		subEnd     string // MM-YYYY
		queryStart string // MM-YYYY
		queryEnd   string // MM-YYYY
		want       int
	}{
		{
			name:       "exact match - 3 months",
			subStart:   "06-2025",
			subEnd:     "08-2025",
			queryStart: "06-2025",
			queryEnd:   "08-2025",
			want:       3, // June, July, August
		},
		{
			name:       "subscription contains query period",
			subStart:   "01-2025",
			subEnd:     "12-2025",
			queryStart: "06-2025",
			queryEnd:   "08-2025",
			want:       3, // June, July, August
		},
		{
			name:       "query contains subscription period",
			subStart:   "06-2025",
			subEnd:     "08-2025",
			queryStart: "01-2025",
			queryEnd:   "12-2025",
			want:       3, // June, July, August
		},
		{
			name:       "partial overlap - start",
			subStart:   "05-2025",
			subEnd:     "07-2025",
			queryStart: "06-2025",
			queryEnd:   "09-2025",
			want:       2, // June, July
		},
		{
			name:       "partial overlap - end",
			subStart:   "07-2025",
			subEnd:     "10-2025",
			queryStart: "05-2025",
			queryEnd:   "08-2025",
			want:       2, // July, August
		},
		{
			name:       "no overlap - subscription before query",
			subStart:   "01-2025",
			subEnd:     "03-2025",
			queryStart: "06-2025",
			queryEnd:   "08-2025",
			want:       0,
		},
		{
			name:       "no overlap - subscription after query",
			subStart:   "10-2025",
			subEnd:     "12-2025",
			queryStart: "06-2025",
			queryEnd:   "08-2025",
			want:       0,
		},
		{
			name:       "single month overlap",
			subStart:   "06-2025",
			subEnd:     "06-2025",
			queryStart: "06-2025",
			queryEnd:   "06-2025",
			want:       1, // June only
		},
		{
			name:       "one year period",
			subStart:   "01-2025",
			subEnd:     "12-2025",
			queryStart: "01-2025",
			queryEnd:   "12-2025",
			want:       12,
		},
		{
			name:       "cross-year overlap",
			subStart:   "11-2024",
			subEnd:     "02-2025",
			queryStart: "01-2025",
			queryEnd:   "03-2025",
			want:       2, // January, February 2025
		},
		{
			name:       "adjacent periods - no overlap",
			subStart:   "01-2025",
			subEnd:     "03-2025",
			queryStart: "04-2025",
			queryEnd:   "06-2025",
			want:       0,
		},
		{
			name:       "touching at boundary",
			subStart:   "01-2025",
			subEnd:     "03-2025",
			queryStart: "03-2025",
			queryEnd:   "06-2025",
			want:       1, // March
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subStart, _ := time.Parse("01-2006", tt.subStart)
			subEnd, _ := time.Parse("01-2006", tt.subEnd)
			queryStart, _ := time.Parse("01-2006", tt.queryStart)
			queryEnd, _ := time.Parse("01-2006", tt.queryEnd)

			got := CalculateMonths(subStart, subEnd, queryStart, queryEnd)
			if got != tt.want {
				t.Errorf("CalculateMonths(%s, %s, %s, %s) = %d, want %d",
					tt.subStart, tt.subEnd, tt.queryStart, tt.queryEnd, got, tt.want)
			}
		})
	}
}

func TestCalculateMonthsEdgeCases(t *testing.T) {
	// Test with time components (should only consider month/year)
	subStart := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	subEnd := time.Date(2025, 8, 20, 18, 45, 0, 0, time.UTC)
	queryStart := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	queryEnd := time.Date(2025, 8, 31, 23, 59, 59, 0, time.UTC)

	got := CalculateMonths(subStart, subEnd, queryStart, queryEnd)
	want := 3 // June, July, August

	if got != want {
		t.Errorf("CalculateMonths with time components = %d, want %d", got, want)
	}
}

func TestCalculateMonthsSymmetry(t *testing.T) {
	// Swapping subscription and query periods should give same result when they match
	subStart := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	subEnd := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	queryStart := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	queryEnd := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)

	result1 := CalculateMonths(subStart, subEnd, queryStart, queryEnd)
	result2 := CalculateMonths(queryStart, queryEnd, subStart, subEnd)

	if result1 != result2 {
		t.Errorf("CalculateMonths not symmetric: %d vs %d", result1, result2)
	}
}
