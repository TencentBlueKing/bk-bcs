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

package v1

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"
	metricTypes "github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

type Metric interface {
	List(clusterID string) (MetricList, error)
	Inspect(clusterID, namespace, name string) (*metricTypes.Metric, error)
	Delete(clusterType, clusterID, namespace, name string) error
	Upsert(clusterType string, data []byte) error

	ListTask(clusterID string) (MetricTaskList, error)
	InspectTask(clusterID, namespace, name string) (*metricTypes.MetricTask, error)
	DeleteTask(clusterID, namespace, name string) error
	UpsertTask(clusterID string, data []byte) error
}

const (
	BcsMetricListURI   = "%s/bcsapi/v4/metric/metrics"
	BcsMetricDeleteURI = "%s/bcsapi/v4/metric/clustertype/%s/clusters/%s/namespaces/%s/metrics?name=%s"
	BcsMetricUpsertURI = "%s/bcsapi/v4/metric/clustertype/%s/metrics"

	BcsMetricTaskListURI     = "%s/bcsapi/v4/metric/tasks/clusters/%s"
	BcsMetricTaskResourceURI = "%s/bcsapi/v4/metric/tasks/clusters/%s/namespaces/%s/name/%s"
)

var (
	MetricNotFound = fmt.Errorf("metric no found")
)

type bcsMetric struct {
	bcsApiAddress string
	requester     utils.ApiRequester
}

func NewBcsMetric(options types.ClientOptions) Metric {
	return &bcsMetric{
		bcsApiAddress: options.BcsApiAddress,
		requester:     utils.NewApiRequester(options.ClientSSL, options.BcsToken),
	}
}

func (bm *bcsMetric) List(clusterID string) (MetricList, error) {
	return bm.list(clusterID)
}

func (bm *bcsMetric) Inspect(clusterID, namespace, name string) (*metricTypes.Metric, error) {
	return bm.inspect(clusterID, namespace, name)
}

func (bm *bcsMetric) Delete(clusterType, clusterID, namespace, name string) error {
	return bm.delete(clusterType, clusterID, namespace, name)
}

func (bm *bcsMetric) Upsert(clusterType string, data []byte) error {
	return bm.upsert(clusterType, data)
}

func (bm *bcsMetric) ListTask(clusterID string) (MetricTaskList, error) {
	return bm.listTask(clusterID)
}

func (bm *bcsMetric) InspectTask(clusterID, namespace, name string) (*metricTypes.MetricTask, error) {
	return bm.inspectTask(clusterID, namespace, name)
}

func (bm *bcsMetric) DeleteTask(clusterID, namespace, name string) error {
	return bm.deleteTask(clusterID, namespace, name)
}

func (bm *bcsMetric) UpsertTask(clusterID string, data []byte) error {
	return bm.upsertTask(clusterID, data)
}

func (bm *bcsMetric) list(clusterID string) (MetricList, error) {
	// generate list condition param
	var param []byte
	_ = codec.EncJson(listMetricQuery{
		ClusterID: []string{clusterID},
	}, &param)

	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricListURI, bm.bcsApiAddress),
		http.MethodPost,
		param,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("list metric failed: %s", msg)
	}

	var result MetricList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bm *bcsMetric) inspect(clusterID, namespace, name string) (*metricTypes.Metric, error) {
	// generate list condition param
	var param []byte
	_ = codec.EncJson(listMetricQuery{
		ClusterID: []string{clusterID},
		Name:      name,
	}, &param)

	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricListURI, bm.bcsApiAddress),
		http.MethodPost,
		param,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("inspect metric failed: %s", msg)
	}

	var result MetricList
	if err = codec.DecJson(data, &result); err != nil {
		return nil, err
	}

	for _, m := range result {
		if m.Namespace == namespace {
			return m, nil
		}
	}
	return nil, MetricNotFound
}

func (bm *bcsMetric) delete(clusterType, clusterID, namespace, name string) error {
	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricDeleteURI, bm.bcsApiAddress, clusterType, clusterID, namespace, name),
		http.MethodDelete,
		nil,
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("delete metric failed: %s", msg)
	}

	return nil
}

func (bm *bcsMetric) upsert(clusterType string, data []byte) error {
	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricUpsertURI, bm.bcsApiAddress, clusterType),
		http.MethodPost,
		data,
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("upsert metric failed: %s", msg)
	}

	return nil
}

func (bm *bcsMetric) listTask(clusterID string) (MetricTaskList, error) {
	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricTaskListURI, bm.bcsApiAddress, clusterID),
		http.MethodGet,
		nil,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("list metric task failed: %s", msg)
	}

	var result MetricTaskList
	err = codec.DecJson(data, &result)
	return result, err
}

func (bm *bcsMetric) inspectTask(clusterID, namespace, name string) (*metricTypes.MetricTask, error) {
	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricTaskResourceURI, bm.bcsApiAddress, clusterID, namespace, name),
		http.MethodGet,
		nil,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("inspect metric task failed: %s", msg)
	}

	var result metricTypes.MetricTask
	if err = codec.DecJson(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (bm *bcsMetric) deleteTask(clusterID, namespace, name string) error {
	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricTaskResourceURI, bm.bcsApiAddress, clusterID, namespace, name),
		http.MethodDelete,
		nil,
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("delete metric task failed: %s", msg)
	}

	return nil
}

func (bm *bcsMetric) upsertTask(clusterID string, data []byte) error {
	var task metricTypes.MetricTask
	if err := codec.DecJson(data, &task); err != nil {
		return fmt.Errorf("metric task data format error, decode failed: %v", err)
	}

	resp, err := bm.requester.Do(
		fmt.Sprintf(BcsMetricTaskResourceURI, bm.bcsApiAddress, clusterID, task.Namespace, task.Name),
		http.MethodPut,
		data,
	)

	if err != nil {
		return err
	}

	code, msg, _, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("upsert metric task failed: %s", msg)
	}

	return nil
}
