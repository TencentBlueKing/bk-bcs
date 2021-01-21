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

package v4http

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"

	//"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
)

//CreateApplication create application implementation
func (s *Scheduler) CreateApplication(body []byte) (string, error) {
	blog.Info("create application. param(%s)", string(body))
	var param bcstype.ReplicaController
	//encoding param by json
	if err := json.Unmarshal(body, &param); err != nil {
		blog.Error("parse parameters failed. param(%s), err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr)
		return err.Error(), err
	}

	// bcs-mesos-scheduler version
	version, err := s.newVersionWithParam(&param)
	if err != nil {
		return err.Error(), err
	}

	//version.RawJson = &param
	// post version to bcs-mesos-scheduler, /v1/apps
	data, err := json.Marshal(version)
	if err != nil {
		blog.Error("marshal parameter version by json failed. err:%s", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+"encode version by json")
		return err.Error(), err
	}

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/apps"
	blog.Info("post a request to url(%s), request:%s", url, string(data))

	//reply, err := bhttp.Request(url, "POST", nil, strings.NewReader(string(data)))
	reply, err := s.client.POST(url, nil, data)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) newVersionWithParam(param *bcstype.ReplicaController) (*types.Version, error) {
	//check ObjectMeta is valid
	err := param.MetaIsValid()
	if err != nil {
		return nil, err
	}

	//var version types.Version
	version := &types.Version{
		Kind:        param.Kind,
		ID:          "",
		Instances:   0,
		RunAs:       "",
		Container:   []*types.Container{},
		Process:     []*bcstype.Process{},
		Labels:      make(map[string]string),
		Constraints: nil,
		Uris:        []string{},
		Ip:          []string{},
		Mode:        "",
	}

	//store ReplicaController original definition
	version.RawJson = param
	version.ObjectMeta = param.ObjectMeta
	version.KillPolicy = &param.KillPolicy

	version.RestartPolicy = &param.RestartPolicy
	if version.RestartPolicy.Policy == "" {
		version.RestartPolicy.Policy = bcstype.RestartPolicy_ONFAILURE
	}

	if version.RestartPolicy.Policy != bcstype.RestartPolicy_ONFAILURE && version.RestartPolicy.Policy != bcstype.RestartPolicy_ALWAYS && version.RestartPolicy.Policy != bcstype.RestartPolicy_NEVER {
		blog.Error("error restart policy: %s", version.RestartPolicy.Policy)
		replyErr := bhttp.InternalError(common.BcsErrMesosDriverParameterErr, common.BcsErrMesosDriverParameterErrStr+"restart policy error")
		return nil, replyErr
	}

	version.ID = param.Name
	version.RunAs = param.NameSpace
	if version.RunAs == "" {
		version.RunAs = "defaultGroup"
	}
	version.Instances = int32(param.ReplicaControllerSpec.Instance)
	version.Constraints = param.Constraints

	for k, v := range param.Labels {
		version.Labels[k] = v
	}

	version, err = s.setVersionWithPodSpec(version, param.ReplicaControllerSpec.Template)
	if err != nil {
		return nil, err
	}

	version.PodObjectMeta.NameSpace = param.ObjectMeta.NameSpace

	for k, v := range version.Labels {
		if strings.Contains(k, "io.tencent.bcs.netsvc.requestip.") {
			val := strings.Replace(v, " ", "", -1)
			version.Labels[k] = val
		}
	}
	for k, v := range version.ObjectMeta.Annotations {
		if strings.Contains(k, "io.tencent.bcs.netsvc.requestip.") {
			val := strings.Replace(v, " ", "", -1)
			version.ObjectMeta.Annotations[k] = val
		}
	}

	return version, nil
}

func (s *Scheduler) setVersionWithPodSpec(version *types.Version, spec *bcstype.PodTemplateSpec) (*types.Version, error) {
	if spec == nil {
		blog.Errorf("spec.template can't be empty")
		replyErr := bhttp.InternalError(common.BcsErrMesosDriverParameterErr, common.BcsErrMesosDriverParameterErrStr+" no template")
		return nil, replyErr
	}

	NumContainer := len(spec.PodSpec.Containers)
	NumProcess := len(spec.PodSpec.Processes)
	if NumContainer <= 0 && NumProcess <= 0 {
		blog.Warn("there is no container or Process parameters.")
		replyErr := bhttp.InternalError(common.BcsErrMesosDriverParameterErr, common.BcsErrMesosDriverParameterErrStr+"no containers and processes")
		return nil, replyErr
	}
	if NumContainer > 0 && NumProcess > 0 {
		blog.Warn("containers and Processes can not coexist.")
		replyErr := bhttp.InternalError(common.BcsErrMesosDriverParameterErr, common.BcsErrMesosDriverParameterErrStr+"containers and processes cannot coexist")
		return nil, replyErr
	}
	//version belong to application
	if version.Kind == "" && NumContainer > 0 {
		version.Kind = commtypes.BcsDataType_APP
	}
	//version belong to process
	if version.Kind == "" && NumProcess > 0 {
		version.Kind = commtypes.BcsDataType_PROCESS
	}

	version.PodObjectMeta = spec.ObjectMeta
	version.PodObjectMeta.NameSpace = version.ObjectMeta.NameSpace
	version.PodObjectMeta.Name = version.ObjectMeta.Name

	// Added  20180808, add labels from version.ObjectsMeta to version.PodObjectMeta
	// version.PodObjectMeta will effect the labels of taskgroups that store in zookeeper
	// And the existing label key in PodObjectMeta will not be recovered.
	if version.PodObjectMeta.Labels == nil {
		version.PodObjectMeta.Labels = version.ObjectMeta.Labels
	} else {
		for k, v := range version.ObjectMeta.Labels {
			if _, ok := version.PodObjectMeta.Labels[k]; !ok {
				version.PodObjectMeta.Labels[k] = v
			}
		}
	}

	for k, v := range spec.Labels {
		version.Labels[k] = v
	}

	for i := 0; i < NumProcess; i++ {
		process := &spec.PodSpec.Processes[i]
		version.Process = append(version.Process, process)
	}

	for i := 0; i < NumContainer; i++ {
		container := new(types.Container)
		c := spec.PodSpec.Containers[i]

		if c.Resources.Requests.Cpu == "" && c.Resources.Limits.Cpu != "" {
			c.Resources.Requests.Cpu = c.Resources.Limits.Cpu
			c.Resources.Requests.Mem = c.Resources.Limits.Mem
			c.Resources.Requests.Storage = c.Resources.Limits.Storage
		}

		container.Type = c.Type
		//Resources
		//request
		container.Resources = new(types.Resource)
		container.Resources.Cpus, _ = strconv.ParseFloat(c.Resources.Requests.Cpu, 64)
		container.Resources.Mem, _ = strconv.ParseFloat(c.Resources.Requests.Mem, 64)
		container.Resources.Disk, _ = strconv.ParseFloat(c.Resources.Requests.Storage, 64)
		//limit
		container.LimitResoures = new(types.Resource)
		container.LimitResoures.Cpus, _ = strconv.ParseFloat(c.Resources.Limits.Cpu, 64)
		container.LimitResoures.Mem, _ = strconv.ParseFloat(c.Resources.Limits.Mem, 64)
		container.LimitResoures.Disk, _ = strconv.ParseFloat(c.Resources.Limits.Storage, 64)
		container.DataClass = &types.DataClass{
			Resources: new(types.Resource),
			Msgs:      []*types.BcsMessage{},
		}
		//extended resources
		container.DataClass.ExtendedResources = c.Resources.ExtendedResources
		//request resources
		container.DataClass.Resources = container.Resources
		//limit resources
		container.DataClass.LimitResources = container.LimitResoures

		//set network flow limit parameters
		container.NetLimit = spec.PodSpec.NetLimit
		container.DataClass.NetLimit = container.NetLimit

		//docker
		container.Docker = new(types.Docker)
		container.Docker.Image = c.Image

		container.Docker.Hostname = c.Hostname
		container.Docker.ImagePullUser = c.ImagePullUser
		container.Docker.ImagePullPasswd = c.ImagePullPasswd

		container.Docker.ForcePullImage = false
		if c.ImagePullPolicy == bcstype.ImagePullPolicy_ALWAYS {
			container.Docker.ForcePullImage = true
		}

		container.Docker.Privileged = c.Privileged
		container.Docker.Network = spec.PodSpec.NetworkMode
		container.Docker.NetworkType = spec.PodSpec.NetworkType
		container.Docker.Command = c.Command
		container.Docker.Arguments = c.Args

		//parameter
		container.Docker.Parameters = []*types.Parameter{}
		for _, ps := range c.Parameters {
			container.Docker.Parameters = append(container.Docker.Parameters, &types.Parameter{Key: ps.Key, Value: ps.Value})
		}

		//portmaping
		container.Docker.PortMappings = []*types.PortMapping{}
		for _, port := range c.Ports {
			portMap := new(types.PortMapping)
			portMap.ContainerPort = int32(port.ContainerPort)

			portMap.HostPort = int32(port.HostPort)
			portMap.Name = port.Name
			portMap.Protocol = port.Protocol

			container.Docker.PortMappings = append(container.Docker.PortMappings, portMap)
		}

		//env
		container.Docker.Env = make(map[string]string)
		for _, env := range c.Env {
			if env.ValueFrom != nil && env.ValueFrom.ResourceFieldRef != nil && env.ValueFrom.ResourceFieldRef.Resource != "" {
				switch env.ValueFrom.ResourceFieldRef.Resource {
				case "requests.cpu":
					container.Docker.Env[env.Name] = fmt.Sprintf("%f", container.Resources.Cpus*1000)
				case "requests.memory":
					container.Docker.Env[env.Name] = fmt.Sprintf("%f", container.Resources.Mem)
				case "limits.cpu":
					container.Docker.Env[env.Name] = fmt.Sprintf("%f", container.LimitResoures.Cpus*1000)
				case "limits.memory":
					container.Docker.Env[env.Name] = fmt.Sprintf("%f", container.LimitResoures.Mem)
				default:
					blog.Errorf("Deployment(%s:%s) Env(%s) ValueFrom(%s) is invalid",
						version.ObjectMeta.NameSpace, version.ObjectMeta.Name, env.Name, env.ValueFrom.ResourceFieldRef.Resource)
				}
			} else {
				container.Docker.Env[env.Name] = env.Value
			}
		}

		//volume
		container.Volumes = []*types.Volume{}
		for _, volUnit := range c.Volumes {
			vol := new(types.Volume)
			vol.ContainerPath = volUnit.Volume.MountPath
			vol.HostPath = volUnit.Volume.HostPath
			vol.Mode = "RW"
			if volUnit.Volume.ReadOnly {
				vol.Mode = "R"
			}

			container.Volumes = append(container.Volumes, vol)
		}

		//configmap
		container.ConfigMaps = c.ConfigMaps

		//secret
		container.Secrets = c.Secrets

		container.HealthChecks = c.HealthChecks
		for _, oneCheck := range container.HealthChecks {
			if oneCheck.DelaySeconds <= 0 {
				oneCheck.DelaySeconds = 10
			}
			if oneCheck.IntervalSeconds <= 0 {
				oneCheck.IntervalSeconds = 60
			}
			if oneCheck.TimeoutSeconds <= 0 {
				oneCheck.TimeoutSeconds = 20
			}
			// if oneCheck.ConsecutiveFailures < 0 {
			// 	oneCheck.ConsecutiveFailures = 0
			// }
			if oneCheck.GracePeriodSeconds <= 0 {
				oneCheck.GracePeriodSeconds = 300
			}
		}

		version.Container = append(version.Container, container)
	} //end for i

	return version, nil
}
