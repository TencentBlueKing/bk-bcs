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

package meta

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/driver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"

	"github.com/bitly/go-simplejson"
)

type MetricMeta struct {
	metric *types.Metric

	storage storage.Storage

	driver driver.ClusterDriver
}

func NewMetricMeta(metric *types.Metric, config *config.Config, storage storage.Storage, route route.Route, zk zk.Zk) (mm *MetricMeta, err error) {
	d, err := driver.GetClusterDriver(metric, config, storage, route, zk)
	if err != nil {
		return
	}
	mm = &MetricMeta{
		metric:  metric,
		storage: storage,
		driver:  d,
	}
	return
}

func (mm *MetricMeta) GetIpMeta() (map[string]btypes.ObjectMeta, error) {
	return mm.driver.GetIPMeta()
}

// Put or update metric collector settings to storage
func (mm *MetricMeta) SetCollectorSettings(ipMeta map[string]btypes.ObjectMeta) error {
	blog.Infof("set collector settings: version(%s) clusterID(%s) namespace(%s) name(%s)",
		mm.metric.Version,
		mm.metric.ClusterID,
		mm.metric.Namespace,
		mm.metric.Name)

	// add extra ipMeta from metric task
	if extraIpMeta := mm.collectTaskIPMeta(); extraIpMeta != nil {
		for k, v := range extraIpMeta {
			ipMeta[k] = v
		}
	}

	cfg := make([]types.CollectorCfg, 0)
	for ip, meta := range ipMeta {
		port := mm.metric.Port
		if pair := strings.Split(ip, driver.IPPortGap); len(pair) > 1 {
			ip = pair[0]
			p, _ := strconv.Atoi(pair[1])
			port = uint(p)
		}

		scheme := "http"
		if mm.metric.TLSConfig.IsTLS {
			scheme = "https"
		}
		if s, ok := meta.Annotations[types.BcsComponentsSchemeKey]; ok && s != "" {
			scheme = s
			delete(meta.Annotations, types.BcsComponentsSchemeKey)
		}

		tlsConfig := mm.metric.TLSConfig
		tlsConfig.IsTLS = scheme == "https"

		address := fmt.Sprintf("%s://%s:%d%s", scheme, ip, port, path.Clean("/"+mm.metric.URI))
		cfg = append(cfg, types.CollectorCfg{
			CfgKey:                fmt.Sprintf("%s_%s", mm.metric.Name, address),
			Meta:                  meta,
			Version:               mm.metric.Version,
			IP:                    ip,
			Port:                  port,
			Scheme:                scheme,
			Address:               address,
			Head:                  mm.metric.Head,
			Parameters:            mm.metric.Parameters,
			Method:                mm.metric.Method,
			DataID:                mm.metric.DataID,
			Frequency:             mm.metric.Frequency,
			Timeout:               mm.metric.Timeout,
			TLSConfig:             tlsConfig,
			MetricType:            mm.metric.MetricType,
			PrometheusConstLabels: mm.metric.PrometheusConstLabels,
		})
	}

	return mm.storage.SaveMetric(&storage.Param{
		Name:      mm.metric.Name,
		Namespace: mm.metric.Namespace,
		ClusterID: mm.metric.ClusterID,
		Type:      types.ResourceCollectorType,
		Data: &types.ApplicationCollectorCfg{
			Version:   mm.metric.Version,
			Name:      mm.metric.Name,
			Namespace: mm.metric.Namespace,
			Cfg:       cfg,
		},
	})
}

func (mm *MetricMeta) DeleteCollectorSettings() error {
	return mm.storage.DeleteMetric(&storage.Param{
		ClusterID: mm.metric.ClusterID,
		Type:      types.ResourceCollectorType,
		Namespace: mm.metric.Namespace,
		Name:      mm.metric.Name,
	})
}

