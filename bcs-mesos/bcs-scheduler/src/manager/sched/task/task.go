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

package task

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	offerP "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	"github.com/golang/protobuf/proto"
)

// ReBuildTaskGroupID Build a new taskgroup ID from an old grouptask ID: the ID format is index.application.namespace.cluster.timestamp
// The new ID is different from old ID in timestamp.
func ReBuildTaskGroupID(taskGroupID string) (string, error) {
	splitID := strings.Split(taskGroupID, ".")
	if len(splitID) < 4 {
		return "", fmt.Errorf("taskGroupID %s format error", taskGroupID)
	}

	appInstances := splitID[0]
	appID := splitID[1]
	appRunAs := splitID[2]
	appClusterID := splitID[3]
	idTime := time.Now().UnixNano()
	newTaskGroupID := fmt.Sprintf("%s.%s.%s.%s.%d", appInstances, appID, appRunAs, appClusterID, idTime)

	return newTaskGroupID, nil
}

// CreateTaskGroup Create a taskgroup for an application, there are two methods to create a taskgroup:
// 1: you can create a taskgroup with the input of version(application definition), ID(taskgroup ID), reason and store
// 2: you can create a taskgroup with the input of version(application definition), appInstances, appClusterID, reason and store.
// taskgroup ID will be appInstances.$appname(version).$namespace(version).appClusterID.$timestamp
func CreateTaskGroup(version *types.Version, ID string, appInstances uint64, appClusterID string, reason string, store store.Store) (*types.TaskGroup, error) {
	appID := version.ID
	runAs := version.RunAs
	var taskgroup types.TaskGroup
	buildID := false
	if ID == "" {
		buildID = true
	}
	idTime := time.Now().UnixNano()
	generateTaskID := func(index int) (taskID, taskInstance string) {
		if buildID {
			taskID = fmt.Sprintf("%d.%d.%d.%s.%s.%s", idTime, index, appInstances, appID, runAs, appClusterID)
			taskInstance = strconv.Itoa(int(appInstances))
		} else {
			splitID := strings.Split(ID, ".")
			appInstances := splitID[0]
			appID := splitID[1]
			appRunAs := splitID[2]
			appClusterID := splitID[3]
			idTimeStr := splitID[4]
			taskID = fmt.Sprintf("%s.%d.%s.%s.%s.%s", idTimeStr, index, appInstances, appID, appRunAs, appClusterID)
			taskInstance = appInstances
		}
		return taskID, taskInstance
	}

	//branch for process or container
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		for index, process := range version.Process {
			var task types.Task
			// basic info
			task.Kind = version.Kind
			taskgroup.Kind = version.Kind
			task.ID, _ = generateTaskID(index)
			task.AppId = version.ID
			task.RunAs = version.RunAs
			task.Name = fmt.Sprintf("%d-%s", idTime, task.ID)
			task.Env = make(map[string]string)
			if process.Env != nil {
				for _, item := range process.Env {
					task.Env[item.Name] = item.Value
				}
			}
			if version.KillPolicy != nil {
				task.KillPolicy = version.KillPolicy
			}
			if version.Labels != nil {
				task.Labels = make(map[string]string)
				if task.Labels == nil {
					blog.Error("task.Labels == nil")
					return nil, fmt.Errorf("task.Labels == nil")
				}
				for k, v := range version.Labels {
					task.Labels[k] = v
				}
			}
			task.Network = "HOST"
			for _, port := range process.Ports {
				task.PortMappings = append(task.PortMappings, &types.PortMapping{
					HostPort: int32(port.HostPort),
					Name:     port.Name,
					Protocol: port.Protocol,
				})
			}
			// resources
			resource := &types.Resource{}
			resource.Cpus, _ = strconv.ParseFloat(process.Resources.Limits.Cpu, 64)
			resource.Mem, _ = strconv.ParseFloat(process.Resources.Limits.Mem, 64)
			resource.Disk, _ = strconv.ParseFloat(process.Resources.Limits.Storage, 64)
			// take new pointer to store uris
			uris := make([]*commtypes.Uri, 0)
			for _, uri := range process.Uris {
				newURI := new(commtypes.Uri)
				*newURI = *uri
				uris = append(uris, newURI)
			}
			// process info
			procInfo := &types.ProcDef{
				ProcName:         process.ProcName,
				WorkPath:         process.WorkPath,
				PidFile:          process.PidFile,
				StartCmd:         process.StartCmd,
				CheckCmd:         process.CheckCmd,
				StopCmd:          process.StopCmd,
				RestartCmd:       process.RestartCmd,
				ReloadCmd:        process.ReloadCmd,
				KillCmd:          process.KillCmd,
				LogPath:          process.LogPath,
				CfgPath:          process.CfgPath,
				StartGracePeriod: process.StartGracePeriod,
				Uris:             uris,
			}

			// dataClass
			task.DataClass = &types.DataClass{
				Resources: resource,
				Msgs:      []*types.BcsMessage{},
				ProcInfo:  procInfo,
			}
			if err := createTaskConfigMaps(&task, process.ConfigMaps, store); err != nil {
				return nil, err
			}
			if err := createTaskSecrets(&task, process.Secrets, store); err != nil {
				return nil, err
			}
			createTaskHealthChecks(&task, process.HealthChecks)
			task.Status = types.TASK_STATUS_STAGING
			task.UpdateTime = time.Now().Unix()

			taskgroup.Taskgroup = append(taskgroup.Taskgroup, &task)
		}
	case commtypes.BcsDataType_APP, "", commtypes.BcsDataType_Daemonset:
		// build container tasks
		for index, container := range version.Container {
			var task types.Task
			var taskInstance string
			if buildID {
				task.ID = fmt.Sprintf("%d.%d.%d.%s.%s.%s", idTime, index, appInstances, appID, runAs, appClusterID)
				taskInstance = strconv.Itoa(int(appInstances))
			} else {
				splitID := strings.Split(ID, ".")
				appInstances := splitID[0]
				appID := splitID[1]
				appRunAs := splitID[2]
				appClusterID := splitID[3]
				idTimeStr := splitID[4]
				task.ID = fmt.Sprintf("%s.%d.%s.%s.%s.%s", idTimeStr, index, appInstances, appID, appRunAs, appClusterID)
				taskInstance = appInstances
			}
			task.Kind = version.Kind
			taskgroup.Kind = version.Kind
			task.AppId = version.ID
			task.RunAs = version.RunAs
			task.Name = fmt.Sprintf("%d-%s", idTime, task.ID)
			task.Hostame = container.Docker.Hostname
			task.Image = container.Docker.Image
			task.ImagePullUser = container.Docker.ImagePullUser
			task.ImagePullPasswd = container.Docker.ImagePullPasswd
			task.Network = container.Docker.Network
			task.NetworkType = container.Docker.NetworkType
			task.NetLimit = container.NetLimit
			//task.Cpuset = container.Cpuset
			if container.Docker.Parameters != nil {
				for _, parameter := range container.Docker.Parameters {
					task.Parameters = append(task.Parameters, &types.Parameter{
						Key:   parameter.Key,
						Value: parameter.Value,
					})
				}
			}
			if container.Docker.PortMappings != nil {
				for _, portMapping := range container.Docker.PortMappings {
					task.PortMappings = append(task.PortMappings, &types.PortMapping{
						ContainerPort: portMapping.ContainerPort,
						HostPort:      portMapping.HostPort,
						Name:          portMapping.Name,
						Protocol:      portMapping.Protocol,
					})
				}
			}
			task.Privileged = container.Docker.Privileged
			task.ForcePullImage = container.Docker.ForcePullImage
			task.Env = container.Docker.Env //version.Env
			task.Volumes = container.Volumes
			task.Command = container.Docker.Command
			task.Arguments = container.Docker.Arguments
			task.DataClass = container.DataClass
			task.DataClass.Msgs = make([]*types.BcsMessage, 0)
			if err := createTaskConfigMaps(&task, container.ConfigMaps, store); err != nil {
				return nil, err
			}
			if err := createTaskSecrets(&task, container.Secrets, store); err != nil {
				return nil, err
			}
			requestIPLabel := "io.tencent.bcs.netsvc.requestip." + taskInstance
			if version.Labels != nil {
				task.Labels = make(map[string]string)
				if task.Labels == nil {
					blog.Error("task.Labels == nil")
					return nil, fmt.Errorf("task.Labels == nil")
				}
				for k, v := range version.Labels {
					if k == requestIPLabel {
						k = "io.tencent.bcs.netsvc.requestip"
						splitV := strings.Split(v, "|")
						if len(splitV) >= 1 {
							v = splitV[0]
						}

						blog.Info("task(%s) set io.tencent.bcs.netsvc.requestip = %s", task.ID, v)
					}
					task.Labels[k] = v
				}
			}
			if version.ObjectMeta.Annotations != nil {
				if task.Labels == nil {
					task.Labels = make(map[string]string)
				}
				for k, v := range version.ObjectMeta.Annotations {
					if k == requestIPLabel {
						k = "io.tencent.bcs.netsvc.requestip"
						splitV := strings.Split(v, "|")
						if len(splitV) >= 1 {
							v = splitV[0]
						}

						blog.Info("task(%s) set io.tencent.bcs.netsvc.requestip = %s", task.ID, v)
					}
					task.Labels[k] = v
				}
			}

			task.Status = types.TASK_STATUS_STAGING
			task.UpdateTime = time.Now().Unix()
			if version.KillPolicy != nil {
				task.KillPolicy = version.KillPolicy
			}
			createTaskHealthChecks(&task, container.HealthChecks)

			taskgroup.Taskgroup = append(taskgroup.Taskgroup, &task)
		}
	}

	if buildID {
		taskgroup.ID = fmt.Sprintf("%d.%s.%s.%s.%d", appInstances, appID, runAs, appClusterID, idTime)
		taskgroup.InstanceID = appInstances
		taskgroup.Name = fmt.Sprintf("%s-%d", appID, appInstances)
	} else {
		taskgroup.ID = ID
		splitID := strings.Split(ID, ".")
		appInstances, _ := strconv.Atoi(splitID[0])
		taskgroup.InstanceID = uint64(appInstances)
		taskgroup.Name = fmt.Sprintf("%s-%s", splitID[1], splitID[0])
	}

	taskgroup.AppID = appID
	taskgroup.VersionName = version.Name
	taskgroup.RunAs = version.RunAs
	taskgroup.LaunchResource = version.AllResource()
	taskgroup.CurrResource = version.AllResource()
	taskgroup.Status = types.TASKGROUP_STATUS_STAGING
	taskgroup.UpdateTime = time.Now().Unix()
	taskgroup.LastUpdateTime = time.Now().Unix()
	taskgroup.ReschededTimes = 0
	taskgroup.LastReschedTime = 0
	taskgroup.ObjectMeta = version.PodObjectMeta
	taskgroup.ObjectMeta.Name = taskgroup.ID
	if version.KillPolicy != nil {
		taskgroup.KillPolicy = version.KillPolicy
	}

	taskgroup.RestartPolicy = version.RestartPolicy

	return &taskgroup, nil
}

