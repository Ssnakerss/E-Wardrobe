package auth

import "time"

// nullableString returns a pointer to s if non-nil, nil otherwise.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullableTime returns a pointer to t if non-nil, nil otherwise.
func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
