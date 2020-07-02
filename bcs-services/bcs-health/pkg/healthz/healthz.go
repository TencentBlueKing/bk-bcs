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

package healthz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func NewHealthCtrl(client Client) (*HealthzCtrl, error) {
	return &HealthzCtrl{
		cli: client,
	}, nil
}

type HealthzCtrl struct {
	cli Client
}

func (h HealthzCtrl) PackageHealthResult() (*Status, error) {

	platform, components, err := h.GetPlatformAndClustersDetails()
	if err != nil {
		return nil, fmt.Errorf("get platform and clusters details failed. err: %v", err)
	}

	if platform == nil || len(components) == 0 {
		return nil, errors.New("got platform or component failed")
	}

	// check the platform component first.
	platformStatus := new(PlatformStatus)
	platformStatus.Status = make(map[string]*HealthResult)
	for comp, ips := range platform.Detail {
		result, err := h.RetrievalComponentStatus(comp, "", ips)
		if err != nil {
			blog.Errorf("got component[%s] status failed, err: %v", comp, err)
			platformStatus.Status[comp] = &HealthResult{
				// Component: comp,
				Status:  Unknown,
				Message: MsgDetail{fmt.Sprintf("retrieval status failed, err: %v", err)},
			}
			continue
		}

		platformStatus.Status[comp] = result
	}

	var pHealthyStatus []HealthStatus
	for _, s := range platformStatus.Status {
		pHealthyStatus = append(pHealthyStatus, s.Status)
	}
	platformStatus.Healthy = AggregateHealthStatus(pHealthyStatus...)

	// check the component health check result.
	clsStatus := make([]*ClusterStatus, 0)
	for _, cls := range components {
		if len(cls.ClusterID) == 0 {
			return nil, errors.New("try to retrival cluster's component health status, but got empty clusterid")
		}
		clusterStatus := new(ClusterStatus)
		clusterStatus.Type = cls.Type
		clusterStatus.ClusterID = cls.ClusterID
		clusterStatus.Status = make(map[string]*HealthResult)

		for comp, ips := range cls.Detail {
			result, err := h.RetrievalComponentStatus(comp, cls.ClusterID, ips)
			if err != nil {
				blog.Errorf("retrival component[%s] status failed. err: %v", comp, err)
				clusterStatus.Status[comp] = &HealthResult{
					// Component: comp,
					Status:  Unknown,
					Message: MsgDetail{fmt.Sprintf("retrieval status failed, err: %v", err)},
				}
				continue
			}
			clusterStatus.Status[comp] = result
		}
		var cHealthStatus []HealthStatus
		for _, comp := range clusterStatus.Status {
			cHealthStatus = append(cHealthStatus, comp.Status)
		}
		clusterStatus.Healthy = AggregateHealthStatus(cHealthStatus...)

		clsStatus = append(clsStatus, clusterStatus)
	}

	var clsHealthStatus []HealthStatus
	for _, cls := range clsStatus {
		clsHealthStatus = append(clsHealthStatus, cls.Healthy)
	}

	allHealthStatus := append(clsHealthStatus, platformStatus.Healthy)

	return &Status{
		Healthy:        AggregateHealthStatus(allHealthStatus...),
		PlatformStatus: platformStatus,
		ClustersStatus: clsStatus,
	}, nil

}

