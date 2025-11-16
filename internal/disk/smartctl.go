package disk

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/covrom/smart-control/internal/smartdata"
)

// getSmartDevices вызывает smartctl --scan-open и возвращает список устройств
func getSmartDevices() ([]smartdata.SMARTDevice, error) {
	cmd := exec.Command("smartctl", "--scan-open")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl --scan-open failed: %w, output: %s", err, string(out))
	}

	var devices []smartdata.SMARTDevice
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
			dev := smartdata.SMARTDevice{
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

func runSmartctlCommands(device smartdata.SMARTDevice) (string, error) {
	// Команда для получения информации об устройстве
	args := []string{"-d", device.Type, "-a", device.Device}
	if device.Type == "" {
		args = []string{"-a", device.Device}
	}
	cmdInfo := exec.Command("smartctl", args...)
	outputInfo, err := cmdInfo.CombinedOutput()
	if err != nil {
		return string(outputInfo), err
	}
	return string(outputInfo), nil
}

func getMountPaths(deviceName string) ([]string, error) {
	// Используем cat /proc/mounts вместо lsblk
	cmd := exec.Command("cat", "/host/proc/1/mounts")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cat /host/proc/1/mounts failed: %w", err)
	}

	// Парсим вывод /proc/mounts
	var mountPaths []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		device := fields[0]
		mountpoint := fields[1]

		// Проверяем, что устройство соответствует устройству, для которого мы ищем монтирования
		if strings.HasPrefix(device, deviceName) {
			mountPaths = append(mountPaths, mountpoint)
		}
	}

	slog.Info("mount", "mountPaths", mountPaths, "deviceName", deviceName)
	return mountPaths, nil
}

func SmartReportOnAllDevices(ctx context.Context, hostname string) smartdata.CommonSMARTReport {
	report := smartdata.CommonSMARTReport{
		Hostname:  hostname,
		OS:        "linux",
		Timestamp: time.Now(),
	}

	devices, err := getSmartDevices()
	if err != nil {
		report.RawError = fmt.Sprintf("Ошибка получения списка дисков: %s", err.Error())
		slog.Error("getSmartDevices error", "err", err)
		return report
	}

	for _, device := range devices {
		if device.Device != "" {
			result, err := runSmartctlCommands(device)
			if err != nil {
				slog.Error("smartctl error", "device", device, "err", err)
				device.RawError = fmt.Sprintf("Ошибка анализа %s(%s):\n%s\n%s", device.Device, device.Type, err.Error(), result)
			} else {
				slog.Info("smartctl analysis done", "device", device)
				device.SMARTData = result
			}

			// Получаем список примонтированных разделов для этого устройства
			mounts, err := getMountPaths(device.Device)
			if err != nil {
				slog.Error("getMountPaths error", "device", device.Device, "err", err)
			} else {
				device.MountPaths = mounts
			}

			report.Devices = append(report.Devices, device)
		}
	}

	return report
}
