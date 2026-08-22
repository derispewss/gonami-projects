package ai

import (
	"errors"
	"sync"
	"time"
)

var ErrBudgetExceeded = errors.New("kuota harian llm habis")

var wibLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}()

type TokenSaver struct {
	mu        sync.Mutex
	day       string
	used      int
	maxPerDay int
}

func NewTokenSaver(maxPerDay int) *TokenSaver {
	return &TokenSaver{maxPerDay: maxPerDay}
}

func (ts *TokenSaver) Allow() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	today := time.Now().In(wibLoc).Format("2006-01-02")
	if today != ts.day {
		ts.day, ts.used = today, 0
	}
	if ts.maxPerDay > 0 && ts.used >= ts.maxPerDay {
		return false
	}
	ts.used++
	return true
}

func (ts *TokenSaver) Used() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	today := time.Now().In(wibLoc).Format("2006-01-02")
	if today != ts.day {
		return 0
	}
	return ts.used
}

func (ts *TokenSaver) SetDayForTest(day string) {
	ts.mu.Lock()
	ts.day = day
	ts.mu.Unlock()
}
