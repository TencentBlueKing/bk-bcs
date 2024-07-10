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

package renderer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/renderer/testdata/formdata"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/path"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

var deployManifest4RenderTest = map[string]interface{}{
	"apiVersion": "",
	"kind":       "Deployment",
	"metadata": map[string]interface{}{
		"name":      "deployment-test-" + stringx.Rand(example.RandomSuffixLength, example.SuffixCharset),
		"namespace": envs.TestNamespace,
		"labels": map[string]interface{}{
			"app": "busybox",
		},
	},
	"spec": map[string]interface{}{
		"replicas": int64(2),
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"app": "busybox",
			},
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "busybox",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "busybox",
						"image": "busybox:latest",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": int64(80),
							},
						},
						"command": []interface{}{
							"/bin/sh",
							"-c",
						},
						"args": []interface{}{
							"echo hello",
						},
					},
				},
			},
		},
	},
}

func TestManifestRenderer(t *testing.T) {
	formData := workload.ParseDeploy(deployManifest4RenderTest)
	ctx := context.WithValue(context.TODO(), ctxkey.UsernameKey, envs.AnonymousUsername)
	manifest, err := NewManifestRenderer(
		ctx, formData, envs.TestClusterID, "", resCsts.Deploy, resCsts.UpdateAction, false,
	).Render()
	assert.Nil(t, err)

	assert.Equal(t, "busybox", mapx.GetStr(manifest, "metadata.labels.app"))
	assert.Equal(t, 2, mapx.Get(manifest, "spec.replicas", 0))
	assert.Equal(t, "busybox", mapx.GetStr(manifest, "spec.selector.matchLabels.app"))

	// 注入信息检查
	assert.Equal(t, "apps/v1", mapx.GetStr(manifest, "apiVersion"))

	assert.Equal(t, resCsts.EditModeForm, mapx.GetStr(
		manifest, []string{"metadata", "annotations", resCsts.EditModeAnnoKey},
	))
}

type manifestRenderTestData struct {
	filePath   string
	formData   interface{}
	applyToK8s bool
}

var testCaseData = []manifestRenderTestData{
	// 工作负载类
	{"workload/deploy_complex.yaml", formdata.DeployComplex, true},
	{"workload/deploy_simple.yaml", formdata.DeploySimple, true},
	{"workload/sts_complex.yaml", formdata.STSComplex, true},
	{"workload/ds_complex.yaml", formdata.DSComplex, true},
	{"workload/cj_complex.yaml", formdata.CJComplex, true},
	{"workload/job_complex.yaml", formdata.JobComplex, true},
	{"workload/po_complex.yaml", formdata.PodComplex, true},
	// 网络类
	{"network/ing_v1.yaml", formdata.IngV1, true},
	{"network/ing_v1beta1.yaml", formdata.IngV1beta1, false},
	{"network/svc_complex.yaml", formdata.SVCComplex, true},
	{"network/ep_complex.yaml", formdata.EPComplex, true},
	// 配置类
	{"config/cm_complex.yaml", formdata.CMComplex, true},
	{"config/secret_opaque.yaml", formdata.SecretOpaque, true},
	{"config/secret_docker.yaml", formdata.SecretSocker, true},
	{"config/secret_basic_auth.yaml", formdata.SecretBasicAuth, true},
	{"config/secret_ssh_auth.yaml", formdata.SecretSSHAuth, true},
	{"config/secret_tls.yaml", formdata.SecretTLS, true},
	{"config/secret_sa_token.yaml", formdata.SecretSAToken, false},
	// 存储类
	{"storage/pv_complex.yaml", formdata.PVComplex, false},
	{"storage/pvc_complex.yaml", formdata.PVCComplex, true},
	{"storage/sc_complex.yaml", formdata.SCComplex, true},
	// HPA
	{"hpa/hpa_complex.yaml", formdata.HPAComplex, true},
	{"hpa/hpa_simple.yaml", formdata.HPASimple, true},
	// 自定义资源
	{"custom/gdeploy_complex.yaml", formdata.GDeployComplex, false},
	{"custom/gdeploy_simple.yaml", formdata.GDeploySimple, false},
	{"custom/gsts_complex.yaml", formdata.GSTSComplex, false},
	{"custom/gsts_simple.yaml", formdata.GSTSSimple, false},
	{"custom/hook_tmpl_complex.yaml", formdata.HookTmplComplex, false},
}

