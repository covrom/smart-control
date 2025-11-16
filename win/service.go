//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"smart-control-win/internal/cron"
	"time"

	"golang.org/x/sys/windows/svc"
)

type smartService struct {
	Hostname    string `json:"hostname"`
	ApiURL      string `json:"apiUrl"`
	Token       string `json:"token"`
	Cron        string `json:"cron"`
	schedule    *cron.CronSchedule
	programData string
}

func (m *smartService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	today := time.Now().Add(5 * time.Second)

loop:
	for {
		timer := time.NewTimer(time.Until(today))
		select {
		case <-timer.C:
			m.collectAndSaveSMARTData()
			today = m.schedule.NextRun(today)
			logEvent("next run at %s", today)
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				timer.Stop()
				break loop
			default:
				elog.Warning(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}

	changes <- svc.Status{State: svc.Stopped, Accepts: cmdsAccepted}
	return
}

func (m *smartService) collectAndSaveSMARTData() {
	report := smartReportOnAllDevices(m.Hostname, m.programData)
	if err := sendReport(context.Background(), m.ApiURL, m.Token, report); err != nil {
		logErrorf("collectAndSaveSMARTData завершилась с ошибкой: %v", err)
		return
	}
	logEvent("отчет успешно отправлен")
}
