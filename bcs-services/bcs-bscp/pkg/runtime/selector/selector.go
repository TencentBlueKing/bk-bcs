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

// Package selector NOTES
package selector

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/criteria/validator"
	"bscp.io/pkg/runtime/jsoni"
)

// Selector defines a group's working scope.
type Selector struct {
	// MatchAll is true means this strategy match all the target.
	// 1. if MatchAll is true, then LabelsOr and LabelsAnd field must be empty.
	// 2. if MatchAll is false, then LabelsOr and LabelsAnd field can not be empty at the same time.
	MatchAll bool `json:"match_all,omitempty"`

	// LabelsOr is instance labels in strategy which control "OR".
	LabelsOr Label `json:"labels_or,omitempty"`

	// LabelsAnd is instance labels in strategy which control "AND".
	LabelsAnd Label `json:"labels_and,omitempty"`

	// NOTE: when LabelsOr(OR) and LabelsAnd(AND) both exist, the strategy need IN(OR) logical relationship,
	// eg. (IN(LabelsOr, LabelsAnd), the strategy matched when any labels logical matched.
}

// Scan is used to decode raw message which is read from db into a structured Selector instance.
func (s *Selector) Scan(raw interface{}) error {
	if s == nil {
		return errors.New("scope selector is not initialized")
	}

	if raw == nil {
		return errors.New("raw is nil, can not be decoded")
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("decode into scope selector failed, err: %v", err)

		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return fmt.Errorf("decode into scope selector failed, err: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported scope selector raw type: %T", v)
	}
}

// Value encode the scope selector to a json raw, so that it can be stored to db with json raw.
func (s *Selector) Value() (driver.Value, error) {
	if s == nil {
		return nil, errors.New("selector is not initialized, can not be encoded")
	}

	return json.Marshal(s)
}

// Unmarshal json to Selector.
func (s *Selector) Unmarshal(bytes []byte) error {
	if err := ValidateBeforeUnmarshal(bytes); err != nil {
		return err
	}

	if err := jsoni.Unmarshal(bytes, &s); err != nil {
		return err
	}

	// validate the selector.
	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

// MarshalPB marshal selector to pb struct.
func (s *Selector) MarshalPB() (*pbstruct.Struct, error) {
	if s == nil {
		return nil, errf.New(errf.InvalidParameter, "selector is nil")
	}

	marshal, err := jsoni.Marshal(s)
	if err != nil {
		return nil, err
	}

	st := new(pbstruct.Struct)
	if err = st.UnmarshalJSON(marshal); err != nil {
		return nil, err
	}

	return st, nil
}

// IsEmpty test this selector is empty or not.
func (s *Selector) IsEmpty() bool {
	if s == nil {
		return true
	}

	if !s.MatchAll && (len(s.LabelsOr) == 0) && (len(s.LabelsAnd) == 0) {
		return true
	}

	return false
}

// Equal check if this selector is equal to another one.
func (s *Selector) Equal(other *Selector) bool {
	if s == nil && other == nil {
		return true
	}

	if s == nil || other == nil {
		return false
	}

	if s.MatchAll != other.MatchAll {
		return false
	}

	if !s.LabelsOr.Equal(other.LabelsOr) {
		return false
	}

	if !s.LabelsAnd.Equal(other.LabelsAnd) {
		return false
	}

	return true
}

// MatchLabels matches strategy base on labels info.
func (s *Selector) MatchLabels(labels map[string]string) (bool, error) {
	if s.MatchAll {
		return true, nil
	}

	if len(labels) == 0 {
		return false, nil
	}

	// match IN multi LabelsOr...
	matched, err := s.matchLabelsOr(s.LabelsOr, labels)
	if err != nil {
		return false, err
	}

	if matched {
		return true, nil
	}

	if len(s.LabelsAnd) == 0 {
		return false, nil
	}

	// match IN multi LabelsAnd...
	matched, err = s.matchLabelsAnd(s.LabelsAnd, labels)
	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

	return true, nil
}

// Validate validate a strategy is valid or not
func (s *Selector) Validate() error {
	if s == nil {
		return errors.New("strategy is nil")
	}

	if s.MatchAll {
		if len(s.LabelsOr) != 0 || len(s.LabelsAnd) != 0 {
			return errors.New("match_all is true, but labels_or or labels_and is not empty")
		}
		return nil
	}

	// not match all, at least one of labels_and or labels_or labels should not be empty.
	if len(s.LabelsOr) == 0 && len(s.LabelsAnd) == 0 {
		return errors.New("match_all is false, but both labels_or and labels_and is empty")
	}

	// validate and labels
	if len(s.LabelsAnd) != 0 {

		if len(s.LabelsAnd) > validator.MaxLabelKeyCount {
			return fmt.Errorf("labels_and contains oversize labels, should be less than %d labels",
				validator.MaxLabelKeyCount)
		}

		if err := s.LabelsAnd.Validate(); err != nil {
			return err
		}

	}

	// validate or labels
	if len(s.LabelsOr) != 0 {

		if len(s.LabelsOr) > validator.MaxLabelKeyCount {
			return fmt.Errorf("labels_or contains oversize labels, should be less than %d labels",
				validator.MaxLabelKeyCount)
		}

		if err := s.LabelsOr.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Selector) matchLabelsOr(labelsOr Label, labels map[string]string) (bool, error) {
	if len(labelsOr) == 0 {
		return false, nil
	}

	var exist bool
	for _, one := range labelsOr {
		if _, exist = labels[one.Key]; !exist {
			continue
		}

		matched, err := one.Match(labels)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	// all labels OR not matched.
	return false, nil
}

func (s *Selector) matchLabelsAnd(labelsAnd Label, labels map[string]string) (bool, error) {
	if len(labelsAnd) == 0 {
		return false, nil
	}

	var exist bool
	for _, one := range labelsAnd {
		if _, exist = labels[one.Key]; !exist {
			return false, nil
		}

		matched, err := one.Match(labels)
		if err != nil {
			return false, err
		}

		if !matched {
			return false, nil
		}
	}

	// all labels AND matched.
	return true, nil
}
