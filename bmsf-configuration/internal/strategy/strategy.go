/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package strategy

import (
	"encoding/json"
	"fmt"
	"strings"

	pbcommon "bk-bscp/internal/protocol/common"
)

const (
	// EmptySidecarLabels is empty json string for sidecar labels.
	EmptySidecarLabels = "{}"

	// EmptyStrategy is empty json string for release.
	EmptyStrategy = "{}"

	// EmptyContentIndex is empty json string for content index.
	EmptyContentIndex = "{}"
)

// SidecarLabels is app sidecar instance labels struct.
type SidecarLabels struct {
	Labels map[string]string
}

/*
	example:
	{
		"Labels":
			{
				"k1":"v1",
				"k2":"v2",
				"k3":"v3"
			}
	}
*/

// ContentIndex is config commit content index struct.
type ContentIndex struct {
	// LabelsOr is content index labels in strategy which control "OR".
	LabelsOr []map[string]string

	// LabelsAnd is content index labels in strategy which control "AND".
	LabelsAnd []map[string]string

	/*
	   	example:
	   	{
	   		"LabelsOr": [
	   			{
	   				"k1":"1",        // k1=1
	   				"k2":"1,2,3",    // k2 in (1,2,3)
	   				"k3":"ne|1,2,3", // k3 not in (1,2,3)
	   				"k4":"gt|1",     // k4 > 1
	   			},
	   			{
	   				"k1":"1",        // k1=1
	   				"k2":"1,2,3",    // k2 in (1,2,3)
	   				"k3":"ne|1,2,3", // k3 not in (1,2,3)
	   				"k4":"gt|1",     // k4 > 1
	   			}
	           ],
	   		"LabelsAnd": [
	   			{
	   				"k3":"1",	  // k3=1
	   				"k4":"1,2,3", // k4 in (1,2,3)
	   			},
	   			{
	   				"k3":"1",	  // k3=1
	   				"k4":"1,2,3", // k4 in (1,2,3)
	   			}
	           ]
	   	}

	   	Labels kv relation ops,
	   	    format:
	   		    "KEY": "OP|VALUE"

	   	    OP:
	   		    1.=: empty or eq
	   		    2.!=: ne
	   		    3.>: gt
	   		    4.<: lt
	   		    5.>=: ge
	   		    6.<=: le
	*/
}

// MatchLabels matchs content index strategy base on labels info.
func (index *ContentIndex) MatchLabels(labels map[string]string) bool {
	if len(index.LabelsOr) == 0 && len(index.LabelsAnd) == 0 {
		// NOTE: empty labels is the content which could be found by any instance.
		return true
	}

	// match IN multi LabelsOr...
	for _, labelsOr := range index.LabelsOr {
		if index.matchLabelsOr(labelsOr, labels) {
			return true
		}
	}

	// match IN multi LabelsAnd...
	for _, labelsAnd := range index.LabelsAnd {
		if index.matchLabelsAnd(labelsAnd, labels) {
			return true
		}
	}
	return false
}

func (index *ContentIndex) matchLabelsOr(labelsOr, labels map[string]string) bool {
	if len(labelsOr) == 0 {
		return false
	}

	for k, v := range labelsOr {
		// parse v from strategy like k=gt|1,2,3.
		op, vals, err := ParseLabelVals(v)
		if err != nil {
			return false
		}

		if IsLabelMatch(labels[k], op, vals) {
			return true
		}
	}

	// all labels OR not matched.
	return false
}

func (index *ContentIndex) matchLabelsAnd(labelsAnd, labels map[string]string) bool {
	if len(labelsAnd) == 0 {
		return false
	}

	for k, v := range labelsAnd {
		// parse v from strategy like k=gt|1,2,3.
		op, vals, err := ParseLabelVals(v)
		if err != nil {
			return false
		}

		if !IsLabelMatch(labels[k], op, vals) {
			return false
		}
	}

	// all labels AND matched.
	return true
}

