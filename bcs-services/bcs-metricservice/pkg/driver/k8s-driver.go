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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

// K8S metric driver
type K8SDriver struct {
	metric *types.Metric
	config *config.Config

	storage storage.Storage
	route   route.Route
}

func NewK8SDriver(m *types.Metric, config *config.Config, s storage.Storage, r route.Route) ClusterDriver {
	return &K8SDriver{metric: m, config: config, storage: s, route: r}
}

func (kd *K8SDriver) GetCollectorTypeName() string {
	return "Deployment"
}

func (kd *K8SDriver) GetIPMeta() (map[string]btypes.ObjectMeta, error) {
	data, err := kd.storage.QueryDynamic(&storage.Param{Namespace: kd.metric.Namespace, ClusterID: kd.metric.ClusterID, ClusterType: types.ClusterK8S, Type: types.ClusterK8S.GetContainerTypeName()})
	if err != nil {
		blog.Error("failed to query pod dynamic: %v", err)
		return nil, err
	}

	return GetIPMetaFromDynamic(data, kd.metric)
}

func (kd *K8SDriver) GetApplicationJson(imageBase string) (js *simplejson.Json, err error) {
	if js, err = LoadResourceJson(kd.config.TempDir, kd.metric.Namespace, types.ClusterK8S); err != nil {
		blog.Errorf("failed to load k8s template json: %v", err)
		return
	}

	name := GetApplicationName(kd.metric)
	namespace := kd.metric.Namespace
	clusterId := kd.metric.ClusterID
	label := fmt.Sprintf("%s_%s", kd.metric.Namespace, kd.metric.Name)

	// metadata info
	js.Get("metadata").Set("name", name)
	js.Get("metadata").Set("namespace", namespace)
	js.Get("metadata").Get("labels").Set("io.tencent.bcs.metric.collector", label)

	// spec/template/metadata info
	js.GetPath("spec", "template", "metadata", "labels").Set("io.tencent.bcs.metric.collector", label)

	// spec/selector/matchLabels info
	js.GetPath("spec", "selector", "matchLabels").Set("io.tencent.bcs.metric.collector", label)

	// spec/template/spec/containers info
	containerJs := js.GetPath("spec", "template", "spec", "containers")
	containers, err := containerJs.Array()
	if err != nil {
		blog.Errorf("k8s template json error: %v", err)
		return
	}

	envSettings := []btypes.EnvVar{
		{Name: "MetricClusterType", Value: types.ClusterK8S.String()},
		{Name: "MetricClusterID", Value: clusterId},
		{Name: "ZkAddress", Value: kd.config.BCSZk},
		{Name: "MetricApplicationName", Value: name},
		{Name: "MetricApplicationNamespace", Value: namespace},
	}

	// ENV
	for index, item := range containers {
		container, ok := item.(map[string]interface{})
		if !ok {
			containerJs.GetIndex(index).Set("env", envSettings)
			continue
		}

		env, ok := container["env"]
		if !ok {
			containerJs.GetIndex(index).Set("env", envSettings)
			continue
		}

		envItem, ok := env.([]interface{})
		if !ok {
			containerJs.GetIndex(index).Set("env", envSettings)
			continue
		}

		for _, es := range envSettings {
			envItem = append(envItem, es)
		}
		containerJs.GetIndex(index).Set("env", envItem)
	}

	// IMAGE
	for index, item := range containers {
		container, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		env, ok := container["image"]
		if !ok {
			continue
		}

		envItem, ok := env.(string)
		if !ok {
			continue
		}

		containerJs.GetIndex(index).Set("image", imageBase+envItem[strings.Index(envItem, "/"):])
	}

	// spec/template/spec network info
	js.GetPath("spec", "template", "spec").Set("hostNetwork", kd.metric.HostNetwork)
	js.GetPath("spec", "template", "spec").Set("dnsPolicy", kd.metric.DnsPolicy)
	js.GetPath("spec", "template", "spec").Set("imagePullSecrets", kd.metric.Secrets)
	return
}

