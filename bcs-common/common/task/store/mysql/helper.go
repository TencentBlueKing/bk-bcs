package mysql

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

func GetStepRecords(t *types.Task) []*StepRecords {
	records := make([]*StepRecords, 0, len(t.Steps))
	for _, step := range t.Steps {
		record := &StepRecords{
			TaskID:              t.TaskID,
			Name:                step.Name,
			Alias:               step.Alias,
			Extras:              step.Extras,
			Status:              step.Status,
			Message:             step.Message,
			SkipOnFailed:        step.SkipOnFailed,
			RetryCount:          step.RetryCount,
			Params:              step.Params,
			Start:               step.Start,
			End:                 step.End,
			ExecutionTime:       step.ExecutionTime,
			MaxExecutionSeconds: step.MaxExecutionSeconds,
			LastUpdate:          step.LastUpdate,
		}
		records = append(records, record)
	}

	return records
}
func GetTaskRecord(t *types.Task) *TaskRecords {
	stepSequence := make([]string, 0, len(t.Steps))
	for i := range t.Steps {
		stepSequence = append(stepSequence, t.Steps[i].Name)
	}

	record := &TaskRecords{
		TaskID:              t.TaskID,
		TaskType:            t.TaskType,
		TaskName:            t.TaskName,
		CurrentStep:         t.CurrentStep,
		StepSequence:        stepSequence,
		CallBackFuncName:    t.CallBackFuncName,
		CommonParams:        t.CommonParams,
		ExtraJson:           t.ExtraJson,
		Status:              t.Status,
		Message:             t.Message,
		ForceTerminate:      t.ForceTerminate,
		Start:               t.Start,
		End:                 t.End,
		ExecutionTime:       t.ExecutionTime,
		MaxExecutionSeconds: t.MaxExecutionSeconds,
		Creator:             t.Creator,
		LastUpdate:          t.LastUpdate,
		Updater:             t.Updater,
	}
	return record
}

func ToTask(task *TaskRecords, steps []*StepRecords) *types.Task {
	t := &types.Task{
		TaskID:              task.TaskID,
		TaskType:            task.TaskType,
		TaskName:            task.TaskName,
		CurrentStep:         task.CurrentStep,
		CallBackFuncName:    task.CallBackFuncName,
		CommonParams:        task.CommonParams,
		ExtraJson:           task.ExtraJson,
		Status:              task.Status,
		Message:             task.Message,
		ForceTerminate:      task.ForceTerminate,
		Start:               task.Start,
		End:                 task.End,
		ExecutionTime:       task.ExecutionTime,
		MaxExecutionSeconds: task.MaxExecutionSeconds,
		Creator:             task.Creator,
		LastUpdate:          task.LastUpdate,
		Updater:             task.Updater,
	}

	t.Steps = make([]*types.Step, 0, len(steps))
	for _, step := range steps {
		t.Steps = append(t.Steps, step.ToStep())
	}
	return t
}

func getUpdateTaskRecord(t *types.Task) *TaskRecords {
	record := &TaskRecords{
		CurrentStep:   t.CurrentStep,
		CommonParams:  t.CommonParams,
		Status:        t.Status,
		Message:       t.Message,
		Start:         t.Start,
		End:           t.End,
		ExecutionTime: t.ExecutionTime,
		Updater:       t.Updater,
	}
	return record
}

func getUpdateStepRecord(t *types.Step) *StepRecords {
	record := &StepRecords{
		Params:        t.Params,
		Status:        t.Status,
		Message:       t.Message,
		Start:         t.Start,
		End:           t.End,
		ExecutionTime: t.ExecutionTime,
		RetryCount:    t.RetryCount,
	}
	return record
}
