package job

import "fmt"

type WorkerPool string

const (
	WorkerPoolAPI     WorkerPool = "api"
	WorkerPoolBrowser WorkerPool = "browser"
)

// StreamName returns the Redis Stream name for the given pool/priority.
// Streams are partitioned by pool to avoid duplicate delivery across consumer groups.
func StreamName(pool WorkerPool, priority Priority) (string, error) {
	switch pool {
	case WorkerPoolAPI, WorkerPoolBrowser:
	default:
		return "", fmt.Errorf("invalid worker pool: %q", pool)
	}

	switch priority {
	case PriorityHigh, PriorityNormal:
	default:
		return "", fmt.Errorf("invalid priority: %q", priority)
	}

	return fmt.Sprintf("synthetic:jobs:%s:%s", pool, priority), nil
}

const DLQStream = "synthetic:jobs:dlq"

// Delayed ZSET used for backoff/scheduled visibility.
const DelayedZSet = "synthetic:jobs:delayed"

const StreamFieldPayload = "payload"
const StreamFieldError = "error"
const StreamFieldErrorType = "error_type"