// added  20180806, create configMaps for task
func createTaskConfigMaps(task *types.Task, configMaps []commtypes.ConfigMap, store store.Store) error {
	for _, configMap := range configMaps {
		blog.Info("configmap:%+v", configMap)
		configMapName := configMap.Name
		configMapNs := task.RunAs
		blog.V(3).Infof("to get bcsconfigmap(Namespace:%s Name:%s)", configMapNs, configMapName)
		bcsConfigMap, err := store.FetchConfigMap(configMapNs, configMapName)
		if err != nil {
			blog.Error("get bcsconfigmap(Namespace:%s Name:%s) err: %s", configMapNs, configMapName, err.Error())
			return fmt.Errorf("get bcsconfigmap(Namespace:%s Name:%s) err: %s", configMapNs, configMapName, err.Error())
		}
		if bcsConfigMap == nil {
			blog.Error("bcsconfigmap(Namespace:%s Name:%s) not exist", configMapNs, configMapName)
			return fmt.Errorf("bcsconfigmap(Namespace:%s Name:%s) not exist", configMapNs, configMapName)
		}

		for _, confItem := range configMap.Items {
			blog.Info("configmap item:%+v", confItem)
			bcsConfigItem, ok := bcsConfigMap.Data[confItem.DataKey]
			if ok == false {
				blog.Warn("bcsconfig item(key:%s) not exist in bcsconfig(%s, %s) ", confItem.DataKey, configMapNs, configMapName)
				continue
			}
			msg := new(types.BcsMessage)
			if confItem.Type == bcstype.DataUsageType_FILE {
				if bcsConfigItem.Type == bcstype.BcsConfigMapSourceType_FILE {
					msg.Type = types.Msg_LOCALFILE.Enum()
					msg.Local = new(types.Msg_LocalFile)
					msg.Local.To = proto.String(confItem.KeyOrPath)
					msg.Local.Base64 = proto.String(bcsConfigItem.Content)
					if confItem.ReadOnly == true {
						msg.Local.Right = proto.String("r")
					} else {
						msg.Local.Right = proto.String("rw")
					}
					msg.Local.User = proto.String(confItem.User)
				} else {
					msg.Type = types.Msg_REMOTE.Enum()
					msg.Remote = new(types.Msg_Remote)
					msg.Remote.To = proto.String(confItem.KeyOrPath)
					msg.Remote.From = proto.String(bcsConfigItem.Content)
					if confItem.ReadOnly == true {
						msg.Remote.Right = proto.String("r")
					} else {
						msg.Remote.Right = proto.String("rw")
					}
					msg.Remote.User = proto.String(confItem.User)
					msg.Remote.Type = proto.String(string(bcsConfigItem.Type))
					msg.Remote.RemoteUser = proto.String(bcsConfigItem.RemoteUser)
					msg.Remote.RemotePasswd = proto.String(bcsConfigItem.RemotePasswd)
				}
			} else if confItem.Type == bcstype.DataUsageType_ENV {
				if bcsConfigItem.Type == bcstype.BcsConfigMapSourceType_FILE {
					msg.Type = types.Msg_ENV.Enum()
					msg.Env = new(types.Msg_Env)
					msg.Env.Name = proto.String(confItem.KeyOrPath)
					msg.Env.Value = proto.String(bcsConfigItem.Content)
				} else {
					msg.Type = types.Msg_ENV_REMOTE.Enum()
					msg.EnvRemote = new(types.Msg_EnvRemote)
					msg.EnvRemote.Name = proto.String(confItem.KeyOrPath)
					msg.EnvRemote.From = proto.String(bcsConfigItem.Content)
					msg.EnvRemote.Type = proto.String(string(bcsConfigItem.Type))
					msg.EnvRemote.RemoteUser = proto.String(bcsConfigItem.RemoteUser)
					msg.EnvRemote.RemotePasswd = proto.String(bcsConfigItem.RemotePasswd)
				}
			} else {
				blog.Warn("unkown configmap type:%s for task:%s", confItem.Type, task.ID)
				continue
			}
			by, _ := json.Marshal(msg)
			blog.Info("add task %s configmap message: %s", task.ID, string(by))
			task.DataClass.Msgs = append(task.DataClass.Msgs, msg)
			by, _ = json.Marshal(task.DataClass)
			blog.Infof("task %s dataclass(%s)", task.ID, string(by))
		}
	}
	return nil
}