func (h HealthzCtrl) RetrievalComponentStatus(component, clusterid string, realIPs []string) (result *HealthResult, err error) {
	result = new(HealthResult)
	// result.Component = component
	members, err := h.cli.GetComponentChildrenPathList(component, clusterid)
	if err != nil {
		blog.Errorf("get component[%s/%s] children path list failed, err: %v", component, clusterid, err)
		result.Message.Append(fmt.Sprintf("get component[%s/%s] children path list failed, err: %v", component, clusterid, err))
		return nil, err
	}

	diff := len(realIPs) - len(members)
	switch {
	case diff > 0:
		result.Message.Append(fmt.Sprintf("lost %d instance", diff))
		result.Status = Unhealthy
		return result, nil
	case diff < 0:
		result.Message.Append(fmt.Sprintf("got %d running instance(s), only need %d", len(members), len(realIPs)))
		result.Status = Unknown
		return result, nil
	}

	// retrival all the real members from zk first.
	memberList := make(map[string]*types.ServerInfo)
	memberIPs := make([]string, 0)
	for _, path := range members {
		s, err := h.cli.RetrievalComponentMemberInfo(path)
		if err != nil {
			blog.Errorf("retrieval %s member[path: %s] failed, err: %v.", component, path, err)
			continue
		}

		svrInfo := new(types.ServerInfo)
		if err := json.Unmarshal([]byte(s), svrInfo); err != nil {
			blog.Errorf("retrieval %s member, but unmarshal failed. err: %v", component, err)
			continue
		}
		memberList[svrInfo.IP] = svrInfo
		memberIPs = append(memberIPs, svrInfo.IP)
	}

	// compare ip values
	for _, ip := range realIPs {
		if _, exist := memberList[ip]; !exist {
			result.Status = Unknown
			result.Message.Append(fmt.Sprintf("online endpoints[%s] is different from the deployed", strings.Join(realIPs, ",")))
			return result, nil
		}
	}

	// retrival metric info
	var healthInfo []*metric.HealthInfo
	for _, m := range memberList {
		addr := fmt.Sprintf("%s:%d", m.IP, m.MetricPort)
		metric, err := h.cli.GetMetric(addr, m.Scheme)
		if err != nil {
			result.Message.Append(fmt.Sprintf("get %s metric failed. err: %v", addr, err))
			continue
		}
		healthInfo = append(healthInfo, metric)
	}

	status, msg := h.WrapHealthInfo(healthInfo, realIPs)
	result.Status = status
	if len(msg) != 0 {
		result.Message.Append(msg)
	}
	return result, nil
}

func (h HealthzCtrl) WrapHealthInfo(info []*metric.HealthInfo, deployedIPs []string) (HealthStatus, string) {
	if len(info) != len(deployedIPs) {
		return Unhealthy, ""
	}

	if len(info) == 0 {
		return Unknown, ""
	}

	mode := info[0]
	switch mode.RunMode {
	case metric.Master_Master_Mode:
		health := true
		var unhealthyList []string
		for _, ins := range info {
			if !ins.IsHealthy {
				health = false
				unhealthyList = append(unhealthyList, ins.IP)
			}
		}

		if health {
			return Healthy, ""
		}

		return Unhealthy, strings.Join(unhealthyList, ",")

	case metric.Master_Slave_Mode:
		var master, slave, unknown []string
		for _, ins := range info {
			switch ins.CurrentRole {
			case metric.MasterRole:
				master = append(master, ins.IP)
			case metric.SlaveRole:
				slave = append(slave, ins.IP)
			case metric.UnknownRole:
				unknown = append(unknown, ins.IP)
			default:
				return Unknown, fmt.Sprintf("got unsupported health role[%s] in master-slave-mode", ins.CurrentRole)
			}
		}

		switch {
		// healthy request
		case len(master) == 1 && (len(slave) == len(deployedIPs)-1) && len(unknown) == 0:
			return Healthy, ""
		case len(master) > 2:
			return Unhealthy, fmt.Sprintf("got %d master", len(master))
		case len(unknown) == len(deployedIPs):
			return Unknown, "all the members is in unknown health mode."
		default:
			return Unknown, fmt.Sprintf("got %d master, %d slave, %d unknown status members", len(master), len(slave), len(unknown))
		}

	default:
		return Unknown, fmt.Sprintf("unsupported run mode: %s", mode.RunMode)
	}
}

func (h HealthzCtrl) GetPlatformAndClustersDetails() (platform *types.DBDataItem, comp []*types.DBDataItem, err error) {
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_CLUSTERKEEPER)
	keepers, err := h.cli.DiscoveryClusterKeeperMembers(path)
	if err != nil {
		return platform, comp, err
	}

	for _, k := range keepers {
		var info types.ClusterKeeperServInfo
		if err := json.Unmarshal([]byte(k), &info); err != nil {
			return platform, comp, fmt.Errorf("unmarshal clusterkeeper info failed. err: %v", err)
		}

		addr := fmt.Sprintf("%s:%d", info.IP, info.Port)
		platform, comp, err = h.cli.GetClusterInfos(addr, info.Scheme)
		if err != nil {
			blog.Errorf("get cluster info failed, err: %v", err)
			continue
		}

		if len(comp) == 0 {
			return platform, comp, errors.New("got empty cluster infos")
		}

		return platform, comp, nil

	}

	return platform, comp, errors.New("try all cluster keepers to get cluster details failed")
}
