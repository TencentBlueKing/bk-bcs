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

package bcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-health/master/app/config"
	"bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"bk-bcs/bcs-services/bcs-health/pkg/role"
	"bk-bcs/bcs-services/bcs-health/pkg/zk"
	"bk-bcs/bcs-services/bcs-health/util"
	"github.com/pborman/uuid"
)

func NewEndpointsAlarm(c config.Config, sendTo utils.AlarmFactory, roleGeter role.RoleInterface) (*EndpointsAlarm, error) {
	t, err := template.New("endpoints").Parse(endpointEventTemplate)
	if nil != err {
		return nil, fmt.Errorf("new endpoint event template failed. err: %v", err)
	}
	alarm := &EndpointsAlarm{
		startTime:        time.Now(),
		sendTo:           sendTo,
		endpointTemplate: t,
		role:             roleGeter,
		users:            c.EndpintsReceivers,
		lbUsers:          c.LBReceivers,
		kubeUsers:        c.KubeReceivers,
		localip:          c.LocalIP,
		backOnlinePool: &backOnlinePool{
			pool: make(map[string]detail),
		},
	}
	watcher, err := zk.NewZkWatcher(strings.Split(c.BCSZk, ","), types.BCS_SERV_BASEPATH, alarm)
	if nil != err {
		return nil, err
	}
	alarm.zkWatcher = watcher
	return alarm, nil
}

type AlarmType string

const (
	AddedAlarmType  AlarmType = "Added"
	UpdateAlarmType AlarmType = "Updated"
	LostAlarmType   AlarmType = "Losted"
	// means it's offline and back to online again.
	BackOnLineType AlarmType = "Backonline"

	GrayTimeInSecond time.Duration = 5

	SecureBackOnlineTime time.Duration = 5 * time.Second
)

type EndpointsAlarm struct {
	localip          string
	zkAddrs          []string
	startTime        time.Time
	sendTo           utils.AlarmFactory
	endpointTemplate *template.Template
	role             role.RoleInterface
	users            string
	lbUsers          string
	kubeUsers        string
	backOnlinePool   *backOnlinePool
	zkWatcher        *zk.ZkWatcher
}

type backOnlinePool struct {
	sync.RWMutex
	// cache event like offline and online again in less than 10s.
	pool map[string]detail
}

type detail struct {
	start   time.Time
	svrInfo *types.ServerInfo
	branch  string
	child   string
}

func (r *EndpointsAlarm) Run(<-chan struct{}) error {
	if err := r.zkWatcher.Run(); nil != err {
		return err
	}

	go func() {
		blog.Info("start to sync back online event.")
		for {
			r.backOnlinePool.Lock()
			for key, d := range r.backOnlinePool.pool {
				if time.Now().Sub(d.start) < SecureBackOnlineTime {
					continue
				}
				delete(r.backOnlinePool.pool, key)
				go func(t detail) {
					if false == r.role.IsMaster() {
						blog.Warnf("lost event has already waited timeout, send event now, but running in slave mode , skip. branch[%s], child:[%s]", t.branch, t.child)
						return
					}

					blog.Warnf("lost event has already waited timeout, send event now. branch[%s], child:[%s]", t.branch, t.child)
					r.SendAlarm(t.branch, t.child, LostAlarmType, t.svrInfo)
				}(d)
			}
			r.backOnlinePool.Unlock()

			time.Sleep(100 * time.Millisecond)
		}
	}()
	return nil
}

func (r *EndpointsAlarm) Filter(branch string, child string, alarmType AlarmType, info *types.ServerInfo) (bool, AlarmType) {
	key := branch + info.IP + strconv.Itoa(int(info.Port))
	r.backOnlinePool.Lock()
	defer r.backOnlinePool.Unlock()
	d, exist := r.backOnlinePool.pool[key]

	if alarmType == LostAlarmType {
		if exist {
			// can not happen in theory.
			blog.Warnf("received a lost alarm, but already cached, branch[%s], child:[%s]", branch, child)
			return true, alarmType
		}

		r.backOnlinePool.pool[key] = detail{start: time.Now(), svrInfo: info, branch: branch, child: child}
		blog.Warnf("received a lost alarm, cached to check backonline event, branch[%s], child:[%s]", branch, child)
		return true, alarmType
	}

	if alarmType == AddedAlarmType {
		if exist {
			// delete the lost event now.
			delete(r.backOnlinePool.pool, key)
			if time.Now().Sub(d.start) < SecureBackOnlineTime {
				blog.Warnf("find a backonline module, branch:[%s], ip:[%s]", branch, info.IP)
				return false, BackOnLineType
			}

			return false, alarmType
		}

		return false, alarmType
	}

	return false, alarmType

}

