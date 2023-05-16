/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package selector

import (
	"encoding/json"
	"errors"
	"fmt"

	"bscp.io/pkg/criteria/validator"

	"github.com/tidwall/gjson"
)

// Label defines the basic label elements
type Label []Element

// Validate validate a label is valid or not.
func (lb Label) Validate() error {
	for _, one := range lb {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Element defines the basic element of a label
type Element struct {
	Key   string      `json:"key"`
	Op    Operator    `json:"op"`
	Value interface{} `json:"value"`
}

// Match matches the element labels
func (e *Element) Match(labels map[string]string) (bool, error) {
	return e.Op.Match(e, labels)
}

// Validate is to validate if this element is valid or not.
func (e *Element) Validate() error {
	if e == nil {
		return errors.New("empty element")
	}

	if err := validator.ValidateLabelKey(e.Key); err != nil {
		return fmt.Errorf("invalid label key: %s, %v", e.Key, err)
	}

	// validate operator
	_, exists := OperatorEnums[e.Op.Name()]
	if !exists {
		return fmt.Errorf("unsupported operator: %v", e.Op.Name())
	}

	if err := e.Op.Validate(e); err != nil {
		return err
	}

	return nil
}

type element struct {
	Key   string      `json:"key"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

// MarshalJSON element to json, the op field requires special handling because it is an interface.
func (e Element) MarshalJSON() ([]byte, error) {
	el := new(element)
	el.Key = e.Key
	el.Op = string(e.Op.Name())
	el.Value = e.Value

	return json.Marshal(el)
}

// UnmarshalJSON unmarshal a json string to a element depends on it's operator type.
func (e *Element) UnmarshalJSON(bytes []byte) error {
	parsed := gjson.GetManyBytes(bytes, "key", "op", "value")
	k := parsed[0].String()
	op := parsed[1].String()
	v := parsed[2]

	if len(k) == 0 {
		return errors.New("invalid key field")
	}

	// set key field
	e.Key = k

	if len(op) == 0 {
		return errors.New("invalid op field")
	}

	operator, exists := OperatorEnums[OperatorType(op)]
	if !exists {
		return fmt.Errorf("unsupported op: %v", op)
	}

	// set op field
	e.Op = operator

	// set value field
	if err := json.Unmarshal([]byte(v.Raw), &e.Value); err != nil {
		return err
	}

	return nil
}
