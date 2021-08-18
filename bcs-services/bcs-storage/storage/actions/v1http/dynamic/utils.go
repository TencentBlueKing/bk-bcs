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

package dynamic

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-micro/v2/broker"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

const (
	urlPrefixK8S    = "/k8s"
	urlPrefixMesos  = "/mesos"
	clusterIDTag    = "clusterId"
	namespaceTag    = "namespace"
	resourceTypeTag = "resourceType"
	resourceNameTag = "resourceName"
	indexNameTag    = "indexName"

	tableTag      = resourceTypeTag
	dataTag       = "data"
	extraTag      = "extra"
	fieldTag      = "field"
	offsetTag     = "offset"
	limitTag      = "limit"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"

	applicationTypeName = "application"
	processTypeName     = "process"
	kindTag             = "data.kind"
)

var needTimeFormatList = []string{updateTimeTag, createTimeTag}
var nsFeatTags = []string{clusterIDTag, namespaceTag, resourceTypeTag, resourceNameTag}
var csFeatTags = []string{clusterIDTag, resourceTypeTag, resourceNameTag}
var nsListFeatTags = []string{clusterIDTag, namespaceTag, resourceTypeTag}
var csListFeatTags = []string{clusterIDTag, resourceTypeTag}
var customResourceFeatTags = []string{}
var customResourceIndexFeatTags = []string{resourceTypeTag, indexNameTag}
var indexKeys = []string{resourceNameTag, namespaceTag}

// Use Mongodb for storage.
const dbConfig = "mongodb/dynamic"

func getSelector(req *restful.Request) []string {
	return lib.GetQueryParamStringArray(req, fieldTag, ",")
}

func getTable(req *restful.Request) string {
	table := req.PathParameter(tableTag)
	// for mesos
	if table == processTypeName {
		table = applicationTypeName
	}
	return table
}

func getExtra(req *restful.Request) operator.M {
	raw := req.QueryParameter(extraTag)
	if raw == "" {
		return nil
	}
	extra := make(operator.M)
	lib.NewExtra(raw).Unmarshal(&extra)
	return extra
}

func getFeatures(req *restful.Request, resourceFeatList []string) operator.M {
	features := make(operator.M)
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}
	return features
}

func getCondition(req *restful.Request, resourceFeatList []string) *operator.Condition {
	features := getFeatures(req, resourceFeatList)
	extras := getExtra(req)
	features.Merge(extras)
	featuresExcept := make(operator.M)
	for key := range features {
		// For historical reasons, mesos process is stored with application in one table(same clusters).
		// And process's construction is almost the same with application, except with field 'data.kind'.
		// If 'data.kind'='process', then this object is a process stored in application-table,
		// If 'data.kind'='application' or '', then this object is an application stored in application-table.
		//
		// For this case, we should:
		// 1. Change the key 'resourceType' from 'process' to 'application' when the caller ask for 'process'.
		// 2. Besides, getFeat() should add an extra condition that
		//    mentions the 'data.kind' to distinguish 'process' and 'application'.
		// 3. Make sure the table is application-table whether the type is 'application' or 'process'. (with getTable())
		if key == resourceTypeTag {
			switch features[key] {
			case applicationTypeName:
				featuresExcept[kindTag] = processTypeName
			case processTypeName:
				features[key] = applicationTypeName
				features[kindTag] = processTypeName
			}
		}
	}
	condition := operator.NewLeafCondition(operator.Eq, features)
	if len(featuresExcept) == 0 {
		notCondition := operator.NewLeafCondition(operator.Ne, featuresExcept)
		condition = operator.NewBranchCondition(operator.And,
			condition, notCondition)
	}
	customCondition := lib.GetCustomCondition(req)
	if customCondition != nil {
		condition = operator.NewBranchCondition(operator.And, condition, customCondition)
	}
	by, _ := json.Marshal(condition)
	blog.Infof("%s", string(by))
	return condition
}

func getNamespaceResources(req *restful.Request) ([]operator.M, error) {
	return getResources(req, nsFeatTags)
}

func getClusterResources(req *restful.Request) ([]operator.M, error) {
	return getResources(req, csFeatTags)
}

func listNamespaceResources(req *restful.Request) ([]operator.M, error) {
	return getResources(req, nsListFeatTags)
}

