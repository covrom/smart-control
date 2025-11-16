package cron

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CronSchedule представляет расписание в формате cron
type CronSchedule struct {
	minutes,
	hours,
	days,
	months,
	weekDays []int
}

// NewCronSchedule создаёт новое cron-расписание из строковых полей.
// Поля: minute, hour, day, month, weekDay (в порядке cron: "min hour dom mon dow")
func NewCronSchedule(minute, hour, day, month, weekDay string) (*CronSchedule, error) {
	minutes, err := parseCronField(minute, 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field %q: %w", minute, err)
	}
	hours, err := parseCronField(hour, 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field %q: %w", hour, err)
	}
	days, err := parseCronField(day, 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid day field %q: %w", day, err)
	}
	months, err := parseCronField(month, 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid month field %q: %w", month, err)
	}
	weekDays, err := parseCronField(weekDay, 0, 6) // 0 = Sunday
	if err != nil {
		return nil, fmt.Errorf("invalid weekday field %q: %w", weekDay, err)
	}

	return &CronSchedule{
		minutes:  minutes,
		hours:    hours,
		days:     days,
		months:   months,
		weekDays: weekDays,
	}, nil
}

// NewCronScheduleFromString создаёт новое cron-расписание из строки в формате cron.
// Принимает строку вида "min hour dom mon dow".
func NewCronScheduleFromString(cronString string) (*CronSchedule, error) {
	parts := strings.Split(strings.TrimSpace(cronString), " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid cron string: %q, expected 5 parts, got %d", cronString, len(parts))
	}

	return NewCronSchedule(parts[0], parts[1], parts[2], parts[3], parts[4])
}

// NextRun возвращает следующее время запуска на основе текущего времени.
// Использует иерархический поиск: месяц → день → час → минута.
// Поддерживает поведение стандартного cron: day и weekday связаны логическим ИЛИ,
// с эвристикой для случая, когда одно из полей — "*".
func (c *CronSchedule) NextRun(now time.Time) time.Time {
	// Начинаем с now + 1 минута
	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute() + 1

	for {
		// Обработка переполнения минут
		if minute > 59 {
			minute = 0
			hour++
		}
		// Обработка переполнения часов
		if hour > 23 {
			hour = 0
			day++
		}

		// Определяем количество дней в текущем месяце
		daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

		// Если день вышел за пределы месяца — переход к следующему месяцу
		if day > daysInMonth {
			day = 1
			month++
			if month > 12 {
				month = 1
				year++
			}
			hour = 0
			minute = 0
			continue
		}

		// Проверка месяца
		if !contains(c.months, int(month)) {
			// Месяц не подходит — переходим к 1-му числу следующего месяца
			month++
			if month > 12 {
				month = 1
				year++
			}
			day = 1
			hour = 0
			minute = 0
			continue
		}

		// Проверка дня с учётом cron-эвристики (ИЛИ + обработка "*")
		tCandidate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		weekDay := int(tCandidate.Weekday())
		if !dayMatches(day, weekDay, c.days, c.weekDays) {
			// День не подходит — к следующему дню
			day++
			hour = 0
			minute = 0
			continue
		}

		// Месяц и день подходят → проверяем час
		if !contains(c.hours, hour) {
			// Ищем ближайший подходящий час >= текущего
			hourIdx := sort.SearchInts(c.hours, hour)
			if hourIdx < len(c.hours) {
				hour = c.hours[hourIdx]
				minute = 0 // сброс минут при переходе к новому часу
			} else {
				// Часов сегодня больше нет — к следующему дню
				day++
				hour = 0
				minute = 0
				continue
			}
		}

		// Час подходит → проверяем минуты
		if !contains(c.minutes, minute) {
			minuteIdx := sort.SearchInts(c.minutes, minute)
			if minuteIdx < len(c.minutes) {
				minute = c.minutes[minuteIdx]
			} else {
				// Минут в этом часу больше нет — к следующему часу
				hourIdx := sort.SearchInts(c.hours, hour+1)
				if hourIdx < len(c.hours) {
					hour = c.hours[hourIdx]
					minute = c.minutes[0]
				} else {
					// Часов сегодня больше нет — к следующему дню
					day++
					hour = 0
					minute = 0
					continue
				}
			}
		}

		// Все поля совпадают — возвращаем результат
		return time.Date(year, month, day, hour, minute, 0, 0, now.Location())
	}
}

