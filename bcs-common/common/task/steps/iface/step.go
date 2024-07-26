package iface

import (
	"context"
)

// A Step processes tasks.
//
// Process should return nil if the processing of a task
// is successful.
//
// If Process returns a non-nil error or panics, the task
// will be retried after delay if retry-count is remaining,
// otherwise the task will be archived.
//
// One exception to this rule is when Process returns a SkipRetry error.
// If the returned error is SkipRetry or an error wraps SkipRetry, retry is
// skipped and the task will be immediately archived instead.
type StepHandler interface {
	Name() string
	Run(context.Context, Step) error
	Close(error)
}

// Step ...
type Step interface {
	GetParam(key string) (string, error)
	GetParams() (map[string]string, error)
	GetOutput(key string) (string, error)
	GetOutputs() (map[string]string, error)
	AddOutput(key, value string) error
	AddOutputs(map[string]string) error
	SetOutputs(map[string]string) error
}
