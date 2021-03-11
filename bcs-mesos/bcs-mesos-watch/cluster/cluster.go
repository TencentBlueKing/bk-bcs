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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"

	"golang.org/x/net/context"
)

//DataExister checker interface for data exist
type DataExister interface {
	IsExist(data interface{}) bool
}

//Reporter for report data
type Reporter interface {
	ReportData(data *types.BcsSyncData) error
	GetClusterID() string
}

//Cluster is interface for reading Cluster info
type Cluster interface {
	Run(cxt context.Context)  //start cluster
	Sync(tp string) error     //ready to sync data, type like services, pods and etc.
	Stop()                    //stop cluster
	GetClusterStatus() string //get curr status
}

//EventHandler hold event interface for All Watch
type EventHandler interface {
	AddEvent(obj interface{})
	DeleteEvent(obj interface{})
	UpdateEvent(old, cur interface{})
}

//ReportFunc report function for handle detail report data
type ReportFunc func(data *types.BcsSyncData) error
