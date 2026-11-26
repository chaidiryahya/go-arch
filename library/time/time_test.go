package time_test

import (
	"testing"
	stdtime "time"

	libtime "go-arch/library/time"
)

func TestTimeImplementsSeam(t *testing.T) {
	var _ interface{ Now() stdtime.Time } = libtime.Time{}
	var _ interface{ Now() stdtime.Time } = &libtime.MockTime{}
}

func TestTimeNowIsCloseToStdlib(t *testing.T) {
	before := stdtime.Now()
	got := libtime.Time{}.Now()
	after := stdtime.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("Time.Now returned %v, outside [%v, %v]", got, before, after)
	}
}

func TestMockTimeStartsAtGivenInstant(t *testing.T) {
	pin := stdtime.Date(2026, stdtime.November, 26, 12, 0, 0, 0, stdtime.UTC)
	m := libtime.NewMockTime(pin)

	if got := m.Now(); !got.Equal(pin) {
		t.Fatalf("Now = %v, want %v", got, pin)
	}
}

func TestMockTimeDoesNotAdvanceOnItsOwn(t *testing.T) {
	pin := stdtime.Date(2026, stdtime.November, 26, 12, 0, 0, 0, stdtime.UTC)
	m := libtime.NewMockTime(pin)

	first := m.Now()
	stdtime.Sleep(2 * stdtime.Millisecond)
	second := m.Now()

	if !second.Equal(first) {
		t.Fatalf("mock clock advanced by itself: %v then %v", first, second)
	}
}

func TestMockTimeSet(t *testing.T) {
	m := libtime.NewMockTime(stdtime.Date(2026, 1, 1, 0, 0, 0, 0, stdtime.UTC))
	target := stdtime.Date(2027, 6, 15, 9, 30, 0, 0, stdtime.UTC)

	m.Set(target)
	if got := m.Now(); !got.Equal(target) {
		t.Fatalf("after Set, Now = %v, want %v", got, target)
	}
}

func TestMockTimeAdvance(t *testing.T) {
	pin := stdtime.Date(2026, stdtime.November, 26, 12, 0, 0, 0, stdtime.UTC)
	m := libtime.NewMockTime(pin)

	m.Advance(90 * stdtime.Second)
	want := pin.Add(90 * stdtime.Second)
	if got := m.Now(); !got.Equal(want) {
		t.Fatalf("after Advance(90s), Now = %v, want %v", got, want)
	}

	m.Advance(-30 * stdtime.Second)
	want = want.Add(-30 * stdtime.Second)
	if got := m.Now(); !got.Equal(want) {
		t.Fatalf("after Advance(-30s), Now = %v, want %v", got, want)
	}
}
