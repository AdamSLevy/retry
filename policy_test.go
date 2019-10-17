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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolicy(t *testing.T) {
	for _, test := range policyTests {
		test := test
		t.Run(test.Name, func(t *testing.T) { testPolicy(t, test) })
	}
	t.Run("Randomize", func(t *testing.T) {
		policy := Randomize{.5, Constant{time.Minute}}
		for i := 0; i < 1000; i++ {
			wait := policy.Wait(0, 0)
			assert.InDelta(t, time.Minute, wait, .5*float64(time.Minute))
		}
	})
	t.Run("Randomize/overflow", func(t *testing.T) {
		policy := Randomize{.5, Constant{math.MaxInt64}}
		wait := policy.Wait(0, 0)
		assert.InDelta(t, math.MaxInt64, wait, .5*float64(math.MaxInt64))
	})
	t.Run("Randomize/stop", func(t *testing.T) {
		policy := Randomize{.5, Constant{Stop}}
		wait := policy.Wait(0, 0)
		assert.Equal(t, Stop, wait)
	})
}
func testPolicy(t *testing.T, test policyTest) {
	assert := assert.New(t)
	for i, arg := range test.Args {
		wait := test.Policy.Wait(arg.Attempts, arg.Total)
		assert.Equalf(test.Wait[i], wait,
			"arg index %v, return wait: %v expected: %v",
			i, wait, test.Wait[i])
	}
}

type policyArgs struct {
	Attempts uint
	Total    time.Duration
}

type policyTest struct {
	Name   string
	Policy Policy
	Args   []policyArgs
	Wait   []time.Duration
}

var policyTests = []policyTest{{
	Name:   "Linear",
	Policy: Linear{time.Minute, 30 * time.Second},
	Args:   []policyArgs{{1, 0}, {2, 30 * time.Second}, {3, time.Hour}},
	Wait:   []time.Duration{time.Minute, 90 * time.Second, 2 * time.Minute},
}, {
	Name:   "Linear/overflow",
	Policy: Linear{math.MaxInt64, 1},
	Args:   []policyArgs{{2, 0}},
	Wait:   []time.Duration{math.MaxInt64},
}, {
	Name:   "Exponential",
	Policy: Exponential{time.Minute, 2},
	Args:   []policyArgs{{1, 0}, {2, 30 * time.Second}, {3, time.Hour}},
	Wait:   []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute},
}, {
	Name:   "Exponential/overflow",
	Policy: Exponential{time.Minute, math.MaxInt64},
	Args:   []policyArgs{{1, 0}, {2, 0}},
	Wait:   []time.Duration{time.Minute, time.Minute},
}, {
	Name:   "LimitTotal",
	Policy: LimitTotal{3 * time.Minute, Constant{time.Minute}},
	Args:   []policyArgs{{1, 0}, {2, 30 * time.Second}, {3, time.Hour}},
	Wait:   []time.Duration{time.Minute, time.Minute, Stop},
}, {
	Name:   "LimitAttempts",
	Policy: LimitAttempts{2, Constant{time.Minute}},
	Args:   []policyArgs{{1, 0}, {2, 30 * time.Second}, {3, time.Hour}},
	Wait:   []time.Duration{time.Minute, Stop, Stop},
}, {
	Name:   "Max",
	Policy: Max{90 * time.Second, Linear{time.Minute, time.Minute}},
	Args:   []policyArgs{{1, 0}, {2, 30 * time.Second}},
	Wait:   []time.Duration{time.Minute, 90 * time.Second},
}}
