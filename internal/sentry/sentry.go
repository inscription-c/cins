package sentry

import (
	"github.com/getsentry/sentry-go"
	"time"
)

func RecoverPanic() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		sentry.Flush(time.Second * 2)
		panic(err)
	}
}