// MatchInstance matchs content index strategy base on instance info.
func (index *ContentIndex) MatchInstance(instance *pbcommon.AppInstance) bool {
	// instance labels.
	insLabels := SidecarLabels{}
	if err := json.Unmarshal([]byte(instance.Labels), &insLabels); err != nil {
		// unknow instance labels.
		return false
	}
	return index.MatchLabels(insLabels.Labels)
}

// Strategy is struct of strategies in release.
type Strategy struct {
	// NOTE: it is OR logical relationship in labels OR level, it is AND logical relationship in labels AND level,
	// it is IN(OR) logical relationship between Labels and LabelsAnd.

	// LabelsOr is instance labels in strategy which control "OR".
	LabelsOr []map[string]string

	// LabelsAnd is instance labels in strategy which control "AND".
	LabelsAnd []map[string]string

	// NOTE: when LabelsOr(OR) and LabelsAnd(AND) both exist, the strategy need IN(OR) logical relationship,
	// eg. (IN(LabelsOr, LabelsAnd), the strategy matched when any labels logical matched.

	/*
	   	example:
	   	{
	   		"LabelsOr": [
	   			{
	   				"k1":"1",        // k1=1
	   				"k2":"1,2,3",    // k2 in (1,2,3)
	   				"k3":"ne|1,2,3", // k3 not in (1,2,3)
	   				"k4":"gt|1",     // k4 > 1
	   			},
	   			{
	   				"k1":"1",        // k1=1
	   				"k2":"1,2,3",    // k2 in (1,2,3)
	   				"k3":"ne|1,2,3", // k3 not in (1,2,3)
	   				"k4":"gt|1",     // k4 > 1
	   			}
	        ],
	   		"LabelsAnd": [
	   			{
	   				"k3":"1",	  // k3=1
	   				"k4":"1,2,3", // k4 in (1,2,3)
	   			},
	   			{
	   				"k3":"1",	  // k3=1
	   				"k4":"1,2,3", // k4 in (1,2,3)
	   			}
	        ]
	   	}

	   	Labels kv relation ops,
	   		format:
	   			"KEY": "OP|VALUE"

	   		OP:
	   			1.=: empty or eq
	   			2.!=: ne
	   			3.>: gt
	   			4.<: lt
	   			5.>=: ge
	   			6.<=: le
	*/
}

// MatchLabels matchs strategy base on labels info.
func (s *Strategy) MatchLabels(labels map[string]string) bool {
	if len(s.LabelsOr) == 0 && len(s.LabelsAnd) == 0 {
		// NOTE: empty labels is the strategy which could be found by any instance.
		return true
	}

	// match IN multi LabelsOr...
	for _, labelsOr := range s.LabelsOr {
		if s.matchLabelsOr(labelsOr, labels) {
			return true
		}
	}

	// match IN multi LabelsAnd...
	for _, labelsAnd := range s.LabelsAnd {
		if s.matchLabelsAnd(labelsAnd, labels) {
			return true
		}
	}
	return false
}

func (s *Strategy) matchLabelsOr(labelsOr, labels map[string]string) bool {
	if len(labelsOr) == 0 {
		return false
	}

	for k, v := range labelsOr {
		// parse v from strategy like k=gt|1,2,3.
		op, vals, err := ParseLabelVals(v)
		if err != nil {
			return false
		}

		if IsLabelMatch(labels[k], op, vals) {
			return true
		}
	}

	// all labels OR not matched.
	return false
}

func (s *Strategy) matchLabelsAnd(labelsAnd, labels map[string]string) bool {
	if len(labelsAnd) == 0 {
		return false
	}

	for k, v := range labelsAnd {
		// parse v from strategy like k=gt|1,2,3.
		op, vals, err := ParseLabelVals(v)
		if err != nil {
			return false
		}

		if !IsLabelMatch(labels[k], op, vals) {
			return false
		}
	}

	// all labels AND matched.
	return true
}

