# retry [![Travis-CI](https://travis-ci.org/AdamSLevy/retry.svg)](https://travis-ci.org/AdamSLevy/retry) [![GoDoc](https://godoc.org/github.com/AdamSLevy/retry?status.svg)](http://godoc.org/github.com/AdamSLevy/retry) [![Report card](https://goreportcard.com/badge/github.com/AdamSLevy/retry)](https://goreportcard.com/report/github.com/AdamSLevy/retry)

Package retry provides a reliable, simple way to retry operations.

`go get -u github.com/AdamSLevy/retry`

See [godoc.org](http://godoc.org/github.com/AdamSLevy/retry) for examples.

This package was inspired by
[github.com/cenkalti/backoff](https://github.com/cenkalti/backoff) but improves
on the design by providing Policy types that are composable, re-usable and safe
for repeated or concurrent calls to Run.
