package billing

import (
	"time"
)

// CalculateMonths вычисляет количество полных месяцев в пересечении двух периодов
// Пример: (2025-06-01, 2025-08-31) и (2025-06-01, 2025-08-31) → 3 месяца (июнь, июль, август)
func CalculateMonths(subStart, subEnd, queryStart, queryEnd time.Time) int {
	// Найти пересечение периодов
	start := maxTime(subStart, queryStart)
	end := minTime(subEnd, queryEnd)

	// Если нет пересечения
	if start.After(end) {
		return 0
	}

	// Считаем количество месяцев включительно
	// Формула: (год_конца - год_начала) * 12 + (месяц_конца - месяц_начала) + 1
	months := (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1

	if months < 0 {
		return 0
	}

	return months
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
