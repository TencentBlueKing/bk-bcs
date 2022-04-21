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

package datajob

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
)

func TestNewProjectDayPolicy(t *testing.T) {
	db := newTestMongo()
	projectPolicy := NewProjectDayPolicy(&metric.MetricGetter{}, db)
	opts := &common.JobCommonOpts{
		ObjectType:  common.ProjectType,
		ProjectID:   "test",
		Dimension:   common.DimensionDay,
		CurrentTime: time.Time{},
	}
	ctx := context.Background()
	projectPolicy.ImplementPolicy(ctx, opts, nil)
}

func TestProjectDayPolicy_CalculateCpu(t *testing.T) {
	type fields struct {
		MetricGetter metric.Server
		store        store.Server
	}
	type args struct {
		clusters []*common.ClusterData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
		want1  float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectDayPolicy{
				MetricGetter: tt.fields.MetricGetter,
				store:        tt.fields.store,
			}
			got, got1 := p.CalculateCpu(tt.args.clusters)
			if got != tt.want {
				t.Errorf("CalculateCpu() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CalculateCpu() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_ProjectDayPolicy(t *testing.T) {
	db := newTestMongo()
	projectPolicy := NewProjectDayPolicy(&metric.MetricGetter{}, db)
	opts := &common.JobCommonOpts{
		ObjectType:  common.ProjectType,
		ProjectID:   "test",
		Dimension:   common.DimensionDay,
		CurrentTime: time.Now(),
	}
	ctx := context.Background()
	projectPolicy.ImplementPolicy(ctx, opts, nil)
}

func Test_calculateMemory(t *testing.T) {
	type fields struct {
		MetricGetter metric.Server
		store        store.Server
	}
	type args struct {
		clusters []*common.ClusterData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectDayPolicy{
				MetricGetter: tt.fields.MetricGetter,
				store:        tt.fields.store,
			}
			got, got1 := p.calculateMemory(tt.args.clusters)
			if got != tt.want {
				t.Errorf("calculateMemory() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("calculateMemory() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_calculateProjectNodeCount(t *testing.T) {
	type fields struct {
		MetricGetter metric.Server
		store        store.Server
	}
	type args struct {
		clusters []*common.ClusterData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProjectDayPolicy{
				MetricGetter: tt.fields.MetricGetter,
				store:        tt.fields.store,
			}
			got, got1 := p.calculateProjectNodeCount(tt.args.clusters)
			if got != tt.want {
				t.Errorf("calculateProjectNodeCount() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("calculateProjectNodeCount() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func newTestMongo() store.Server {
	mongoOptions := &mongo.Options{
		Hosts:                 []string{"127.0.0.1:27017"},
		ConnectTimeoutSeconds: 3,
		Database:              "datamanager_test",
		Username:              "data",
		Password:              "test1234",
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		fmt.Println(err)
	}
	err = mongoDB.Ping()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("init mongo db successfully")
	return store.NewServer(mongoDB)
}