func listClusterResources(req *restful.Request) ([]operator.M, error) {
	return getResources(req, csListFeatTags)
}

func getCustomResources(req *restful.Request) ([]operator.M, operator.M, error) {
	return getResourcesWithPageInfo(req, customResourceFeatTags)
}

func getStoreOption(req *restful.Request, resourceFeatList []string) (*lib.StoreGetOption, error) {
	condition := getCondition(req, resourceFeatList)
	offset, err := lib.GetQueryParamInt64(req, offsetTag, 0)
	if err != nil {
		return nil, err
	}
	limit, err := lib.GetQueryParamInt64(req, limitTag, 0)
	if err != nil {
		return nil, err
	}
	return &lib.StoreGetOption{
		Fields: getSelector(req),
		Cond:   condition,
		Offset: offset,
		Limit:  limit,
	}, nil
}

func getResources(req *restful.Request, resourceFeatList []string) ([]operator.M, error) {
	getOption, err := getStoreOption(req, resourceFeatList)
	if err != nil {
		return nil, err
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	mList, err := store.Get(req.Request.Context(), getTable(req), getOption)
	if err != nil {
		return nil, err
	}
	lib.FormatTime(mList, needTimeFormatList)
	return mList, err
}

func getResourcesWithPageInfo(req *restful.Request, resourceFeatList []string) (data []operator.M, extra operator.M, err error) {
	getOption, err := getStoreOption(req, resourceFeatList)
	if err != nil {
		return nil, nil, err
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	count, err := store.Count(req.Request.Context(), getTable(req), getOption)
	if err != nil {
		return nil, nil, err
	}
	mList, err := store.Get(req.Request.Context(), getTable(req), getOption)
	if err != nil {
		return nil, nil, err
	}
	lib.FormatTime(mList, needTimeFormatList)

	extra = operator.M{
		"total":    count,
		"pageSize": getOption.Limit,
		"offset":   getOption.Offset,
	}
	return mList, extra, err
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

func putNamespaceResources(req *restful.Request) error {
	data, err := putResources(req, nsFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		err = publishDynamicResourceToQueue(data, nsFeatTags, msgqueue.EventTypeUpdate)
		if err != nil {
			blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putNamespaceResources", err)
		}
	}

	return nil
}

func putClusterResources(req *restful.Request) error {
	data, err := putResources(req, csFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		err = publishDynamicResourceToQueue(data, csFeatTags, msgqueue.EventTypeUpdate)
		if err != nil {
			blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putClusterResources", err)
		}
	}

	return nil
}

func putCustomResources(req *restful.Request) error {
	// Obtain table index
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	index, err := store.GetIndex(req.Request.Context(), getTable(req))
	if err != nil {
		return err
	}

	// resolve data waiting to be put
	dataRaw := make(operator.M)
	if err = codec.DecJsonReader(req.Request.Body, &dataRaw); err != nil {
		return err
	}

	putOption := &lib.StorePutOption{
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	var uniIdx drivers.Index
	if index != nil {
		uniIdx = *index
	}
	conds := make([]*operator.Condition, 0)
	if len(uniIdx.Key) != 0 {
		for _, bsonElem := range uniIdx.Key {
			key := bsonElem.Key
			conds = append(conds, operator.NewLeafCondition(operator.Eq, operator.M{key: dataRaw[key]}))
		}
	}
	if len(conds) != 0 {
		putOption.Cond = operator.NewBranchCondition(operator.And, conds...)
	}
	return store.Put(req.Request.Context(), getTable(req), dataRaw, putOption)
}

func putResources(req *restful.Request, resourceFeatList []string) (operator.M, error) {
	features := getFeatures(req, resourceFeatList)
	extras := getExtra(req)
	features.Merge(extras)
	data, err := getReqData(req, features)
	if err != nil {
		return nil, err
	}
	putOption := &lib.StorePutOption{
		UniqueKey:     resourceFeatList,
		Cond:          operator.NewLeafCondition(operator.Eq, features),
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	err = store.Put(req.Request.Context(), getTable(req), data, putOption)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func deleteNamespaceResources(req *restful.Request) error {
	mList, err := deleteResources(req, nsFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		go func(mList []operator.M, featTags []string) {
			for _, data := range mList {
				err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
				if err != nil {
					blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteNamespaceResources", err)
				}
			}
		}(mList, nsFeatTags)
	}

	return nil
}

func deleteClusterResources(req *restful.Request) error {
	mList, err := deleteResources(req, csFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		go func(mList []operator.M, featTags []string) {
			for _, data := range mList {
				err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
				if err != nil {
					blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterResources", err)
				}
			}
		}(mList, csFeatTags)
	}

	return nil
}

func deleteCustomResources(req *restful.Request) error {
	_, err := deleteResources(req, []string{})
	if err != nil {
		return err
	}
	return nil
}

func deleteResources(req *restful.Request, resourceFeatList []string) ([]operator.M, error) {
	condition := getCondition(req, resourceFeatList)

	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}

	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
		// when resource to be deleted not found, do not return error
		IgnoreNotFound: true,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	mList, err := store.Get(req.Request.Context(), getTable(req), getOption)
	if err != nil {
		return nil, err
	}
	lib.FormatTime(mList, needTimeFormatList)

	err = store.Remove(req.Request.Context(), getTable(req), rmOption)
	if err != nil {
		return nil, err
	}

	return mList, nil
}

func getTimeCondition(req *restful.Request) *operator.Condition {
	var data types.BcsStorageDynamicBatchDeleteIf
	if err := codec.DecJsonReader(req.Request.Body, &data); err != nil {
		return operator.EmptyCondition
	}

	condList := make([]*operator.Condition, 0)
	if data.UpdateTimeBegin > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Gt, operator.M{
			updateTimeTag: time.Unix(data.UpdateTimeBegin, 0)}))
	}
	if data.UpdateTimeEnd > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Lt, operator.M{
			updateTimeTag: time.Unix(data.UpdateTimeEnd, 0)}))
	}
	if len(condList) == 0 {
		return operator.EmptyCondition
	}
	return operator.NewBranchCondition(operator.And, condList...)
}

