package hello

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// Hello hello
type Hello struct{}

// Name method name
func (s Hello) Name() string {
	return "hello"
}

// DoWork for worker exec task
func (s Hello) DoWork(ctx context.Context, task *types.Step) error {
	fmt.Println("Hello")
	return nil
}

// Close ...
func (s Hello) Close(_ error) {
}