// ValidateLabels checks labels formats for strategy or content index.
func ValidateLabels(labels map[string]string) error {
	if len(labels) == 0 {
		// just empty labels.
		return nil
	}

	// range and check all labels.
	for _, v := range labels {
		// OP|VALUES.
		opVals := strings.Split(v, "|")

		if len(opVals) == 2 {
			op := opVals[0]

			if op != "eq" && op != "ne" && op != "gt" &&
				op != "lt" && op != "ge" && op != "le" {
				// invalid op.
				return fmt.Errorf("invalid labels values op type: %s", v)
			}

			values := opVals[1]

			if len(values) == 0 {
				// invalid values.
				return fmt.Errorf("invalid labels values: %s", v)
			}

		} else if len(opVals) == 1 {
			values := opVals[0]

			if len(values) == 0 {
				// invalid values.
				return fmt.Errorf("invalid labels values: %s", v)
			}
		} else {
			// invalid formats.
			return fmt.Errorf("invalid labels formats: %s", v)
		}
	}

	// valid label values.
	return nil
}

// ParseLabelVals returns label values from k=gt|1,2,3.
func ParseLabelVals(labelValue string) (string, []string, error) {
	// final result op, empty means 'eq'.
	op := "eq"

	// final result vals.
	vals := []string{}

	// split op and vals.
	opVals := strings.Split(labelValue, "|")

	// real values after split.
	realValues := ""

	if len(opVals) == 2 {
		// OP|VALUES.
		op = opVals[0]
		realValues = opVals[1]

	} else if len(opVals) == 1 {
		// VALUES.
		op = "eq"
		realValues = opVals[0]

	} else {
		// invalid formats.
		return "", nil, fmt.Errorf("invalid labels values format: %s", labelValue)
	}

	if op != "eq" && op != "ne" && op != "gt" &&
		op != "lt" && op != "ge" && op != "le" {
		// invalid op.
		return "", nil, fmt.Errorf("invalid labels values op type: %s", op)
	}

	// split real values.
	values := strings.Split(realValues, ",")
	for _, v := range values {
		vals = append(vals, v)
	}
	return op, vals, nil
}

// IsLabelMatch matchs labels KV level base on op and values.
// op is eq/ne/gt/lt/ge/le.
func IsLabelMatch(insValue, op string, targetVals []string) bool {
	// TODO use map to make eq/ne compare action faster.

	switch op {
	case "eq":
		for _, val := range targetVals {
			if insValue == val {
				return true
			}
		}
		return false

	case "ne":
		for _, val := range targetVals {
			if insValue == val {
				return false
			}
		}
		return true

	case "gt":
		for _, val := range targetVals {
			if insValue <= val {
				return false
			}
		}
		return true

	case "lt":
		for _, val := range targetVals {
			if insValue >= val {
				return false
			}
		}
		return true

	case "ge":
		for _, val := range targetVals {
			if insValue < val {
				return false
			}
		}
		return true

	case "le":
		for _, val := range targetVals {
			if insValue > val {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// Matcher matches the strategy base on information in app sidecar instance.
type Matcher func(strategy *Strategy, instance *pbcommon.AppInstance) bool

// Handler handles the strategy matching.
type Handler struct {
	matcher Matcher
}

// NewHandler creates a new strategy handler.
func NewHandler(matcher Matcher) *Handler {
	if matcher == nil {
		matcher = defaultMatcher
	}
	return &Handler{matcher: matcher}
}

// Matcher returns matcher in handler.
func (h *Handler) Matcher() Matcher {
	return h.matcher
}

// default matcher impl.
var defaultMatcher = func(strategy *Strategy, instance *pbcommon.AppInstance) bool {
	if strategy == nil || instance == nil {
		return false
	}
	return defaultMatch(strategy, instance)
}

func defaultMatch(strategy *Strategy, instance *pbcommon.AppInstance) bool {
	// #lizard forgives
	insLabels := SidecarLabels{}
	if err := json.Unmarshal([]byte(instance.Labels), &insLabels); err != nil {
		// unknow instance labels.
		return false
	}
	return strategy.MatchLabels(insLabels.Labels)
}
