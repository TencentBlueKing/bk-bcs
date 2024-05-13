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

package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

func TestIsSameRuleCondition(t *testing.T) {
	conditionA := []types.RuleCondition{
		{
			Field:            toStrPtr("host-header"),
			HostHeaderConfig: &types.HostHeaderConditionConfig{Values: []string{"www.qq.com"}},
		},
		{
			Field:             toStrPtr("path-pattern"),
			PathPatternConfig: &types.PathPatternConditionConfig{Values: []string{"/"}},
		},
	}

	conditionB := []types.RuleCondition{
		{
			Field:            toStrPtr("host-header"),
			HostHeaderConfig: &types.HostHeaderConditionConfig{Values: []string{"www.qq.com"}},
		},
		{
			Field:             toStrPtr("path-pattern"),
			PathPatternConfig: &types.PathPatternConditionConfig{Values: []string{"/grafana/pracing"}},
		},
	}

	conditionC := []types.RuleCondition{
		{
			Field:             toStrPtr("path-pattern"),
			PathPatternConfig: &types.PathPatternConditionConfig{Values: []string{"/"}},
		},
		{
			Field:            toStrPtr("host-header"),
			HostHeaderConfig: &types.HostHeaderConditionConfig{Values: []string{"www.qq.com"}},
		},
	}

	if isSameRuleCondition(conditionA, conditionB) {
		t.Error("different condition")
	}
	if !isSameRuleCondition(conditionA, conditionC) {
		t.Error("same condition")
	}
}

func toStrPtr(str string) *string {
	return &str
}
