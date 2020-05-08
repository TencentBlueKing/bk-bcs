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
 *
 */

package types

type ConstraintValue_Scalar struct {
	Value float64 `json:"value"`
}

type ConstraintValue_Range struct {
	Begin uint64 `json:"begin"`
	End   uint64 `json:"end"`
}

type ConstraintValue_Text struct {
	Value string `json:"value"`
}

type ConstraintValue_Set struct {
	Item []string `json:"item"`
}

type ConstraintValue_Type int32

const (
	ConstValueType_UNKNOW ConstraintValue_Type = 0
	ConstValueType_Scalar ConstraintValue_Type = 1
	ConstValueType_Range  ConstraintValue_Type = 2
	ConstValueType_Text   ConstraintValue_Type = 3
	ConstValueType_Set    ConstraintValue_Type = 4
)

const (
	Constraint_Type_UNIQUE     = "UNIQUE"
	Constraint_Type_CLUSTER    = "CLUSTER"
	Constraint_Type_GROUP_BY   = "GROUPBY"
	Constraint_Type_MAX_PER    = "MAXPER"
	Constraint_Type_LIKE       = "LIKE"
	Constraint_Type_UNLIKE     = "UNLIKE"
	Constraint_Type_EXCLUDE    = "EXCLUDE"
	Constraint_Type_GREATER    = "GREATER"
	Constraint_Type_TOLERATION = "TOLERATION"
)

type ConstraintData struct {
	Name    string                   `json:"name"`
	Operate string                   `json:"operate"`
	Type    ConstraintValue_Type     `json:"type"`
	Scalar  *ConstraintValue_Scalar  `json:"scalar"`
	Ranges  []*ConstraintValue_Range `json:"ranges"`
	Text    *ConstraintValue_Text    `json:"text"`
	Set     *ConstraintValue_Set     `json:"set"`
}

type ConstraintDataItem struct {
	UnionData []*ConstraintData `json:"unionData"`
}

type Constraint struct {
	IntersectionItem []*ConstraintDataItem `json:"intersectionItem,omitempty"`
	NodeSelector     map[string]string     `json:"nodeSelector,omitempty"`
}
