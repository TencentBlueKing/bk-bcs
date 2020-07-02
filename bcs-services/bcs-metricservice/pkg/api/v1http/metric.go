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

package v1http

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"io/ioutil"
	"net/http"
	"strings"

	restful "github.com/emicklei/go-restful"
)

const (
	clusterTypeTag = "clusterType"
	clusterIdTag   = "clusterId"
	namespaceTag   = "namespace"
	nameTag        = "name"
)

var metricStorage storage.Storage

func SetMetrics(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterType := strings.ToLower(req.PathParameter(clusterTypeTag))
	t := types.GetClusterType(clusterType)
	if t == types.ClusterUnknown {
		blog.Errorf("unknown cluster type: %s", clusterType)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricUnknownClusterType, Message: common.BcsErrMetricUnknownClusterTypeStr})
		return
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("fail to read request body. err: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommHttpReadReqBody, Message: common.BcsErrCommHttpReadReqBodyStr})
		return
	}

	param := make([]*types.Metric, 0)
	if err = codec.DecJson(body, &param); err != nil {
		blog.Errorf("fail to decode data from metric set request body: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommJsonDecode, Message: common.BcsErrCommJsonDecodeStr})
		return
	}

	for i, m := range param {
		if err = isValidMetric(m); err != nil {
			blog.Errorf("metric(No.%d) invalid: %v | %v", i, err, m)
			api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricInvalidData, Message: common.BcsErrMetricInvalidDataStr})
			return
		}
	}

	for i, m := range param {
		m.ClusterType = clusterType
		if err := metricStorage.SaveMetric(&storage.Param{
			ClusterID: m.ClusterID,
			Type:      types.ResourceMetricType,
			Name:      m.Name,
			Namespace: m.Namespace,
			Data:      m,
		}); err != nil {
			blog.Errorf("save metric failed(No.%d), clusterId(%s) namespace(%s) name(%s): %v", i, m.ClusterID, m.Namespace, m.Name, err)
			api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricSetMetricFailed, Message: common.BcsErrMetricSetMetricFailedStr})
			return
		}
	}
	api.ReturnRest(&api.RestResponse{Resp: resp})
}

func DeleteMetrics(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterType := strings.ToLower(req.PathParameter(clusterTypeTag))
	t := types.GetClusterType(clusterType)
	if t == types.ClusterUnknown {
		blog.Errorf("unknown cluster type: %s", clusterType)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricUnknownClusterType, Message: common.BcsErrMetricUnknownClusterTypeStr})
		return
	}

	clusterId := req.PathParameter(clusterIdTag)
	namespace := req.PathParameter(namespaceTag)
	names := strings.Split(req.QueryParameter(nameTag), ",")

	for i, n := range names {
		if err := metricStorage.DeleteMetric(&storage.Param{
			ClusterID: clusterId,
			Type:      types.ResourceMetricType,
			Name:      n,
			Namespace: namespace,
		}); err != nil {
			blog.Errorf("delete metric failed(No.%d), clusterId(%s) namespace(%s) name(%s): %v", i, clusterId, namespace, n, err)
			api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricDeleteMetricFailed, Message: common.BcsErrMetricDeleteMetricFailedStr})
			return
		}
	}
	api.ReturnRest(&api.RestResponse{Resp: resp})
}

func GetCollector(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterId := req.PathParameter(clusterIdTag)
	namespace := req.PathParameter(namespaceTag)

	r, err := metricStorage.QueryMetric(&storage.Param{
		ClusterID: clusterId,
		Type:      types.ResourceCollectorType,
		Namespace: namespace,
	})
	if err != nil {
		blog.Errorf("get collector failed, clusterId(%s) namespace(%s): %v", clusterId, namespace, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetCollectorFailed, Message: common.BcsErrMetricGetCollectorFailedStr})
		return
	}

	var cfg []collectorCfg
	if err = codec.DecJson(r, &cfg); err != nil {
		blog.Errorf("decode collector config, failed clusterId(%s) namespace(%s): %v", clusterId, namespace, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetCollectorFailed, Message: common.BcsErrMetricGetCollectorFailedStr})
		return
	}

	result := make([]interface{}, 0)
	version := ""
	for _, c := range cfg {
		version += c.Data.Version
		result = append(result, c.Data.Cfg)
	}
	api.ReturnRest(&api.RestResponse{Resp: resp, Data: map[string]interface{}{
		"version": version,
		"cfgs":    result,
	}})
}

