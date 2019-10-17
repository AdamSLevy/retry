// Copyright 2019 Adam S Levy
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

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
