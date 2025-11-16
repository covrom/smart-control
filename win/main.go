//go:build windows
// +build windows

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"smart-control-win/internal/cron"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"gopkg.in/ini.v1"
)

const (
	serviceName = "SMARTDataCollector"
)

//go:embed smartmontools/*
var smtFS embed.FS

func main() {
	isWS, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal(err)
	}

	if !isWS {
		elog = debug.New(serviceName)
	} else {
		elog, err = eventlog.Open(serviceName)
		if err != nil {
			log.Fatal(err)
		}
		defer elog.Close()
	}

	run := svc.Run

	if !isWS {
		run = debug.Run
	}

	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = "C:\\"
		fmt.Println("Переменная PROGRAMDATA не найдена (возможно, не Windows).")
	}
	programData = filepath.Join(programData, serviceName)
	os.MkdirAll(programData, 0755)

	// Сохранить все файлы из smtFS в папку programData
	err = copyDir(smtFS, "smartmontools", programData)
	if err != nil {
		log.Fatalf("Ошибка сохранения файлов из smtFS: %v", err)
	}

	hostname, _ := os.Hostname()

	sms := &smartService{
		Hostname: hostname,
		ApiURL:   "",
		Token:    "",
		Cron:     "",
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatal("Ошибка получения пути к исполняемому файлу:", err)
	}

	// Получаем директорию, в которой находится исполняемый файл
	exeDir := filepath.Dir(exe)

	// Загрузить настройки из settings.ini
	settingsFile := filepath.Join(exeDir, "settings.ini")
	if _, err := os.Stat(settingsFile); err == nil {
		cfg, err := ini.Load(settingsFile)
		if err != nil {
			log.Fatalf("Ошибка чтения settings.ini: %v", err)
		}

		section, err := cfg.GetSection("Config")
		if err != nil {
			log.Fatalf("Ошибка получения секции Config из settings.ini: %v", err)
		}

		sms.Hostname = section.Key("hostname").String()
		sms.ApiURL = section.Key("apiUrl").String()
		sms.Token = section.Key("token").String()
		sms.Cron = section.Key("cron").String()
	}

	sch, err := cron.NewCronScheduleFromString(sms.Cron)
	if err != nil {
		log.Fatal(fmt.Errorf("NewCronScheduleFromString error: %w", err))
	}

	sms.schedule = sch
	sms.programData = programData

	err = run(serviceName, sms)
	if err != nil {
		logErrorf("%s завершилась с ошибкой: %v", serviceName, err)
		return
	}
	elog.Info(1, fmt.Sprintf("%s остановлена.", serviceName))
}
