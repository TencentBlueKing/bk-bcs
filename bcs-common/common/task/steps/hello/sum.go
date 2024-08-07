/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	// time.Sleep(30 * time.Second)

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
