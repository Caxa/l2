package main

import (
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	c := NewCalendar()
	e := Event{UserID: 1, Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), Text: "New Year Party"}
	if err := c.Create(e); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	evts, _ := c.EventsForDay("2023-12-31")
	if len(evts) != 1 {
		t.Errorf("expected 1 event, got %d", len(evts))
	}
	if evts[0].Text != "New Year Party" {
		t.Errorf("expected 'New Year Party', got %s", evts[0].Text)
	}
}

func TestCreateEventInvalid(t *testing.T) {
	c := NewCalendar()
	err := c.Create(Event{UserID: 0, Date: time.Now(), Text: ""})
	if err == nil {
		t.Errorf("expected error for invalid event, got nil")
	}
}

func TestUpdateEvent(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	c.Create(Event{UserID: 1, Date: date, Text: "Old Text"})

	err := c.Update(Event{UserID: 1, Date: date, Text: "Updated Text"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	evts, _ := c.EventsForDay("2023-12-31")
	if evts[0].Text != "Updated Text" {
		t.Errorf("expected 'Updated Text', got %s", evts[0].Text)
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	c := NewCalendar()
	err := c.Update(Event{UserID: 99, Date: time.Now(), Text: "No Event"})
	if err == nil {
		t.Errorf("expected error for updating non-existing event, got nil")
	}
}

func TestDeleteEvent(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	c.Create(Event{UserID: 1, Date: date, Text: "Some Event"})

	err := c.Delete(Event{UserID: 1, Date: date})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	evts, _ := c.EventsForDay("2023-12-31")
	if len(evts) != 0 {
		t.Errorf("expected 0 events, got %d", len(evts))
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	c := NewCalendar()
	err := c.Delete(Event{UserID: 99, Date: time.Now()})
	if err == nil {
		t.Errorf("expected error for deleting non-existing event, got nil")
	}
}

func TestEventsForWeek(t *testing.T) {
	c := NewCalendar()
	start := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	c.Create(Event{UserID: 1, Date: start, Text: "Day 1"})
	c.Create(Event{UserID: 2, Date: start.AddDate(0, 0, 3), Text: "Day 4"})

	evts := c.EventsForWeek(start)
	if len(evts) != 2 {
		t.Errorf("expected 2 events, got %d", len(evts))
	}
}

func TestEventsForMonth(t *testing.T) {
	c := NewCalendar()
	c.Create(Event{UserID: 1, Date: time.Date(2023, 12, 5, 0, 0, 0, 0, time.UTC), Text: "Mid Month"})
	c.Create(Event{UserID: 2, Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), Text: "End Month"})

	evts := c.EventsForMonth(2023, time.December)
	if len(evts) != 2 {
		t.Errorf("expected 2 events, got %d", len(evts))
	}
}

func TestEventsForDayEmpty(t *testing.T) {
	c := NewCalendar()
	evts, err := c.EventsForDay("2023-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evts) != 0 {
		t.Errorf("expected 0 events, got %d", len(evts))
	}
}