func (mm *MetricMeta) CreateApplication() (err error) {
	if types.GetClusterType(mm.metric.ClusterType) == types.BcsComponents {
		return
	}

	js, err := mm.driver.GetApplicationJson(mm.metric.ImageBase)
	if err != nil {
		blog.Errorf("failed to create application: %v", err)
		return
	}

	// add version info to json
	mm.tagVersion(js)

	b, err := js.MarshalJSON()
	if err != nil {
		blog.Errorf("application json marshal failed: %v", err)
		return
	}

	if err = mm.DeleteApplication(); err != nil {
		if err != types.DeleteCollectorNotExist {
			blog.Errorf("delete application before creation failed: %v", err)
			return
		}
		blog.Warnf("application is no exist, can not be delete: clusterId(%s) namespace(%s) name(%s)", mm.metric.ClusterID, mm.metric.Namespace, mm.metric.Name)
	}

	if err = mm.driver.CreateApplication(b); err != nil {
		blog.Errorf("create application failed: %v", err)
		return
	}

	return mm.storage.SaveMetric(&storage.Param{
		ClusterID: mm.metric.ClusterID,
		Type:      types.ResourceApplicationType,
		Namespace: mm.metric.Namespace,
		Name:      driver.GetApplicationName(mm.metric),
		Data:      string(b),
	})
}

func (mm *MetricMeta) DeleteApplication() (err error) {
	if types.GetClusterType(mm.metric.ClusterType) == types.BcsComponents {
		return
	}

	isAvailable, err := mm.isApplicationAvailable()
	if err != nil {
		blog.Errorf("check before delete application failed: %v", err)
		return
	}
	if isAvailable {
		if err = mm.driver.DeleteApplication(nil); err != nil {
			blog.Errorf("delete application failed: %v", err)
			return
		}
	} else {
		blog.Warnf("application is no exist, can not be delete: clusterId(%s) namespace(%s) name(%s)", mm.metric.ClusterID, mm.metric.Namespace, mm.metric.Name)
	}

	return mm.storage.DeleteMetric(&storage.Param{
		ClusterID: mm.metric.ClusterID,
		Type:      types.ResourceApplicationType,
		Namespace: mm.metric.Namespace,
		Name:      driver.GetApplicationName(mm.metric),
	})
}

func (mm *MetricMeta) tagVersion(js *simplejson.Json) {
	js.Get("metadata").Get("labels").Set("io.tencent.bcs.metric.version", mm.metric.Version)
}

func (mm *MetricMeta) isApplicationAvailable() (b bool, err error) {
	r, err := mm.storage.GetDynamicNs(&storage.Param{
		ClusterType: types.GetClusterType(mm.metric.ClusterType),
		ClusterID:   mm.metric.ClusterID,
		Namespace:   mm.metric.Namespace,
		Name:        driver.GetApplicationName(mm.metric),
		Type:        mm.driver.GetCollectorTypeName(),
	})

	if err != nil {
		if strings.Contains(err.Error(), "resource does not exist") {
			return false, nil
		}
		blog.Errorf("check if application is available failed: %v", err)
		return
	}

	var l []interface{}
	if err = codec.DecJson(r, &l); err != nil {
		blog.Errorf("check if application is available failed: %v", err)
		return
	}

	b = len(l) > 0
	return
}

func (mm *MetricMeta) collectTaskIPMeta() map[string]btypes.ObjectMeta {
	r, err := mm.storage.QueryMetric(&storage.Param{
		ClusterID: mm.metric.ClusterID,
		Namespace: mm.metric.Namespace,
		Type:      types.ResourceTaskType})

	if err != nil {
		blog.Errorf("collectTaskIPMeta(%s) get metric task failed: %v", mm.metric.ClusterID, err)
		return nil
	}

	var taskList []*types.StorageTaskIf
	if err = codec.DecJson(r, &taskList); err != nil {
		blog.Errorf("collectTaskIPMeta(%s) decode metric task failed: %v", mm.metric.ClusterID, err)
		return nil
	}

	ipMeta := make(map[string]btypes.ObjectMeta)
	for _, task := range taskList {
		if task.Data.Selector == nil {
			continue
		}

		// check if the task is fit the metric
		match := true
		for selectKey, selectVal := range mm.metric.Selector {
			if val, ok := task.Data.Selector[selectKey]; !ok || val != selectVal {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		for _, pod := range task.Data.Pods {
			key := pod.IP
			if pod.Port > 0 {
				key = fmt.Sprintf("%s%s%d", pod.IP, driver.IPPortGap, pod.Port)
			}

			// override the namespace of task pod, in case of inconsistency
			pod.Meta.NameSpace = mm.metric.Namespace
			ipMeta[key] = pod.Meta
		}
	}
	return ipMeta
}
