package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/covrom/smart-control/internal/llmdesc"
	"github.com/covrom/smart-control/internal/smartdata"
)

func workerRecvReports(ctx context.Context, wg *sync.WaitGroup, hostname string, chTgMsg chan<- string, chReps <-chan smartdata.CommonSMARTReport, llmDescriber *llmdesc.LLMSmartDescriber) {
	defer wg.Done()

	slog.Info("workerRecvReports started")

	for {
		select {
		case <-ctx.Done():
			return
		case report := <-chReps:
			slog.Info("receive report", "hostname", report.Hostname)
			if report.RawError != "" {
				chTgMsg <- fmt.Sprintf("❌ Ошибка для %s (%s)\n%s",
					report.Hostname, report.OS, report.RawError)
			} else {
				for _, d := range report.Devices {
					if d.RawError != "" {
						chTgMsg <- fmt.Sprintf("❌ Ошибка для %s (%s)\nУстройство: %s\n%s",
							report.Hostname, report.OS, d.Device, d.RawError)
					} else {
						var prev smartdata.SMARTDevice
						// сохраняем анализ (переменную d) в файл json с именем, соответствующим report.Hostname и d.Device (с заменой небезопасных символов на знак '_')
						// предварительно загружаем из этого файла предыдущую версию (если файл существует) в переменную prev
						filename := fmt.Sprintf("%s_%s.json", report.Hostname, d.Device)
						// заменяем небезопасные символы на '_'
						for _, r := range filename {
							if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
								filename = strings.ReplaceAll(filename, string(r), "_")
							}
						}

						filename = filepath.Join("/var/lib/smart_reports_data/", filename)

						// загружаем предыдущую версию, если файл существует
						if data, err := os.ReadFile(filename); err == nil {
							if err := json.Unmarshal(data, &prev); err != nil {
								slog.Error("failed to unmarshal prev data", "err", err)
							}
						}
						// сохраняем текущую версию в файл
						data, err := json.Marshal(d)
						if err != nil {
							slog.Error("failed to marshal current data", "err", err)
						} else {
							if err := os.WriteFile(filename, data, 0644); err != nil {
								slog.Error("failed to write file", "err", err)
							}
						}

						msg := fmt.Sprintf("💻 Анализ для %s (%s)\n📀 Устройство: %s, точки монтирования:\n%s\n\n%s",
							report.Hostname, report.OS, d.Device, strings.Join(d.MountPaths, "\n"), llmDescriber.Describe(ctx, hostname, d, prev))
						if len(d.MountPaths) == 0 {
							msg = fmt.Sprintf("💻 Анализ для %s (%s)\n📀 Устройство: %s, точки монтирования отсутствуют\n\n%s",
								report.Hostname, report.OS, d.Device, llmDescriber.Describe(ctx, hostname, d, prev))
						}
						chTgMsg <- msg
					}
				}
			}
		}
	}
}
