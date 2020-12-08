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
 *
 */

package evaluate

import (
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/antonmedv/expr"
	"k8s.io/klog"
)

func EvaluateResult(result interface{}, metric v1alpha1.Metric) v1alpha1.HookPhase {
	successCondition := false
	failCondition := false
	var err error

	if metric.SuccessCondition != "" {
		successCondition, err = EvalCondition(result, metric.SuccessCondition)
		if err != nil {
			klog.Warning(err.Error())
			return v1alpha1.HookPhaseError
		}
	}
	if metric.FailureCondition != "" {
		failCondition, err = EvalCondition(result, metric.FailureCondition)
		if err != nil {
			klog.Warning(err.Error())
			return v1alpha1.HookPhaseError
		}
	}

	switch {
	case metric.SuccessCondition == "" && metric.FailureCondition == "":
		//Always return success unless there is an error
		return v1alpha1.HookPhaseSuccessful
	case metric.SuccessCondition != "" && metric.FailureCondition == "":
		// Without a failure condition, a measurement is considered a failure if the measurement's success condition is not true
		failCondition = !successCondition
	case metric.SuccessCondition == "" && metric.FailureCondition != "":
		// Without a success condition, a measurement is considered a successful if the measurement's failure condition is not true
		successCondition = !failCondition
	}

	if failCondition {
		return v1alpha1.HookPhaseFailed
	}

	if !failCondition && !successCondition {
		return v1alpha1.HookPhaseInconclusive
	}

	// If we reach this code path, failCondition is false and successCondition is true
	return v1alpha1.HookPhaseSuccessful
}

// EvalCondition evaluates the condition with the resultValue as an input
func EvalCondition(resultValue interface{}, condition string) (bool, error) {
	var err error

	env := map[string]interface{}{
		"result":  resultValue,
		"asInt":   asInt,
		"asFloat": asFloat,
	}

	// Setup a clean recovery in case the eval code panics.
	// TODO: this actually might not be nessary since it seems evaluation lib handles panics from functions internally
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("evaluation logic panicked: %v", r)
		}
	}()

	program, err := expr.Compile(condition, expr.Env(env), expr.AsBool())
	if err != nil {
		return false, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return output.(bool), err
}

func asInt(in string) int64 {
	inAsInt, err := strconv.ParseInt(in, 10, 64)
	if err == nil {
		return inAsInt
	}
	panic(err)
}

func asFloat(in string) float64 {
	inAsFloat, err := strconv.ParseFloat(in, 64)
	if err == nil {
		return inAsFloat
	}
	panic(err)
}
