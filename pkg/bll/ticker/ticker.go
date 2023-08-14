package ticker

import (
	"context"
	"time"
)

func Every(ctx context.Context, d time.Duration) <-chan time.Time {
	c := make(chan time.Time)

	go func() {
		defer close(c)
		for {
			c <- time.Now()

			select {
			case <-ctx.Done():
				return
			case <-time.After(d):
			}
		}
	}()

	return c

}
