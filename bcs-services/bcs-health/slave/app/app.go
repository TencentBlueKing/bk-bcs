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
	"fmt"
	"os"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/metric"
	"bk-bcs/bcs-common/common/static"
	"bk-bcs/bcs-common/common/statistic"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-services/bcs-health/pkg/client"
	"bk-bcs/bcs-services/bcs-health/pkg/job/collector"
	"bk-bcs/bcs-services/bcs-health/pkg/register"
	"bk-bcs/bcs-services/bcs-health/pkg/role"
	"bk-bcs/bcs-services/bcs-health/slave/app/config"
	"bk-bcs/bcs-services/bcs-health/util"
)

func Run(c config.Config) error {

	if err := register.Register(c.LocalIP, "", 0, c.MetricPort, c.BCSZk, c.ClusterName); err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	slave := &util.Slave{
		SlaveClusterName: c.ClusterName,
		Zones:            c.Zones,
		ServerInfo: types.ServerInfo{
			IP:       c.LocalIP,
			Port:     0,
			HostName: hostname,
			Scheme:   "",
			Version:  version.GetVersion(),
			Cluster:  "",
			Pid:      os.Getpid(),
		},
	}

	tls := util.TLS{
		CaFile:   c.CAFile,
		CertFile: c.ClientCertFile,
		KeyFile:  c.ClientKeyFile,
		PassWord: static.ClientCertPwd,
	}

	cli, err := client.NewClient(c.BCSZk, tls)
	if err != nil {
		return fmt.Errorf("new client failed. err: %v", err)
	}

	slavePath := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_HEALTH, c.ClusterName)
	role, err := role.NewRoleController(c.LocalIP, c.BCSZk, slavePath)
	if err != nil {
		return fmt.Errorf("new role controller failed. err: %v", err)
	}

	if err := startMetricService(c, role); err != nil {
		return fmt.Errorf("start metric service failed. err: %v", err)
	}

	jc := collector.NewJobCollector(slave, cli, role)
	jc.Run()
	blog.Info("start health slave success.")
	select {}
}

func startMetricService(c config.Config, roleC role.RoleInterface) error {
	mc := metric.Config{
		ModuleName:  "bcs-health-slave",
		RunMode:     metric.Master_Slave_Mode,
		IP:          c.LocalIP,
		MetricPort:  c.MetricPort,
		ClusterID:   "",
		Labels:      make(map[string]string),
		SvrCaFile:   c.CAFile,
		SvrCertFile: c.ServerCertFile,
		SvrKeyFile:  c.ServerKeyFile,
		SvrKeyPwd:   static.ServerCertPwd,
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

	return metric.NewMetricController(mc, health)
}
