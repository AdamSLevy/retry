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

// Package retry provides a reliable, simple way to retry operations.
//
// This package was inspired by github.com/cenkalti/backoff but improves on the
// design by providing Policy types that are composable, re-usable and safe for
// repeated or concurrent calls to Run.
package retry

import (
	"context"
	"errors"
	"time"
)

// Allow for tests to mock the time package.
var (
	timeNow      = time.Now
	timeSince    = time.Since
	timeNewTimer = newTimer
)

func newTimer(d time.Duration) timer {
	return (*timeTimer)(time.NewTimer(d))
}

// timer allows time.Timer to be mocked in tests.
type timer interface {
	Reset(time.Duration) bool
	Stop() bool
	GetC() <-chan time.Time
}

type timeTimer time.Timer

func (t *timeTimer) Reset(d time.Duration) bool {
	return (*time.Timer)(t).Reset(d)
}
func (t *timeTimer) Stop() bool {
	return (*time.Timer)(t).Stop()
}
func (t *timeTimer) GetC() <-chan time.Time {
	return t.C
}

// Run op until one of the following occurs,
//
//      - op returns nil.
//      - op returns context.Canceled or context.DeadlineExceeded.
//      - op returns an error wrapped by ErrorStop.
//      - p.Wait returns Stop.
//      - ctx.Done() is closed.
//
// If the above conditions are not met, then op is retried after waiting
// p.Wait. The total number of attempts and the total time elapsed since Run
// was envoked are passed to p.Wait. See Policy for more details.
//
// If filter is not nil, all calls to op are wrapped by filter:
//
//      op = func() error { return filter(op()) }
//
// Use filter to cause Run to return immediately on certain op errors by either
// returning nil to censor the error, or by wrapping the error with ErrorStop
// to pass the error up the call stack.
//
// Run always returns the latest filtered op return value. If the error was
// wrapped by ErrorStop, it is unwrapped, and the original error is returned.
//
// If notify is not nil, it is called with the latest return values of op and
// p.Wait prior to waiting.
//
// If ctx is nil, context.Background() is used.
//
// If ctx.Done() is closed while waiting, Run returns immediately.
func Run(ctx context.Context,
	p Policy, filter func(error) error,
	notify func(error, uint, time.Duration),
	op func() error) error {

	if ctx == nil {
		ctx = context.Background()
	}

	filterOp := op
	if filter != nil {
		filterOp = func() error { return filter(op()) }
	}

	tmr := timeNewTimer(0)
	defer tmr.Stop()

	start := timeNow()
	var attempt uint
	for {
		err := filterOp()
		if err == nil {
			return nil
		}
		attempt++

		// There is no point in retrying after a context error.
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		// Do not retry after an ErrorStop.
		if err, ok := err.(errorStop); ok {
			// Return the original error.
			return err.err
		}

		// Determine the next wait time.
		wait := p.Wait(attempt, timeSince(start))
		if wait <= Stop {
			return err
		}

		if notify != nil {
			notify(err, attempt, wait)
		}

		if wait == 0 {
			// Skip over the tmr.
			continue
		}

		// Start the tmr.
		tmr.Reset(wait)

		select {
		case <-ctx.Done():
			// Return the op error.
			return err
		case <-tmr.GetC():
		}
	}
}
