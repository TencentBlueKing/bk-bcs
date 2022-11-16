/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package operator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Policy defines a user's authorize policy in IAM.
type Policy struct {
	Operator OpType `json:"op"`
	// Element is a pointer interface point to the implements' struct,
	// which should be one of Content or FieldValue.
	Element
}

// UnmarshalJSON unmarshal the json to the policy.
func (p *Policy) UnmarshalJSON(i []byte) error {
	if string(i) == "{}" {
		return nil
	}

	broker := new(policyBroker)
	err := json.Unmarshal(i, broker)
	if err != nil {
		return err
	}

	p.Operator = broker.Operator

	if broker.Operator == And || broker.Operator == Or {
		content := new(Content)
		if err := json.Unmarshal(broker.Content, &content.Content); err != nil {
			return err
		}
		p.Element = content
		return nil
	}

	if broker.Operator == In || broker.Operator == Nin {
		to := make([]interface{}, 0)
		if err := json.Unmarshal(broker.Value, &to); err != nil {
			return err
		}

		p.Element = &FieldValue{
			Field: broker.Field,
			Value: to,
		}

	} else {
		to := new(interface{})
		if err := json.Unmarshal(broker.Value, &to); err != nil {
			return err
		}

		p.Element = &FieldValue{
			Field: broker.Field,
			Value: *to,
		}
	}

	return nil
}

type policyBroker struct {
	Operator OpType          `json:"op"`
	Content  json.RawMessage `json:"content"`
	Field    Field           `json:"field"`
	Value    json.RawMessage `json:"value"`
}

// MarshalJSON is used to marshal the policy to the standard
// iam policy protocol, which is not correspond to the struct
// we defined here.
// Note: when you marshal the policy, the policy must be a pointer,
// otherwise, the marshaled json struct is wrong.
func (p *Policy) MarshalJSON() ([]byte, error) {
	js, err := json.Marshal(p.Element)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	buf.WriteString(`{"op":"`)
	buf.WriteString(string(p.Operator))
	buf.WriteString(`",`)
	buf.Write(js[1 : len(js)-1])
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Element defines each element of a user's authorize policy.
type Element interface {
	// EleName is the name of each kind of element.
	EleName() string
}

// Content is the content type element of a user's authorize policy.
type Content struct {
	// Content is only exist when OpType is "And" or "OR"
	Content []*Policy `json:"content"`
}

// EleName return the content type name
func (e *Content) EleName() string {
	return "content"
}

// FieldValue is the filed value type element of a user's authorize policy.
type FieldValue struct {
	// Field and Value is only exist when OpType is not
	// one of "And" or "OR"
	Field Field       `json:"field"`
	Value interface{} `json:"value"`
}

// EleName return the filed value type name.
func (f *FieldValue) EleName() string {
	return "field_value"
}

// Field defines the FieldValue's Field
type Field struct {
	Resource  string
	Attribute string
}

// UnmarshalJSON unmarshal a json to the Field.
func (f *Field) UnmarshalJSON(i []byte) error {
	if string(i) == "\"\"" {
		f.Attribute = ""
		f.Resource = ""
		return nil
	}

	index := bytes.IndexByte(i, '.')
	if index < 0 {
		return errors.New("invalid \"field\"")
	}

	f.Resource = string(bytes.TrimLeft(i[:index], "\""))
	f.Attribute = string(bytes.TrimRight(i[index+1:], "\""))

	if f.Resource == "" || f.Attribute == "" {
		return errors.New("invalid \"field\"")
	}

	return nil
}

// MarshalJSON marshal the Field to a json string.
func (f *Field) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s.%s\"", f.Resource, f.Attribute)), nil
}
