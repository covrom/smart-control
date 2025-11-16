package main

import "time"

// CommonSMARTReport — единый формат отчёта от агентов (Windows/Linux)
type CommonSMARTReport struct {
	Hostname  string        `json:"hostname"`
	OS        string        `json:"os"`        // "windows" или "linux"
	Timestamp time.Time     `json:"timestamp"` // RFC3339
	Devices   []SMARTDevice `json:"devices"`
	RawError  string        `json:"raw_error,omitempty"`
}

// SMARTDevice — данные одного устройства
type SMARTDevice struct {
	Device    string   `json:"device"`
	Type      string   `json:"type"`
	SMARTData string   `json:"smart_data"`
	RawError  string   `json:"raw_error,omitempty"`
	MountPaths []string `json:"mount_paths,omitempty"`
}