func GetMetrics(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("fail to read request body. err: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommHttpReadReqBody, Message: common.BcsErrCommHttpReadReqBodyStr})
		return
	}

	var param metricQuery
	if err = codec.DecJson(body, &param); err != nil {
		blog.Errorf("fail to decode data from metric set request body: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommJsonDecode, Message: common.BcsErrCommJsonDecodeStr})
		return
	}

	result := make([]*types.Metric, 0)
	for _, clusterId := range param.ClusterId {
		r, err := metricStorage.QueryMetric(&storage.Param{
			ClusterID: clusterId,
			Type:      types.ResourceMetricType,
			Name:      param.Name,
		})
		if err != nil {
			blog.Errorf("get metric failed, clusterId(%s) : %v", clusterId, err)
			api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetMetricFailed, Message: common.BcsErrMetricGetMetricFailedStr})
			return
		}
		var m []metricData
		if err = codec.DecJson(r, &m); err != nil {
			blog.Errorf("decode metric config, failed clusterId(%s): %v", clusterId, err)
			api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetMetricFailed, Message: common.BcsErrMetricGetMetricFailedStr})
			return
		}
		for _, mi := range m {
			result = append(result, mi.Data)
		}
	}
	api.ReturnRest(&api.RestResponse{Resp: resp, Data: result})
}

func SetMetricTask(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterID := req.PathParameter(clusterIdTag)
	namespace := req.PathParameter(namespaceTag)
	name := req.PathParameter(nameTag)
	if clusterID == "" || namespace == "" || name == "" {
		err := fmt.Errorf("invalid query params clusterID: %s name: %s namespace: %s",
			clusterID, namespace, name)
		blog.Errorf("set metric task failed: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricSetMetricTaskFailed, Message: fmt.Sprintf("%s %v", common.BcsErrMetricSetMetricTaskFailedStr, err)})
		return
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("fail to read request body. err: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommHttpReadReqBody, Message: common.BcsErrCommHttpReadReqBodyStr})
		return
	}

	var task types.MetricTask
	if err = codec.DecJson(body, &task); err != nil {
		blog.Errorf("fail to decode data from metric task request body: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrCommJsonDecode, Message: common.BcsErrCommJsonDecodeStr})
		return
	}

	task.ClusterID = clusterID
	task.Namespace = namespace
	task.Name = name
	for _, pod := range task.Pods {
		pod.Meta.NameSpace = namespace
	}
	if err = metricStorage.SaveMetric(&storage.Param{
		ClusterID: task.ClusterID,
		Type:      types.ResourceTaskType,
		Name:      task.Name,
		Namespace: task.Namespace,
		Data:      task,
	}); err != nil {
		blog.Errorf("save metric task failed, clusterId(%s) namespace(%s) name(%s): %v", task.ClusterID, task.Namespace, task.Name, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricSetMetricTaskFailed, Message: fmt.Sprintf("%s %v", common.BcsErrMetricSetMetricTaskFailedStr, err)})
		return
	}

	api.ReturnRest(&api.RestResponse{Resp: resp})
}

func GetMetricTask(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterID := req.PathParameter(clusterIdTag)
	namespace := req.PathParameter(namespaceTag)
	name := req.PathParameter(nameTag)
	if clusterID == "" || namespace == "" || name == "" {
		err := fmt.Errorf("invalid query params clusterID: %s name: %s namespace: %s",
			clusterID, namespace, name)
		blog.Errorf("delete metric task failed: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetMetricTaskFailed, Message: fmt.Sprintf("%s %v", common.BcsErrMetricGetMetricTaskFailedStr, err)})
		return
	}

	r, err := metricStorage.QueryMetric(&storage.Param{
		ClusterID: clusterID,
		Type:      types.ResourceTaskType,
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		blog.Errorf("get metric task failed, clusterId(%s) namespace(%s) name(%s): %v", clusterID, namespace, name, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetMetricTaskFailed, Message: common.BcsErrMetricGetMetricTaskFailedStr})
		return
	}
	var m []*metricTaskData
	if err = codec.DecJson(r, &m); err != nil {
		blog.Errorf("decode metric task failed clusterId(%s) namespace(%s) name(%s): %v", clusterID, namespace, name, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricGetMetricTaskFailed, Message: common.BcsErrMetricGetMetricTaskFailedStr})
		return
	}
	if len(m) == 0 {
		blog.Errorf("get metric task failed, clusterId(%s) namespace(%s) name(%s): resource not exist", clusterID, namespace, name)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricMetricTaskNotExist, Message: common.BcsErrMetricMetricTaskNotExistStr})
		return
	}

	api.ReturnRest(&api.RestResponse{Resp: resp, Data: m[0].Data})
}

