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
	"testing"

	"github.com/stretchr/testify/assert"

	pbcommon "bk-bscp/internal/protocol/common"
)

// TestContentIndexMatch test content index match.
func TestContentIndexMatch(t *testing.T) {
	assert := assert.New(t)

	var labelsOr []map[string]string
	var labelsAnd []map[string]string

	labelsOrItem := make(map[string]string)
	labelsOrItem["k1"] = "lt|v1-1,v1-2"
	labelsOrItem["k2"] = "ge|v2-1"

	labelsAndItem := make(map[string]string)
	labelsAndItem["k3"] = "le|v3-1,v3-2"
	labelsAndItem["k4"] = "ne|v4-1,v4-2"

	assert.Equal(nil, ValidateLabels(labelsOrItem))
	assert.Equal(nil, ValidateLabels(labelsAndItem))

	labelsOr = append(labelsOr, labelsOrItem)
	labelsAnd = append(labelsAnd, labelsAndItem)

	labels := make(map[string]string)
	newContentIndex := &ContentIndex{LabelsOr: labelsOr, LabelsAnd: labelsAnd}

	// empty labels.
	assert.Equal(false, newContentIndex.MatchLabels(labels))

	// empty index.
	labels = map[string]string{"k": "v"}
	newContentIndex = &ContentIndex{}
	assert.Equal(false, newContentIndex.MatchLabels(labels))

	newContentIndex = &ContentIndex{LabelsOr: labelsOr, LabelsAnd: labelsAnd}

	// match "k1:lt|v1-1,v1-2"
	labels = map[string]string{"k1": "v1-0", "k2": "v2-0", "k3": "v3-3", "k4": "v4-1"}
	assert.Equal(true, newContentIndex.MatchLabels(labels))

	// match "k2:ge|v2-1"
	labels = map[string]string{"k1": "v1-1", "k2": "v2-1", "k3": "v3-3", "k4": "v4-1"}
	assert.Equal(true, newContentIndex.MatchLabels(labels))

	// only match "k3:le|v3-1,v3-2"
	labels = map[string]string{"k1": "v1-1", "k2": "v2-0", "k3": "v3-1", "k4": "v4-1"}
	assert.Equal(false, newContentIndex.MatchLabels(labels))

	// only match "k4:ne|v4-1,v4-2"
	labels = map[string]string{"k1": "v1-1", "k2": "v2-0", "k3": "v3-3", "k4": "v4-3"}
	assert.Equal(false, newContentIndex.MatchLabels(labels))

	// match "k3:le|v3-1,v3-2" and "k4:ne|v4-1,v4-2"
	labels = map[string]string{"k1": "v1-1", "k2": "v2-0", "k3": "v3-1", "k4": "v4-3"}
	assert.Equal(true, newContentIndex.MatchLabels(labels))
}

// TestStrategyMatch test strategy match.
func TestStrategyMatch(t *testing.T) {
	assert := assert.New(t)

	handler := NewHandler(nil)
	match := handler.Matcher()

	var labelsOr []map[string]string
	var labelsAnd []map[string]string

	labelsOrItem := make(map[string]string)
	labelsOrItem["k1"] = "lt|v1-1,v1-2"
	labelsOrItem["k2"] = "ge|v2-1"

	labelsAndItem := make(map[string]string)
	labelsAndItem["k3"] = "le|v3-1,v3-2"
	labelsAndItem["k4"] = "ne|v4-1,v4-2"

	assert.Equal(nil, ValidateLabels(labelsOrItem))
	assert.Equal(nil, ValidateLabels(labelsAndItem))

	labelsOr = append(labelsOr, labelsOrItem)
	labelsAnd = append(labelsAnd, labelsAndItem)
	newStrategy := &Strategy{LabelsOr: labelsOr, LabelsAnd: labelsAnd}

	ins := &pbcommon.AppInstance{
		AppId:   "appid",
		CloudId: "0",
		Ip:      "127.0.0.1",
		Path:    "etc",
	}

	// match "k1:lt|v1-1,v1-2"
	ins.Labels = "{\"Labels\":{\"k1\":\"v1-0\", \"k2\":\"v2-0\", \"k3\":\"v3-3\", \"k4\":\"v4-1\"}}"
	assert.Equal(true, match(newStrategy, ins))

	// match "k2:ge|v2-1"
	ins.Labels = "{\"Labels\":{\"k1\":\"v1-1\", \"k2\":\"v2-1\", \"k3\":\"v3-3\", \"k4\":\"v4-1\"}}"
	assert.Equal(true, match(newStrategy, ins))

	// only match "k3:le|v3-1,v3-2"
	ins.Labels = "{\"Labels\":{\"k1\":\"v1-1\", \"k2\":\"v2-0\", \"k3\":\"v3-1\", \"k4\":\"v4-1\"}}"
	assert.Equal(false, match(newStrategy, ins))

	// only match "k4:ne|v4-1,v4-2"
	ins.Labels = "{\"Labels\":{\"k1\":\"v1-1\", \"k2\":\"v2-0\", \"k3\":\"v3-3\", \"k4\":\"v4-3\"}}"
	assert.Equal(false, match(newStrategy, ins))

	// match "k3:le|v3-1,v3-2" and "k4:ne|v4-1,v4-2"
	ins.Labels = "{\"Labels\":{\"k1\":\"v1-1\", \"k2\":\"v2-0\", \"k3\":\"v3-1\", \"k4\":\"v4-3\"}}"
	assert.Equal(true, match(newStrategy, ins))
}