// added  20180806, create secrets for task
func createTaskSecrets(task *types.Task, secrets []commtypes.Secret, store store.Store) error {
	for _, secret := range secrets {
		blog.Info("secret:%+v", secret)
		secretName := secret.SecretName
		secretNs := task.RunAs
		blog.V(3).Infof("to get bcssecret(Namespace:%s Name:%s)", secretNs, secretName)
		bcsSecret, err := store.FetchSecret(secretNs, secretName)
		if err != nil {
			blog.Error("get bcssecret(Namespace:%s Name:%s) err: %s", secretNs, secretName, err.Error())
			return fmt.Errorf("get bcssecret(Namespace:%s Name:%s) err: %s", secretNs, secretName, err.Error())
		}
		if bcsSecret == nil {
			blog.Error("bcssecret(Namespace:%s Name:%s) not exist", secretNs, secretName)
			return fmt.Errorf("bcssecret(Namespace:%s Name:%s) not exist", secretNs, secretName)
		}

		for _, secretItem := range secret.Items {
			blog.Info("secret item:%+v", secretItem)
			bcsSecretItem, ok := bcsSecret.Data[secretItem.DataKey]
			if ok == false {
				blog.Warn("bcssecret item(key:%s) not exist in bcssecret(%s, %s) ", secretItem.DataKey, secretNs, secretName)
				continue
			}
			msg := new(types.BcsMessage)
			msg.Type = types.Msg_SECRET.Enum()
			msg.Secret = new(types.Msg_Secret)
			switch secretItem.Type {
			case bcstype.DataUsageType_ENV:
				msg.Secret.Type = types.Secret_Env.Enum()
			case bcstype.DataUsageType_FILE:
				msg.Secret.Type = types.Secret_File.Enum()
			default:
				msg.Secret.Type = types.Secret_Unknown.Enum()
			}
			msg.Secret.Name = proto.String(secretItem.KeyOrPath)
			msg.Secret.Value = proto.String(bcsSecretItem.Content)
			blog.Info("add task secret message:%+v", msg)
			task.DataClass.Msgs = append(task.DataClass.Msgs, msg)
		}
	}
	return nil
}

// added  20180806, create healthChecks for task
func createTaskHealthChecks(task *types.Task, healthChecks []*commtypes.HealthCheck) {
	task.Healthy = true
	task.LocalMaxConsecutiveFailures = 0
	excutorCheckNum := 0
	remoteHTTPCheckNum := 0
	remoteTCPCheckNum := 0
	for _, healthCheck := range healthChecks {
		switch healthCheck.Type {
		case bcstype.BcsHealthCheckType_COMMAND:
			excutorCheckNum++
			if excutorCheckNum <= 1 {
				task.HealthChecks = append(task.HealthChecks, healthCheck)
				healthStatus := new(bcstype.BcsHealthCheckStatus)
				healthStatus.Type = healthCheck.Type
				healthStatus.Result = true
				healthStatus.Message = "command check by executor"
				task.HealthCheckStatus = append(task.HealthCheckStatus, healthStatus)
				task.LocalMaxConsecutiveFailures = healthCheck.ConsecutiveFailures
			}
		case bcstype.BcsHealthCheckType_TCP:
			excutorCheckNum++
			if excutorCheckNum <= 1 {
				task.HealthChecks = append(task.HealthChecks, healthCheck)
				healthStatus := new(bcstype.BcsHealthCheckStatus)
				healthStatus.Type = healthCheck.Type
				healthStatus.Result = true
				healthStatus.Message = "tcp check by executor"
				task.HealthCheckStatus = append(task.HealthCheckStatus, healthStatus)
				task.LocalMaxConsecutiveFailures = healthCheck.ConsecutiveFailures
			}
		case bcstype.BcsHealthCheckType_HTTP:
			excutorCheckNum++
			if excutorCheckNum <= 1 {
				task.HealthChecks = append(task.HealthChecks, healthCheck)
				healthStatus := new(bcstype.BcsHealthCheckStatus)
				healthStatus.Type = healthCheck.Type
				healthStatus.Result = true
				healthStatus.Message = "http check by executor"
				task.HealthCheckStatus = append(task.HealthCheckStatus, healthStatus)
				task.LocalMaxConsecutiveFailures = healthCheck.ConsecutiveFailures
			}
		case bcstype.BcsHealthCheckType_REMOTEHTTP:
			remoteHTTPCheckNum++
			if remoteHTTPCheckNum <= 1 {
				task.HealthChecks = append(task.HealthChecks, healthCheck)
				healthStatus := new(bcstype.BcsHealthCheckStatus)
				healthStatus.Type = healthCheck.Type
				healthStatus.Result = true
				healthStatus.Message = "remote http check, not reported"
				task.HealthCheckStatus = append(task.HealthCheckStatus, healthStatus)
			}
		case bcstype.BcsHealthCheckType_REMOTETCP:
			remoteTCPCheckNum++
			if remoteTCPCheckNum <= 1 {
				task.HealthChecks = append(task.HealthChecks, healthCheck)
				healthStatus := new(bcstype.BcsHealthCheckStatus)
				healthStatus.Type = healthCheck.Type
				healthStatus.Result = true
				healthStatus.Message = "remote tcp check, not reported"
				task.HealthCheckStatus = append(task.HealthCheckStatus, healthStatus)
			}
		default:
			blog.Info("task(%s) healthcheck(%s) not supported", task.ID, healthCheck.Type)
		}
	}
}