func deleteBatchNamespaceResource(req *restful.Request) error {
	mList, err := deleteBatchResources(req, nsListFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		go func(mList []operator.M, featTags []string) {
			for _, data := range mList {
				err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
				if err != nil {
					blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteBatchNamespaceResource", err)
				}
			}
		}(mList, nsListFeatTags)
	}

	return nil
}

func deleteClusterNamespaceResource(req *restful.Request) error {
	mList, err := deleteBatchResources(req, csListFeatTags)
	if err != nil {
		return err
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		go func(mList []operator.M, featTags []string) {
			for _, data := range mList {
				err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
				if err != nil {
					blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterNamespaceResource", err)
				}
			}
		}(mList, csListFeatTags)
	}

	return nil
}

func deleteBatchResources(req *restful.Request, resourceFeatList []string) ([]operator.M, error) {
	featCondition := getCondition(req, resourceFeatList)
	timeCondition := getTimeCondition(req)
	condition := operator.NewBranchCondition(operator.And, featCondition, timeCondition)

	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}
	rmOption := &lib.StoreRemoveOption{
		Cond:           condition,
		IgnoreNotFound: true,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	mList, err := store.Get(req.Request.Context(), getTable(req), getOption)
	if err != nil {
		return nil, err
	}
	lib.FormatTime(mList, needTimeFormatList)

	err = store.Remove(req.Request.Context(), getTable(req), rmOption)
	if err != nil {
		return nil, err
	}

	return mList, nil
}

func createCustomResourcesIndex(req *restful.Request) error {
	index := drivers.Index{
		Unique: true,
		Name:   req.PathParameter(indexNameTag),
	}
	keys := bson.D{}
	by, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return err
	}
	err = bson.UnmarshalExtJSON(by, true, &keys)
	if err != nil {
		return err
	}
	index.Key = keys

	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	return store.CreateIndex(req.Request.Context(), getTable(req), index)
}

func deleteCustomResourcesIndex(req *restful.Request) error {
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	return store.DeleteIndex(req.Request.Context(), getTable(req), req.PathParameter(indexNameTag))
}

func urlPathK8S(oldURL string) string {
	return urlPrefixK8S + oldURL
}

func urlPathMesos(oldURL string) string {
	return urlPrefixMesos + oldURL
}

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

func publishDynamicResourceToQueue(data operator.M, featTags []string, event msgqueue.EventKind) error {
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

	exist := isExistResourceQueue(message.Header)
	if !exist {
		return nil
	}

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
