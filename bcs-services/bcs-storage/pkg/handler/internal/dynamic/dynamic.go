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
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamic"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	needTimeFormatList = []string{constants.UpdateTimeTag, constants.CreateTimeTag}
	nsFeatTags         = []string{constants.ClusterIDTag, constants.NamespaceTag, constants.ResourceTypeTag,
		constants.ResourceNameTag}
	customResourceFeatTags = nsFeatTags
	csListFeatTags         = []string{constants.ClusterIDTag, constants.ResourceTypeTag}
	nsListFeatTags         = []string{constants.ClusterIDTag, constants.NamespaceTag, constants.ResourceTypeTag}
	csFeatTags             = []string{constants.ClusterIDTag, constants.ResourceTypeTag, constants.ResourceNameTag}
)

type general struct {
	ctx              context.Context
	resource         *storage.Resources
	resourceMap      map[string]interface{}
	resourceFeatList []string
	//
	data             map[string]interface{}
	extra            string
	offset           uint64
	limit            uint64
	fields           []string
	labelSelector    string
	updateTimeBefore int64
	//
	updateTimeBegin int64
	updateTimeEnd   int64
	//
	indexName string
}

func (g *general) getExtra() operator.M {
	if g.extra == "" {
		return nil
	}
	extra := make(operator.M)
	lib.NewExtra(g.extra).Unmarshal(&extra)
	return extra
}

func (g *general) getCustomBody() map[string][]string {
	body := make(map[string][]string)

	for _, key := range g.resourceFeatList {
		if v, ok := g.resourceMap[key].(string); ok && v != "" {
			body[key] = []string{v}
		}
	}

	if g.labelSelector != "" {
		body[constants.LabelSelectorTag] = []string{g.labelSelector}
	}

	if v := strconv.FormatInt(g.updateTimeBefore, 10); v != "" && g.updateTimeBefore != 0 {
		body[constants.UpdateTimeQueryTag] = []string{v}
	}

	return body
}

func (g *general) getFeatures() operator.M {
	features := make(operator.M)
	for _, key := range g.resourceFeatList {
		if v, ok := g.resourceMap[key].(string); ok && v != "" {
			features[key] = v
		}
	}
	return features
}

func (g *general) getCondition() *operator.Condition {
	features := g.getFeatures()
	extras := g.getExtra()
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
		if key == constants.ResourceTypeTag {
			switch features[key] {
			case constants.ApplicationTypeName:
				featuresExcept[constants.KindTag] = constants.ProcessTypeName
			case constants.ProcessTypeName:
				features[key] = constants.ApplicationTypeName
				features[constants.KindTag] = constants.ProcessTypeName
			}
		}
	}
	condition := operator.NewLeafCondition(operator.Eq, features)
	if len(featuresExcept) == 0 {
		notCondition := operator.NewLeafCondition(operator.Ne, featuresExcept)
		condition = operator.NewBranchCondition(operator.And, condition, notCondition)
	}

	customCondition := lib.GetCustomConditionFromBody(g.getCustomBody())
	if customCondition != nil {
		condition = operator.NewBranchCondition(operator.And, condition, customCondition)
	}

	by, _ := json.Marshal(condition)
	blog.Infof("condition: %s", string(by))
	return condition
}

func (g *general) getSelector() []string {
	return g.fields
}

func (g *general) getStoreOption() *lib.StoreGetOption {
	return &lib.StoreGetOption{
		Fields: g.getSelector(),
		Cond:   g.getCondition(),
		Offset: int64(g.offset),
		Limit:  int64(g.limit),
	}
}

func (g *general) getResources() ([]operator.M, error) {
	opt := g.getStoreOption()

	mList, err := dynamic.GetData(g.ctx, g.resource.ResourceType, opt)
	if err != nil {
		return nil, err
	}

	lib.FormatTime(mList, needTimeFormatList)
	return mList, nil
}

// getData 获取任意类型
func (g *general) getData(features operator.M) (operator.M, error) {
	tmp := util.StructToMap(g.data)

	data := lib.CopyMap(features)
	data[constants.DataTag] = tmp
	return data, nil
}

func (g *general) putResources() (operator.M, error) {
	features := g.getFeatures()
	extras := g.getExtra()
	features.Merge(extras)
	data, err := g.getData(features)
	if err != nil {
		return nil, err
	}
	resourceType := g.resource.ResourceType

	if err = dynamic.PutData(g.ctx, data, features, g.resourceFeatList, resourceType); err != nil {
		return nil, err
	}
	return data, nil
}

func (g *general) deleteResources() ([]operator.M, error) {
	condition := g.getCondition()
	resourceType := g.resource.ResourceType
	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}
	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
		// when resource to be deleted not found, do not return error
		IgnoreNotFound: true,
	}
	return dynamic.DeleteBatchData(g.ctx, resourceType, getOption, rmOption)
}

