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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"testing"
)

func TestNewPolicyFactory(t *testing.T) {

}

func Test_policyFactory_GetPolicy(t *testing.T) {
	mongoOptions := &mongo.Options{
		Hosts:                 []string{"127.0.0.1:27017"},
		ConnectTimeoutSeconds: 3,
		Database:              "datamanager_test",
		Username:              "data",
		Password:              "test1234",
	}
	db, _ := mongo.NewDB(mongoOptions)
	storeCli := store.NewServer(db)
	fatory := NewPolicyFactory(storeCli)
	fatory.Init()
	policy := fatory.GetPolicy(common.ClusterType, common.DimensionDay)
	fmt.Println(policy)
	dataJob := &DataJob{}
	dataJob.SetPolicy(policy)
	ctx := context.Background()
	dataJob.DoPolicy(ctx)
}

func Test_policyFactory_initClusterMap(t *testing.T) {

}

func Test_policyFactory_initNamespaceMap(t *testing.T) {

}

func Test_policyFactory_initProjectMap(t *testing.T) {
}

func Test_policyFactory_initPublicMap(t *testing.T) {
}

func Test_policyFactory_initWorkloadMap(t *testing.T) {
}
