package util

import (
	"time"
)

// TimerUtil provides utility functions for timing operations
type TimerUtil struct{}

// NewTimerUtil creates a new TimerUtil
func NewTimerUtil() *TimerUtil {
	return &TimerUtil{}
}

// Timer is a timer that can be used to measure elapsed time
type Timer struct {
	start time.Time
}

// NewTimer creates a new Timer
func (t *TimerUtil) NewTimer() *Timer {
	return &Timer{
		start: time.Now(),
	}
}

// Elapsed returns the elapsed time since the timer was created
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// ElapsedMilliseconds returns the elapsed time in milliseconds
func (t *Timer) ElapsedMilliseconds() int64 {
	return t.Elapsed().Milliseconds()
}

// ElapsedSeconds returns the elapsed time in seconds
func (t *Timer) ElapsedSeconds() float64 {
	return t.Elapsed().Seconds()
}

// Reset resets the timer
func (t *Timer) Reset() {
	t.start = time.Now()
}

// Sleep sleeps for the specified duration
func (t *TimerUtil) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// SleepMilliseconds sleeps for the specified number of milliseconds
func (t *TimerUtil) SleepMilliseconds(milliseconds int64) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// SleepSeconds sleeps for the specified number of seconds
func (t *TimerUtil) SleepSeconds(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

// After waits for the duration to elapse and then sends the current time on the returned channel
func (t *TimerUtil) After(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine
func (t *TimerUtil) AfterFunc(duration time.Duration, f func()) *time.Timer {
	return time.AfterFunc(duration, f)
}

// NewTicker returns a new Ticker containing a channel that will send the time with a period specified by the duration argument
func (t *TimerUtil) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}

// CreateTimer creates a new Timer that will send the current time on its channel after at least duration d
func (t *TimerUtil) CreateTimer(duration time.Duration) *time.Timer {
	return time.NewTimer(duration)
}
