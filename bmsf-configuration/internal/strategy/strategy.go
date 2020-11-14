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

// Strategy is struct of strategies in release.
type Strategy struct {
	Appid string

	// NOTE: it is OR logical relationship in clusterid zoneid dc ip levels, it is OR logical relationship
	// in labels OR level, it is AND logical relationship in labels AND level, it is AND logical relationship
	// between clusterid zoneid dc ip labels(Labels && LabelsAnd) levels, it is AND logical relationship between
	// Labels and LabelsAnd.

	Clusterids []string
	Zoneids    []string
	Dcs        []string
	IPs        []string

	// Labels is instance labels in strategy which control "OR".
	Labels map[string]string

	// LabelsAnd is instance labels in strategy which control "AND".
	LabelsAnd map[string]string

	// NOTE: when Labels(OR) and LabelsAnd(AND) both exist, the strategy need AND logical relationship,
	// eg. (Labels AND LabelsAnd), the strategy matched when all labels logical matched.
	// The labels support IN (like SQL) grammar, eg. k=1,2,3, mean k in (1,2,3)

	// eg. (clusterid IN(...)) && (zoneid IN(...)) && (dc IN(...)) && (ip IN(...)) && (labels (K IN(...) || K IN(...) ...)) && (labelsAnd (K IN(...) && K IN(...) ...))
}

/*
	example:
	{
		"Appid":"appid",
		"Clusterids":
			["clusterid01","clusterid02","clusterid03"],
		"Zoneids":
			["zoneid01","zoneid02","zoneid03"],
		"Dcs":
			["dc01","dc02","dc03"],
		"IPs":
			["X.X.X.1","X.X.X.2","X.X.X.3"],
		"Labels":
			{
				"k1":"1",        // k1=1
				"k2":"1,2,3",    // k2 in (1,2,3)
				"k3":"ne|1,2,3", // k3 not in (1,2,3)
				"k4":"gt|1",     // k4 > 1
			},
		"LabelsAnd":
			{
				"k3":"1",	  // k3=1
				"k4":"1,2,3", // k4 in (1,2,3)
			}
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

// ValidateLabels checks labels formats for strategy.
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
		if len(v) != 0 {
			vals = append(vals, v)
		}
	}

	return op, vals, nil
}

// IsLabelMatch matchs labels KV level base on op and values.
// op is eq/ne/gt/lt/ge/le.
func IsLabelMatch(insValue, op string, targetVals []string) bool {
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
	return matchWhitelist(strategy, instance)
}

// matchWhitelist matchs strategy.
func matchWhitelist(strategy *Strategy, instance *pbcommon.AppInstance) bool {
	// #lizard forgives
	if strategy.Appid != instance.Appid {
		return false
	}

	// matching results, cluster level.
	isClusteridMatched := false

	// matching results, zoneid level.
	isZoneidMatched := false

	// matching results, dc level.
	isDcMatched := false

	// matching results, ip level.
	isIPMatched := false

	// matching results, labels OR level.
	isLabelsMatched := false

	// matching results, labels AND level.
	isLabelsAndMatched := false

	if len(strategy.Clusterids) != 0 {
		for _, clusterid := range strategy.Clusterids {
			if clusterid == instance.Clusterid {
				isClusteridMatched = true
				break
			}
		}

		// not matched cluster level.
		if !isClusteridMatched {
			return false
		}
	}

	if len(strategy.Zoneids) != 0 {
		for _, zoneid := range strategy.Zoneids {
			if zoneid == instance.Zoneid {
				isZoneidMatched = true
				break
			}
		}

		// not matched zone level.
		if !isZoneidMatched {
			return false
		}
	}

	if len(strategy.Dcs) != 0 {
		for _, dc := range strategy.Dcs {
			if dc == instance.Dc {
				isDcMatched = true
				break
			}
		}

		// not matched dc level.
		if !isDcMatched {
			return false
		}
	}

	if len(strategy.IPs) != 0 {
		for _, ip := range strategy.IPs {
			if ip == instance.IP {
				isIPMatched = true
				break
			}
		}

		// not matched ip level.
		if !isIPMatched {
			return false
		}
	}

	if len(strategy.Labels) == 0 && len(strategy.LabelsAnd) == 0 {
		// empty labels.
		return true
	}

	// instance labels.
	insLabels := SidecarLabels{}
	if err := json.Unmarshal([]byte(instance.Labels), &insLabels); err != nil {
		// unknow instance labels.
		return false
	}

	// Labels is labels for "OR".
	if len(strategy.Labels) != 0 {
		for k, v := range strategy.Labels {
			// parse v from strategy like k=gt|1,2,3.
			op, vals, err := ParseLabelVals(v)
			if err != nil {
				return false
			}

			if IsLabelMatch(insLabels.Labels[k], op, vals) {
				isLabelsMatched = true
				// no need to match other labels in OR mode.
				break
			}
		}
	} else {
		// empty OR labels, mark flag true, and contine check LabelsAnd.
		isLabelsMatched = true
	}

	// LabelsAnd is labels for "AND".
	if len(strategy.LabelsAnd) != 0 {
		// AND mode, try to find and mark FALSE.
		isLabelsAndMatched = true

		for k, v := range strategy.LabelsAnd {
			// parse v from strategy like k=gt|1,2,3.
			op, vals, err := ParseLabelVals(v)
			if err != nil {
				return false
			}

			if !IsLabelMatch(insLabels.Labels[k], op, vals) {
				// no need to match other labels in AND mode.
				isLabelsAndMatched = false
				break
			}
		}
	} else {
		// empty ADN labels, mark flag true, and final result base on OR labels.
		isLabelsAndMatched = true
	}

	// check final OR && AND labels result.
	if isLabelsMatched && isLabelsAndMatched {
		// labels matched.
		return true
	}

	return false
}
