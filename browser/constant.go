package browser

import "time"

const (
	BackoffFactor = time.Second
	MaxRetries    = 3
)
