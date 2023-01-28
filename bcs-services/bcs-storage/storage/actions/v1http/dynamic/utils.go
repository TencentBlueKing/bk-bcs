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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	msgqueue "github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueuev4"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/emicklei/go-restful"
	"go-micro.dev/v4/broker"
	"go.mongodb.org/mongo-driver/bson"
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
	eventTimeTag  = "eventTime"

	applicationTypeName = "application"
	processTypeName     = "process"
	kindTag             = "data.kind"

	eventResourceType = "Event"

	databaseFieldNameForDeletionFlag    = "_isBcsObjectDeleted"
	getDatabaseFieldNameForDeletionFlag = "getDeletionFlag"
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

const eventDBConfig = "mongodb/event"

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
		/*
			中文翻译：
			由于历史原因，mesos 进程与应用程序存储在同一个表中（相同的集群）。除了字段'data.kind'字段之外，进程的构造与应用程序几乎相同。
			如果'data.kind'='process'，那么这个对象就是一个存储在“application-table”中的进程，
			如果'data.kind'='application' 或者 ''(为空)，那么这个对象就是一个存储在“application-table”中的应用。

			对于这种情况，我们应该：
			1. 当调用者请求“进程”时，将键“resourceType”从“进程”更改为“应用程序”。
			2. 此外，getFeat() 应该添加一个额外的条件，提到“data.kind”，以区分“进程”和“应用程序”。
			3.确保表是application-table，无论类型是'application'还是'process'。 （都使用 getTable()）
		*/
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
		condition = operator.NewBranchCondition(operator.And, condition, notCondition)
	}
	customCondition := lib.GetCustomCondition(req)
	if customCondition != nil {
		condition = operator.NewBranchCondition(operator.And, condition, customCondition)
	}
	by, _ := json.Marshal(condition)
	blog.Infof("condition: %s", string(by))
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
	// option
	opt, err := getStoreOption(req, resourceFeatList)
	if err != nil {
		return nil, err
	}
	// 表名
	resourceType := getTable(req)

	mList, err := GetData(req.Request.Context(), resourceType, opt)
	if err != nil {
		return nil, err
	}

	lib.FormatTime(mList, needTimeFormatList)
	return mList, err
}

func getResourcesWithPageInfo(req *restful.Request, resourceFeatList []string) (data []operator.M, extra operator.M, err error) {
	opt, err := getStoreOption(req, resourceFeatList)
	if err != nil {
		return nil, nil, err
	}
	resourceType := getTable(req)

	return GetDataWithPageInfo(req.Request.Context(), resourceType, opt)
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
	PushCreateResourcesToQueue(data)
	return nil
}

func putClusterResources(req *restful.Request) error {
	data, err := putResources(req, csFeatTags)
	if err != nil {
		return err
	}
	PushCreateClusterToQueue(data)
	return nil
}

func putCustomResources(req *restful.Request) error { // resolve data waiting to be put
	dataRaw := make(operator.M)
	if err := codec.DecJsonReader(req.Request.Body, &dataRaw); err != nil {
		return err
	}
	//  表名
	resourceType := getTable(req)
	// option
	opt := &lib.StorePutOption{
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	return PutCustomResourceToDB(req.Request.Context(), resourceType, dataRaw, opt)
}

func putResources(req *restful.Request, resourceFeatList []string) (operator.M, error) {
	// 参数
	features := getFeatures(req, resourceFeatList)
	extras := getExtra(req)
	features.Merge(extras)
	data, err := getReqData(req, features)
	if err != nil {
		return nil, err
	}
	// 表名
	resourceType := getTable(req)

	if err = PutData(req.Request.Context(), data, features, resourceFeatList, resourceType); err != nil {
		return nil, err
	}
	return data, nil
}

func deleteNamespaceResources(req *restful.Request) error {
	mList, err := deleteResources(req, nsFeatTags)
	if err != nil {
		return err
	}
	PushDeleteResourcesToQueue(mList)
	return nil
}

func deleteClusterResources(req *restful.Request) error {
	mList, err := deleteResources(req, csFeatTags)
	if err != nil {
		return err
	}
	PushDeleteClusterToQueue(mList)
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
	// 条件
	condition := getCondition(req, resourceFeatList)
	// 表名
	resourceType := getTable(req)
	// get option
	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}
	// rm option
	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
		// when resource to be deleted not found, do not return error
		IgnoreNotFound: true,
	}
	return DeleteBatchData(req.Request.Context(), resourceType, getOption, rmOption)
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
	PushDeleteBatchResourceToQueue(mList)
	return nil
}

func deleteClusterNamespaceResource(req *restful.Request) error {
	mList, err := deleteBatchResources(req, csListFeatTags)
	if err != nil {
		return err
	}
	PushDeleteBatchClusterToQueue(mList)
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
	// 表名
	resourceType := getTable(req)

	return DeleteBatchData(req.Request.Context(), resourceType, getOption, rmOption)
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

	// 表名
	resourceType := getTable(req)
	return CreateCustomResourceIndex(req.Request.Context(), resourceType, index)
}

func deleteCustomResourcesIndex(req *restful.Request) error {
	resourceType := getTable(req)
	indexName := req.PathParameter(indexNameTag)

	return DeleteCustomResourceIndex(req.Request.Context(), resourceType, indexName)
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
