package main

import (
	"errors"
	"sync"
	"time"
)

type Event struct {
	UserID int       `json:"user_id"`
	Date   time.Time `json:"date"`
	Text   string    `json:"event"`
}

type Calendar struct {
	mu     sync.Mutex
	events map[string][]Event
}

func NewCalendar() *Calendar {
	return &Calendar{events: make(map[string][]Event)}
}

func (c *Calendar) Create(e Event) error {
	if e.UserID <= 0 || e.Text == "" {
		return errors.New("неверные данные события")
	}
	dateKey := e.Date.Format("2006-01-02")
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events[dateKey] = append(c.events[dateKey], e)
	return nil
}

func (c *Calendar) Update(e Event) error {
	dateKey := e.Date.Format("2006-01-02")
	c.mu.Lock()
	defer c.mu.Unlock()
	if evts, ok := c.events[dateKey]; ok {
		for i, existing := range evts {
			if existing.UserID == e.UserID {
				c.events[dateKey][i] = e
				return nil
			}
		}
	}
	return errors.New("событие не найдено")
}

func (c *Calendar) Delete(e Event) error {
	dateKey := e.Date.Format("2006-01-02")
	c.mu.Lock()
	defer c.mu.Unlock()
	if evts, ok := c.events[dateKey]; ok {
		for i, existing := range evts {
			if existing.UserID == e.UserID {
				c.events[dateKey] = append(evts[:i], evts[i+1:]...)
				return nil
			}
		}
	}
	return errors.New("событие не найдено")
}

func (c *Calendar) EventsForDay(date string) ([]Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if evts, ok := c.events[date]; ok {
		return evts, nil
	}
	return []Event{}, nil
}

func (c *Calendar) EventsForWeek(startDate time.Time) []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []Event
	for i := 0; i < 7; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		if evts, ok := c.events[d]; ok {
			result = append(result, evts...)
		}
	}
	return result
}

func (c *Calendar) EventsForMonth(year int, month time.Month) []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []Event
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	for day := 1; day <= daysInMonth; day++ {
		d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		if evts, ok := c.events[d]; ok {
			result = append(result, evts...)
		}
	}
	return result
}