//add branch for process task, to do 20180802
// #lizard forgives createContainerTaskInfo
func createContainerTaskInfo(offer *mesos.Offer, resources []*mesos.Resource, task *types.Task, portUsed int) (*mesos.TaskInfo, int) {
	blog.V(3).Infof("Prepared task for launch with offer %s from %s", *offer.GetId().Value, offer.GetHostname())
	taskInfo := mesos.TaskInfo{
		Name: proto.String(task.Name),
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		AgentId:   offer.AgentId,
		Resources: resources,
		Command: &mesos.CommandInfo{
			Shell:     proto.Bool(false),
			Value:     proto.String(task.Command),
			Arguments: task.Arguments,
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_MESOS.Enum(),
			Mesos: &mesos.ContainerInfo_MesosInfo{
				Image: &mesos.Image{
					Type: mesos.Image_DOCKER.Enum(),
					Docker: &mesos.Image_Docker{
						Name: proto.String(task.Image),
					},
				},
			},
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image:        proto.String(task.Image),
				PortMappings: []*mesos.ContainerInfo_DockerInfo_PortMapping{},
				Parameters:   []*mesos.Parameter{},
			},
		},
	}

	taskInfo.Container.Docker.Privileged = &task.Privileged
	taskInfo.Container.Docker.ForcePullImage = &task.ForcePullImage

	if task.Network != "HOST" && task.Hostame != "" {
		hostname := fmt.Sprintf("%s-%d", task.Hostame, time.Now().UnixNano())
		taskInfo.Container.Hostname = &hostname
	}

	if task.KillPolicy != nil && task.KillPolicy.GracePeriod > 0 {
		durationS := time.Second * time.Duration(task.KillPolicy.GracePeriod)

		durationInfo := &mesos.DurationInfo{
			Nanoseconds: proto.Int64(int64(durationS)),
		}

		taskInfo.KillPolicy = &mesos.KillPolicy{
			GracePeriod: durationInfo,
		}
	}

	for _, parameter := range task.Parameters {
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters, &mesos.Parameter{
			Key:   proto.String(parameter.Key),
			Value: proto.String(parameter.Value),
		})
	}

	for _, volume := range task.Volumes {
		mode := mesos.Volume_RO
		if volume.Mode == "RW" {
			mode = mesos.Volume_RW
		}
		taskInfo.Container.Volumes = append(taskInfo.Container.Volumes, &mesos.Volume{
			ContainerPath: proto.String(volume.ContainerPath),
			HostPath:      proto.String(volume.HostPath),
			Mode:          &mode,
		})
	}

	renderAppTaskVarTemplate(task, offer)
	varEnvs := make([]*mesos.Environment_Variable, 0)
	for k, v := range task.Env {
		varEnvs = append(varEnvs, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	// add 190114 namespace and application name to container labels
	labelsMap := make(map[string]string)
	for k, v := range task.Labels {
		labelsMap[k] = v
	}
	labelsMap["namespace"] = task.RunAs
	//task.ID = 1536138501685462613.0.0.app-name.namespace.clusterid
	taskids := strings.Split(task.ID, ".")
	labelsMap["pod_name"] = fmt.Sprintf("%s-%s", taskids[3], taskids[2])

	labels := make([]*mesos.Label, 0)
	for k, v := range labelsMap {
		labels = append(labels, &mesos.Label{
			Key:   proto.String(k),
			Value: proto.String(v),
		})
	}
	taskInfo.Labels = &mesos.Labels{
		Labels: labels,
	}

	portNum := 0
	blog.V(3).Infof("task(%s) process Network: %s", task.ID, task.Network)
	switch task.Network {
	case "NONE":
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
		for _, m := range task.PortMappings {
			envName := "PORT" + "_" + m.Name
			envValue := strconv.Itoa(int(m.HostPort))
			blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
			varEnvs = append(varEnvs, &mesos.Environment_Variable{
				Name:  proto.String(envName),
				Value: proto.String(envValue),
			})
		}
	case "HOST":
		ports := getOfferPorts(offer)
		blog.V(3).Infof("offer(%s)(%s) port num(%d)", *offer.GetId().Value, offer.GetHostname(), len(ports))
		for idx, m := range task.PortMappings {
			blog.V(3).Infof("task(%s) Port setting[%d]: Network(%s), HostPort(%d), ContainerPort(%d), Protocol(%s) ",
				task.ID, idx, task.Network, m.HostPort, m.ContainerPort, m.Protocol)
			if m.ContainerPort < 0 {
				blog.Error("task(%s) Network error: ContainerPort(%d) error under Network(%s)",
					task.ID, m.ContainerPort, task.Network)
				continue
			}
			randomPort := false
			cPort := m.ContainerPort
			hPort := m.HostPort
			//random port, get from offer
			if m.ContainerPort == 0 {
				if len(ports) <= portUsed+portNum {
					blog.Error("task(%s) Port not enough: offerPorts(%d)<=used(%d)+currUsed(%d)", task.ID, len(ports), portUsed, portNum)
					return nil, portNum
				}
				cPort = int32(ports[portUsed+portNum])
				// set hostPort as the same as containerPort
				randomPort = true
				hPort = cPort
				portNum++
				blog.Info("task(%s) under HOST network, containerPort 0, so get Container and Host Port(%d) from offer", task.ID, cPort)
				m.ContainerPort = cPort
				m.HostPort = hPort
			}

			taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
				&mesos.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(hPort)),
					ContainerPort: proto.Uint32(uint32(cPort)),
					Protocol:      proto.String(m.Protocol),
				},
			)
			// set environmet PORTn
			envName := "PORT" + "_" + m.Name
			envValue := strconv.Itoa(int(cPort))
			blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
			varEnvs = append(varEnvs, &mesos.Environment_Variable{
				Name:  proto.String(envName),
				Value: proto.String(envValue),
			})

			if randomPort && hPort > 0 {
				taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
					Name: proto.String("ports"),
					Type: mesos.Value_RANGES.Enum(),
					Ranges: &mesos.Value_Ranges{
						Range: []*mesos.Value_Range{
							{
								Begin: proto.Uint64(uint64(hPort)),
								End:   proto.Uint64(uint64(hPort)),
							},
						},
					},
				})
			}
		}
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_HOST.Enum()
	case "USER":
		taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters,
			&mesos.Parameter{
				Key:   proto.String("net"),
				Value: proto.String(task.RunAs),
			})
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_USER.Enum()
		for _, m := range task.PortMappings {
			envName := "PORT" + "_" + m.Name
			envValue := strconv.Itoa(int(m.ContainerPort))
			blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
			varEnvs = append(varEnvs, &mesos.Environment_Variable{
				Name:  proto.String(envName),
				Value: proto.String(envValue),
			})
		}

	case "BRIDGE":
		ports := getOfferPorts(offer)
		blog.V(3).Infof("offer(%s)(%s) port num(%d)", *offer.GetId().Value, offer.GetHostname(), len(ports))
		for _, m := range task.PortMappings {
			blog.V(3).Infof("task(%s) Port setting: Network(%s), HostPort(%d), ContainerPort(%d), Protocol(%s) ",
				task.ID, task.Network, m.HostPort, m.ContainerPort, m.Protocol)
			cPort := m.ContainerPort
			hPort := m.HostPort
			randomPort := false
			//random port, get from offer
			if m.HostPort == 0 {
				if len(ports) <= portUsed+portNum {
					blog.Error("task(%s) Port not enough: offerPorts(%d), used(%d), I used(%d)", task.ID, len(ports), portUsed, portNum)
					return nil, portNum
				}
				hPort = int32(ports[portUsed+portNum])
				randomPort = true
				portNum++
				blog.Info("task(%s) under BRIDGE network, HostPort setting is 0, so get HostPort(%d) from offer", task.ID, hPort)

				// set environmet PORTn
				envName := "PORT" + "_" + m.Name
				envValue := strconv.Itoa(int(hPort))
				blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
				varEnvs = append(varEnvs, &mesos.Environment_Variable{
					Name:  proto.String(envName),
					Value: proto.String(envValue),
				})
			}
			if hPort < 0 {
				blog.V(3).Infof("task(%s) HostPort(%d) < 0, set to 0", task.ID, hPort)
				hPort = 0
			}
			//write back data to task
			m.ContainerPort = cPort
			m.HostPort = hPort
			taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
				&mesos.ContainerInfo_DockerInfo_PortMapping{
					HostPort:      proto.Uint32(uint32(hPort)),
					ContainerPort: proto.Uint32(uint32(cPort)),
					Protocol:      proto.String(m.Protocol),
				},
			)
			if randomPort && hPort > 0 {
				taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
					Name: proto.String("ports"),
					Type: mesos.Value_RANGES.Enum(),
					Ranges: &mesos.Value_Ranges{
						Range: []*mesos.Value_Range{
							{
								Begin: proto.Uint64(uint64(hPort)),
								End:   proto.Uint64(uint64(hPort)),
							},
						},
					},
				})
			}
		}
		taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_BRIDGE.Enum()
	default:
		if task.Network == "" {
			taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_NONE.Enum()
			for _, m := range task.PortMappings {
				envName := "PORT" /*+strconv.Itoa(int(idx)) */ + "_" + m.Name
				envValue := strconv.Itoa(int(m.HostPort))
				blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
				varEnvs = append(varEnvs, &mesos.Environment_Variable{
					Name:  proto.String(envName),
					Value: proto.String(envValue),
				})
			}
		} else if strings.ToLower(task.NetworkType) == "cnm" {
			ports := getOfferPorts(offer)
			blog.V(3).Infof("offer(%s)(%s) port num(%d)", *offer.GetId().Value, offer.GetHostname(), len(ports))
			for _, m := range task.PortMappings {
				blog.V(3).Infof("task(%s) Port setting: Network(%s), HostPort(%d), ContainerPort(%d), Protocol(%s) ",
					task.ID, task.Network, m.HostPort, m.ContainerPort, m.Protocol)
				cPort := m.ContainerPort
				hPort := m.HostPort
				randomPort := false
				//random port, get from offer
				if m.HostPort == 0 {
					if len(ports) <= portUsed+portNum {
						blog.Error("task(%s) Port not enough: offerPorts(%d), used(%d), I used(%d)", task.ID, len(ports), portUsed, portNum)
						return nil, portNum
					}
					hPort = int32(ports[portUsed+portNum])
					randomPort = true
					portNum++
					blog.Info("task(%s) under BRIDGE network, HostPort setting is 0, so get HostPort(%d) from offer", task.ID, hPort)
					envName := "PORT" + "_" + m.Name
					envValue := strconv.Itoa(int(hPort))
					blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
					varEnvs = append(varEnvs, &mesos.Environment_Variable{
						Name:  proto.String(envName),
						Value: proto.String(envValue),
					})
				}
				if hPort < 0 {
					blog.V(3).Infof("task(%s) HostPort(%d) < 0, set to 0", task.ID, hPort)
					hPort = 0
				}
				m.ContainerPort = cPort
				m.HostPort = hPort
				taskInfo.Container.Docker.PortMappings = append(taskInfo.Container.Docker.PortMappings,
					&mesos.ContainerInfo_DockerInfo_PortMapping{
						HostPort:      proto.Uint32(uint32(hPort)),
						ContainerPort: proto.Uint32(uint32(cPort)),
						Protocol:      proto.String(m.Protocol),
					},
				)
				if randomPort && hPort > 0 {
					taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
						Name: proto.String("ports"),
						Type: mesos.Value_RANGES.Enum(),
						Ranges: &mesos.Value_Ranges{
							Range: []*mesos.Value_Range{
								{
									Begin: proto.Uint64(uint64(hPort)),
									End:   proto.Uint64(uint64(hPort)),
								},
							},
						},
					})
				}
			}
			taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters,
				&mesos.Parameter{
					Key:   proto.String("net"),
					Value: proto.String(task.Network),
				})
			taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_USER.Enum()
		} else { // as USER
			taskInfo.Container.Docker.Parameters = append(taskInfo.Container.Docker.Parameters,
				&mesos.Parameter{
					Key:   proto.String("net"),
					Value: proto.String(task.Network),
				})
			taskInfo.Container.Docker.Network = mesos.ContainerInfo_DockerInfo_USER.Enum()
			for _, m := range task.PortMappings {
				envName := "PORT" + "_" + m.Name
				envValue := strconv.Itoa(int(m.ContainerPort))
				blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
				varEnvs = append(varEnvs, &mesos.Environment_Variable{
					Name:  proto.String(envName),
					Value: proto.String(envValue),
				})
			}
		}
	}

	blog.V(3).Infof("task(%s) set Environment num: %d", task.ID, len(varEnvs))
	taskInfo.Command.Environment = &mesos.Environment{
		Variables: varEnvs,
	}

	msgData, err := json.Marshal(task.DataClass)
	blog.V(3).Infof("task %s dataclass %s", task.ID, string(msgData))

	if err == nil {
		taskInfo.Data = []byte(base64.StdEncoding.EncodeToString(msgData))
	} else {
		blog.Warn("Prepared data of taskinfo is err, for %s", err.Error())
	}

	createTaskInfoHealth(task, &taskInfo)

	return &taskInfo, portNum
}

