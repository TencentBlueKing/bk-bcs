package blog

import "context"

// LogCrashStack for stack
func LogCrashStack(ctx context.Context, r interface{}) {
	GetTraceFromContext(ctx).Warnf("panic: %v, detail: %s", r, string(Stacks(false)))
}

// HandleCrash xx
func HandleCrash(fn ...func(r interface{})) {
	if r := recover(); r != nil {
		for _, f := range fn {
			f(r)
		}
	}
}