func (g *general) getTimeCondition() *operator.Condition {
	condList := make([]*operator.Condition, 0)

	if g.updateTimeBegin > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Gt, operator.M{
			constants.UpdateTimeTag: time.Unix(g.updateTimeBegin, 0)}))
	}
	if g.updateTimeEnd > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Lt, operator.M{
			constants.UpdateTimeTag: time.Unix(g.updateTimeEnd, 0)}))
	}

	if len(condList) == 0 {
		return operator.EmptyCondition
	}
	return operator.NewBranchCondition(operator.And, condList...)
}

func (g *general) deleteBatchResources() ([]operator.M, error) {
	featCondition := g.getCondition()
	timeCondition := g.getTimeCondition()
	condition := operator.NewBranchCondition(operator.And, featCondition, timeCondition)

	getOption := &lib.StoreGetOption{
		Cond:           condition,
		IsAllDocuments: true,
	}

	rmOption := &lib.StoreRemoveOption{
		Cond:           condition,
		IgnoreNotFound: true,
	}

	resourceType := g.resource.ResourceType

	return dynamic.DeleteBatchData(g.ctx, resourceType, getOption, rmOption)
}

func (g *general) getCustomResources() ([]operator.M, operator.M, error) {
	return dynamic.GetDataWithPageInfo(g.ctx, g.resource.ResourceType, g.getStoreOption())
}

func (g *general) deleteCustomResources() error {
	_, err := g.deleteResources()
	return err
}

func (g *general) putCustomResources() error {
	dataRaw := g.data

	resourceType := g.resource.ResourceType
	opt := &lib.StorePutOption{
		CreateTimeKey: constants.CreateTimeTag,
		UpdateTimeKey: constants.UpdateTimeTag,
	}
	return dynamic.PutCustomResourceToDB(g.ctx, resourceType, dataRaw, opt)
}

func (g *general) createCustomResourcesIndex() error {
	keys := bson.D{}
	index := drivers.Index{
		Unique: true,
		Name:   g.indexName,
	}

	bytes, err := json.Marshal(g.data)
	if err != nil {
		return err
	}

	if err = bson.UnmarshalExtJSON(bytes, true, &keys); err != nil {
		return err
	}

	index.Key = keys

	return dynamic.CreateCustomResourceIndex(g.ctx, g.resource.ResourceType, index)
}

func (g *general) deleteCustomResourcesIndex() error {
	return dynamic.DeleteCustomResourceIndex(g.ctx, g.resource.ResourceType, g.indexName)
}

//k8s namespace resources
//mesos namespace resources

// HandlerGetNsResourcesReq  GetNamespaceResourcesRequest业务方法
func HandlerGetNsResourcesReq(ctx context.Context, req *storage.GetNamespaceResourcesRequest) ([]operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: nsFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		extra:            req.Extra,
		labelSelector:    req.LabelSelector,
		updateTimeBefore: req.UpdateTimeBefore,
		offset:           req.Offset,
		limit:            req.Limit,
		fields:           req.Fields,
	}

	return r.getResources()
}

// HandlerPutNsResourcesReq PutNamespaceResourcesRequest业务方法
func HandlerPutNsResourcesReq(ctx context.Context, req *storage.PutNamespaceResourcesRequest) (operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: nsFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		data:  req.Data.AsMap(),
		extra: req.Extra,
	}

	return r.putResources()
}

// HandlerDelNsResourcesReq DeleteNamespaceResourcesRequest业务方法
func HandlerDelNsResourcesReq(ctx context.Context, req *storage.DeleteNamespaceResourcesRequest) (
	[]operator.M, error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: nsFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
	}

	return r.deleteResources()
}

// HandlerListNsResourcesReq ListNamespaceResourcesRequest业务方法
func HandlerListNsResourcesReq(ctx context.Context, req *storage.ListNamespaceResourcesRequest) (
	[]operator.M, error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: nsListFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		extra:            req.Extra,
		labelSelector:    req.LabelSelector,
		updateTimeBefore: req.UpdateTimeBefore,
		offset:           req.Offset,
		limit:            req.Limit,
		fields:           req.Fields,
	}

	return r.getResources()
}

// HandlerDelBatchNsResourceReq DeleteBatchNamespaceResourceRequest业务方法
func HandlerDelBatchNsResourceReq(ctx context.Context, req *storage.DeleteBatchNamespaceResourceRequest) (
	[]operator.M, error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: nsListFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		updateTimeBegin:  req.UpdateTimeBegin,
		updateTimeEnd:    req.UpdateTimeEnd,
	}

	return r.deleteBatchResources()
}

