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

package local

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

// CalculatorLocal local calculator
type CalculatorLocal struct {
}

// NewCalculatorLocal create the instance of local calculator
func NewCalculatorLocal() calculator.CalculateInterface {
	return &CalculatorLocal{}
}

// Calculate from local
func (c *CalculatorLocal) Calculate(ctx context.Context, req *calculator.CalculateConvergeRequest) (
	plan calculator.ResultPlan, err error) {
	return plan, errors.Errorf("local not implemented")
}
