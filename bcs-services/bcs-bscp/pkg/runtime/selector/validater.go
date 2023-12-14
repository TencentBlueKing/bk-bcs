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

package selector

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

type selectorRaw struct {
	MatchAll  bool       `json:"match_all,omitempty"`
	LabelsOr  []*element `json:"labels_or,omitempty"`
	LabelsAnd []*element `json:"labels_and,omitempty"`
}

func validatEelement(e *element) error {
	if len(e.Key) == 0 {
		return errors.New("selector label key is required")
	}
	if len(e.Op) == 0 {
		return errors.New("selector label op is required")
	}

	operator, exists := OperatorEnums[OperatorType(e.Op)]
	if !exists {
		return fmt.Errorf("unsupported op: %s", e.Op)
	}

	if operator == &InOperator || operator == &NotInOperator {
		values, ok := e.Value.([]interface{})
		if !ok {
			return fmt.Errorf("selector label value %v must be list", e.Value)
		}
		if len(values) == 0 {
			return fmt.Errorf("selector label value %v is empty", e.Value)
		}
		for _, value := range values {
			v, ok := value.(string)
			if !ok {
				return fmt.Errorf("selector label value %v element must be string", e.Value)
			}
			if v == "" {
				return fmt.Errorf("selector label value %v element is empty", e.Value)
			}
		}
		return nil
	}

	if operator == &EqualOperator || operator == &NotEqualOperator {
		value, ok := e.Value.(string)
		if !ok {
			return fmt.Errorf("selector label value %s must be string", e.Value)
		}
		if len(value) == 0 {
			return fmt.Errorf("selector label value %v is empty", e.Value)
		}
		return nil
	}

	if operator == &RegexOperator || operator == &NotRegexOperator {
		value, ok := e.Value.(string)
		if !ok {
			return fmt.Errorf("selector label value %s must be string", e.Value)
		}
		if len(value) == 0 {
			return fmt.Errorf("selector label value %v is empty", e.Value)
		}
		if _, err := regexp.Compile(value); err != nil {
			return fmt.Errorf("selector label value %v is invalid regex: %v", value, err)
		}
		return nil
	}

	_, ok := e.Value.(float64)
	if !ok {
		return fmt.Errorf("selector label value %v must be int/float", e.Value)
	}

	return nil
}

// ValidateBeforeUnmarshal 可读校验
func ValidateBeforeUnmarshal(bytes []byte) error {
	r := selectorRaw{}
	if err := jsoni.Unmarshal(bytes, &r); err != nil {
		return err
	}

	if len(r.LabelsAnd) == 0 && len(r.LabelsOr) == 0 {
		return errors.New("selector labels is required")
	}

	for _, v := range r.LabelsOr {
		if err := validatEelement(v); err != nil {
			return err
		}
	}

	for _, v := range r.LabelsAnd {
		if err := validatEelement(v); err != nil {
			return err
		}
	}

	return nil
}
