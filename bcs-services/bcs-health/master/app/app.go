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

package app

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/check"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/master/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/role"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/topendpoints"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/statistic"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm"
	etcdc "github.com/coreos/etcd/client"
)

func Run(c config.Config) error {
	scheme := "http"
	if len(c.CAFile) != 0 &&
		len(c.ServerCertFile) != 0 &&
		len(c.ServerKeyFile) != 0 {
		scheme = "https"
	}

	if err := register.Register(c.LocalIP, scheme, c.Port, c.MetricPort, c.BCSZk, "master"); nil != err {
		return fmt.Errorf("register bcs health failed. err: %v", err)
	}

	masterPath := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_HEALTH, "master")
	roleController, err := role.NewRoleController(c.LocalIP, c.BCSZk, masterPath)
	if nil != err {
		return err
	}

	if err := startMetricService(c, roleController); err != nil {
		return fmt.Errorf("start metric service failed. err: %v", err)
	}

	bcsAlarm, err := newBcsAlarm(c, roleController)
	if nil != err {
		return err
	}
	if err := bcsAlarm.Run(); nil != err {
		return err
	}
	check.Succeed()
	select {}
}

type AlarmObjectInterface interface {
	Run(<-chan struct{}) error
}

func newBcsAlarm(c config.Config, roleC role.RoleInterface) (*BcsAlarm, error) {
	alarm, err := alarm.NewAlarmProxy(c)
	if err != nil {
		return nil, err
	}

	tls := util.TLS{
		CaFile:   c.ETCD.CaFile,
		CertFile: c.ETCD.CertFile,
		KeyFile:  c.ETCD.KeyFile,
		PassWord: c.ETCD.PassWord,
	}

	etcdCli, err := newEtcdClient(c.ETCD.EtcdEndpoints, tls)
	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %v", err)
	}

	endpointAlarm, err := bcs.NewEndpointsAlarm(c, alarm, roleC)
	if nil != err {
		return nil, err
	}

	httpAlarm, err := server.NewHttpAlarm(c, alarm, etcdCli, roleC)
	if nil != err {
		return nil, err
	}

	bcsAlarm := &BcsAlarm{
		Objects: map[string]AlarmObjectInterface{
			"endpoints": endpointAlarm,
			"httpAlarm": httpAlarm,
		},
	}

	return bcsAlarm, nil
}

type BcsAlarm struct {
	Objects map[string]AlarmObjectInterface
}

func (b *BcsAlarm) Run() error {
	neverStop := make(chan struct{})
	for _, obj := range b.Objects {
		if err := obj.Run(neverStop); nil != err {
			return err
		}
	}
	return nil
}

func newEtcdClient(endpoints string, stls util.TLS) (etcdc.KeysAPI, error) {
	var tlsConf *tls.Config
	var err error
	if len(stls.CaFile) != 0 && len(stls.CertFile) != 0 && len(stls.KeyFile) != 0 {
		tlsConf, err = ssl.ServerTslConfVerityClient(stls.CaFile, stls.CertFile, stls.KeyFile, stls.PassWord)
		if err != nil {
			return nil, err
		}
	}

	ends := strings.Split(endpoints, ",")
	etcdCfg := etcdc.Config{
		Endpoints: ends,
		Transport: newHTTPSTransport(tlsConf),
	}
	cli, err := etcdc.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	return etcdc.NewKeysAPI(cli), nil
}

func newHTTPSTransport(cc *tls.Config) *http.Transport {
	// this seems like a bad idea but was here in the previous version
	if cc != nil {
		cc.InsecureSkipVerify = true
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     cc,
		MaxIdleConnsPerHost: 25,
	}

	return tr
}

func startMetricService(c config.Config, roleC role.RoleInterface) error {

	mc := metric.Config{
		ModuleName:          "bcs-health-master",
		RunMode:             metric.Master_Slave_Mode,
		IP:                  c.LocalIP,
		MetricPort:          c.MetricPort,
		ClusterID:           "",
		DisableGolangMetric: false,
		Labels:              make(map[string]string),
		SvrCaFile:           c.CAFile,
		SvrCertFile:         c.ServerCertFile,
		SvrKeyFile:          c.ServerKeyFile,
		SvrKeyPwd:           static.ServerCertPwd,
	}

	health := func() metric.HealthMeta {
		currentRole := metric.SlaveRole
		if roleC.IsMaster() {
			currentRole = metric.MasterRole
		}
		msg := "bcs-health-slave is healthy."
		status, health := statistic.Status()
		if !health {
			msg = status
		}

		return metric.HealthMeta{
			CurrentRole: currentRole,
			IsHealthy:   true,
			Message:     msg,
		}
	}

	totalRequest := &metric.MetricContructor{
		GetMeta: func() *metric.MetricMeta {
			return &metric.MetricMeta{
				Name: "total_requests",
				Help: "the total requests numbers statistics from the health-master is started",
			}
		},
		GetResult: func() (*metric.MetricResult, error) {
			value, err := metric.FormFloatOrString(statistic.GetTotalAccess())
			if err != nil {
				blog.Errorf("format get total access numbers failed, err: %v", err)
				return nil, err
			}
			return &metric.MetricResult{
				Value:          value,
				VariableLabels: make(map[string]string),
			}, nil
		},
	}

	return metric.NewMetricController(mc, health, totalRequest)
}
