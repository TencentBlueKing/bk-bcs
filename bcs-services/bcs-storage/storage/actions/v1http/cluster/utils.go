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
 */

package cluster

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"
	"go-micro.dev/v4/broker"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

const (
	clusterIDTag    = "clusterID"
	clusterNameTag  = "clusterName"
	resourceTypeTag = "resourceType"

	tableCluster  = "Cluster"
	dataTag       = "data"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
)

var clusterFeatTags = []string{clusterIDTag}

// Use Mongodb for storage.
const dbConfig = "mongodb/dynamic"

func isExistResourceQueue(features map[string]string) bool {
	if len(features) == 0 {
		return false
	}

	resourceType, ok := features[resourceTypeTag]
	if !ok {
		return false
	}

	if _, ok := apiserver.GetAPIResource().GetMsgQueue().ResourceToQueue[resourceType]; !ok {
		return false
	}

	return true
}

func publishClusterInfoToQueue(data operator.M, featTags []string, event msgqueue.EventKind) error {
	var (
		err     error
		message = &broker.Message{
			Header: map[string]string{},
		}
	)

	startTime := time.Now()
	for _, feat := range featTags {
		if v, ok := data[feat].(string); ok {
			message.Header[feat] = v
		}
	}
	message.Header[string(msgqueue.EventType)] = string(event)
	message.Header[resourceTypeTag] = tableCluster

	exist := isExistResourceQueue(message.Header)
	if !exist {
		return nil
	}

	// NOCC:revive/early-return(设计如此:)
	// nolint
	if v, ok := data[dataTag]; ok {
		codec.EncJson(v, &message.Body)
	} else {
		blog.Infof("object[%v] not exist data", data[dataTag])
		return nil
	}

	err = apiserver.GetAPIResource().GetMsgQueue().MsgQueue.Publish(message)
	if err != nil {
		return err
	}

	if queueName, ok := message.Header[resourceTypeTag]; ok {
		metrics.ReportQueuePushMetrics(queueName, err, startTime)
	}

	return nil
}

func urlPath(oldURL string) string {
	return oldURL
}

func getReqData(req *restful.Request, features operator.M) (operator.M, error) {
	var tmp types.BcsStorageDynamicIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	data := lib.CopyMap(features)
	data[dataTag] = tmp.Data
	return data, nil
}

func getFeatures(req *restful.Request, resourceFeatList []string) operator.M {
	features := make(operator.M)
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}
	return features
}

func putResources(req *restful.Request, resourceFeatList []string) (operator.M, error) {
	features := getFeatures(req, resourceFeatList)
	data, err := getReqData(req, features)
	if err != nil {
		return nil, err
	}
	err = PutData(req.Request.Context(), data, features, resourceFeatList)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// putClusterInfo put cluster info
func putClusterInfo(req *restful.Request) error {

	data, err := putResources(req, clusterFeatTags)
	if err != nil {
		return err
	}
	PushCreateClusterInfoToQueue(data)
	return nil
}

func deleteResources(req *restful.Request) ([]operator.M, error) {
	condition := getClusterFeat(req)
	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}
	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
		// when resource to be deleted not found, do not return error
		IgnoreNotFound: true,
	}
	return DeleteBatchData(req.Request.Context(), getOption, rmOption)
}

// deleteClusterInfo delete cluster info
func deleteClusterInfo(req *restful.Request) error {

	data, err := deleteResources(req)
	if err != nil {
		return err
	}
	if len(data) == 1 {
		PushDeleteClusterInfoToQueue(data[0])
	}
	return nil
}

func getClusterFeat(req *restful.Request) *operator.Condition {
	return getFeat(req, clusterFeatTags)
}

func getFeat(req *restful.Request, resourceFeatList []string) *operator.Condition {
	features := make(operator.M, len(resourceFeatList))
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}
	return operator.NewLeafCondition(operator.Eq, features)
}
