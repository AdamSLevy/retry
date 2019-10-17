package retry

import "time"

func init() {
	useMockTime()
}
func useMockTime() {
	timeNow = mockNow
	timeNewTimer = mockNewTimer
}
func useActualTime() {
	timeNow = time.Now
	timeNewTimer = newTimer
}

var now = time.Unix(0, 0)

func mockNow() time.Time {
	return now
}

type mockTimer struct {
	C chan time.Time
}

func mockNewTimer(d time.Duration) timer {
	t := mockTimer{C: make(chan time.Time, 1)}
	t.Reset(d)
	return &t
}

func (t *mockTimer) Reset(d time.Duration) bool {
	// Clear the channel.
	select {
	case <-t.C:
	default:
	}
	// Advance "Now" and load the channel.
	now = now.Add(d)
	t.C <- now
	return true
}

func (t *mockTimer) Stop() bool {
	// Clear the channel.
	select {
	case <-t.C:
	default:
	}
	return true
}
func (t *mockTimer) GetC() <-chan time.Time {
	return t.C
}