func ListMetricTask(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterID := req.PathParameter(clusterIdTag)
	namespace := req.QueryParameter(namespaceTag)

	r, err := metricStorage.QueryMetric(&storage.Param{
		ClusterID: clusterID,
		Type:      types.ResourceTaskType,
		Namespace: namespace,
	})
	if err != nil {
		blog.Errorf("list metric task failed, clusterId(%s) namespace(%s): %v", clusterID, namespace, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricListMetricTaskFailed, Message: common.BcsErrMetricListMetricTaskFailedStr})
		return
	}

	var m []*metricTaskData
	if err = codec.DecJson(r, &m); err != nil {
		blog.Errorf("decode metric config, failed clusterId(%s) namespace(%s) data(%s): %v", clusterID, namespace, r, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricListMetricTaskFailed, Message: common.BcsErrMetricListMetricTaskFailedStr})
		return
	}

	result := make([]*types.MetricTask, 0)
	for _, mi := range m {
		result = append(result, mi.Data)
	}
	api.ReturnRest(&api.RestResponse{Resp: resp, Data: result})
}

func DeleteMetricTask(req *restful.Request, resp *restful.Response) {
	if metricStorage == nil {
		blog.Errorf("metric storage not init")
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricStorageNoFound, Message: common.BcsErrMetricStorageNoFoundStr})
		return
	}

	clusterID := req.PathParameter(clusterIdTag)
	namespace := req.PathParameter(namespaceTag)
	name := req.PathParameter(nameTag)
	if clusterID == "" || namespace == "" || name == "" {
		err := fmt.Errorf("invalid query params clusterID: %s name: %s namespace: %s",
			clusterID, namespace, name)
		blog.Errorf("delete metric task failed: %v", err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricDeleteMetricTaskFailed, Message: fmt.Sprintf("%s %v", common.BcsErrMetricDeleteMetricTaskFailedStr, err)})
		return
	}

	if err := metricStorage.DeleteMetric(&storage.Param{
		ClusterID: clusterID,
		Type:      types.ResourceTaskType,
		Name:      name,
		Namespace: namespace,
	}); err != nil {
		blog.Errorf("delete metric task failed, clusterId(%s) namespace(%s) name(%s): %v", clusterID, namespace, name, err)
		api.ReturnRest(&api.RestResponse{Resp: resp, ErrCode: common.BcsErrMetricDeleteMetricTaskFailed, Message: common.BcsErrMetricDeleteMetricTaskFailedStr})
		return
	}

	api.ReturnRest(&api.RestResponse{Resp: resp})
}

type metricQuery struct {
	ClusterId []string `json:"clusterID"`
	Name      string   `json:"name"`
}

type metricData struct {
	Data *types.Metric `json:"data"`
}

type metricTaskData struct {
	Data *types.MetricTask `json:"data"`
}

type collectorCfg struct {
	Data struct {
		Version string      `json:"version"`
		Cfg     interface{} `json:"cfg"`
	} `json:"data"`
}

func isValidMetric(metric *types.Metric) error {
	if metric.ClusterID == "" {
		return fmt.Errorf("clusterid is empty")
	}
	if metric.Namespace == "" {
		return fmt.Errorf("namespace is empty")
	}
	if metric.Name == "" {
		return fmt.Errorf("name is empty")
	}
	if _, err := metric.TLSConfig.GetTLSConfig(); err != nil {
		return err
	}
	return nil
}

func InitMetricStorage() (err error) {
	resource := api.GetAPIResource()
	metricStorage, err = storage.New(resource.Conf, resource.Rd)
	return
}

func init() {
	api.RegisterV1Action(api.Action{Verb: http.MethodPost, Path: "/metric/clustertype/{clusterType}/metrics", Params: nil, Handler: SetMetrics})
	api.RegisterV1Action(api.Action{Verb: http.MethodDelete, Path: "/metric/clustertype/{clusterType}/clusters/{clusterId}/namespaces/{namespace}/metrics", Params: nil, Handler: DeleteMetrics})
	api.RegisterV1Action(api.Action{Verb: http.MethodGet, Path: "/metric/collector/{clusterType}/{clusterId}/{namespace}/{name}", Params: nil, Handler: GetCollector})
	api.RegisterV1Action(api.Action{Verb: http.MethodPost, Path: "/metric/metrics", Params: nil, Handler: GetMetrics})
	api.RegisterV1Action(api.Action{Verb: http.MethodGet, Path: "/metric/tasks/clusters/{clusterId}", Params: nil, Handler: ListMetricTask})
	api.RegisterV1Action(api.Action{Verb: http.MethodGet, Path: "/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/{name}", Params: nil, Handler: GetMetricTask})
	api.RegisterV1Action(api.Action{Verb: http.MethodPut, Path: "/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/{name}", Params: nil, Handler: SetMetricTask})
	api.RegisterV1Action(api.Action{Verb: http.MethodDelete, Path: "/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/{name}", Params: nil, Handler: DeleteMetricTask})

	api.RegisterInitFunc(InitMetricStorage)
}
