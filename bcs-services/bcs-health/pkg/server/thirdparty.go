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

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	bresp "bk-bcs/bcs-common/common/http"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"bk-bcs/bcs-services/bcs-health/pkg/healthz"
	"github.com/emicklei/go-restful"
	"github.com/pborman/uuid"
)

func (r *HttpAlarm) Create(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("read request body failed. err: %v", err).Error()})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}
	uuid := uuid.NewUUID().String()
	blog.Infof("received an alarm request, uuid: %s, source info: [ %s ], data: %s.", uuid, req.Request.RemoteAddr, string(data))
	options := utils.AlarmOptions{}
	if err = json.Unmarshal(data, &options); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("unmarshal healthinfo failed. err: %v", err).Error()})
		blog.Errorf("received an alarm request, but unmarshal health info failed, err: %v", err)
		return
	}

	options.UUID = uuid

	err = r.s.sendTo.SendAlarm(&options, req.Request.RemoteAddr)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusInternalServerError, Message: err.Error()})
		blog.Errorf("received an http alarm request, but send alarm failed. err: %v", err)
		return
	}
	resp.WriteEntity(bresp.APIRespone{Result: true, Code: 0, Message: "success"})
}

func (r *HttpAlarm) CreateAlarm(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("read request body failed. err: %v", err).Error()})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}
	uuid := uuid.NewUUID().String()
	blog.Infof("received an alarm request, uuid: %s, source info: [ %s ], data: %s.", uuid, req.Request.RemoteAddr, string(data))
	options := utils.AlarmOptions{}
	if err := json.Unmarshal(data, &options); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("unmarshal healthinfo failed. err: %v", err).Error()})
		blog.Errorf("received an alarm request, but unmarshal health info failed, err: %v", err)
		return
	}

	options.UUID = uuid
	err = r.s.sendTo.SendAlarm(&options, req.Request.RemoteAddr)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusInternalServerError, Message: err.Error()})
		blog.Errorf("received an http alarm request, but send alarm failed. err: %v", err)
		return
	}
	resp.WriteEntity(bresp.APIRespone{Result: true, Code: 0, Message: "success"})
}

type MaintenanceConfig struct {
	AlarmName      string `json:"alarmName,omitempty"`
	ClusterID      string `json:"clusterID"`
	StopForSeconds int64  `json:"stopForSeconds"`
}

func (r *HttpAlarm) SetMaintenance(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("read request body failed. err: %v", err).Error()})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}
	blog.Infof("received an *set maintenance* request, source: [ %s ], data: %s.", req.Request.RemoteAddr, string(data))
	cfg := MaintenanceConfig{}
	if err := json.Unmarshal(data, &cfg); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("unmarshal maintenance config failed. err: %v", err).Error()})
		blog.Errorf("received an *set maintenance* request, but unmarshal failed, err: %v", err)
		return
	}

	if len(cfg.ClusterID) == 0 || 0 == cfg.StopForSeconds {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: "clusterID can not be null and stopForSeconds can not be zero."})
		blog.Errorf("received an *set maintenance* request, but got invalid clusterID or stopForSeconds.")
		return
	}

	//defer r.server.zkClient.Close()
	if err := r.s.zkClient.Connect(); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusInternalServerError, Message: fmt.Sprintf("connect to zk failed. err: %v", err)})
		blog.Errorf("connect to zk server failed. err: %v", err)
		return
	}
	var path string
	if len(cfg.AlarmName) == 0 {
		path = fmt.Sprintf("/bcs/services/healthinfo/%s/bcsDefaultConfig", cfg.ClusterID)
	} else {
		path = fmt.Sprintf("/bcs/services/healthinfo/%s/%s", cfg.ClusterID, cfg.AlarmName)
	}
	now := time.Now()
	record := MaintRecord{
		AlarmName:    cfg.AlarmName,
		ClusterID:    cfg.ClusterID,
		StartTime:    now.Unix(),
		DeadlineTime: now.Add(time.Duration(cfg.StopForSeconds) * time.Second).Unix(),
	}

	js, err := json.MarshalIndent(record, "", "\t")
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusInternalServerError, Message: err.Error()})
		blog.Errorf("marshal failed. err: %v", err)
		return
	}

	if err := r.s.zkClient.Update(path, string(js)); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusInternalServerError, Message: err.Error()})
		blog.Errorf("write maint info to zk failed, path: %s. err: %v", path, err)
		return
	}
	resp.WriteEntity(bresp.APIRespone{Result: true, Code: 0, Message: "set maintenance success."})
}

