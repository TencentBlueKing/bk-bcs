package iface

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
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
type Step interface {
	Run(context.Context, *types.Step) error
}

// The StepFunc type is an adapter to allow the use of
// ordinary functions as a Step. If f is a function
// with the appropriate signature, StepFunc(f) is a
// Step that calls f.
type StepFunc func(context.Context, *types.Step) error

// Run calls fn(ctx, task)
func (fn StepFunc) Run(ctx context.Context, task *types.Step) error {
	return fn(ctx, task)
}
