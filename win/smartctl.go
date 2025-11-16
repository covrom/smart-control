//go:build windows
// +build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// getSmartDevices вызывает smartctl --scan-open и возвращает список устройств
func getSmartDevices(programData string) ([]SMARTDevice, error) {
	cmd := exec.Command(filepath.Join(programData, "smartctl.exe"), "--scan-open")
	cmd.Dir = programData
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl --scan-open failed: %w, output: %s", err, string(out))
	}

	var devices []SMARTDevice
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			devname := parts[0]

			// Убираем возможные кавычки (иногда smartctl их добавляет)
			devname = strings.Trim(devname, `"'`)
			dev := SMARTDevice{
				Device: devname,
			}
			if len(parts) > 2 && parts[1] == "-d" {
				dev.Type = parts[2]
			}

			devices = append(devices, dev)
		}
	}
	return devices, nil
}

func runSmartctlCommands(programData string, device SMARTDevice) (string, error) {
	// Команда для получения информации об устройстве
	args := []string{"-a", device.Device, "-d", device.Type, "-T", "permissive"}
	if device.Type == "" {
		args = []string{"-a", device.Device, "-T", "permissive"}
	}
	cmdInfo := exec.Command(filepath.Join(programData, "smartctl.exe"), args...)
	cmdInfo.Dir = programData
	outputInfo, err := cmdInfo.CombinedOutput()
	if err != nil {
		return string(outputInfo), err
	}
	return string(outputInfo), nil
}

func smartReportOnAllDevices(hostname, programData string) CommonSMARTReport {
	report := CommonSMARTReport{
		Hostname:  hostname,
		OS:        "windows",
		Timestamp: time.Now(),
	}

	devices, err := getSmartDevices(programData)
	if err != nil {
		report.RawError = fmt.Sprintf("Ошибка получения списка дисков: %s", err.Error())
		logErrorf("getSmartDevices error: %v", err)
		return report
	}

	for _, device := range devices {
		if device.Device != "" {
			result, err := runSmartctlCommands(programData, device)
			if err != nil {
				logErrorf("smartctl error for device %q: %v\n%s", device, err, result)
				device.RawError = fmt.Sprintf("Ошибка анализа %s(%s):\n%s\n%s", device.Device, device.Type, err.Error(), result)
			} else {
				logEvent("smartctl analysis done for device %q", device)
				device.SMARTData = result
			}

			report.Devices = append(report.Devices, device)
		}
	}

	return report
}
