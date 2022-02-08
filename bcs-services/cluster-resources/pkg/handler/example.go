/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package handler example.go K8S 资源配置示例接口实现
package handler

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// GetK8SResTemplate ...
func (crh *ClusterResourcesHandler) GetK8SResTemplate(
	_ context.Context, req *clusterRes.GetK8SResTemplateReq, resp *clusterRes.CommonResp,
) (err error) {
	if !util.StringInSlice(req.Kind, example.HasDemoManifestResKinds) {
		return fmt.Errorf("资源类型 %s 暂无参考示例", req.Kind)
	}
	conf, err := example.LoadResConf(req.Kind)
	if err != nil {
		return err
	}
	conf["references"], err = example.LoadResRefs(req.Kind)
	if err != nil {
		return err
	}
	for _, t := range conf["items"].([]interface{}) {
		t, _ := t.(map[interface{}]interface{})
		t = make(map[interface{}]interface{})
		t["manifest"], _ = example.LoadDemoManifest(fmt.Sprintf("%s/%s", conf["class"], t["name"]))
	}
	resp.Data, err = util.Map2pbStruct(conf)
	return err
}
