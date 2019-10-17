package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdamSLevy/retry"
)

func tryWork(context.Context) error { return nil }

func ExampleRun() {
	// The provided policies can be composed to a custom Policy.
	policy := retry.LimitTotal{25 * time.Minute,
		retry.LimitAttempts{10,
			retry.Max{10 * time.Minute,
				retry.Randomize{.5,
					retry.Exponential{5 * time.Second, 2}}}}}

	// A notify function is called before each wait period.
	notify := func(err error, attempt uint, d time.Duration) {
		fmt.Printf("Attempt %v returned %v. Retrying in %v...\n",
			attempt, err, d)
	}

	// A filter function can be used to omit or wrap certain errors to tell
	// Run to stop immediately.
	filter := func(err error) error {
		if errors.Is(err, errors.New("unrecoverable err")) {
			return retry.ErrorStop(err)
		}
		return err
	}

	// A context.Context may be passed so that waits can be canceled.
	var ctx = context.TODO()

	err := retry.Run(ctx, policy, filter, notify, func() error {
		return tryWork(ctx)
	})
	if err != nil {
		return
	}
}
