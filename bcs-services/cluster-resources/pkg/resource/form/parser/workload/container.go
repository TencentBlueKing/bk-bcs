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

package workload

import (
	"github.com/mitchellh/mapstructure"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseContainerGroup ...
func ParseContainerGroup(manifest map[string]interface{}, cGroup *model.ContainerGroup) {
	// 初始容器
	if cs, _ := mapx.GetItems(manifest, "spec.template.spec.initContainers"); cs != nil {
		for _, c := range cs.([]interface{}) {
			cGroup.InitContainers = append(cGroup.InitContainers, parseContainer(c.(map[string]interface{})))
		}
	}
	// 标准容器
	if cs, _ := mapx.GetItems(manifest, "spec.template.spec.containers"); cs != nil {
		for _, c := range cs.([]interface{}) {
			cGroup.Containers = append(cGroup.Containers, parseContainer(c.(map[string]interface{})))
		}
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
	basic.PullPolicy = mapx.Get(raw, "imagePullPolicy", "").(string)
}

func parseContainerCommand(raw map[string]interface{}, command *model.ContainerCommand) {
	_ = mapstructure.Decode(raw, command)
}

func parseContainerService(raw map[string]interface{}, service *model.ContainerService) {
	_ = mapstructure.Decode(raw["ports"], &service.Ports)
}

func parseContainerEnvs(raw map[string]interface{}, cEnvs *model.ContainerEnvs) {
	// container.env
	if envs, ok := raw["env"]; ok {
		for _, env := range envs.([]interface{}) {
			e, _ := env.(map[string]interface{})
			if value, ok := e["value"]; ok {
				envVar := model.EnvVar{Name: e["name"].(string), Type: EnvVarTypeKeyVal, Value: value.(string)}
				cEnvs.Vars = append(cEnvs.Vars, envVar)
			} else if valFrom, ok := e["valueFrom"]; ok {
				envVar := genValueFormEnvVar(valFrom.(map[string]interface{}), e["name"].(string))
				cEnvs.Vars = append(cEnvs.Vars, envVar)
			}
		}
	}
	// container.envFrom
	if envFroms, ok := raw["envFrom"]; ok {
		for _, envFrom := range envFroms.([]interface{}) {
			envVar := genEnvFromEnvVar(envFrom.(map[string]interface{}))
			cEnvs.Vars = append(cEnvs.Vars, envVar)
		}
	}
}

func genValueFormEnvVar(valFrom map[string]interface{}, name string) model.EnvVar {
	var varType, value, source string
	if fieldRef, ok := valFrom["fieldRef"]; ok {
		// 来源于 Pod 本身字段信息
		varType = EnvVarTypePodField
		value = fieldRef.(map[string]interface{})["fieldPath"].(string)
	} else if resFieldRef, ok := valFrom["resourceFieldRef"]; ok {
		// 来源于资源配额信息
		varType = EnvVarTypeResource
		source = resFieldRef.(map[string]interface{})["containerName"].(string)
		value = resFieldRef.(map[string]interface{})["resource"].(string)
	} else if cmKeyRef, ok := valFrom["configMapKeyRef"]; ok {
		// 来源于 ConfigMap 键
		varType = EnvVarTypeCMKey
		source = cmKeyRef.(map[string]interface{})["name"].(string)
		value = cmKeyRef.(map[string]interface{})["key"].(string)
	} else if secRef, ok := valFrom["secretKeyRef"]; ok {
		// 来源于 Secret 键
		varType = EnvVarTypeSecretKey
		source = secRef.(map[string]interface{})["name"].(string)
		value = secRef.(map[string]interface{})["key"].(string)
	}
	return model.EnvVar{Name: name, Type: varType, Source: source, Value: value}
}

func genEnvFromEnvVar(envFrom map[string]interface{}) model.EnvVar {
	envVar := model.EnvVar{Name: envFrom["prefix"].(string)}
	if cmRef, ok := envFrom["configMapRef"]; ok {
		// 来源于 ConfigMap
		envVar.Type = EnvVarTypeCM
		envVar.Source = cmRef.(map[string]interface{})["name"].(string)
	} else if secRef, ok := envFrom["secretRef"]; ok {
		// 来源于 Secret
		envVar.Type = EnvVarTypeSecret
		envVar.Source = secRef.(map[string]interface{})["name"].(string)
	}
	return envVar
}

func parseContainerHealthz(raw map[string]interface{}, healthz *model.ContainerHealthz) {
	if readinessProbe, ok := raw["readinessProbe"]; ok {
		parseProbe(readinessProbe.(map[string]interface{}), &healthz.ReadinessProbe)
	}
	if livenessProbe, ok := raw["livenessProbe"]; ok {
		parseProbe(livenessProbe.(map[string]interface{}), &healthz.LivenessProbe)
	}
}

func parseProbe(raw map[string]interface{}, probe *model.Probe) {
	probe.PeriodSecs = mapx.Get(raw, "periodSeconds", int64(0)).(int64)
	probe.InitialDelaySecs = mapx.Get(raw, "initialDelaySeconds", int64(0)).(int64)
	probe.TimeoutSecs = mapx.Get(raw, "timeoutSeconds", int64(0)).(int64)
	probe.SuccessThreshold = mapx.Get(raw, "successThreshold", int64(0)).(int64)
	probe.FailureThreshold = mapx.Get(raw, "failureThreshold", int64(0)).(int64)
	if httpGet, ok := raw["httpGet"]; ok {
		probe.Type = ProbeTypeHTTPGet
		probe.Path = httpGet.(map[string]interface{})["path"].(string)
		probe.Port = httpGet.(map[string]interface{})["port"].(int64)
	} else if tcpSocket, ok := raw["tcpSocket"]; ok {
		probe.Type = ProbeTypeTCPSocket
		probe.Port = tcpSocket.(map[string]interface{})["port"].(int64)
	} else if exec, ok := raw["exec"]; ok {
		probe.Type = ProbeTypeExec
		for _, command := range exec.(map[string]interface{})["command"].([]interface{}) {
			probe.Command = append(probe.Command, command.(string))
		}
	}
}

func parseContainerRes(raw map[string]interface{}, res *model.ContainerRes) {
	res.Requests.CPU = util.ConvertCPUUnit(mapx.Get(raw, "resources.requests.cpu", "").(string))
	res.Requests.Memory = util.ConvertMemoryUnit(mapx.Get(raw, "resources.requests.memory", "").(string))
	res.Limits.CPU = util.ConvertCPUUnit(mapx.Get(raw, "resources.limits.cpu", "").(string))
	res.Limits.Memory = util.ConvertMemoryUnit(mapx.Get(raw, "resources.limits.memory", "").(string))
}

func parseContainerSecurity(raw map[string]interface{}, security *model.SecurityCtx) {
	_ = mapstructure.Decode(raw["securityContext"], security)
}

func parseContainerMount(raw map[string]interface{}, mount *model.ContainerMount) {
	_ = mapstructure.Decode(raw["volumeMounts"], &mount.Volumes)
}
