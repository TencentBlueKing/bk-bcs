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
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

// Mesos metric driver
type MesosDriver struct {
	metric *types.Metric
	config *config.Config

	storage storage.Storage
	route   route.Route
}

func NewMesosDriver(m *types.Metric, config *config.Config, s storage.Storage, r route.Route) ClusterDriver {
	return &MesosDriver{metric: m, config: config, storage: s, route: r}
}

func (md *MesosDriver) GetCollectorTypeName() string {
	return "application"
}

func (md *MesosDriver) GetIPMeta() (map[string]btypes.ObjectMeta, error) {
	data, err := md.storage.QueryDynamic(&storage.Param{Namespace: md.metric.Namespace, ClusterID: md.metric.ClusterID, ClusterType: types.ClusterMesos, Type: types.ClusterMesos.GetContainerTypeName()})
	if err != nil {
		blog.Error("failed to query taskgroup dynamic: %v", err)
		return nil, err
	}

	return GetIPMetaFromDynamic(data, md.metric)
}

func (md *MesosDriver) GetApplicationJson(imageBase string) (js *simplejson.Json, err error) {
	if js, err = LoadResourceJson(md.config.TempDir, md.metric.Namespace, types.ClusterMesos); err != nil {
		blog.Errorf("failed to load mesos template json: %v", err)
		return
	}

	name := GetApplicationName(md.metric)
	namespace := md.metric.Namespace
	clusterId := md.metric.ClusterID
	label := fmt.Sprintf("%s_%s", md.metric.Namespace, md.metric.Name)

	// metadata info
	js.Get("metadata").Set("name", name)
	js.Get("metadata").Set("namespace", namespace)
	js.Get("metadata").Get("labels").Set("io.tencent.bcs.metric.collector", label)

	// spec/template/metadata info
	js.GetPath("spec", "template", "metadata").Set("name", name)
	js.GetPath("spec", "template", "metadata").Set("namespace", namespace)
	js.GetPath("spec", "template", "metadata", "labels").Set("io.tencent.bcs.metric.collector", label)

	// spec/template/spec/containers info
	containerJs := js.GetPath("spec", "template", "spec", "containers")
	containers, err := containerJs.Array()
	if err != nil {
		blog.Errorf("mesos template json error: %v", err)
		return
	}

	envSettings := []btypes.EnvVar{
		{Name: "MetricClusterType", Value: types.ClusterMesos.String()},
		{Name: "MetricClusterID", Value: clusterId},
		{Name: "ZkAddress", Value: md.config.BCSZk},
		{Name: "MetricApplicationName", Value: name},
		{Name: "MetricApplicationNamespace", Value: namespace},
	}

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
		continue
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
	js.GetPath("spec", "template", "spec").Set("networkMode", md.metric.NetworkMode)
	js.GetPath("spec", "template", "spec").Set("networktype", md.metric.NetworkType)

	return
}

func (md *MesosDriver) CreateApplication(data []byte) (err error) {
	if err = md.route.CreateMesos(md.metric.ClusterID, md.metric.Namespace, data); err != nil {
		blog.Errorf("create mesos application failed: %v", err)
	}
	return
}

func (md *MesosDriver) DeleteApplication(data []byte) (err error) {
	if err = md.route.DeleteMesos(md.metric.ClusterID, md.metric.Namespace, GetApplicationName(md.metric), data); err != nil {
		blog.Errorf("delete mesos application failed: %v", err)
	}
	return
}
