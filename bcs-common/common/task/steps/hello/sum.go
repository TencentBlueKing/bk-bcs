package hello

import (
	"context"
	"fmt"
	"strconv"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
)

var (
	SumA istep.ParamKey = "sumA"
	SumB istep.ParamKey = "sumB"
	SumC istep.ParamKey = "sumC"
)

// Sum ...
func Sum(ctx context.Context, step *istep.Work) error {
	a, ok := step.GetParam(SumA.String())
	if !ok {
		return fmt.Errorf("%w: param=%s", istep.ErrParamNotFound, SumA.String())
	}

	b, ok := step.GetParam(SumB.String())
	if !ok {
		return fmt.Errorf("%w: param=%s", istep.ErrParamNotFound, SumB.String())
	}

	a1, err := strconv.Atoi(a)
	if err != nil {
		return err
	}

	b1, err := strconv.Atoi(b)
	if err != nil {
		return err
	}

	c := a1 + b1
	step.AddCommonParams(SumC.String(), fmt.Sprintf("%v", c))

	fmt.Printf("%s %s %s sumC: %v\n", step.GetTaskID(), step.GetTaskType(), step.GetName(), c)
	return nil
}
