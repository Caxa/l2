package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func makeHandlers(c *Calendar) {
	http.HandleFunc("/create_event", logRequest(func(w http.ResponseWriter, r *http.Request) {
		var e Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSONError(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		if err := c.Create(e); err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"result": "Событие создано"})
	}))

	http.HandleFunc("/update_event", logRequest(func(w http.ResponseWriter, r *http.Request) {
		var e Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSONError(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		if err := c.Update(e); err != nil {
			writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"result": "Событие обновлено"})
	}))

	http.HandleFunc("/delete_event", logRequest(func(w http.ResponseWriter, r *http.Request) {
		var e Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSONError(w, "Некорректные данные", http.StatusBadRequest)
			return
		}
		if err := c.Delete(e); err != nil {
			writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"result": "Событие удалено"})
	}))

	http.HandleFunc("/events_for_day", logRequest(func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			writeJSONError(w, "Не указана дата", http.StatusBadRequest)
			return
		}
		evts, _ := c.EventsForDay(date)
		writeJSON(w, http.StatusOK, evts)
	}))

	http.HandleFunc("/events_for_week", logRequest(func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		start, err := time.Parse("2006-01-02", date)
		if err != nil {
			writeJSONError(w, "Некорректная дата", http.StatusBadRequest)
			return
		}
		evts := c.EventsForWeek(start)
		writeJSON(w, http.StatusOK, evts)
	}))

	http.HandleFunc("/events_for_month", logRequest(func(w http.ResponseWriter, r *http.Request) {
		monthYear := r.URL.Query().Get("month_year")
		parts := []rune(monthYear)
		if len(parts) != 7 || parts[4] != '-' {
			writeJSONError(w, "Некорректный формат, используйте YYYY-MM", http.StatusBadRequest)
			return
		}
		year, _ := strconv.Atoi(string(parts[:4]))
		month, _ := strconv.Atoi(string(parts[5:]))
		evts := c.EventsForMonth(year, time.Month(month))
		writeJSON(w, http.StatusOK, evts)
	}))
}