// k8s cluster resources
// mesos Cluster resources.

// HandlerGetClusterResourcesReq GetClusterResourcesRequest业务方法
func HandlerGetClusterResourcesReq(ctx context.Context, req *storage.GetClusterResourcesRequest) ([]operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		ResourceName: req.ResourceName,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: csFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		extra:            req.Extra,
		labelSelector:    req.LabelSelector,
		updateTimeBefore: req.UpdateTimeBefore,
		offset:           req.Offset,
		limit:            req.Limit,
		fields:           req.Fields,
	}

	return r.getResources()
}

// HandlerPutClusterResourcesReq PutClusterResourcesRequest业务方法
func HandlerPutClusterResourcesReq(ctx context.Context, req *storage.PutClusterResourcesRequest) (operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		ResourceName: req.ResourceName,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: csFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		data:  req.Data.AsMap(),
		extra: req.Extra,
	}

	return r.putResources()
}

// HandlerDelClusterResourcesReq DeleteClusterResourcesRequest业务方法
func HandlerDelClusterResourcesReq(ctx context.Context, req *storage.DeleteClusterResourcesRequest) (
	[]operator.M, error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		ResourceName: req.ResourceName,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: csFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
	}

	return r.deleteResources()
}

// HandlerListClusterResourcesReq ListClusterResourcesRequest业务方法
func HandlerListClusterResourcesReq(ctx context.Context, req *storage.ListClusterResourcesRequest) ([]operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: csListFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		extra:            req.Extra,
		labelSelector:    req.LabelSelector,
		updateTimeBefore: req.UpdateTimeBefore,
		offset:           req.Offset,
		limit:            req.Limit,
		fields:           req.Fields,
	}

	return r.getResources()
}

// HandlerDelBatchClusterResourceReq DeleteBatchClusterResourceRequest业务方法
func HandlerDelBatchClusterResourceReq(ctx context.Context, req *storage.DeleteBatchClusterResourceRequest) (
	[]operator.M, error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resourceFeatList: csListFeatTags,
		resource:         resource,
		resourceMap:      util.StructToMap(resource),
		updateTimeBegin:  req.UpdateTimeBegin,
		updateTimeEnd:    req.UpdateTimeEnd,
	}

	return r.deleteBatchResources()
}

// Custom resource
// Custom resources OPs

// HandlerGetCustomResources GetCustomResourcesRequest业务方法
func HandlerGetCustomResources(ctx context.Context, req *storage.GetCustomResourcesRequest) ([]operator.M, operator.M,
	error) {
	resource := &storage.Resources{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceName: req.ResourceName,
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resource:         resource,
		resourceFeatList: customResourceFeatTags,
		resourceMap:      util.StructToMap(resource),
		// --- ----
		extra:            req.Extra,
		labelSelector:    req.LabelSelector,
		updateTimeBefore: req.UpdateTimeBefore,
		offset:           req.Offset,
		limit:            req.Limit,
		fields:           req.Fields,
	}

	return r.getCustomResources()
}

// HandlerDelCustomResources DeleteCustomResourcesRequest业务方法
func HandlerDelCustomResources(ctx context.Context, req *storage.DeleteCustomResourcesRequest) error {
	resource := &storage.Resources{
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resource:         resource,
		resourceFeatList: customResourceFeatTags,
	}

	return r.deleteCustomResources()
}

// HandlerPutCustomResources PutCustomResourcesRequest业务方法
func HandlerPutCustomResources(ctx context.Context, req *storage.PutCustomResourcesRequest) error {
	resource := &storage.Resources{
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resource:         resource,
		resourceFeatList: customResourceFeatTags,
		// --- ----
		data: req.Data.AsMap(),
	}

	return r.putCustomResources()
}

// HandlerCreateCustomResourcesIndex CreateCustomResourcesIndexRequest业务方法
func HandlerCreateCustomResourcesIndex(ctx context.Context, req *storage.CreateCustomResourcesIndexRequest) error {
	resource := &storage.Resources{
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resource:         resource,
		indexName:        req.IndexName,
		resourceFeatList: customResourceFeatTags,
		// --- ----
		data: req.Data.AsMap(),
	}

	return r.createCustomResourcesIndex()
}

// HandlerDelCustomResourcesIndex DeleteCustomResourcesIndexRequest业务方法
func HandlerDelCustomResourcesIndex(ctx context.Context, req *storage.DeleteCustomResourcesIndexRequest) error {
	resource := &storage.Resources{
		ResourceType: req.ResourceType,
	}
	r := &general{
		ctx:              ctx,
		resource:         resource,
		resourceFeatList: customResourceFeatTags,
	}

	return r.deleteCustomResourcesIndex()
}
