/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"os"
	"sync"
	"text/template"

	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/renderengine"
	"bk-bscp/pkg/common"
)

type engine struct {
	input  *renderengine.RenderInConf
	output *renderengine.RenderOutConf
	mutex  sync.Mutex
	wgroup sync.WaitGroup
}

func newEngine(input *renderengine.RenderInConf) *engine {
	return &engine{
		input: input,
		output: &renderengine.RenderOutConf{
			ErrCode: pbcommon.ErrCode_E_OK,
			ErrMsg:  "OK",
		},
		wgroup: sync.WaitGroup{},
	}
}

func (e *engine) add(ins *renderengine.RenderOutInstance) {
	e.mutex.Lock()
	e.output.Instances = append(e.output.Instances, ins)
	e.mutex.Unlock()
}

func mergeVars(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	vars := make(map[string]interface{})
	for k, v := range m1 {
		vars[k] = v
	}
	for k, v := range m2 {
		vars[k] = v
	}
	return vars
}

func (e *engine) realRender(cluster string, clusterLabels map[string]string, zone, index string, vars map[string]interface{}) {
	defer e.wgroup.Done()

	for key, value := range clusterLabels {
		vars[key] = value
	}

	t, err := template.New("").Parse(e.input.Template)
	if err != nil {
		renderengine.SetStdout(pbcommon.ErrCode_E_TPL_RENDER_FAILED, err.Error(), nil)
		os.Exit(0)
	}

	// the final configs content size may over the limit, don't block it here,
	// it would be checked at datamanager level.
	buffer := bytes.NewBuffer(nil)

	// rendering template.
	if err := t.Execute(buffer, vars); err != nil {
		renderengine.SetStdout(pbcommon.ErrCode_E_TPL_RENDER_FAILED, err.Error(), nil)
		os.Exit(0)
	}

	e.add(&renderengine.RenderOutInstance{
		Cluster:       cluster,
		ClusterLabels: clusterLabels,
		Zone:          zone,
		Index:         index,
		Content:       buffer.String(),
	})
}

func (e *engine) renderForZone(cluster string, clusterLabels map[string]string, z *renderengine.RenderInZone, vars map[string]interface{}) {
	tmpVars := common.MergeVars(vars, z.Vars)
	if len(z.Instances) == 0 {
		e.wgroup.Add(1)
		go e.realRender(cluster, clusterLabels, z.Zone, "", tmpVars)
	}
	for _, ins := range z.Instances {
		insVars := common.MergeVars(tmpVars, ins.Vars)
		e.wgroup.Add(1)
		go e.realRender(cluster, clusterLabels, z.Zone, ins.Index, insVars)
	}
}

func (e *engine) renderForCluster(c *renderengine.RenderInCluster, vars map[string]interface{}) {
	tmpVars := common.MergeVars(vars, c.Vars)
	if len(c.Zones) == 0 {
		e.wgroup.Add(1)
		go e.realRender(c.Cluster, c.ClusterLabels, "", "", tmpVars)
	}
	for _, z := range c.Zones {
		e.renderForZone(c.Cluster, c.ClusterLabels, z, tmpVars)
	}
}

func (e *engine) render() {
	if len(e.input.Clusters) != 0 {
		for _, c := range e.input.Clusters {
			e.renderForCluster(c, e.input.Vars)
		}
	}
	e.wgroup.Wait()
}

func main() {
	in, err := renderengine.LoadStdin()
	if err != nil {
		renderengine.SetStdout(pbcommon.ErrCode_E_TPL_RENDER_FAILED, err.Error(), nil)
	}

	if len(in.Clusters) == 0 {
		renderengine.SetStdout(pbcommon.ErrCode_E_TPL_RENDER_FAILED, "no cluster info", nil)
	}

	en := newEngine(in)
	en.render()

	renderengine.SetStdout(en.output.ErrCode, en.output.ErrMsg, en.output.Instances)
}