func (kd *K8SDriver) CreateApplication(data []byte) (err error) {
	if err = kd.route.CreateK8S(kd.metric.ClusterID, kd.metric.Namespace, data); err != nil {
		blog.Errorf("create k8s deployment failed: %v", err)
	}
	return
}

func (kd *K8SDriver) DeleteApplication(data []byte) (err error) {
	rs, err := kd.getRS()
	if err != nil {
		return
	}

	pod, err := kd.getPods(rs)
	if err != nil {
		return
	}

	// first delete Deployment
	if err = kd.deleteDeployment(); err != nil {
		return
	}

	// then delete the RSs belong to this Deployment
	for _, rsi := range rs {
		if err = kd.deleteRS(rsi); err != nil {
			return
		}
	}

	// finally delete the Pods belong to all these RSs
	for _, podI := range pod {
		if err = kd.deletePods(podI); err != nil {
			return
		}
	}
	return
}

func (kd *K8SDriver) deleteDeployment() (err error) {
	if err = kd.route.DeleteK8S(kd.metric.ClusterID, kd.metric.Namespace, "deployments", GetApplicationName(kd.metric), nil); err != nil {
		blog.Errorf("delete k8s collector deployment failed: %v", err)
	}
	return
}

func (kd *K8SDriver) getRS() (rsDataList []*ResourceData, err error) {
	rsDataList = make([]*ResourceData, 0)
	r, err := kd.storage.GetDynamicNs(&storage.Param{
		ClusterType: types.GetClusterType(kd.metric.ClusterType),
		ClusterID:   kd.metric.ClusterID,
		Namespace:   kd.metric.Namespace,
		Type:        "ReplicaSet",
		Field:       []string{"resourceName", "namespace"},
		Extra:       map[string]string{"data.metadata.ownerReferences.name": GetApplicationName(kd.metric)},
	})

	if err != nil {
		blog.Errorf("get k8s collector rs failed: %v | %s", err, r)
	}

	if err = codec.DecJson(r, &rsDataList); err != nil {
		blog.Errorf("decode k8s collector rs failed: %v | %s", err, r)
		return
	}
	blog.Infof("get k8s collector rs: %v", rsDataList)
	return
}

func (kd *K8SDriver) deleteRS(rsData *ResourceData) (err error) {
	if err = kd.route.DeleteK8S(kd.metric.ClusterID, kd.metric.Namespace, "replicasets", rsData.Name, nil); err != nil {
		blog.Errorf("delete k8s collector rs failed: %v", err)
	}
	return
}

func (kd *K8SDriver) getPods(rsDataList []*ResourceData) (podDataList []*ResourceData, err error) {
	podDataList = make([]*ResourceData, 0)
	var r []byte
	for _, rsi := range rsDataList {
		r, err = kd.storage.GetDynamicNs(&storage.Param{
			ClusterType: types.GetClusterType(kd.metric.ClusterType),
			ClusterID:   kd.metric.ClusterID,
			Namespace:   kd.metric.Namespace,
			Type:        "Pod",
			Field:       []string{"resourceName", "namespace"},
			Extra:       map[string]string{"data.metadata.ownerReferences.name": rsi.Name},
		})
		if err != nil {
			blog.Errorf("get k8s collector pod failed: %v | %s", err, r)
			return
		}
		if err = codec.DecJson(r, &podDataList); err != nil {
			blog.Errorf("decode k8s collector pod failed: %v | %s", err, r)
			return
		}
		blog.Infof("get k8s collector pod: %v", podDataList)
	}
	return
}

func (kd *K8SDriver) deletePods(podData *ResourceData) (err error) {
	if err = kd.route.DeleteK8S(kd.metric.ClusterID, kd.metric.Namespace, "pods", podData.Name, nil); err != nil {
		blog.Errorf("delete k8s collector pod failed: %v", err)
	}
	return
}

type ResourceData struct {
	Name string `json:"resourceName"`
}
