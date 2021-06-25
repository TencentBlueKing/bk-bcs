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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// CreateDeployment create deployment, call scheduler create deployment api
func (s *Scheduler) CreateDeployment(body []byte) (string, error) {
	blog.Info("create deployment. param(%s)", string(body))
	var param bcstype.BcsDeployment

	//encoding param by json
	if err := json.Unmarshal(body, &param); err != nil {
		blog.Error("parse parameters failed. param(%s), err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr)
		return err.Error(), err
	}

	// bcs-mesos-scheduler deploymentDef
	deploymentDef, err := s.newDeploymentDefWithParam(&param)
	if err != nil {
		return err.Error(), err
	}
	//store BcsDeployment original definition
	deploymentDef.RawJson = &param

	// post deploymentdef to bcs-mesos-scheduler,
	data, err := json.Marshal(deploymentDef)
	if err != nil {
		blog.Error("marshal parameter deploymentDef by json failed. err:%s", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+"encode deploymentDef by json")
		return err.Error(), err
	}

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	name := deploymentDef.ObjectMeta.Name
	namespace := deploymentDef.ObjectMeta.NameSpace

	url := fmt.Sprintf("%s/v1/deployment/%s/%s", s.GetHost(), namespace, name)
	blog.Info("post a request to url(%s), request:%s", url, string(data))

	reply, err := s.client.POST(url, nil, data)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

// UpdateDeployment do update deployment, call scheduler update deployment api
func (s *Scheduler) UpdateDeployment(body []byte, args string) (string, error) {
	blog.Info("udpate deployment. param(%s)", string(body))
	var param bcstype.BcsDeployment

	//encoding param by json
	if err := json.Unmarshal(body, &param); err != nil {
		blog.Error("parse parameters failed. param(%s), err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr)
		return err.Error(), err
	}

	// bcs-mesos-scheduler deploymentDef
	deploymentDef, err := s.newDeploymentDefWithParam(&param)
	if err != nil {
		return err.Error(), err
	}

	//store BcsDeployment original definition
	deploymentDef.RawJson = &param

	// post deploymentdef to bcs-mesos-scheduler,
	data, err := json.Marshal(deploymentDef)
	if err != nil {
		blog.Error("marshal parameter deploymentDef by json failed. err:%s", err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonEncode, common.BcsErrCommJsonEncodeStr+"encode deploymentDef by json")
		return err.Error(), err
	}

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	name := deploymentDef.ObjectMeta.Name
	namespace := deploymentDef.ObjectMeta.NameSpace

	url := fmt.Sprintf("%s/v1/deployment/%s/%s?args=%s", s.GetHost(), namespace, name, args)
	blog.Info("post a request to url(%s), request:%s", url, string(data))

	reply, err := s.client.PUT(url, nil, data)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) deleteDeployment(ns, name string, enforce string) (string, error) {
	blog.Info("delete deployment namespace %s name %s", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/deployment/%s/%s?enforce=%s", s.GetHost(), ns, name, enforce)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) cancelupdateDeployment(ns, name string) (string, error) {
	blog.Info("cancelupdate deployment namespace %s name %s", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/deployment/%s/%s/cancelupdate", s.GetHost(), ns, name)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) pauseupdateDeployment(ns, name string) (string, error) {
	blog.Info("pauseupdate deployment namespace %s name %s", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/deployment/%s/%s/pauseupdate", s.GetHost(), ns, name)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) resumeupdateDeployment(ns, name string) (string, error) {
	blog.Info("resumeupdate deployment namespace %s name %s", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/deployment/%s/%s/resumeupdate", s.GetHost(), ns, name)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) scaleDeployment(ns, name string, instances int) (string, error) {
	blog.Info("scaleDeployment deployment namespace %s name %s instances %d", ns, name, instances)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/deployment/%s/%s/scale/%d", s.GetHost(), ns, name, instances)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.PUT(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) newDeploymentDefWithParam(param *bcstype.BcsDeployment) (*types.DeploymentDef, error) {
	//check ObjectMeta is valid
	err := param.MetaIsValid()
	if err != nil {
		return nil, err
	}

	deploymentDef := &types.DeploymentDef{
		ObjectMeta: param.ObjectMeta,
		Selector:   param.Spec.Selector,
		Version:    nil,
		Strategy:   param.Spec.Strategy,
	}

	//if template is nil, then this deployment is for binding old application
	if param.Spec.Template == nil {
		return deploymentDef, nil
	}

	//var version types.Version
	version := &types.Version{
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

	//blog.V(3).Infof("param: +%v", *param)

	version.ObjectMeta = param.ObjectMeta
	version.ID = param.Name

	version.KillPolicy = &param.KillPolicy

	version.RestartPolicy = &param.RestartPolicy
	if version.RestartPolicy == nil {
		version.RestartPolicy = &bcstype.RestartPolicy{}
	}

	if version.RestartPolicy.Policy == "" {
		version.RestartPolicy.Policy = bcstype.RestartPolicy_ONFAILURE
	}

	if version.RestartPolicy.Policy != bcstype.RestartPolicy_ONFAILURE &&
		version.RestartPolicy.Policy != bcstype.RestartPolicy_ALWAYS &&
		version.RestartPolicy.Policy != bcstype.RestartPolicy_NEVER {
		blog.Error("error restart policy: %s", version.RestartPolicy.Policy)
		replyErr := bhttp.InternalError(common.BcsErrMesosDriverParameterErr,
			common.BcsErrMesosDriverParameterErrStr+"restart policy error")
		return nil, replyErr
	}

	version.RunAs = param.NameSpace
	version.Instances = int32(param.Spec.Instance)
	version.Constraints = param.Constraints

	for k, v := range param.Labels {
		version.Labels[k] = v
	}

	version, err = s.setVersionWithPodSpec(version, param.Spec.Template)
	if err != nil {
		return nil, err
	}

	deploymentDef.Version = version

	return deploymentDef, nil
}
