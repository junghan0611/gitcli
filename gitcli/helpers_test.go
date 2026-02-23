package main

import "testing"

func TestResolveDate(t *testing.T) {
	tests := []struct {
		date, yearsAgo, daysAgo string
		wantErr                 bool
		wantLen                 int // 10 = YYYY-MM-DD
	}{
		{"2025-10-10", "", "", false, 10},
		{"20251010", "", "", false, 10},
		{"20251010T143000", "", "", false, 10},
		{"", "", "", false, 10},    // today
		{"", "1", "", false, 10},   // 1 year ago
		{"", "", "7", false, 10},   // 7 days ago
		{"bad", "", "", true, 0},
		{"", "-1", "", true, 0},
	}

	for _, tt := range tests {
		got, err := resolveDate(tt.date, tt.yearsAgo, tt.daysAgo)
		if (err != nil) != tt.wantErr {
			t.Errorf("resolveDate(%q,%q,%q) error=%v, wantErr=%v", tt.date, tt.yearsAgo, tt.daysAgo, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && len(got) != tt.wantLen {
			t.Errorf("resolveDate(%q,%q,%q) = %q, want len %d", tt.date, tt.yearsAgo, tt.daysAgo, got, tt.wantLen)
		}
	}
}

func TestResolveDateValues(t *testing.T) {
	got, err := resolveDate("2025-10-10", "", "")
	if err != nil || got != "2025-10-10" {
		t.Errorf("got %q, err %v", got, err)
	}

	got, err = resolveDate("20251010", "", "")
	if err != nil || got != "2025-10-10" {
		t.Errorf("got %q, err %v", got, err)
	}

	got, err = resolveDate("20251010T143022", "", "")
	if err != nil || got != "2025-10-10" {
		t.Errorf("got %q, err %v", got, err)
	}
}

func TestParseHHMM(t *testing.T) {
	if parseHHMM("02:30") != 150 {
		t.Error("02:30 should be 150 minutes")
	}
	if parseHHMM("14:00") != 840 {
		t.Error("14:00 should be 840 minutes")
	}
}

func TestCalcActiveHours(t *testing.T) {
	h := calcActiveHours("02:00", "14:00")
	if h != 12.0 {
		t.Errorf("expected 12.0, got %f", h)
	}
}
