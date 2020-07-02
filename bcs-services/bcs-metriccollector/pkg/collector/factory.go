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

package collector

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	bcsType "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/output"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/role"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

// New create
func New(cfg *config.Config, outputMgr output.Output) (Collector, error) {
	var name, namespace, clusterid, clustertype string
	switch cfg.RunMode {
	case config.ContainerType:
		name = os.Getenv("MetricApplicationName")
		namespace = os.Getenv("MetricApplicationNamespace")
		clusterid = os.Getenv("MetricClusterID")
		clustertype = os.Getenv("MetricClusterType")
	case config.TraditionalType:
		name = TraditionalName
		namespace = TraditionalNamespace
		clusterid = TraditionalClusterID
		clustertype = types.BcsComponents.String()
	default:
		return nil, fmt.Errorf("unsupported run mode: %s", cfg.RunMode)
	}

	watchPath := fmt.Sprintf("%s/%s", bcsType.BCS_SERV_BASEPATH, bcsType.BCS_MODULE_METRICCOLLECTOR)

	var err error
	roleC := role.RoleInterface(&role.Role{CurrentRole: role.MasterRole})
	if cfg.RunMode == config.TraditionalType {
		roleC, err = role.NewRoleController(cfg.LocalIP, cfg.MetricPort, cfg.BCSZk, watchPath)
		if err != nil {
			return nil, fmt.Errorf("new role controller failed, err:%v", err)
		}
		if err = initialMetricServer(*cfg, roleC); err != nil {
			return nil, fmt.Errorf("star metric server failed, err: %v", err)
		}
		blog.Infof("start metric server success.")
	}

	tmp := &collector{
		output:      outputMgr,
		role:        roleC,
		name:        name,
		namespace:   namespace,
		clusterID:   clusterid,
		clusterType: clustertype,
		cfg:         cfg,
	}

	tmp.metricClient = httpclient.NewHttpClient()
	tmp.metricClient.SetHeader("Content-Type", "application/json")
	tmp.metricClient.SetHeader("Accept", "application/json")

	if cfg.MetricClientCert.IsSSL {
		if err := tmp.metricClient.SetTlsVerity(cfg.MetricClientCert.CAFile, cfg.MetricClientCert.CertFile, cfg.MetricClientCert.KeyFile, cfg.MetricClientCert.CertPasswd); nil != err {
			blog.Error("failed to set tls ")
			return nil, err
		}
	}
	return tmp, nil
}

func initialMetricServer(cfg config.Config, roleC role.RoleInterface) error {

	c := metric.Config{
		ModuleName:  bcsType.BCS_MODULE_METRICCOLLECTOR,
		RunMode:     metric.Master_Slave_Mode,
		IP:          cfg.LocalIP,
		MetricPort:  cfg.MetricPort,
		ClusterID:   "",
		Labels:      make(map[string]string),
		SvrCaFile:   cfg.CAFile,
		SvrCertFile: cfg.ServerCertFile,
		SvrKeyFile:  cfg.ServerKeyFile,
		SvrKeyPwd:   static.ServerCertPwd,
	}

	health := func() metric.HealthMeta {
		var currentRole metric.RoleType
		if roleC.IsMaster() {
			currentRole = metric.MasterRole
		} else {
			currentRole = metric.SlaveRole
		}

		return metric.HealthMeta{
			CurrentRole: currentRole,
			IsHealthy:   true,
			Message:     "is healthy",
		}
	}

	return metric.NewMetricController(c, health)
}