// added  180807, create TaskInfo for process. By default, the process network is always HOST, and the port can be
// specific or given by scheduler when port=0. The var template in Task.DataClass.ProDef and Task.Env will be rendered, and finally
// the taskgroup in zookeeper(or other storage) and the taskinfo down to mesos are all rendered.
func createProcessTaskInfo(offer *mesos.Offer, resources []*mesos.Resource, task *types.Task, portUsed int) (*mesos.TaskInfo, int) {
	blog.V(3).Infof("Prepared task for launch with offer %s from %s", *offer.GetId().Value, offer.GetHostname())
	taskInfo := mesos.TaskInfo{
		Name: proto.String(task.Name),
		TaskId: &mesos.TaskID{
			Value: proto.String(task.ID),
		},
		AgentId:   offer.AgentId,
		Resources: resources,
		Command:   &mesos.CommandInfo{},
	}

	if task.KillPolicy != nil && task.KillPolicy.GracePeriod > 0 {
		durationS := time.Second * time.Duration(task.KillPolicy.GracePeriod)

		durationInfo := &mesos.DurationInfo{
			Nanoseconds: proto.Int64(int64(durationS)),
		}

		taskInfo.KillPolicy = &mesos.KillPolicy{
			GracePeriod: durationInfo,
		}
	}

	if task.Labels != nil {
		labels := make([]*mesos.Label, 0)
		for k, v := range task.Labels {
			labels = append(labels, &mesos.Label{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}

		taskInfo.Labels = &mesos.Labels{
			Labels: labels,
		}
	}

	portNum := 0
	portEnvs := make([]*mesos.Environment_Variable, 0)
	ports := getOfferPorts(offer)
	blog.V(3).Infof("offer(%s)(%s) port num(%d)", *offer.GetId().Value, offer.GetHostname(), len(ports))

	for idx, m := range task.PortMappings {
		blog.V(3).Infof("task(%s) Port setting[%d]: Network(%s), HostPort(%d), ContainerPort(%d), Protocol(%s) ",
			task.ID, idx, task.Network, m.HostPort, m.ContainerPort, m.Protocol)

		randomPort := false
		hPort := m.HostPort
		if hPort <= 0 {
			if len(ports) <= portUsed+portNum {
				blog.Error("task(%s) Port not enough: offerPorts(%d)<=used(%d)+currUsed(%d)", task.ID, len(ports), portUsed, portNum)
				return nil, portNum
			}
			hPort = int32(ports[portUsed+portNum])
			randomPort = true
			portNum++
			blog.Info("task(%s) under HOST network, hostPort 0, so get Container and Host Port(%d) from offer", task.ID, hPort)
			// write back data to task
			// set containerPort as the same as hostPort
			m.HostPort = hPort
		}
		m.ContainerPort = hPort

		// set environmet PORTn
		envName := "PORT" + "_" + m.Name
		envValue := strconv.Itoa(int(hPort))
		blog.V(3).Infof("task(%s) set env(%s:%s)", task.ID, envName, envValue)
		portEnvs = append(portEnvs, &mesos.Environment_Variable{
			Name:  proto.String(envName),
			Value: proto.String(envValue),
		})

		if randomPort {
			taskInfo.Resources = append(taskInfo.Resources, &mesos.Resource{
				Name: proto.String("ports"),
				Type: mesos.Value_RANGES.Enum(),
				Ranges: &mesos.Value_Ranges{
					Range: []*mesos.Value_Range{
						{
							Begin: proto.Uint64(uint64(hPort)),
							End:   proto.Uint64(uint64(hPort)),
						},
					},
				},
			})
		}
	}

	// render var templates in process Task data
	renderProcessTaskVarTemplate(task, offer)
	varEnvs := make([]*mesos.Environment_Variable, 0)
	for k, v := range task.Env {
		varEnvs = append(varEnvs, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}
	// added  20180808, append portEnvs after user-specific env, to keep the same sequence as before.
	varEnvs = append(varEnvs, portEnvs...)

	blog.V(3).Infof("task(%s) set Environment num: %d", task.ID, len(varEnvs))
	taskInfo.Command.Environment = &mesos.Environment{
		Variables: varEnvs,
	}

	msgData, err := json.Marshal(task.DataClass)
	blog.V(3).Infof("task %s dataclass %s", task.ID, string(msgData))

	if err == nil {
		taskInfo.Data = []byte(base64.StdEncoding.EncodeToString(msgData))
	} else {
		blog.Warn("Prepared data of taskinfo is err, for %s", err.Error())
	}

	createTaskInfoHealth(task, &taskInfo)

	return &taskInfo, portNum
}

func createTaskInfoHealth(task *types.Task, taskInfo *mesos.TaskInfo) {

	for _, healthCheck := range task.HealthChecks {
		switch healthCheck.Type {
		case bcstype.BcsHealthCheckType_COMMAND:
			if healthCheck.Command == nil {
				blog.Error("task(%s) healthcheck(%s) data is nil", task.ID, healthCheck.Type)
				break
			}
			if healthCheck.Command.Value == "" {
				blog.Error("task(%s) healthcheck MesosCommand.Value empty", task.ID)
				break
			}
			taskInfo.HealthCheck = &mesos.HealthCheck{
				Type: mesos.HealthCheck_COMMAND.Enum(),
				Command: &mesos.CommandInfo{
					Value: proto.String(healthCheck.Command.Value),
				},
				DelaySeconds:        proto.Float64(float64(healthCheck.DelaySeconds)),
				IntervalSeconds:     proto.Float64(float64(healthCheck.IntervalSeconds)),
				TimeoutSeconds:      proto.Float64(float64(healthCheck.TimeoutSeconds)),
				ConsecutiveFailures: proto.Uint32(healthCheck.ConsecutiveFailures),
				GracePeriodSeconds:  proto.Float64(float64(healthCheck.GracePeriodSeconds)),
			}
		case bcstype.BcsHealthCheckType_TCP:
			if healthCheck.Tcp == nil {
				blog.Error("task(%s) healthcheck(%s) data is nil", task.ID, healthCheck.Type)
				break
			}
			checkPort := healthCheck.Tcp.Port
			if checkPort <= 0 {
				checkPort, _ = getTaskHealthCheckPort(task, healthCheck.Tcp.PortName)
				if checkPort <= 0 {
					blog.Error("task(%s) healthcheck(%s) no port", task.ID, healthCheck.Type)
					break
				}
			}
			taskInfo.HealthCheck = &mesos.HealthCheck{
				Type: mesos.HealthCheck_TCP.Enum(),
				Tcp: &mesos.HealthCheck_TCPCheckInfo{
					Port: proto.Uint32(uint32(checkPort)),
				},
				DelaySeconds:        proto.Float64(float64(healthCheck.DelaySeconds)),
				IntervalSeconds:     proto.Float64(float64(healthCheck.IntervalSeconds)),
				TimeoutSeconds:      proto.Float64(float64(healthCheck.TimeoutSeconds)),
				ConsecutiveFailures: proto.Uint32(healthCheck.ConsecutiveFailures),
				GracePeriodSeconds:  proto.Float64(float64(healthCheck.GracePeriodSeconds)),
			}
		case bcstype.BcsHealthCheckType_HTTP:
			if healthCheck.Http == nil {
				blog.Error("task(%s) healthcheck(%s) data is nil", task.ID, healthCheck.Type)
				break
			}
			checkPort := healthCheck.Http.Port
			if checkPort <= 0 {
				checkPort, _ = getTaskHealthCheckPort(task, healthCheck.Http.PortName)
				if checkPort <= 0 {
					blog.Error("task(%s) healthcheck(%s) no port", task.ID, healthCheck.Type)
					break
				}
			}
			taskInfo.HealthCheck = &mesos.HealthCheck{
				Type: mesos.HealthCheck_HTTP.Enum(),
				Http: &mesos.HealthCheck_HTTPCheckInfo{
					Port:   proto.Uint32(uint32(checkPort)),
					Path:   proto.String(healthCheck.Http.Path),
					Scheme: proto.String(healthCheck.Http.Scheme),
				},
				DelaySeconds:        proto.Float64(float64(healthCheck.DelaySeconds)),
				IntervalSeconds:     proto.Float64(float64(healthCheck.IntervalSeconds)),
				TimeoutSeconds:      proto.Float64(float64(healthCheck.TimeoutSeconds)),
				ConsecutiveFailures: proto.Uint32(healthCheck.ConsecutiveFailures),
				GracePeriodSeconds:  proto.Float64(float64(healthCheck.GracePeriodSeconds)),
			}
		default:
			blog.Info("task(%s) healthcheck(%s) is remote check or not supported", task.ID, healthCheck.Type)
		}
	}

	return
}

// added  180807, render the var templates in *types.Task, it will effect taskgroup and the following copy
// operations from taskgroup to taskinfo.
func renderProcessTaskVarTemplate(task *types.Task, offer *mesos.Offer) {
	m := make(map[string]string)

	// set application name
	m[types.TASK_TEMPLATE_KEY_PROCESSNAME] = task.AppId

	// set namespace
	m[types.TASK_TEMPLATE_KEY_NAMESPACE] = task.RunAs

	// set this taskgroup(which this task belong to) index in all instances of application
	m[types.TASK_TEMPLATE_KEY_INSTANCEID] = strings.Split(task.ID, ".")[2]

	// set InnerIP
	ip, ok := offerP.GetOfferIp(offer)
	if ok {
		m[types.TASK_TEMPLATE_KEY_HOSTIP] = ip
	}

	// set port by name
	for _, port := range task.PortMappings {
		key := fmt.Sprintf(types.TASK_TEMPLATE_KEY_PORT_FORMAT, port.Name)
		m[key] = fmt.Sprintf("%d", port.HostPort)
	}

	// render workPath first and then add it into vars for rendering others
	for k, v := range m {
		task.DataClass.ProcInfo.WorkPath = renderString(task.DataClass.ProcInfo.WorkPath, k, v)
		task.DataClass.ProcInfo.PidFile = renderString(task.DataClass.ProcInfo.PidFile, k, v)
	}
	m[types.TASK_TEMPLATE_KEY_WORKPATH] = task.DataClass.ProcInfo.WorkPath
	m[types.TASK_TEMPLATE_KEY_PIDFILE] = task.DataClass.ProcInfo.PidFile

	for k, v := range m {
		// render all keys and values in Env
		for envKey, envVal := range task.Env {
			newKey := renderString(envKey, k, v)
			newVal := renderString(envVal, k, v)
			delete(task.Env, envKey)
			task.Env[newKey] = newVal
		}

		// render process info
		var temp []byte
		_ = codec.EncJson(task.DataClass.ProcInfo, &temp)
		temp = []byte(renderString(string(temp), k, v))
		_ = codec.DecJson(temp, task.DataClass.ProcInfo)

		for _, msg := range task.DataClass.Msgs {
			switch *msg.Type {
			case types.Msg_LOCALFILE:
				*msg.Local.To = renderString(*msg.Local.To, k, v)
			case types.Msg_REMOTE:
				*msg.Remote.To = renderString(*msg.Remote.To, k, v)
			case types.Msg_ENV:
				*msg.Env.Name = renderString(*msg.Env.Name, k, v)
			case types.Msg_ENV_REMOTE:
				*msg.EnvRemote.Name = renderString(*msg.EnvRemote.Name, k, v)
			case types.Msg_SECRET:
				*msg.Secret.Name = renderString(*msg.Secret.Name, k, v)
			default:
				continue
			}
		}
	}
}

func renderAppTaskVarTemplate(task *types.Task, offer *mesos.Offer) {
	//task.ID = 1536138501685462613.0.0.app-name.namespace.clusterid
	taskids := strings.Split(task.ID, ".")
	timestamp := taskids[0]
	instanceid := taskids[2]
	appname := taskids[3]
	namespace := taskids[4]
	clusterid := taskids[5]

	m := make(map[string]string)

	// set application name
	m[types.APP_TASK_TEMPLATE_KEY_APPNAME] = task.AppId

	// set namespace
	m[types.APP_TASK_TEMPLATE_KEY_NAMESPACE] = task.RunAs

	// set this taskgroup(which this task belong to) index in all instances of application
	m[types.APP_TASK_TEMPLATE_KEY_INSTANCEID] = strings.Split(task.ID, ".")[2]

	// set InnerIP
	ip, ok := offerP.GetOfferIp(offer)
	if ok {
		m[types.APP_TASK_TEMPLATE_KEY_HOSTIP] = ip
	}

	//set taskgroup id
	m[types.APP_TASK_TEMPLATE_KEY_PODID] = fmt.Sprintf("%s.%s.%s.%s.%s", instanceid, appname, namespace, clusterid, timestamp)
	//set taskgroup name
	m[types.APP_TASK_TEMPLATE_KEY_PODNAME] = fmt.Sprintf("%s-%s", appname, instanceid)

	// set port by name
	for _, port := range task.PortMappings {
		key := fmt.Sprintf(types.APP_TASK_TEMPLATE_KEY_PORT_FORMAT, port.Name)
		m[key] = fmt.Sprintf("%d", port.HostPort)
	}

	for k, v := range m {
		// render all keys and values in Env
		for envKey, envVal := range task.Env {
			newKey := renderString(envKey, k, v)
			newVal := renderString(envVal, k, v)
			delete(task.Env, envKey)
			task.Env[newKey] = newVal
		}
	}
}

// Replace all K in s with v and return the replaced string r.
// K is the string after formatting k with TASK_TEMPLATE_KEY_FORMAT
func renderString(s, k, v string) (r string) {
	k = fmt.Sprintf(types.TASK_TEMPLATE_KEY_FORMAT, k)
	return strings.Replace(s, k, v, -1)
}

func getTaskHealthCheckPort(task *types.Task, name string) (int32, error) {
	blog.V(3).Infof("get health check port for name:%s, task:%s Network:%s", name, task.ID, task.Network)

	switch task.Network {
	case "NONE":
		for _, m := range task.PortMappings {
			if m.Name == name {
				return m.ContainerPort, nil
			}
		}
		return 0, fmt.Errorf("task(%s) has not portname(%s)", task.ID, name)
	case "HOST":
		for _, m := range task.PortMappings {
			if m.Name == name {
				return m.ContainerPort, nil
			}
		}
		return 0, fmt.Errorf("task(%s) has not portname(%s)", task.ID, name)
	case "USER":
		for _, m := range task.PortMappings {
			if m.Name == name {
				return m.ContainerPort, nil
			}
		}
		return 0, fmt.Errorf("task(%s) has not portname(%s)", task.ID, name)
	case "BRIDGE":
		for _, m := range task.PortMappings {
			if m.Name == name {
				return m.ContainerPort, nil
			}
		}
		return 0, fmt.Errorf("task(%s) has not portname(%s)", task.ID, name)
	default:
		for _, m := range task.PortMappings {
			if m.Name == name {
				return m.ContainerPort, nil
			}
		}
		return 0, fmt.Errorf("task(%s) has not portname(%s)", task.ID, name)
	}
}

// CreateTaskGroupInfo Create taskgroup information with offered resource
// the information include: ports, slave attributions, health-check information etc.
func CreateTaskGroupInfo(offer *mesos.Offer, version *types.Version,
	resources []*mesos.Resource, taskgroup *types.TaskGroup) *mesos.TaskGroupInfo {
	blog.Info("build taskgroup(%s) with offer %s||%s", taskgroup.ID, offer.GetHostname(), *offer.GetId().Value)

	taskgroup.AgentID = *offer.AgentId.Value
	taskgroup.HostName = offer.GetHostname()
	taskgroup.StartTime = time.Now().Unix()
	// build taskgroup's attributes according to version and offer attributes,
	// for UNIQUE and other constraints
	if version.Constraints != nil {
		for _, oneConstraint := range version.Constraints.IntersectionItem {
			if oneConstraint == nil {
				continue
			}
			for _, oneData := range oneConstraint.UnionData {
				if oneData == nil {
					continue
				}
				blog.V(3).Infof("version(RunAs:%s ID:%s), Constraint attribute(%s)",
					version.RunAs, version.ID, oneData.Name)
				// copy attribute from offer to taskgroup
				isIn := false
				for _, currAttribute := range taskgroup.Attributes {
					if currAttribute.GetName() == oneData.Name {
						isIn = true
						blog.V(3).Infof("attribute(%s) is already in taskgroup", oneData.Name)
						break
					}
				}
				if isIn == false {
					var attribute *mesos.Attribute
					if oneData.Name == "hostname" {
						blog.V(3).Infof("create attribute(%s) for taskgroup", oneData.Name)
						var attr mesos.Attribute
						var attrName = "hostname"
						attr.Name = &attrName
						var attrType mesos.Value_Type = mesos.Value_TEXT
						attr.Type = &attrType
						var attrValue mesos.Value_Text
						var host string = offer.GetHostname()
						attrValue.Value = &host
						attr.Text = &attrValue
						attribute = &attr
					} else {
						attribute, _ = offerP.GetOfferAttribute(offer, oneData.Name)
					}
					if attribute != nil {
						blog.V(3).Infof("add attribute(%s) to taskgroup", oneData.Name)
						taskgroup.Attributes = append(taskgroup.Attributes, attribute)
					} else {
						blog.Warn("get attribute(%s) for taskgroup return nil", oneData.Name)
					}
				}
			}
		}
	}

	if len(taskgroup.Taskgroup) <= 0 {
		blog.Errorf("build taskgroup(%s) failed: taskgroup.Taskgroup is empty", taskgroup.ID)
		return nil
	}

	var taskgroupinfo mesos.TaskGroupInfo

	portTotal := 0
	executorResourceDone := false
	for _, task := range taskgroup.Taskgroup {

		//update task offer info
		task.OfferId = *offer.GetId().Value
		task.AgentId = *offer.AgentId.Value
		task.AgentHostname = *offer.Hostname
		task.AgentIPAddress, _ = offerP.GetOfferIp(offer)
		//if task contains extended resources, then set device plugin socket address int it
		for _, ex := range task.DataClass.ExtendedResources {
			for _, re := range offer.GetResources() {
				if re.GetName() == ex.Name {
					//device plugin socket setted in role parameter
					ex.Socket = re.GetRole()
				}
			}
		}

		resource := *task.DataClass.Resources
		if !executorResourceDone && resource.Cpus >= 10*types.CPUS_PER_EXECUTOR {
			resource.Cpus -= types.CPUS_PER_EXECUTOR + 0.01
			executorResourceDone = true
		}
		taskResource := BuildResources(&resource)

		var taskInfo *mesos.TaskInfo
		var portNum int
		// if task.Kind is Process, make taskInfo by createProcessTaskInfo
		// else(default) make taskInfo by createContainerTaskInfo
		if task.Kind == commtypes.BcsDataType_PROCESS {
			taskInfo, portNum = createProcessTaskInfo(offer, taskResource, task, portTotal)
		} else {
			taskInfo, portNum = createContainerTaskInfo(offer, taskResource, task, portTotal)
		}

		if taskInfo == nil {
			blog.Error("build taskinfo return nil")
			return nil
		}
		portTotal += portNum
		taskgroupinfo.Tasks = append(taskgroupinfo.Tasks, taskInfo)
	}

	return &taskgroupinfo
}

func getOfferCpusetResources(o *mesos.Offer) *mesos.Resource {
	for _, i := range o.GetResources() {
		if i.GetName() == "cpuset" {
			return i
		}
	}

	return nil
}

// GetTaskGroupID get taskgroup id from mesos information
func GetTaskGroupID(taskGroupInfo *mesos.TaskGroupInfo) *string {
	defID := proto.String("default")
	if len(taskGroupInfo.Tasks) <= 0 {
		return defID
	}

	taskID := taskGroupInfo.Tasks[0].TaskId.Value
	splitID := strings.Split(*taskID, ".")

	if len(splitID) != 6 {
		return defID
	}

	//ID := strings.Join(splitID[2:], ".")

	// appInstances, appID, appRunAs, appClusterID, idTime
	ID := fmt.Sprintf("%s.%s.%s.%s.%s", splitID[2], splitID[3], splitID[4], splitID[5], splitID[0])
	return &ID
}

// IsTaskGroupEnd Whether an taskgroup is in ending statuses
func IsTaskGroupEnd(taskGroup *types.TaskGroup) bool {
	for _, task := range taskGroup.Taskgroup {
		status := task.Status
		if status == types.TASK_STATUS_LOST || status == types.TASK_STATUS_STAGING || status == types.TASK_STATUS_STARTING || status == types.TASK_STATUS_RUNNING || status == types.TASK_STATUS_KILLING { //nolint
			blog.Info("task %s status(%s), not end status", task.ID, status)
			return false
		}
	}

	return true
}

// CanTaskGroupShutdown Whether an taskgroup can be shutdown currently
func CanTaskGroupShutdown(taskGroup *types.TaskGroup) bool {
	for _, task := range taskGroup.Taskgroup {
		status := task.Status
		if status == types.TASK_STATUS_KILLING {
			blog.Info("task %s status(%s), cannot do shutdown now", task.ID, status)
			return false
		}
	}

	return true
}

// CanTaskGroupReschedule Whether an taskgroup can be rescheduled currently
func CanTaskGroupReschedule(taskGroup *types.TaskGroup) bool {

	for _, task := range taskGroup.Taskgroup {
		status := task.Status
		if status == types.TASK_STATUS_STAGING || status == types.TASK_STATUS_STARTING || status == types.TASK_STATUS_RUNNING || status == types.TASK_STATUS_KILLING {
			blog.Info("task %s status(%s), can not reschedule", task.ID, status)
			return false
		}
	}

	return true
}

// CheckVersion Check whether the version definition is correct
func CheckVersion(version *types.Version, store store.Store) error {

	for _, container := range version.Container {
		for _, configMap := range container.ConfigMaps {
			configMapName := configMap.Name
			configMapNs := version.RunAs
			blog.V(3).Infof("to get bcsconfigmap(Namespace:%s Name:%s)", configMapNs, configMapName)
			bcsConfigMap, err := store.FetchConfigMap(configMapNs, configMapName)
			if err != nil {
				blog.Warn("get bcsconfigmap(Namespace:%s Name:%s) err: %s", configMapNs, configMapName, err.Error())
				return fmt.Errorf("get bcsconfigmap(Namespace:%s Name:%s) err: %s", configMapNs, configMapName, err.Error())
			}
			if bcsConfigMap == nil {
				blog.Warn("bcsconfigmap(Namespace:%s Name:%s) not exist", configMapNs, configMapName)
				return fmt.Errorf("bcsconfigmap(Namespace:%s Name:%s) not exist", configMapNs, configMapName)
			}

			for _, confItem := range configMap.Items {
				_, ok := bcsConfigMap.Data[confItem.DataKey]
				if ok == false {
					blog.Warn("bcsconfig item(key:%s) not exist in bcsconfig(%s, %s) ", confItem.DataKey, configMapNs, configMapName)
					return fmt.Errorf("bcsconfig item(key:%s) not exist in bcsconfig(%s, %s) ", confItem.DataKey, configMapNs, configMapName)
				}
			}
		}

		for _, secret := range container.Secrets {
			secretName := secret.SecretName
			secretNs := version.RunAs
			blog.V(3).Infof("to get bcssecret(Namespace:%s Name:%s)", secretNs, secretName)
			bcsSecret, err := store.FetchSecret(secretNs, secretName)
			if err != nil {
				blog.Warn("get bcssecret(Namespace:%s Name:%s) err: %s", secretNs, secretName, err.Error())
				return fmt.Errorf("get bcssecret(Namespace:%s Name:%s) err: %s", secretNs, secretName, err.Error())
			}
			if bcsSecret == nil {
				blog.Warn("bcssecret(Namespace:%s Name:%s) not exist", secretNs, secretName)
				return fmt.Errorf("bcssecret(Namespace:%s Name:%s) not exist", secretNs, secretName)
			}

			for _, secretItem := range secret.Items {
				_, ok := bcsSecret.Data[secretItem.DataKey]
				if ok == false {
					blog.Warn("bcssecret item(key:%s) not exist in bcssecret(%s, %s) ", secretItem.DataKey, secretNs, secretName)
					return fmt.Errorf("bcssecret item(key:%s) not exist in bcssecret(%s, %s) ", secretItem.DataKey, secretNs, secretName)
				}
			}
		}

		//check image pull secret
		//"imagePullUser": "secret::imagesecret||user"
		//"imagePullPasswd": "secret::imagesecret||pwd"
		err := checkImageSecret(store, version.RunAs, container.Docker.ImagePullUser)
		if err != nil {
			blog.Errorf(err.Error())
			return err
		}
		err = checkImageSecret(store, version.RunAs, container.Docker.ImagePullPasswd)
		if err != nil {
			blog.Errorf(err.Error())
			return err
		}
	}

	//check requestIP labels "io.tencent.bcs.netsvc.requestip.*"
	requestIPLabelNum := 0
	for k := range version.Labels {
		splitK := strings.Split(k, ".")
		if len(splitK) == 6 && splitK[3] == "netsvc" && splitK[4] == "requestip" {
			//if strconv.Itoa(requestIPLabelNum) != splitK[5] {
			//	return fmt.Errorf("label netsvc.requestip.%d not exist or not in correct position", requestIPLabelNum)
			//}
			requestIPLabelNum++
		}
	}
	if requestIPLabelNum > 0 && requestIPLabelNum < int(version.Instances) {
		return fmt.Errorf("label netsvc.requestip count(%d) < version.Instances(%d)", requestIPLabelNum, version.Instances)
	}

	return nil
}
func checkImageSecret(store store.Store, ns, secret string) error {
	if strings.HasPrefix(secret, "secret::") {
		secretSplit := strings.Split(secret, "::")
		if len(secretSplit) != 2 {
			return fmt.Errorf("image secret(%s) format is invalid", secret)
		}
		userSplit := strings.Split(secretSplit[1], "||")
		if len(userSplit) != 2 {
			return fmt.Errorf("image secret(%s) format is invalid", secret)
		}

		secretName := strings.TrimSpace(userSplit[0])
		secretKey := strings.TrimSpace(userSplit[1])
		blog.Infof("to get user from secret(%s.%s::%s)", ns, secretName, secretKey)
		bcsSecret, err := store.FetchSecret(ns, secretName)
		if err != nil {
			return fmt.Errorf("get bcssecret(%s.%s) err: %s", ns, secretName, err.Error())
		}
		if bcsSecret == nil {
			return fmt.Errorf("bcssecret(%s.%s) not exist", ns, secretName)
		}
		_, ok := bcsSecret.Data[secretKey]
		if ok == false {
			return fmt.Errorf("bcssecret item(key:%s) not exist in bcssecret(%s.%s)", secretKey, ns, secretName)
		}
	}
	return nil
}

// GetVersionRequestIpCount Get reserve IP number in version definition
func GetVersionRequestIpCount(version *types.Version) int {
	requestIPLabelNum := 0
	for k := range version.Labels {
		splitK := strings.Split(k, ".")
		if len(splitK) == 6 && splitK[3] == "netsvc" && splitK[4] == "requestip" {
			requestIPLabelNum++
		}
	}

	return requestIPLabelNum
}
