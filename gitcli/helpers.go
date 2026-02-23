// helpers.go — date resolution and parsing
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// resolveDate converts user input to YYYY-MM-DD.
func resolveDate(date, yearsAgo, daysAgo string) (string, error) {
	now := time.Now()

	if yearsAgo != "" {
		n, err := strconv.Atoi(yearsAgo)
		if err != nil || n < 0 {
			return "", fmt.Errorf("invalid --years-ago: %s", yearsAgo)
		}
		t := now.AddDate(-n, 0, 0)
		return t.Format("2006-01-02"), nil
	}

	if daysAgo != "" {
		n, err := strconv.Atoi(daysAgo)
		if err != nil || n < 0 {
			return "", fmt.Errorf("invalid --days-ago: %s", daysAgo)
		}
		t := now.AddDate(0, 0, -n)
		return t.Format("2006-01-02"), nil
	}

	if date == "" {
		// default: today
		return now.Format("2006-01-02"), nil
	}

	// YYYYMMDD → YYYY-MM-DD
	if len(date) == 8 && !strings.Contains(date, "-") {
		return date[:4] + "-" + date[4:6] + "-" + date[6:8], nil
	}

	// YYYYMMDDT... (Denote ID) → YYYY-MM-DD
	if len(date) >= 8 && strings.Contains(date, "T") {
		d := date[:strings.Index(date, "T")]
		if len(d) == 8 {
			return d[:4] + "-" + d[4:6] + "-" + d[6:8], nil
		}
	}

	// Already YYYY-MM-DD
	if len(date) == 10 && date[4] == '-' && date[7] == '-' {
		return date, nil
	}

	return "", fmt.Errorf("unrecognized date format: %s (use YYYY-MM-DD or YYYYMMDD)", date)
}

// parseDate parses YYYY-MM-DD to time.Time.
func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

// dayOfWeek returns English day name for a date string.
func dayOfWeek(dateStr string) string {
	t, err := parseDate(dateStr)
	if err != nil {
		return ""
	}
	return t.Weekday().String()
}
