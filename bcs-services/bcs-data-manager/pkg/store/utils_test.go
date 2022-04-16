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

package store

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_distinctSlice(t *testing.T) {
	type args struct {
		key   string
		slice *[]map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := distinctSlice(tt.args.key, tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("distinctSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensure(t *testing.T) {
	type args struct {
		ctx       context.Context
		db        drivers.DB
		tableName string
		indexes   []drivers.Index
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ensure(tt.args.ctx, tt.args.db, tt.args.tableName, tt.args.indexes); (err != nil) != tt.wantErr {
				t.Errorf("ensure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ensureTable(t *testing.T) {
	type args struct {
		ctx    context.Context
		public *Public
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ensureTable(tt.args.ctx, tt.args.public); (err != nil) != tt.wantErr {
				t.Errorf("ensureTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getPublicData(t *testing.T) {
	type args struct {
		ctx  context.Context
		db   drivers.DB
		cond *operator.Condition
	}
	tests := []struct {
		name string
		args args
		want *common.PublicData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPublicData(tt.args.ctx, tt.args.db, tt.args.cond); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPublicData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStartTime(t *testing.T) {
	fmt.Println(time.Now())
	fmt.Println(getStartTime(common.DimensionHour))
	fmt.Println(getStartTime(common.DimensionDay))
	fmt.Println(getStartTime(common.DimensionMinute))
	fmt.Println(primitive.NewDateTimeFromTime(getStartTime(common.DimensionDay)))
	fmt.Println(primitive.NewDateTimeFromTime(getStartTime(common.DimensionHour)))
	fmt.Println(primitive.NewDateTimeFromTime(getStartTime(common.DimensionMinute)))
	fmt.Println(primitive.NewDateTimeFromTime(getStartTime(common.DimensionDay)).Time().String())
}
