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

package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdamSLevy/retry"
)

func workToRetry(context.Context) error { return nil }

func ExampleRun() {
	// The provided policies can be composed to a custom Policy. The
	// following Policy implements exponential backoff. The policy
	// increases exponentially by a factor of 1.5 starting from 500
	// milliseconds with some random variation, up to a max wait time of a
	// minute. Additionally, the Policy limits retries to 15 attempts or 20
	// minutes.
	policy := retry.LimitTotal{20 * time.Minute,
		retry.LimitAttempts{15,
			retry.Max{time.Minute,
				retry.Randomize{.5,
					retry.Exponential{500 * time.Millisecond, 1.5}}}}}

	// A notify function is called before each wait period.
	notify := func(err error, attempt uint, d time.Duration) {
		fmt.Printf("Attempt %v returned %v. Retrying in %v...\n",
			attempt, err, d)
	}

	// A filter function can be used to omit or wrap certain errors to tell
	// Run to stop immediately.
	filter := func(err error) error {
		if errors.Is(err, errors.New("unrecoverable")) {
			// Run will return err.
			return retry.ErrorStop(err)
		}
		if errors.Is(err, errors.New("ignorable")) {
			// Run will return nil.
			return nil
		}
		return err
	}

	// A context.Context may be passed so that waits can be canceled.
	var ctx = context.TODO()

	err := retry.Run(ctx, policy, filter, notify, func() error {
		// If your op requires a context.Context you should create a
		// closure around it. If tryWork returns context.Canceled or
		// context.DeadlinExceeded Run will return immediately.
		return workToRetry(ctx)
	})
	if err != nil {
		return
	}
}