func (r *EndpointsAlarm) TryToSendAlarm(branch string, child string, alarmType AlarmType, data string) {
	svrinfo := new(types.ServerInfo)
	if err := json.Unmarshal([]byte(data), svrinfo); nil != err {
		blog.Errorf("unmarshal leaf %s/%s failed. data: [%s] err: %v", branch, child, data, err)
		return
	}

	skip, handleType := r.Filter(branch, child, alarmType, svrinfo)
	if skip {
		blog.V(5).Infof("skip send alarm, branch[%s], child:[%s]", branch, child)
		return
	}

	if false == r.role.IsMaster() {
		blog.Warnf("received endpoints alarm, type: %s, branch: %s, children: %s, "+
			"but drop it because of running in slave mode", alarmType, branch, child)
		return
	}

	r.SendAlarm(branch, child, handleType, svrinfo)

}

func (r *EndpointsAlarm) SendAlarm(branch, child string, alarmType AlarmType, srvInfo *types.ServerInfo) {
	endpointAlarm := EndpointEvent{
		Type:       string(alarmType),
		Endpoint:   child,
		Path:       branch,
		IP:         srvInfo.IP,
		Cluster:    srvInfo.Cluster,
		Version:    srvInfo.Version,
		ReportTime: time.Now().Format(util.TimeFormat),
	}

	if time.Since(r.startTime) < time.Duration(GrayTimeInSecond*time.Second) {
		blog.Warnf("received alarm but now is in gray time, skip! path: %s", branch+"/"+child)
		return
	}

	w := bytes.Buffer{}
	if err := r.endpointTemplate.Execute(&w, endpointAlarm); nil != err {
		blog.Errorf("format endpoint alarm failed. alarm: %#v, err: %v", endpointAlarm, err)
		return
	}

	labels := make(map[string]string)
	labels[utils.EndpointsEventKindLabelKey] = string(alarmType)
	var users string
	if true == strings.Contains(branch, fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_LOADBALANCE)) {
		users = r.lbUsers
		labels[utils.EndpointsNameLabelKey] = utils.LBEndpoints
	} else if true == strings.Contains(branch, fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, "kube")) {
		users = r.kubeUsers
		labels[utils.EndpointsNameLabelKey] = utils.KubeEndpoints
	} else {
		labels[utils.EndpointsNameLabelKey] = utils.DefaultBcsEndpoints
		users = r.users
	}
	var name string
	ele := strings.Split(branch, "/")
	if len(ele) >= 5 {
		name = ele[4]
	}

	options := &utils.AlarmOptions{
		Receivers: users,
		AlarmName: "Module " + string(alarmType),
		Module:    "bcs-health-master",
		ClusterID: srvInfo.Cluster,
		AlarmMsg:  w.String(),

		EventMessage:  fmt.Sprintf("Module %s is %s ", name, string(alarmType)),
		ModuleIP:      srvInfo.IP,
		ModuleVersion: srvInfo.Version,
		AtTime:        time.Now().Unix(),
		UUID:          uuid.NewUUID().String(),
	}

	labels[utils.VoiceAlarmLabelKey] = "false"
	switch alarmType {
	case AddedAlarmType:
		options.AlarmKind = utils.INFO_ALARM
	case UpdateAlarmType:
		options.AlarmKind = utils.INFO_ALARM
	case BackOnLineType:
		options.AlarmKind = utils.INFO_ALARM
	case LostAlarmType:
		readVoice := fmt.Sprintf("BCS 组件异常告警, 组件：%s, IP地址: %s", branch[strings.LastIndex(branch, "/")+1:], srvInfo.IP)
		options.AlarmKind = utils.ERROR_ALARM
		options.VoiceReadMsg = readVoice
		labels[utils.VoiceAlarmLabelKey] = "true"
		labels[utils.VoiceMsgLabelKey] = fmt.Sprintf("%s %s", alarmType, name)
	default:
		options.AlarmKind = utils.INFO_ALARM
	}

	options.Labels = labels

	err := r.sendTo.SendAlarm(options, r.localip)
	if nil != err {
		blog.Errorf("send endpoints alarm failed. err :%v; content: %s", err, w.String())
		return
	}
	blog.Info("send endpoints alarm success, uuid %s, msg: %s.", options.UUID, w.String())
}

func (e *EndpointsAlarm) OnAddLeaf(branch, leaf, value string) {
	e.TryToSendAlarm(branch, leaf, AddedAlarmType, value)
}

func (e *EndpointsAlarm) OnUpdateLeaf(branch, leaf, oldvalue, newvalue string) {
	e.TryToSendAlarm(branch, leaf, UpdateAlarmType, newvalue)
}

func (e *EndpointsAlarm) OnDeleteLeaf(branch, leaf, value string) {
	e.TryToSendAlarm(branch, leaf, LostAlarmType, value)
}
