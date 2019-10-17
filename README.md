# retry [![GoDoc](https://godoc.org/github.com/AdamSLevy/retry?status.svg)](http://godoc.org/github.com/AdamSLevy/retry) [![Report card](https://goreportcard.com/badge/github.com/AdamSLevy/retry)](https://goreportcard.com/report/github.com/AdamSLevy/retry) [![Travis-CI](https://travis-ci.org/AdamSLevy/retry.svg)](https://travis-ci.org/AdamSLevy/retry) [![Coverage Status](https://coveralls.io/repos/github/AdamSLevy/retry/badge.svg?branch=master)](https://coveralls.io/github/AdamSLevy/retry?branch=master)

Package retry provides a reliable, simple way to retry operations.

`go get -u github.com/AdamSLevy/retry`

```golang
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
```

This package was inspired by
[github.com/cenkalti/backoff](https://github.com/cenkalti/backoff) but improves
on the design by providing Policy types that are composable, re-usable and safe
for repeated or concurrent calls to Run.
