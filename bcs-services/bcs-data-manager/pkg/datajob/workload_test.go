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
	"net/http"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
)

func TestNewWorkloadDayPolicy(t *testing.T) {

}

func TestNewWorkloadHourPolicy(t *testing.T) {

}

func TestNewWorkloadMinutePolicy(t *testing.T) {

}

func TestWorkloadDayPolicy(t *testing.T) {

}

func TestWorkloadHourPolicy(t *testing.T) {

}

func TestWorkloadMinutePolicy(t *testing.T) {
	opts := &common.JobCommonOpts{
		ObjectType:   common.WorkloadType,
		ClusterID:    "BCS-K8S-15091",
		ClusterType:  common.Kubernetes,
		Dimension:    common.DimensionMinute,
		Namespace:    "bcs-system",
		WorkloadType: common.DeploymentType,
		Name:         "bcs-k8s-watch",
		CurrentTime:  common.FormatTime(time.Now(), common.DimensionMinute),
	}
	getter := &metric.MetricGetter{}
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
	db := store.NewServer(mongoDB)
	policy := NewWorkloadMinutePolicy(getter, db)
	ctx := context.TODO()
	opt := bcsmonitor.BcsMonitorClientOpt{
		Schema:   "http",
		Endpoint: "",
		UserName: "",
		Password: "",
	}
	requester := bcsmonitor.NewRequester()
	monitorCli := bcsmonitor.NewBcsMonitorClient(opt, requester)
	monitorCli.SetCompleteEndpoint()
	header := http.Header{}
	monitorCli.SetDefaultHeader(header)

	clients := &Clients{
		monitorClient: monitorCli,
	}
	policy.ImplementPolicy(ctx, opts, clients)
}
