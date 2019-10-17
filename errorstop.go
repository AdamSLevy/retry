package retry

// ErrorStop wraps err such that when returned from an op or filter, it will
// cause Run to stop immediately and return err.
func ErrorStop(err error) error {
	return errorStop{err}
}

type errorStop struct{ err error }

func (e errorStop) Error() string {
	return e.err.Error()
}
