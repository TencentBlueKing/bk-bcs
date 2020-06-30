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

package driver

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"io/ioutil"
	"os"
	"path/filepath"

	simplejson "github.com/bitly/go-simplejson"
)

const (
	jsonCustomFormat  = "%s_custom_%s.json"
	jsonDefaultFormat = "%s_default.json"

	IPPortGap = "+"
)

func GetFilePath(tempDir, namespace string, ct types.ClusterType) (target string, err error) {
	// custom json with namespace
	target = filepath.Join(tempDir, fmt.Sprintf(jsonCustomFormat, ct, namespace))
	if _, err = os.Stat(target); !os.IsNotExist(err) {
		return
	}

	// default json
	target = filepath.Join(tempDir, fmt.Sprintf(jsonDefaultFormat, ct.String()))
	if _, err = os.Stat(target); !os.IsNotExist(err) {
		return
	}

	return "", fmt.Errorf("%s: %s", common.BcsErrMetricResourceFileNotExistStr, tempDir)
}

func LoadResourceJson(tempDir, namespace string, ct types.ClusterType) (*simplejson.Json, error) {
	target, err := GetFilePath(tempDir, namespace, ct)
	if err != nil {
		return nil, err
	}

	template, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}

	return simplejson.NewJson(template)
}

type StorageTaskGroupIf struct {
	Namespace string              `json:"namespace"`
	Data      btypes.BcsPodStatus `json:"data"`
}

type StoragePodIf struct {
	Namespace string `json:"namespace"`
	Data      struct {
		Metadata btypes.ObjectMeta `json:"metadata"`
		Status   struct {
			HostIP string `json:"hostIP"`
			PodIP  string `json:"podIP"`
			Phase  string `json:"phase"`
		} `json:"status"`
	} `json:"data"`
}

func GetIPMetaFromDynamic(raw []byte, metric *types.Metric) (ipMeta map[string]btypes.ObjectMeta, err error) {
	switch types.GetClusterType(metric.ClusterType) {
	case types.ClusterMesos:
		return GetMesosIPMeta(raw, metric)
	case types.ClusterK8S:
		return GetK8SIPMeta(raw, metric)
	default:
		err = fmt.Errorf("unknown cluster type: %s", metric.ClusterType)
		return
	}
}

func GetMesosIPMeta(raw []byte, metric *types.Metric) (ipMeta map[string]btypes.ObjectMeta, err error) {
	var data []StorageTaskGroupIf
	if err = codec.DecJson(raw, &data); err != nil {
		return
	}
	ipMeta = make(map[string]btypes.ObjectMeta)

	for _, item := range data {
		if item.Namespace != metric.Namespace {
			continue
		}
		if item.Data.Status != btypes.Pod_Running {
			continue
		}
		if len(item.Data.ContainerStatuses) == 0 {
			continue
		}
		taskInfo := item.Data.ContainerStatuses[0]

		match := true
		for selectKey, selectVal := range metric.Selector {
			if val, ok := taskInfo.Labels[selectKey]; !ok || val != selectVal {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		key := ""
		// No-Bridge mode
		if taskInfo.Network != "BRIDGE" {
			if taskInfo.Network == "HOST" {
				if item.Data.HostIP != "" {
					key = fmt.Sprintf("%s%s%d", item.Data.HostIP, IPPortGap, metric.Port)
				}
			} else {
				if item.Data.PodIP != "" {
					key = fmt.Sprintf("%s%s%d", item.Data.PodIP, IPPortGap, metric.Port)
				}
			}
		}

		// Bridge mode
		if taskInfo.Network == "BRIDGE" && item.Data.HostIP != "" {
			key = findMesosNetworkIPKey(item.Data.ContainerStatuses, item.Data.HostIP, int(metric.Port))
		}

		if key != "" {
			ipMeta[key] = item.Data.ObjectMeta
		}
	}
	return
}

func findMesosNetworkIPKey(containerStatuses []*btypes.BcsContainerStatus, hostIP string, metricPort int) string {
	for _, cStatus := range containerStatuses {
		for _, pStatus := range cStatus.Ports {
			if pStatus.ContainerPort == metricPort && pStatus.HostPort > 0 {
				return fmt.Sprintf("%s%s%d", hostIP, IPPortGap, pStatus.HostPort)
			}
		}
	}
	return ""
}

func GetK8SIPMeta(raw []byte, metric *types.Metric) (ipMeta map[string]btypes.ObjectMeta, err error) {
	var data []StoragePodIf
	if err = codec.DecJson(raw, &data); err != nil {
		return
	}
	ipMeta = make(map[string]btypes.ObjectMeta)

	for _, item := range data {
		if item.Namespace != metric.Namespace {
			continue
		}
		if item.Data.Status.Phase != "Running" {
			continue
		}
		match := true
		for selectKey, selectVal := range metric.Selector {
			if val, ok := item.Data.Metadata.Labels[selectKey]; !ok || val != selectVal {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		if item.Data.Status.PodIP != "" {
			ipMeta[item.Data.Status.PodIP] = item.Data.Metadata
			continue
		}

		if item.Data.Status.HostIP != "" {
			ipMeta[item.Data.Status.HostIP] = item.Data.Metadata
			continue
		}
	}
	return
}

func GetApplicationName(metric *types.Metric) string {
	return fmt.Sprintf("bcs-collector-%s", metric.Namespace)
}
