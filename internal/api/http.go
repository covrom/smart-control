package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/covrom/smart-control/internal/smartdata"
)

func NewHttpServer(token string, ch chan<- smartdata.CommonSMARTReport) *http.Server {
	srv := &http.Server{
		Addr:         ":8000",
		Handler:      nil,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	http.Handle("POST /smart/report", handleSmartReport(token, ch))
	go srv.ListenAndServe()
	slog.Info("http server started")
	return srv
}

func handleSmartReport(token string, chReps chan<- smartdata.CommonSMARTReport) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedAuthHeader := fmt.Sprintf("Bearer %s", token)

		// 1. Проверка Authorization
		auth := r.Header.Get("Authorization")
		if auth != expectedAuthHeader {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			slog.Error("Unauthorized request", "from", r.RemoteAddr)
			return
		}

		// 2. Чтение тела
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			slog.Error("Body request error", "err", err)
			return
		}
		defer r.Body.Close()

		// 3. Парсинг JSON в общую структуру
		var report smartdata.CommonSMARTReport
		if err := json.Unmarshal(body, &report); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			log.Printf("Неверный JSON от %s: %v", r.RemoteAddr, err)
			return
		}

		// 4. Валидация
		if report.Hostname == "" || report.OS == "" || report.Timestamp.IsZero() {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			slog.Error("Missing required fields", "from", r.RemoteAddr)
			return
		}

		// 5. Сохранение
		chReps <- report

		// 6. Ответ
		slog.Info("Received info", "from", r.RemoteAddr, "host", report.Hostname, "devices", len(report.Devices))
	})
}
