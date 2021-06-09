/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	comm "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"

	"github.com/emicklei/go-restful"
)

// list transactions
func (r *Router) listTransaction(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	objectKind := req.QueryParameter("objKind")
	objectName := req.QueryParameter("objName")
	namespace := req.PathParameter("namespace")
	blog.Infof("request to list transaction with args objKind=%s, objName=%s, namespace=%s",
		objectKind, objectName, namespace)

	transactionList, err := r.backend.ListTransaction(namespace)
	if err != nil {
		blog.Errorf("request list transaction failed, err %s", err.Error())
		data := createResponseDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}
	retList := make([]*types.Transaction, 0)
	for _, trans := range transactionList {
		if len(objectKind) != 0 && trans.ObjectKind != objectKind {
			continue
		}
		if len(objectName) != 0 && trans.ObjectName != objectName {
			continue
		}
		retList = append(retList, trans)
	}
	data := createResponseData(nil, "success", retList)
	resp.Write([]byte(data))
	blog.Info("query transaction finish")
	return
}

// delete transactions
func (r *Router) deleteTransaction(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Infof("request delete transaction %s/%s", ns, name)

	var data string
	if err := r.backend.DeleteTransaction(ns, name); err != nil {
		blog.Errorf("failed to delete transaction %s/%s, err %s", ns, name, err.Error())
		data = createResponseDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}
	data = createResponseData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Infof("request delete transaction %s/%s successfully", ns, name)
	return
}