// dayMatches проверяет, удовлетворяет ли день условиям cron с учётом
// стандартного поведения: day и weekday связаны ИЛИ,
// но если одно поле — "*", а другое — ограничено, используется только ограничено.
func dayMatches(day int, weekDay int, days, weekDays []int) bool {
	// Эвристика: определяем, были ли поля эквивалентны "*"
	daysAll := len(days) == 31 && days[0] == 1 && days[30] == 31
	weekDaysAll := len(weekDays) == 7 && weekDays[0] == 0 && weekDays[6] == 6

	if daysAll && !weekDaysAll {
		// Пример: "* * * * 0" → использовать только день недели
		return contains(weekDays, weekDay)
	}
	if weekDaysAll && !daysAll {
		// Пример: "* * 1 * *" → использовать только день месяца
		return contains(days, day)
	}
	// Иначе — стандартное ИЛИ (включая случай, когда оба "*")
	return contains(days, day) || contains(weekDays, weekDay)
}

// parseCronField парсит одно поле cron-выражения (например, "0,30", "1-5", "*/2")
// min и max задают допустимый диапазон значений.
func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		result := make([]int, 0, max-min+1)
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
		return result, nil
	}

	if field == "" {
		return []int{}, nil
	}

	result := make([]int, 0, max-min+1)
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") && !strings.Contains(part, "/") {
			// Простой диапазон: "1-5"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid range values: %s", part)
			}
			if start < min || end > max || start > end {
				return nil, fmt.Errorf("range out of bounds: %s", part)
			}
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else if strings.Contains(part, "/") {
			// Шаг: "*/2", "1-10/3"
			stepParts := strings.Split(part, "/")
			if len(stepParts) != 2 {
				return nil, fmt.Errorf("invalid step format: %s", part)
			}

			base := stepParts[0]
			stepStr := stepParts[1]
			step, err := strconv.Atoi(stepStr)
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step value: %s", stepStr)
			}

			var allValues []int
			if base == "*" {
				allValues = make([]int, 0, max-min+1)
				for i := min; i <= max; i++ {
					allValues = append(allValues, i)
				}
			} else if strings.Contains(base, "-") {
				rangeParts := strings.Split(base, "-")
				if len(rangeParts) != 2 {
					return nil, fmt.Errorf("invalid range in step: %s", base)
				}
				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 != nil || err2 != nil {
					return nil, fmt.Errorf("invalid range in step: %s", base)
				}
				if start < min || end > max || start > end {
					return nil, fmt.Errorf("range out of bounds in step: %s", base)
				}
				allValues = make([]int, 0, end-start+1)
				for i := start; i <= end; i++ {
					allValues = append(allValues, i)
				}
			} else {
				// Одиночное число в шаге: "5/2" → [5]
				val, err := strconv.Atoi(base)
				if err != nil {
					return nil, fmt.Errorf("invalid step base: %s", base)
				}
				if val < min || val > max {
					return nil, fmt.Errorf("step base out of range: %s", base)
				}
				allValues = []int{val}
			}

			// Берём каждый step-ый элемент
			for i := 0; i < len(allValues); i += step {
				result = append(result, allValues[i])
			}
		} else {
			// Одиночное значение
			value, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", part)
			}
			if value < min || value > max {
				return nil, fmt.Errorf("value out of range: %s", part)
			}
			result = append(result, value)
		}
	}

	// Убираем дубликаты и сортируем
	unique := make(map[int]bool)
	uniqueResult := make([]int, 0, len(result))
	for _, v := range result {
		if !unique[v] {
			unique[v] = true
			uniqueResult = append(uniqueResult, v)
		}
	}
	sort.Ints(uniqueResult)
	return uniqueResult, nil
}

// contains проверяет наличие значения в отсортированном срезе.
// Использует бинарный поиск для эффективности.
func contains(sorted []int, value int) bool {
	i := sort.SearchInts(sorted, value)
	return i < len(sorted) && sorted[i] == value
}
