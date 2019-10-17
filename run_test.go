package retry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	for _, test := range runTests {
		test := test
		t.Run(test.Name, func(t *testing.T) { testRun(t, test) })
	}
	t.Run("ctx canceled", func(t *testing.T) {
		useActualTime()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		testRun(t, runTest{
			Name:        "ctx canceled",
			Ctx:         ctx,
			Policy:      Constant{time.Second},
			Op:          func() error { return fmt.Errorf("failed") },
			Err:         "failed",
			NotifyCount: 2,
		})
	})
}
func testRun(t *testing.T, test runTest) {
	assert := assert.New(t)

	if test.Policy == nil {
		test.Policy = Constant{100}
	}

	var opCount uint = 1
	var notified bool
	notify := func(_ error, _ uint, d time.Duration) {
		opCount++
		notified = true
		//fmt.Println(test.Name, "notify", opCount, d)
	}

	err := Run(test.Ctx, test.Policy, test.Filter, notify, test.Op)

	if len(test.Err) == 0 {
		assert.NoError(err)
	} else {
		assert.EqualError(err, test.Err)
	}

	assert.True(notified, "notified")
	assert.Equal(test.NotifyCount, opCount)
}

func testOp(attempts uint, final error) func() error {
	var opCount uint
	return func() error {
		opCount++
		if opCount >= attempts {
			return final
		}
		return fmt.Errorf("failed")
	}
}

type runTest struct {
	Name   string
	Ctx    context.Context
	Policy Policy
	Filter func(error) error
	Op     func() error

	NotifyCount uint
	Err         string
}

var runTests = []runTest{
	{
		Name:        "op()==nil",
		Op:          testOp(5, nil),
		NotifyCount: 5,
	}, {
		Name: "op()==context.Canceled",
		Op: testOp(5,
			fmt.Errorf("wrapped: %w", context.Canceled)),
		Err:         "wrapped: " + context.Canceled.Error(),
		NotifyCount: 5,
	}, {
		Name: "op()==context.DeadlineExceeded",
		Op: testOp(5,
			fmt.Errorf("wrapped: %w", context.DeadlineExceeded)),
		Err:         "wrapped: " + context.DeadlineExceeded.Error(),
		NotifyCount: 5,
	}, {
		Name: "filter ErrorStop",
		Op:   testOp(8, nil),
		Filter: func() func(err error) error {
			var count int = 1
			return func(err error) error {
				if count >= 2 {
					return ErrorStop(fmt.Errorf("filtered"))
				}
				count++
				return err
			}
		}(),
		NotifyCount: 2,
		Err:         "filtered",
	}, {
		Name:        "policy stop",
		Op:          testOp(8, nil),
		Policy:      LimitAttempts{2, Immediate{}},
		NotifyCount: 2,
		Err:         "failed",
	},
}
