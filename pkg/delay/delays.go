package delay

import (
	"math"
	"math/rand/v2"
	"time"
)

// -----------------------------------------------------------------------------
// Retry strategies
// -----------------------------------------------------------------------------

// RetryDelay defines a strategy for how long to wait before the next retry.
type RetryDelay interface {
	Wait(taskName string, attempt int)
}

// ConstantDelay waits a constant number of seconds between retries.
type ConstantDelay struct{ Period int }

func (d ConstantDelay) Wait(task string, attempt int) {
	time.Sleep(time.Duration(d.Period) * time.Second)
}

// ExponentialBackoff waits exponentially longer between each retry attempt.
type ExponentialBackoff struct{}

func (d ExponentialBackoff) Wait(task string, attempt int) {
	backoff := time.Duration(math.Min(2*math.Pow(2, float64(attempt)), 10))
	// Add jitter to the backoff to avoid retry collisions.
	jitter := time.Duration(rand.Float64() * float64(backoff) * 0.5)
	time.Sleep((backoff + jitter) * time.Second)
}
