/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFloatData(t *testing.T) {

}

func TestGetInt64Data(t *testing.T) {

}

func TestGetIntData(t *testing.T) {

}

func Test_generateMesosPodCondition(t *testing.T) {

}

func Test_generatePodCondition(t *testing.T) {

}

func Test_getDimensionPromql(t *testing.T) {

}

func Test_getIncreasingInterDifference(t *testing.T) {
	testNums := []int{0, 1, 2, 1, 2, 3, 4, 0}
	result := getIncreasingIntervalDifference(testNums)
	assert.Equal(t, 5, result)
}

func Test_fillMetrics(t *testing.T) {
	initial := make([][]interface{}, 0)
	initial = append(initial, []interface{}{float64(16593460), "1"}, []interface{}{float64(16593490), "1"},
		[]interface{}{float64(16593590), "1"})
	result := fillMetrics(float64(16593460), initial, 30)
	assert.Equal(t, 4, len(result))
}
