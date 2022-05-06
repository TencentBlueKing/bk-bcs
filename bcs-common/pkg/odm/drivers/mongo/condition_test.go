/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// TestMongoQueryCondtionCombine test mongo query condition combine
func TestMongoQueryCondtionCombine(t *testing.T) {
	testCases := []struct {
		title        string
		getCondition func() *operator.Condition
		expectedData bson.M
	}{
		{
			"test Eq",
			func() *operator.Condition {
				eqCondition := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				return eqCondition
			},
			bson.M{
				"key1": "value1",
			},
		},
		{
			"test Ne",
			func() *operator.Condition {
				eqCondition := operator.NewLeafCondition(
					operator.Ne,
					operator.M{"key1": "value1"})
				return eqCondition
			},
			bson.M{
				"key1": bson.M{"$ne": "value1"},
			},
		},
		{
			"test Or",
			func() *operator.Condition {
				condition1 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				condition2 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key2": "value2"})
				orCondition := operator.NewBranchCondition(
					operator.Or,
					condition1, condition2)
				return orCondition
			},
			bson.M{
				"$or": bson.A{
					bson.M{
						"key1": "value1",
					},
					bson.M{
						"key2": "value2",
					},
				},
			},
		},
		{
			"test Nor",
			func() *operator.Condition {
				condition1 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				condition2 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key2": "value2"})
				orCondition := operator.NewBranchCondition(
					operator.Nor,
					condition1, condition2)
				return orCondition
			},
			bson.M{
				"$nor": bson.A{
					bson.M{
						"key1": "value1",
					},
					bson.M{
						"key2": "value2",
					},
				},
			},
		},
		{
			"test not",
			func() *operator.Condition {
				condition1 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				notCondition := operator.NewBranchCondition(
					operator.Not,
					condition1)
				return notCondition
			},
			bson.M{
				"$not": bson.M{
					"key1": "value1",
				},
			},
		},
		{
			"test match",
			func() *operator.Condition {
				condition1 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				condition2 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key2": "value2"})
				orCondition := operator.NewBranchCondition(
					operator.Or,
					condition1, condition2)
				notCondition := operator.NewBranchCondition(
					operator.Mat,
					orCondition)
				return notCondition
			},
			bson.M{
				"$match": bson.M{
					"$or": bson.A{
						bson.M{
							"key1": "value1",
						},
						bson.M{
							"key2": "value2",
						},
					},
				},
			},
		},
		{
			"test And",
			func() *operator.Condition {
				eqCondition1 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key1": "value1"})
				eqCondition2 := operator.NewLeafCondition(
					operator.Eq,
					operator.M{"key2": "value2"})
				eqCondtion12 := operator.NewBranchCondition(
					operator.And,
					eqCondition1, eqCondition2)

				eqCondition3 := operator.NewLeafCondition(
					operator.Lt,
					operator.M{"key3": 5})
				eqCondition4 := operator.NewLeafCondition(
					operator.Lte,
					operator.M{"key4": 10})
				eqCondition5 := operator.NewLeafCondition(
					operator.Gt,
					operator.M{"key5": 10})
				eqCondition6 := operator.NewLeafCondition(
					operator.Gte,
					operator.M{"key6": 10})
				eqCondition7 := operator.NewLeafCondition(
					operator.In,
					operator.M{"key7": 10})
				eqCondition8 := operator.NewLeafCondition(
					operator.Nin,
					operator.M{"key8": 10})
				eqCondtion345678 := operator.NewBranchCondition(
					operator.And,
					eqCondition3, eqCondition4, eqCondition5, eqCondition6, eqCondition7, eqCondition8)
				eqCondition := operator.NewBranchCondition(
					operator.And,
					eqCondtion12, eqCondtion345678)
				return eqCondition
			},
			bson.M{
				"$and": bson.A{
					bson.M{
						"$and": bson.A{
							bson.M{
								"key1": "value1",
							},
							bson.M{
								"key2": "value2",
							},
						},
					},
					bson.M{
						"$and": bson.A{
							bson.M{
								"key3": bson.M{"$lt": 5},
							},
							bson.M{
								"key4": bson.M{"$lte": 10},
							},
							bson.M{
								"key5": bson.M{"$gt": 10},
							},
							bson.M{
								"key6": bson.M{"$gte": 10},
							},
							bson.M{
								"key7": bson.M{"$in": 10},
							},
							bson.M{
								"key8": bson.M{"$nin": 10},
							},
						},
					},
				},
			},
		},
		{
			"test Con when value is string",
			func() *operator.Condition {
				conCondition := operator.NewLeafCondition(
					operator.Con,
					operator.M{"key1": "value1"})
				return conCondition
			},
			bson.M{
				"key1": primitive.Regex{Pattern: ".*value1.*"},
			},
		},
		{
			"test Con when value is primitive.Regex",
			func() *operator.Condition {
				conCondition := operator.NewLeafCondition(
					operator.Con,
					operator.M{"key1": primitive.Regex{Pattern: ".*value1.*", Options: "i"}})
				return conCondition
			},
			bson.M{
				"key1": primitive.Regex{Pattern: ".*value1.*", Options: "i"},
			},
		},
	}

	for index, tCase := range testCases {
		t.Logf("test %d - title: %s", index, tCase.title)
		tmpCondition := tCase.getCondition()
		tmpData := tmpCondition.Combine(leafNodeProcessor, branchNodeProcessor)
		expectBytes, err := bson.MarshalExtJSON(tCase.expectedData, true, true)
		if err != nil {
			t.Errorf("marshal expected bson %+v to json failed, err %s", tCase.expectedData, err.Error())
		}
		tmpBytes, err := bson.MarshalExtJSON(tmpData, true, true)
		if err != nil {
			t.Errorf("marshal tmp bson %+v to json failed, err %s", tmpData, err.Error())
		}
		if string(expectBytes) != string(tmpBytes) {
			t.Errorf("expect %s, but get %s", string(expectBytes), string(tmpBytes))
		}
	}

}
