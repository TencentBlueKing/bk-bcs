package task

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// TaskBuilder ...
type TaskBuilder interface { // nolint
	TaskInfo() types.TaskInfo
	Steps() ([]*types.Step, error) // Steps init step and define StepSequence
	BuildTask(t types.Task) (types.Task, error)
}

// NewByTaskBuilder init task from builder
func NewByTaskBuilder(builder TaskBuilder, opts ...types.TaskOption) (*types.Task, error) {
	// 声明step
	steps, err := builder.Steps()
	if err != nil {
		return nil, err
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	task := types.NewTask(builder.TaskInfo(), opts...)
	task.Steps = steps
	task.CurrentStep = steps[0].GetName()

	// 自定义extraJson等
	newTask, err := builder.BuildTask(*task)
	if err != nil {
		return nil, err
	}

	return &newTask, nil
}
