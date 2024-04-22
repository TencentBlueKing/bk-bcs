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

package workload

import (
	"github.com/mitchellh/mapstructure"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseContainerGroup xxx
func ParseContainerGroup(manifest map[string]interface{}, cGroup *model.ContainerGroup) {
	prefix := "spec.template.spec."
	switch mapx.GetStr(manifest, "kind") {
	case resCsts.CJ:
		prefix = "spec.jobTemplate.spec.template.spec."
	case resCsts.Po:
		prefix = "spec."
	}
	// 初始容器
	for _, c := range mapx.GetList(manifest, prefix+"initContainers") {
		cGroup.InitContainers = append(cGroup.InitContainers, parseContainer(c.(map[string]interface{})))
	}
	// 标准容器
	for _, c := range mapx.GetList(manifest, prefix+"containers") {
		cGroup.Containers = append(cGroup.Containers, parseContainer(c.(map[string]interface{})))
	}
}

func parseContainer(raw map[string]interface{}) model.Container {
	c := model.Container{}
	parseContainerBasic(raw, &c.Basic)
	parseContainerCommand(raw, &c.Command)
	parseContainerService(raw, &c.Service)
	parseContainerEnvs(raw, &c.Envs)
	parseContainerHealthz(raw, &c.Healthz)
	parseContainerRes(raw, &c.Resource)
	parseContainerSecurity(raw, &c.Security)
	parseContainerMount(raw, &c.Mount)
	return c
}

func parseContainerBasic(raw map[string]interface{}, basic *model.ContainerBasic) {
	basic.Name = raw["name"].(string)
	basic.Image = raw["image"].(string)
	basic.PullPolicy = mapx.GetStr(raw, "imagePullPolicy")
}

func parseContainerCommand(raw map[string]interface{}, command *model.ContainerCommand) {
	_ = mapstructure.Decode(raw, command)
}

func parseContainerService(raw map[string]interface{}, service *model.ContainerService) {
	_ = mapstructure.Decode(raw["ports"], &service.Ports)
}

func parseContainerEnvs(raw map[string]interface{}, cEnvs *model.ContainerEnvs) {
	// container.env
	for _, env := range mapx.GetList(raw, "env") {
		e, _ := env.(map[string]interface{})
		if value, ok := e["value"]; ok {
			envVar := model.EnvVar{Name: e["name"].(string), Type: resCsts.EnvVarTypeKeyVal, Value: value.(string)}
			cEnvs.Vars = append(cEnvs.Vars, envVar)
		} else if valFrom, ok := e["valueFrom"]; ok {
			envVar := genValueFormEnvVar(valFrom.(map[string]interface{}), e["name"].(string))
			cEnvs.Vars = append(cEnvs.Vars, envVar)
		}
	}

	// container.envFrom
	for _, envFrom := range mapx.GetList(raw, "envFrom") {
		envVar := genEnvFromEnvVar(envFrom.(map[string]interface{}))
		cEnvs.Vars = append(cEnvs.Vars, envVar)
	}
}

func genValueFormEnvVar(valFrom map[string]interface{}, name string) model.EnvVar {
	var varType, value, source string
	if fieldRef, ok := valFrom["fieldRef"]; ok {
		// 来源于 Pod 本身字段信息
		varType = resCsts.EnvVarTypePodField
		value = fieldRef.(map[string]interface{})["fieldPath"].(string)
	} else if resFieldRef, ok := valFrom["resourceFieldRef"]; ok {
		// 来源于资源配额信息
		varType = resCsts.EnvVarTypeResource
		source = resFieldRef.(map[string]interface{})["containerName"].(string)
		value = resFieldRef.(map[string]interface{})["resource"].(string)
	} else if cmKeyRef, ok := valFrom["configMapKeyRef"]; ok {
		// 来源于 ConfigMap 键
		varType = resCsts.EnvVarTypeCMKey
		source = cmKeyRef.(map[string]interface{})["name"].(string)
		value = cmKeyRef.(map[string]interface{})["key"].(string)
	} else if secRef, ok := valFrom["secretKeyRef"]; ok {
		// 来源于 Secret 键
		varType = resCsts.EnvVarTypeSecretKey
		source = secRef.(map[string]interface{})["name"].(string)
		value = secRef.(map[string]interface{})["key"].(string)
	}
	return model.EnvVar{Name: name, Type: varType, Source: source, Value: value}
}

func genEnvFromEnvVar(envFrom map[string]interface{}) model.EnvVar {
	envVar := model.EnvVar{Name: envFrom["prefix"].(string)}
	if cmRef, ok := envFrom["configMapRef"]; ok {
		// 来源于 ConfigMap
		envVar.Type = resCsts.EnvVarTypeCM
		envVar.Source = cmRef.(map[string]interface{})["name"].(string)
	} else if secRef, ok := envFrom["secretRef"]; ok {
		// 来源于 Secret
		envVar.Type = resCsts.EnvVarTypeSecret
		envVar.Source = secRef.(map[string]interface{})["name"].(string)
	}
	return envVar
}

func parseContainerHealthz(raw map[string]interface{}, healthz *model.ContainerHealthz) {
	if readinessProbe, ok := raw["readinessProbe"]; ok {
		parseProbe(readinessProbe.(map[string]interface{}), &healthz.ReadinessProbe)
	} else {
		// 默认不启用探针，会预设初始的延时，成功失败阈值等
		setDefaultProbe(&healthz.ReadinessProbe)
	}
	if livenessProbe, ok := raw["livenessProbe"]; ok {
		parseProbe(livenessProbe.(map[string]interface{}), &healthz.LivenessProbe)
	} else {
		setDefaultProbe(&healthz.LivenessProbe)
	}
}

func parseProbe(raw map[string]interface{}, probe *model.Probe) {
	probe.PeriodSecs = mapx.GetInt64(raw, "periodSeconds")
	probe.InitialDelaySecs = mapx.GetInt64(raw, "initialDelaySeconds")
	probe.TimeoutSecs = mapx.GetInt64(raw, "timeoutSeconds")
	probe.SuccessThreshold = mapx.GetInt64(raw, "successThreshold")
	probe.FailureThreshold = mapx.GetInt64(raw, "failureThreshold")
	if httpGet, ok := raw["httpGet"]; ok {
		probe.Enabled = true
		probe.Type = resCsts.ProbeTypeHTTPGet
		probe.Path = httpGet.(map[string]interface{})["path"].(string)
		probe.Port = httpGet.(map[string]interface{})["port"].(int64)
	} else if tcpSocket, ok := raw["tcpSocket"]; ok {
		probe.Enabled = true
		probe.Type = resCsts.ProbeTypeTCPSocket
		probe.Port = tcpSocket.(map[string]interface{})["port"].(int64)
	} else if exec, ok := raw["exec"]; ok {
		probe.Enabled = true
		probe.Type = resCsts.ProbeTypeExec
		for _, command := range mapx.GetList(exec.(map[string]interface{}), "command") {
			probe.Command = append(probe.Command, command.(string))
		}
	}
}

// setDefaultProbe 预设探针默认值，但是不会启用
func setDefaultProbe(probe *model.Probe) {
	probe.Enabled = false
	probe.Type = resCsts.ProbeTypeHTTPGet
	probe.PeriodSecs = 10
	probe.InitialDelaySecs = 0
	probe.TimeoutSecs = 1
	probe.SuccessThreshold = 1
	probe.FailureThreshold = 3
}

func parseContainerRes(raw map[string]interface{}, res *model.ContainerRes) {
	res.Requests.CPU = util.ConvertCPUUnit(mapx.GetStr(raw, "resources.requests.cpu"))
	res.Requests.Memory = util.ConvertMemoryUnit(mapx.GetStr(raw, "resources.requests.memory"))
	res.Requests.EphemeralStorage = util.ConvertStorageUnit(mapx.GetStr(raw, "resources.requests.ephemeral-storage"))
	res.Requests.Extra = genResExtra(mapx.GetMap(raw, "resources.requests"))
	res.Limits.CPU = util.ConvertCPUUnit(mapx.GetStr(raw, "resources.limits.cpu"))
	res.Limits.Memory = util.ConvertMemoryUnit(mapx.GetStr(raw, "resources.limits.memory"))
	res.Limits.EphemeralStorage = util.ConvertStorageUnit(mapx.GetStr(raw, "resources.limits.ephemeral-storage"))
	res.Limits.Extra = genResExtra(mapx.GetMap(raw, "resources.limits"))
}

func genResExtra(requirement map[string]interface{}) []model.ResExtra {
	extra := []model.ResExtra{}
	for key, value := range requirement {
		// cpu, memory 作为固定的指标，不会加入到 extra
		if key == resCsts.MetricResCPU || key == resCsts.MetricResMem || key == resCsts.MetricResEphemeralStorage {
			continue
		}
		extra = append(extra, model.ResExtra{
			Key: key, Value: value.(string),
		})
	}
	return extra
}

func parseContainerSecurity(raw map[string]interface{}, security *model.SecurityCtx) {
	_ = mapstructure.Decode(raw["securityContext"], security)
}

func parseContainerMount(raw map[string]interface{}, mount *model.ContainerMount) {
	_ = mapstructure.Decode(raw["volumeMounts"], &mount.Volumes)
}
