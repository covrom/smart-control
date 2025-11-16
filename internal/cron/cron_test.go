package cron

import (
	"reflect"
	"testing"
	"time"
)

func TestCronSchedule_NextRun(t *testing.T) {
	// Фиксируем время для тестов
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	type fields struct {
		minutes  []int
		hours    []int
		days     []int
		months   []int
		weekDays []int
	}
	type args struct {
		now time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
		{
			name: "test every minute",
			fields: fields{
				minutes:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59},
				hours:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
				days:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				months:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				weekDays: []int{0, 1, 2, 3, 4, 5, 6},
			},
			args: args{
				now: now,
			},
			want: time.Date(2025, 1, 1, 12, 1, 0, 0, time.UTC),
		},
		{
			name: "test specific minute",
			fields: fields{
				minutes:  []int{30},
				hours:    []int{12},
				days:     []int{1},
				months:   []int{1},
				weekDays: []int{3},
			},
			args: args{
				now: now,
			},
			want: time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC),
		},
		{
			name: "test specific hour and minute",
			fields: fields{
				minutes:  []int{15},
				hours:    []int{14},
				days:     []int{1},
				months:   []int{1},
				weekDays: []int{4},
			},
			args: args{
				now: now,
			},
			want: time.Date(2025, 1, 1, 14, 15, 0, 0, time.UTC),
		},
		{
			name: "test next day",
			fields: fields{
				minutes:  []int{0},
				hours:    []int{0},
				days:     []int{1,2},
				months:   []int{1},
				weekDays: []int{3,4},
			},
			args: args{
				now: now,
			},
			want: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "test next week",
			fields: fields{
				minutes:  []int{0},
				hours:    []int{0},
				days:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				months:   []int{1},
				weekDays: []int{0}, // Sunday
			},
			args: args{
				now: now,
			},
			want: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CronSchedule{
				minutes:  tt.fields.minutes,
				hours:    tt.fields.hours,
				days:     tt.fields.days,
				months:   tt.fields.months,
				weekDays: tt.fields.weekDays,
			}
			if got := c.NextRun(tt.args.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CronSchedule.NextRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCronSchedule_NextRun_Week(b *testing.B) {
	// Фиксируем время для тестов
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	c, err := NewCronSchedule("55", "23", "*", "*", "0")
	if err != nil {
		b.Fatalf("Failed to create CronSchedule: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.NextRun(now)
	}
}

func BenchmarkCronSchedule_NextRun_Year(b *testing.B) {
	// Фиксируем время для тестов
	now := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	c, err := NewCronSchedule("55", "23", "1", "1", "*")
	if err != nil {
		b.Fatalf("Failed to create CronSchedule: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.NextRun(now)
	}
}