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

package types

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// ClusteredNamespace limit query range of clusters and namespaces
type ClusteredNamespace struct {
	ClusterId  string   `json:"clusterId"`
	Namespaces []string `json:"namespaces"`
}

// MulticlusterListReqParams request params of multi cluster resource list
type MulticlusterListReqParams struct {
	Offset              int64                 `json:"offset"`
	Limit               int64                 `json:"limit"`
	ClusteredNamespaces []*ClusteredNamespace `json:"clusteredNamespaces"`
	Field               string                `json:"field"`
	LabelSelector       string                `json:"labelSelector"`
	Conditions          []*operator.Condition `json:"conditions"`
	Sort                map[string]int        `json:"sort"`
}

// EnsureConditions ensure conditions's value is operator.M instead of map[string]interface{}
func (r *MulticlusterListReqParams) EnsureConditions() error {
	// check custom conditions, and convert value in custom conditions into operator.M
	if len(r.Conditions) != 0 {
		newConditions, err := r.ensureConditionValueSlice(r.Conditions)
		if err != nil {
			return err
		}
		r.Conditions = newConditions

	}
	return nil
}

func (r *MulticlusterListReqParams) ensureConditionValueSlice(conds []*operator.Condition) (
	[]*operator.Condition, error) {
	newConditions := []*operator.Condition{}
	for _, oldCond := range conds {
		newCond, err := r.ensureConditionValue(oldCond)
		if err != nil {
			return nil, err
		}
		newConditions = append(newConditions, newCond)
	}
	return newConditions, nil
}

func (r *MulticlusterListReqParams) ensureConditionValue(cond *operator.Condition) (*operator.Condition, error) {
	if cond == nil {
		return cond, nil
	}

	// convert Value into operator.M from map[string]interface{}
	if cond.Value != nil {
		m, ok := (cond.Value).(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Value %+v can not convert into map[string]interface{}", cond.Value)
		}
		newValue := operator.M{}
		for k, v := range m {
			newValue[k] = v
		}
		cond.Value = newValue
	}

	// ensure children
	if cond.Children != nil {
		var newChildren []*operator.Condition
		for _, child := range cond.Children {
			newChild, err := r.ensureConditionValue(child)
			if err != nil {
				return nil, err
			}
			newChildren = append(newChildren, newChild)
		}
		cond.Children = newChildren
	}

	return cond, nil
}