func (r *HttpAlarm) CancelMaintenance(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("read request body failed. err: %v", err).Error()})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}
	blog.Infof("received an *cancel maintenance* request, source: [ %s ], data: %s.", req.Request.RemoteAddr, string(data))
	cfg := MaintenanceConfig{}
	if err := json.Unmarshal(data, &cfg); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("unmarshal cancel maintenance config failed. err: %v", err).Error()})
		blog.Errorf("received an *cancel maintenance* request, but unmarshal failed, err: %v", err)
		return
	}

	var path string
	if len(cfg.AlarmName) == 0 {
		//path = fmt.Sprintf("/bcs/services/healthinfo/%s/bcsDefaultConfig", cfg.ClusterID)
		path = NewHealthPath().ClusterPath(cfg.ClusterID).String()
	} else {
		//path = fmt.Sprintf("/bcs/services/healthinfo/%s/%s", cfg.ClusterID, cfg.AlarmName)
		path = NewHealthPath().AlarmNamePath(cfg.ClusterID, cfg.AlarmName).String()
	}

	if err := r.s.zkClient.Del(path, -1); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf(" cancel maintenance config failed. err: %v", err).Error()})
		blog.Errorf("received an *cancel maintenance* request, but cancel failed, err: %v", err)
		return
	}

	resp.WriteEntity(bresp.APIRespone{Result: true, Code: 0, Message: "cancel maintenance success."})
	return
}

func (r *HttpAlarm) GetMaintenance(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("read request body failed. err: %v", err).Error()})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}
	blog.Infof("received an *get maintenance* request, data: %s.", string(data))
	cfg := MaintenanceConfig{}
	if err := json.Unmarshal(data, &cfg); nil != err {
		resp.WriteEntity(bresp.APIRespone{Result: false, Code: http.StatusBadRequest, Message: fmt.Errorf("unmarshal get maintenance config failed. err: %v", err).Error()})
		blog.Errorf("received an *get maintenance* request, but unmarshal failed, err: %v", err)
		return
	}

	resp.WriteEntity(bresp.APIRespone{Result: true, Code: 0, Message: "cancel maintenance success."})
	return
}

func (r *HttpAlarm) GetPlatformAndComponentHealthz(req *restful.Request, resp *restful.Response) {
	blog.Infof("received GetPlatformAndComponentHealthz request, remote ip: %s", req.Request.RemoteAddr)
	result, err := r.s.healthzCtrl.PackageHealthResult()
	if err != nil {
		blog.Errorf("received GetPlatformAndComponentHealthz request, but failed with err: %v", err)
		resp.WriteEntity(healthz.HealthResponse{Code: common.BcsErrHealthGetHealthzInfoErr, OK: false, Message: err.Error()})
		return
	}

	resp.WriteEntity(healthz.HealthResponse{Code: 0, OK: true, Message: "success", Data: *result})
	return
}

type MaintRecord struct {
	AlarmName    string `json:"alarmName"`
	ClusterID    string `json:"clusterID"`
	StartTime    int64  `json:"startTime"`
	DeadlineTime int64  `json:"deadlineTime"`
}

func NewHealthPath() HealthPath {
	return HealthPath("/bcs/services/healthinfo")
}

type HealthPath string

func (h HealthPath) ClusterPath(clusterid string) HealthPath {
	return HealthPath(fmt.Sprintf("%s/%s/bcsDefaultConfig", h, clusterid))
}

func (h HealthPath) AlarmNamePath(clusterid, alarmname string) HealthPath {
	return HealthPath(fmt.Sprintf("%s/%s/%s", h, clusterid, alarmname))
}

func (h HealthPath) Parent() HealthPath {
	arr := strings.Split(string(h), "/")
	return HealthPath(strings.Join(arr[:len(arr)-1], "/"))
}

func (h HealthPath) String() string {
	return string(h)
}

func (h HealthPath) IsClusterPath() bool {
	return strings.HasSuffix(string(h), "bcsDefaultConfig")
}