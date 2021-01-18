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
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

func (s *Scheduler) createDaemonsetHandler(req *restful.Request, resp *restful.Response) {
	//get http request body data
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}
	//check whether daemonset type
	err = util.CheckKind(bcstype.BcsDataType_Daemonset, body)
	if err != nil {
		blog.Error("fail to create daemonset(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateDaemonset(body)
	if err != nil {
		blog.Error("fail to create daemonset. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateDaemonset(body []byte) (string, error) {
	blog.Info("create daemonset. param(%s)", string(body))
	var param bcstype.BcsDaemonset
	//encoding param by json
	if err := json.Unmarshal(body, &param); err != nil {
		blog.Error("parse daemonset failed. param(%s), err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommJsonDecode, common.BcsErrCommJsonDecodeStr)
		return err.Error(), err
	}

	// bcs-mesos-scheduler daemonset definition
	def, err := s.newDaemonsetDefWithParam(&param)
	if err != nil {
		return err.Error(), err
	}
	// post daemonset definition to bcs-mesos-scheduler,
	data, _ := json.Marshal(def)
	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/daemonsets", s.GetHost())
	blog.Info("post a request to url(%s), request:%s", url, string(data))

	reply, err := s.client.POST(url, nil, data)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) newDaemonsetDefWithParam(param *bcstype.BcsDaemonset) (*types.BcsDaemonsetDef, error) {
	//check ObjectMeta is valid
	err := param.MetaIsValid()
	if err != nil {
		return nil, err
	}

	//new daemonset definition
	def := &types.BcsDaemonsetDef{
		ObjectMeta: param.ObjectMeta,
	}

	//var version types.Version
	version := &types.Version{
		ID:          "",
		Instances:   0,
		RunAs:       "",
		Container:   []*types.Container{},
		Labels:      make(map[string]string),
		Constraints: nil,
		Uris:        []string{},
		Ip:          []string{},
		Mode:        "",
	}
	//init version parameters
	version.ObjectMeta = param.ObjectMeta
	version.ID = param.Name
	version.KillPolicy = &param.KillPolicy
	version.RestartPolicy = &param.RestartPolicy
	if version.RestartPolicy == nil {
		version.RestartPolicy = &bcstype.RestartPolicy{}
	}
	//default onfailure restart policy
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
	for k, v := range param.Labels {
		version.Labels[k] = v
	}
	//the version belongs daemonset
	version.Kind = bcstype.BcsDataType_Daemonset
	version, err = s.setVersionWithPodSpec(version, param.Spec.Template)
	if err != nil {
		return nil, err
	}

	def.Version = version
	return def, nil
}

func (s *Scheduler) deleteDaemonsetHandler(req *restful.Request, resp *restful.Response) {
	//namespace
	ns := req.PathParameter("ns")
	//daemonset name
	name := req.PathParameter("name")
	//whether enforce delete daemonset
	enforce := req.QueryParameter("enforce")
	reply, err := s.deleteDaemonset(ns, name, enforce)
	if err != nil {
		blog.Error("fail to delete daemonset namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteDaemonset(ns, name string, enforce string) (string, error) {
	blog.Info("delete daemonset namespace %s name %s", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/daemonsets/%s/%s?enforce=%s", s.GetHost(), ns, name, enforce)
	blog.Info("post a request to url(%s), request: null", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}
