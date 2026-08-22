package ai_test

import (
	"testing"

	"github.com/derispewss/gonami-projects/internal/ai"
)

func TestTokenSaverBudget(t *testing.T) {
	ts := ai.NewTokenSaver(2)
	if !ts.Allow() || !ts.Allow() {
		t.Fatal("panggilan dalam budget harus diizinkan")
	}
	if ts.Allow() {
		t.Fatal("panggilan melebihi budget harus ditolak")
	}
	if ts.Used() != 2 {
		t.Fatalf("Used() = %d, want 2", ts.Used())
	}
}

func TestTokenSaverUnlimited(t *testing.T) {
	ts := ai.NewTokenSaver(0)
	for i := 0; i < 1000; i++ {
		if !ts.Allow() {
			t.Fatalf("budget unlimited tidak boleh menolak (iterasi %d)", i)
		}
	}
}

func TestTokenSaverResetsOnNewDay(t *testing.T) {
	ts := ai.NewTokenSaver(1)
	if !ts.Allow() {
		t.Fatal("panggilan pertama harus diizinkan")
	}

	ts.SetDayForTest("2000-01-01")
	if !ts.Allow() {
		t.Fatal("counter harus reset saat ganti hari")
	}
}
