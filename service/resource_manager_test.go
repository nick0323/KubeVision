package service

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterEventsByTime(t *testing.T) {
	now := time.Now()
	recent := v1.Event{LastTimestamp: metav1.NewTime(now.Add(-5 * time.Minute))}
	old := v1.Event{LastTimestamp: metav1.NewTime(now.Add(-3 * time.Hour))}
	emptyTimestamp := v1.Event{}

	tests := []struct {
		name     string
		events   []v1.Event
		since    string
		expected int
	}{
		{"empty since returns all", []v1.Event{recent, old}, "", 2},
		{"filter by 1h", []v1.Event{recent, old}, "1h", 1},
		{"all events within time", []v1.Event{recent}, "1h", 1},
		{"empty timestamp events", []v1.Event{recent, emptyTimestamp}, "1h", 1},
		{"invalid since defaults to 1h", []v1.Event{recent, old}, "invalid", 1},
		{"no events", []v1.Event{}, "1h", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEventsByTime(tt.events, tt.since)
			if len(result) != tt.expected {
				t.Errorf("filterEventsByTime() got %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestFilterEventsByTimeOrder(t *testing.T) {
	now := time.Now()
	e1 := v1.Event{LastTimestamp: metav1.Time{Time: now.Add(-30 * time.Minute)}}
	e2 := v1.Event{LastTimestamp: metav1.Time{Time: now.Add(-5 * time.Minute)}}

	result := filterEventsByTime([]v1.Event{e1, e2}, "")
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
}
