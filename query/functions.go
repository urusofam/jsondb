package query

import (
	"math"
	"regexp"
	"strings"
	"time"
)

// StringFunctions предоставляет функции для работы со строками
type StringFunctions struct{}

// NewStringFunctions создает новый экземпляр StringFunctions
func NewStringFunctions() *StringFunctions {
	return &StringFunctions{}
}

// Length возвращает длину строки
func (sf *StringFunctions) Length(s string) int {
	return len(s)
}

// ToUpper преобразует строку в верхний регистр
func (sf *StringFunctions) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower преобразует строку в нижний регистр
func (sf *StringFunctions) ToLower(s string) string {
	return strings.ToLower(s)
}

// Substring возвращает подстроку
func (sf *StringFunctions) Substring(s string, start, length int) string {
	if start < 0 {
		start = 0
	}
	if start >= len(s) {
		return ""
	}
	
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	
	return s[start:end]
}

// Replace заменяет вхождения подстроки
func (sf *StringFunctions) Replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// Match проверяет, соответствует ли строка регулярному выражению
func (sf *StringFunctions) Match(s, pattern string) bool {
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match
}

// NumberFunctions предоставляет функции для работы с числами
type NumberFunctions struct{}

// NewNumberFunctions создает новый экземпляр NumberFunctions
func NewNumberFunctions() *NumberFunctions {
	return &NumberFunctions{}
}

// Abs возвращает абсолютное значение числа
func (nf *NumberFunctions) Abs(n float64) float64 {
	return math.Abs(n)
}

// Round округляет число до ближайшего целого
func (nf *NumberFunctions) Round(n float64) float64 {
	return math.Round(n)
}

// Ceil возвращает наименьшее целое число, большее или равное n
func (nf *NumberFunctions) Ceil(n float64) float64 {
	return math.Ceil(n)
}

// Floor возвращает наибольшее целое число, меньшее или равное n
func (nf *NumberFunctions) Floor(n float64) float64 {
	return math.Floor(n)
}

// Pow возвращает x^y
func (nf *NumberFunctions) Pow(x, y float64) float64 {
	return math.Pow(x, y)
}

// Sqrt возвращает квадратный корень из n
func (nf *NumberFunctions) Sqrt(n float64) float64 {
	return math.Sqrt(n)
}

// Min возвращает минимальное значение из двух чисел
func (nf *NumberFunctions) Min(a, b float64) float64 {
	return math.Min(a, b)
}

// Max возвращает максимальное значение из двух чисел
func (nf *NumberFunctions) Max(a, b float64) float64 {
	return math.Max(a, b)
}

// DateFunctions предоставляет функции для работы с датами
type DateFunctions struct{}

// NewDateFunctions создает новый экземпляр DateFunctions
func NewDateFunctions() *DateFunctions {
	return &DateFunctions{}
}

// Parse разбирает строку даты
func (df *DateFunctions) Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// Format форматирует дату
func (df *DateFunctions) Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// Now возвращает текущее время
func (df *DateFunctions) Now() time.Time {
	return time.Now()
}

// AddDays добавляет дни к дате
func (df *DateFunctions) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths добавляет месяцы к дате
func (df *DateFunctions) AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears добавляет годы к дате
func (df *DateFunctions) AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// DaysBetween возвращает количество дней между двумя датами
func (df *DateFunctions) DaysBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Hours() / 24)
}

// MonthsBetween возвращает приблизительное количество месяцев между двумя датами
func (df *DateFunctions) MonthsBetween(t1, t2 time.Time) int {
	months := 0
	
	// Убедимся, что t1 < t2
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()
	
	months = int(m2) - int(m1) + (y2-y1)*12
	
	return months
}

// YearsBetween возвращает количество лет между двумя датами
func (df *DateFunctions) YearsBetween(t1, t2 time.Time) int {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	
	years := t2.Year() - t1.Year()
	
	// Проверяем, прошла ли годовщина в этом году
	anniversary := time.Date(t2.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC)
	if anniversary.After(t2) {
		years--
	}
	
	return years
}

// FunctionRegistry содержит все доступные функции
type FunctionRegistry struct {
	StringFunctions *StringFunctions
	NumberFunctions *NumberFunctions
	DateFunctions   *DateFunctions
}

// NewFunctionRegistry создает новый реестр функций
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		StringFunctions: NewStringFunctions(),
		NumberFunctions: NewNumberFunctions(),
		DateFunctions:   NewDateFunctions(),
	}
}
