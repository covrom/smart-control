package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/covrom/smart-control/internal/api"
	"github.com/covrom/smart-control/internal/cron"
	"github.com/covrom/smart-control/internal/disk"
)

func workerSendReports(ctx context.Context, wg *sync.WaitGroup, apiUrl, token, hostname string, schedule *cron.CronSchedule) {
	defer wg.Done()

	today := time.Now().Add(5 * time.Second)

	for {
		timer := time.NewTimer(time.Until(today))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			report := disk.SmartReportOnAllDevices(ctx, hostname)
			if err := api.SendReport(ctx, apiUrl, token, report); err != nil {
				slog.Error("report sending error", "err", err)
			}
			today = schedule.NextRun(today)
			slog.Info("next run", "at", today)
		}
	}
}
