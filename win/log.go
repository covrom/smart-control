//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/svc/debug"
)

var elog debug.Log

func logEvent(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if elog != nil {
		elog.Info(1, msg)
	}
	log.Println(msg)
}

func logErrorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if elog != nil {
		elog.Error(1, msg)
	}
	log.Println(msg)
}
