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

package reflector

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
)

const (
	// applicationNode is the zk node name of the application
	applicationNode string = "application"
	// bcsRootNode is the root node name
	bcsRootNode string = "blueking"
	// Deployment zk node
	deploymentNode string = "deployment"
	//crd zk node
	crdNode string = "crd"
	// crd zk node
	autoscalerNode = "autoscaler"
)

type zkReflector struct {
	//hpa controller config
	config *config.Config

	zk *zkclient.ZkClient
}

func NewZkReflector(conf *config.Config) Reflector {
	reflector := &zkReflector{
		config: conf,
	}

	zkservs := strings.Split(conf.ClusterZkAddr, ",")
	reflector.zk = zkclient.NewZkClient(zkservs)
	err := reflector.zk.ConnectEx(time.Second * 5)
	if err != nil {
		blog.Error("Connect cluster zk %s error %s", conf.ClusterZkAddr, err.Error())
		os.Exit(1)
	}

	return reflector
}

//list all namespace autoscaler
func (reflector *zkReflector) ListAutoscalers() ([]*commtypes.BcsAutoscaler, error) {
	//get /blueking/crd/autoscaler children list
	nsKey := fmt.Sprintf("/%s/%s/%s", bcsRootNode, crdNode, autoscalerNode)
	nsChildren, err := reflector.zk.GetChildren(nsKey)
	if err != nil {
		blog.Errorf("store zk get %s children error %s", nsKey, err.Error())
		return nil, err
	}

	scalers := make([]*commtypes.BcsAutoscaler, 0)
	for _, ns := range nsChildren {
		//get /blueking/crd/autoscaler/ns children list
		path := fmt.Sprintf("%s/%s", nsKey, ns)
		scalerChildren, err := reflector.zk.GetChildren(path)
		if err != nil {
			blog.Errorf("store zk get %s data error %s", path, err.Error())
			continue
		}
		for _, child := range scalerChildren {
			//get /blueking/crd/autoscaler/ns/xxx scaler
			key := fmt.Sprintf("%s/%s", path, child)
			data, err := reflector.zk.Get(key)
			if err != nil {
				blog.Errorf("store zk get %s data error %s", key, err.Error())
				continue
			}

			var scaler *commtypes.BcsAutoscaler
			err = json.Unmarshal([]byte(data), &scaler)
			if err != nil {
				blog.Errorf("Unmarshal data %s to commtypes.BcsAutoscaler error %s", data, err.Error())
				continue
			}

			scalers = append(scalers, scaler)
		}
	}

	return scalers, nil
}

// update autoscaler in zk
func (reflector *zkReflector) UpdateAutoscaler(autoscaler *commtypes.BcsAutoscaler) error {
	zkScaler, err := reflector.FetchAutoscalerByUuid(autoscaler.GetUuid())
	if err != nil {
		return err
	}
	if zkScaler.GetUuid() != autoscaler.GetUuid() {
		return fmt.Errorf("autoscaler %s not found", autoscaler.GetUuid())
	}

	reflector.StoreAutoscaler(autoscaler)
	return err
}

func (reflector *zkReflector) StoreAutoscaler(autoscaler *commtypes.BcsAutoscaler) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", bcsRootNode, crdNode, autoscalerNode, autoscaler.ObjectMeta.NameSpace, autoscaler.ObjectMeta.Name)
	by, _ := json.Marshal(autoscaler)
	err := reflector.zk.Set(key, string(by), -1)
	return err
}

// fetch autoscaler from zk
func (reflector *zkReflector) FetchAutoscalerByUuid(uuid string) (*commtypes.BcsAutoscaler, error) {
	uids := strings.Split(uuid, "_")
	if len(uids) != 3 {
		return nil, fmt.Errorf("uuid %s is invalid", uuid)
	}

	key := fmt.Sprintf("/%s/%s/%s/%s/%s", bcsRootNode, crdNode, autoscalerNode, uids[0], uids[1])
	data, err := reflector.zk.Get(key)
	if err != nil {
		return nil, err
	}

	var scaler *commtypes.BcsAutoscaler
	err = json.Unmarshal([]byte(data), &scaler)
	if err != nil {
		return nil, err
	}

	return scaler, nil
}

//fetch deployment info, if deployment status is not Running, then can't autoscale this deployment
func (reflector *zkReflector) FetchDeploymentInfo(namespace, name string) (*schedtypes.Deployment, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s", bcsRootNode, deploymentNode, namespace, name)
	data, err := reflector.zk.Get(key)
	if err != nil {
		return nil, err
	}

	var deploy *schedtypes.Deployment
	err = json.Unmarshal([]byte(data), &deploy)
	return deploy, err
}

//fetch application info, if application status is not Running or Abnormal, then can't autoscale this application
func (reflector *zkReflector) FetchApplicationInfo(namespace, name string) (*schedtypes.Application, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s", bcsRootNode, applicationNode, namespace, name)
	data, err := reflector.zk.Get(key)
	if err != nil {
		blog.Errorf("get zk %s error %s", key, err.Error())
		return nil, err
	}

	var app *schedtypes.Application
	err = json.Unmarshal([]byte(data), &app)
	return app, err
}

//list selectorRef deployment taskgroup
func (reflector *zkReflector) ListTaskgroupRefDeployment(namespace, name string) ([]*schedtypes.TaskGroup, error) {
	path := fmt.Sprintf("/%s/%s/%s/%s", bcsRootNode, deploymentNode, namespace, name)
	data, err := reflector.zk.Get(path)
	if err != nil {
		blog.Errorf("get zk %s error %s", path, err.Error())
		return nil, err
	}

	var deploy *schedtypes.Deployment
	err = json.Unmarshal([]byte(data), &deploy)
	if err != nil {
		return nil, err
	}

	return reflector.ListTaskgroupRefApplication(namespace, deploy.Application.ApplicationName)
}

//list selectorRef application taskgroup
func (reflector *zkReflector) ListTaskgroupRefApplication(namespace, name string) ([]*schedtypes.TaskGroup, error) {
	path := fmt.Sprintf("/%s/%s/%s/%s", bcsRootNode, applicationNode, namespace, name)
	children, err := reflector.zk.GetChildren(path)
	if err != nil {
		return nil, err
	}

	blog.Infof("%v", children)

	taskgroups := make([]*schedtypes.TaskGroup, 0)
	for _, child := range children {
		key := fmt.Sprintf("%s/%s", path, child)
		data, err := reflector.zk.Get(key)
		if err != nil {
			blog.Errorf("store zk get %s error %s", key, err.Error())
			continue
		}

		var taskgroup *schedtypes.TaskGroup
		err = json.Unmarshal([]byte(data), &taskgroup)
		if err != nil {
			blog.Errorf("Unmarshal data %s to schedtypes.TaskGroup error %s", data, err.Error())
			continue
		}

		taskgroups = append(taskgroups, taskgroup)
	}

	return taskgroups, nil
}
