package utils

import "time"

func Retry(f func() bool, retryLimit int, retryDelay time.Duration) {
	if retryLimit <= 0 {
		return
	}
	shouldRetry := f()
	if shouldRetry {
		time.Sleep(retryDelay)
		Retry(f, retryLimit-1, retryDelay)
	}
}