func TestManifestRenderByPipe(t *testing.T) {
	ctx := handler.NewInjectedContext("", "", "")
	pathPrefix := path.GetCurPKGPath() + "/testdata/manifest/"
	clusterConf := res.NewClusterConf(envs.TestClusterID)

	// 考虑下拆分函数，逻辑太多了可读性差
	for _, data := range testCaseData {
		// 先加载预设的结果
		yamlFile, err := os.ReadFile(pathPrefix + data.filePath)
		assert.Nil(t, err, "load excepted manifest from file [%s] failed: %v", data.filePath, err)

		excepted := map[string]interface{}{}
		err = yaml.Unmarshal(yamlFile, &excepted)
		assert.Nil(t, err, "yaml unmarshal file [%s] content failed: %v", data.filePath, err)

		// 根据表单数据渲染的结果
		formDataMap := structs.Map(data.formData)
		resKind := mapx.GetStr(formDataMap, "metadata.kind")
		result, err := NewManifestRenderer(ctx, formDataMap, envs.TestClusterID, "", resKind, resCsts.UpdateAction,
			false).Render()
		assert.Nil(t, err, "kind [%s] manifest render failed: %v", resKind, err)

		// 做对比确保一致（在同步随机名称后）
		resName := mapx.GetStr(result, "metadata.name")
		_ = mapx.SetItems(excepted, "metadata.name", resName)
		assert.Equal(t, excepted, result)

		// 部分资源资源测试集群中可能不存在，不进行下发检查
		if !data.applyToK8s {
			continue
		}

		// 下发到集群，确保不会报错后删除
		apiVersion := mapx.GetStr(formDataMap, "metadata.apiVersion")
		k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, apiVersion)
		assert.Nil(t, err, "GetGroupVersionResource failed [apiVersion: %s, kind: %s]", apiVersion, resKind)

		namespace := mapx.GetStr(result, "metadata.namespace")
		resCli := cli.NewResClient(clusterConf, k8sRes)

		_, err = resCli.Create(ctx, result, namespace != "", metav1.CreateOptions{})
		assert.Nil(
			t, err, "create k8s res [apiVersion: %s, kind: %s, name: %s] failed: %v", apiVersion, resKind, resName, err,
		)

		// 获取集群中的数据，与下发的配置做对比
		var ret *unstructured.Unstructured
		ret, err = resCli.Get(ctx, namespace, resName, metav1.GetOptions{})
		assert.Nil(t, err, "kind [%s] manifest render failed: %v", resKind, err)
		for _, diffRet := range mapx.NewDiffer(result, ret.UnstructuredContent()).Do() {
			if !diffRetCanIgnored(diffRet) {
				assert.Fail(
					t, "some unignorable diff result find",
					"filepath: %s, action: %s, dotted: %s, oldVal: %v, newVal %v",
					data.filePath, diffRet.Action, diffRet.Dotted, diffRet.OldVal, diffRet.NewVal,
				)
			}
		}

		// 删除测试时创建的 k8s 资源
		_ = resCli.Delete(ctx, namespace, resName, metav1.DeleteOptions{})
	}
}

// 判断 Diff 结果中不相同的能否被忽略
func diffRetCanIgnored(ret mapx.DiffRet) bool {
	// 允许新增的字段，比如 status, k8s 资源默认值等等
	if ret.Action == mapx.ActionAdd {
		return true
	}
	// yaml.Unmarshal 中整数类型为 int，k8s api 中返回的是 in64，这里用字面值做比较，忽略类型的影响
	if fmt.Sprintf("%v", ret.OldVal) == fmt.Sprintf("%v", ret.NewVal) {
		return true
	}
	// cpu, memory 字段兼容单位转换
	if strings.HasSuffix(ret.Dotted, resCsts.MetricResCPU) &&
		util.ConvertCPUUnit(ret.OldVal.(string)) == util.ConvertCPUUnit(ret.NewVal.(string)) {
		return true
	}
	if strings.HasSuffix(ret.Dotted, resCsts.MetricResMem) &&
		util.ConvertMemoryUnit(ret.OldVal.(string)) == util.ConvertMemoryUnit(ret.NewVal.(string)) {
		return true
	}
	return false
}
