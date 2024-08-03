package hello

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
)

// hello hello
type hello struct{}

// NewHello ...
func NewHello() iface.StepWorkerInterface {
	return &hello{}
}

// DoWork for worker exec task
func (s *hello) DoWork(ctx context.Context, work *istep.Work) error {
	fmt.Println("Hello")
	// time.Sleep(30 * time.Second)
	if err := work.AddCommonParams("name", "hello"); err != nil {
		return err
	}
	return nil
}

func init() {
	// 使用结构体注册
	istep.Register("hello", NewHello())

	// 使用函数注册
	istep.Register("sum", istep.StepWorkerFunc(Sum))
}
