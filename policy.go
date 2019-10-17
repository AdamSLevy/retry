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

import (
	"math"
	"math/rand"
	"time"

	"github.com/JohnCGriffin/overflow"
)

// Stop can be returned by Policy.Wait to tell Run to stop.
const Stop = time.Duration(-1)

// Policy tells Run how long to wait before the next retry.
type Policy interface {
	// Wait returns a wait time based on the number of previous attempts
	// and the total amount of time elapsed since the first attempt within
	// a Run call.
	//
	// For the first call to Wait made by a Call to Run, attempts = 1.
	//
	// Wait returns 0 to retry immediately.
	//
	// Wait returns Stop to tell Run to return without any further retries.
	//
	// In order to ensure that a Policy is re-usable across concurrent
	// calls to Run, Wait should not have any side-effects such as mutating
	// any internal state of Policy. The one exception to this is the use
	// of math/rand.Float64() the in Randomize Policy.
	Wait(attempts uint, total time.Duration) (wait time.Duration)
}

// Immediate is a Policy that always returns a zero wait time.
type Immediate struct{}

// Wait always returns c.Fixed.
func (i Immediate) Wait(uint, time.Duration) time.Duration { return 0 }

// Constant is a Policy that always returns a fixed waited time.
type Constant time.Duration

// Wait always returns c.Fixed.
func (c Constant) Wait(uint, time.Duration) time.Duration { return time.Duration(c) }

// Linear is a Policy that increases wait time linearly starting from Initial
// and adding Increment for each additional attempt.
type Linear struct {
	Initial   time.Duration
	Increment time.Duration
}

// Wait returns l.Initial + (attempts-1)*l.Increment or math.MaxInt64 if any
// integer overflow occurs.
func (l Linear) Wait(attempts uint, total time.Duration) time.Duration {
	if mx, ok := overflow.Mul64(int64(l.Increment), int64(attempts-1)); ok {
		if wait, ok := overflow.Add64(int64(l.Initial), mx); ok {
			return time.Duration(wait)
		}
	}
	return math.MaxInt64
}

// Exponential is a Policy that increases wait time exponentially starting from
// Initial and multiplying Multiplier for each additional attempt.
//
// Initial must be non-zero and Multiplier must be greater than 1 in order for
// the wait time to increase.
type Exponential struct {
	Initial    time.Duration
	Multiplier float64
}

// Wait returns e.Initial * math.Pow(e.Multiplier, attempts) up to the number
// of attempts that would cause overflow, at which point the largest value that
// does not overflow is returned.
func (e Exponential) Wait(attempts uint, total time.Duration) time.Duration {
	wait := float64(e.Initial)
	overflow := math.MaxInt64 / e.Multiplier
	for i := uint(1); i < attempts; i++ {
		if wait == 0 || wait > overflow {
			break
		}
		wait *= e.Multiplier
	}
	return time.Duration(wait)
}

// LimitAttempts wraps a Policy such that Run will return after Limit attempts.
type LimitAttempts struct {
	Limit uint
	Policy
}

// Wait returns Stop if attempts >= l.Limit, otherwise the result of
// l.Policy.Wait(attempts, total) is returned.
func (l LimitAttempts) Wait(attempts uint, total time.Duration) time.Duration {
	if attempts >= l.Limit {
		return Stop
	}
	return l.Policy.Wait(attempts, total)
}

// LimitTotal wraps a Policy such that Run will stop after total time meets or
// exceeds Limit.
type LimitTotal struct {
	Limit time.Duration
	Policy
}

// Wait returns Stop if total >= l.Limit, otherwise the result of
// l.Policy.Wait(attempts, total) is returned.
func (l LimitTotal) Wait(attempts uint, total time.Duration) time.Duration {
	if total >= l.Limit {
		return Stop
	}
	return l.Policy.Wait(attempts, total)
}

// Max wraps a Policy such that wait time is capped to Cap.
type Max struct {
	Cap time.Duration
	Policy
}

// Wait returns the minimum between m.Max and the result of
// m.Policy.Wait(attempts, total).
func (m Max) Wait(attempts uint, total time.Duration) time.Duration {
	wait := m.Policy.Wait(attempts, total)
	if wait > m.Cap {
		return m.Cap
	}
	return wait
}

// Randomize wraps a Policy such that its wait time is randomly selected from
// the range [wait * (1 - Factor), wait * (1 + Factor)].
type Randomize struct {
	Factor float64
	Policy
}

// Wait returns a wait time randomly selected from the range
//
//      [wait * (1 - r.Factor), wait * (1 + r.Factor)]
//
// such that wait will not overflow, where wait is the return value of
// r.Policy.Wait(attempts, total).
//
// If wait is 0 or Stop, it is returned directly.
func (r Randomize) Wait(attempts uint, total time.Duration) time.Duration {
	wait := r.Policy.Wait(attempts, total)
	if wait <= 0 {
		return wait
	}

	min := float64(wait) * (1 - r.Factor)
	max := float64(wait) * (1 + r.Factor)
	if max > math.MaxInt64 {
		// Prevent overflows.
		max = math.MaxInt64
	}

	// The formula below uses a +1 to account for truncation of float64
	// into int64. If the min is 1 and the max is 3 then we want a 33%
	// chance for selecting either 1, 2 or 3.
	return time.Duration(min + (rand.Float64() * (max - min + 1)))
}
