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

package suanlicpu

import (
	"context"
	"fmt"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// SeriesQuery series query
type SeriesQuery struct {
	podType   string
	namespace string
	podName   string
	vmId      string
}

// GetSeriesQueryList get series query list
func (p *Store) GetSeriesQueryList(ctx context.Context, r *storepb.SeriesRequest) ([]*SeriesQuery, error) {
	seriesQueryList := []*SeriesQuery{}

	// 支持的过滤 labels
	podType, err := clientutil.GetLabelMatchValue("pod_type", r.Matchers)
	if err != nil {
		return nil, err
	}
	namespace, err := clientutil.GetLabelMatchValue("namespace", r.Matchers)
	if err != nil {
		return nil, err
	}
	podNames, err := clientutil.GetLabelMatchValues("pod_name", r.Matchers)
	if err != nil {
		return nil, err
	}
	vmIds, err := clientutil.GetLabelMatchValues("vm_id", r.Matchers)
	if err != nil {
		return nil, err
	}

	if len(vmIds) > 0 {
		for _, vmId := range vmIds {
			seriesQueryList = append(seriesQueryList, &SeriesQuery{
				podType: podType,
				vmId:    vmId,
			})
		}
		return seriesQueryList, nil
	}

	if namespace != "" && len(podNames) > 0 {
		for _, podName := range podNames {
			seriesQueryList = append(seriesQueryList, &SeriesQuery{
				podType:   podType,
				namespace: namespace,
				podName:   podName,
			})
		}
		return seriesQueryList, nil
	}

	return nil, errors.Errorf("namespace,pod_name or vm_id label not found")

}

// FetchAndSendGPU 单个数据查询
func (p *Store) FetchAndSendGPU(ctx context.Context, r *storepb.SeriesRequest, s storepb.Store_SeriesServer,
	query *SeriesQuery) error {
	matchers := make([]storepb.LabelMatcher, 0, len(r.Matchers))
	appendGPULabels := map[string]string{
		"pod_type":   query.podType,
		"cluster_id": p.config.ClusterID,
	}

	for _, m := range r.Matchers {
		if _, ok := IgnoreGPULabels[m.Name]; ok {
			continue
		}
		matchers = append(matchers, m)
	}

	if query.vmId != "" {
		matchers = append(matchers, storepb.LabelMatcher{
			Type:  storepb.LabelMatcher_EQ,
			Name:  "pod_name", // pod_name 是算力固定值
			Value: query.vmId,
		})
		appendGPULabels["vm_id"] = query.vmId

	} else if query.namespace != "" && query.podName != "" {
		vmId, err := k8sclient.GetPodEntryValue(ctx, p.config.ClusterID, query.namespace, query.podName, "lowerPodID")
		if err != nil {
			return err
		}

		matchers = append(matchers, storepb.LabelMatcher{
			Type:  storepb.LabelMatcher_EQ,
			Name:  "pod_name", // pod_name 是算力固定值
			Value: vmId,
		})

		appendGPULabels["namespace"] = query.namespace
		appendGPULabels["pod_name"] = query.podName
		appendGPULabels["vm_id"] = vmId

	} else {
		return errors.Errorf("namespace,pod_name or vm_id label not found")
	}

	namespaceMatcher, err := p.MakeNamespaceMatcher(ctx)
	if err != nil {
		return err
	}

	// 添加算力集群ID限制， 使用正则表达式
	matchers = append(matchers, *namespaceMatcher)

	data, _, err := p.QueryRangeInGRPC(ctx, p.base, matchers, clientutil.GetPromQueryTime(r))
	if err != nil {
		return err
	}

	series := clientutil.SampleStreamToSeries(data, IgnoreGPULabels, appendGPULabels)
	for _, serie := range series {
		if err := p.SendSeries(serie, s, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

// FetchAndSendGPUSeries 并发数据查询
func (p *Store) FetchAndSendGPUSeries(ctx context.Context, r *storepb.SeriesRequest,
	s storepb.Store_SeriesServer) error {
	seriesQueryList, err := p.GetSeriesQueryList(ctx, r)
	if err != nil {
		return err
	}

	var (
		wg              sync.WaitGroup
		seriesQueryChan = make(chan *SeriesQuery)
		multiErrors     *multierror.Error
		mtx             sync.Mutex
	)

	expectConcurrency := concurrency
	if expectConcurrency > len(seriesQueryList) {
		expectConcurrency = len(seriesQueryList)
	}

	for i := 0; i < expectConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seriesQuery := range seriesQueryChan {
				err := p.FetchAndSendGPU(ctx, r, s, seriesQuery)
				if err != nil {
					mtx.Lock()
					multiErrors = multierror.Append(multiErrors, errors.Wrapf(err, fmt.Sprintf("fetch %s", seriesQuery)))
					mtx.Unlock()
				}
			}
		}()
	}

	for _, seriesQuery := range seriesQueryList {
		seriesQueryChan <- seriesQuery
	}

	close(seriesQueryChan)
	wg.Wait()

	// 如果全部错误, 直接返回异常请求
	if len(multiErrors.WrappedErrors()) == len(seriesQueryList) {
		return multiErrors.ErrorOrNil()
	}

	// 部分错误, 返回 warning 信息
	if len(multiErrors.WrappedErrors()) > 0 {
		_ = s.Send(storepb.NewWarnSeriesResponse(multiErrors.ErrorOrNil()))
	}

	return nil
}
