package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("не удалось открыть файл для логов: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	logger = log.New(multiWriter, "", log.LstdFlags)
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}
