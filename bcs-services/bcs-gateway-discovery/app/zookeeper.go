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
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
)

//service event notification
func (s *DiscoveryServer) moduleEventNotifycation(module string) {
	if !s.bcsRegister.IsMaster() {
		blog.Infof("gateway-discovery instance is not master, skip module %s event notification", module)
		return
	}
	//get event notification
	event := &ModuleEvent{
		Module: module,
	}
	s.evtCh <- event
}

func (s *DiscoveryServer) handleModuleChange(event *ModuleEvent) error {
	//get specified module info and construct data for refresh
	svcs, err := s.formatBCSServerInfo(event.Module)
	if err != nil {
		blog.Errorf("discovery handle module %s changed failed, %s", event.Module, err.Error())
		return err
	}
	event.Svc = svcs
	//update service route
	if err := s.gatewayServiceSync(event); err != nil {
		blog.Errorf("hanlde zookeeper event failed, it had better confirm latest data synchronization")
		return err
	}
	return nil
}

// formatBCSServerInfo format bcs zookeeper server info according to event module name
//@param: module, mesosdriver/cluster-xxxx, storage
func (s *DiscoveryServer) formatBCSServerInfo(module string) (*register.Service, error) {
	originals, err := s.discovery.GetModuleServers(module)
	if err != nil {
		//* if get no ServerInfo, GetModuleServers return error
		blog.Errorf("get module %s information from module-discovery failed, %s", module, err.Error())
		return nil, err
	}
	blog.V(5).Infof("get module %s string detail: %+v", module, originals)
	var svcs []*types.ServerInfo
	skip := false
	for _, info := range originals {
		data := info.(string)
		svc := new(types.ServerInfo)
		if err := json.Unmarshal([]byte(data), svc); err != nil {
			blog.Errorf("handle module %s json %s unmarshal failed, %s", module, data, err.Error())
			continue
		}
		//! compatible code here, when mesos driver already start etcd registry feature
		//! discovery stop adopt zookeeper registry information
		if s.isClusterRestriction(svc.Cluster) {
			blog.Warnf("discovery check that cluster %s[%s] mesosdriver change to etcd registry, skip", svc.Cluster, svc.IP)
			skip = true
			continue
		}
		svcs = append(svcs, svc)
	}
	if len(svcs) == 0 && !skip {
		blog.Errorf("convert module %s all json info failed, pay more attention", module)
		return nil, fmt.Errorf("module %s all json err", module)
	}
	//data structure conversion
	rSvcs, err := s.adapter.GetService(module, svcs)
	if err != nil {
		blog.Errorf("converts module %s ServerInfo to api-gateway info failed in synchronization, %s", module, err.Error())
		return nil, err
	}
	return rSvcs, nil
}

//formatDriverServerInfo format mesosdriver & kubernetedriver server information
//@param: module, module info with clusterID, mesosdriver/BCS-MESOS-10032
func (s *DiscoveryServer) formatDriverServerInfo(module string) ([]*register.Service, error) {
	originals, err := s.discovery.GetModuleServers(module)
	if err != nil {
		blog.Errorf("get module %s information from module-discovery failed, %s", module, err.Error())
		return nil, err
	}
	blog.V(5).Infof("get module %s string detail: %+v", module, originals)
	svcs := make(map[string][]*types.ServerInfo)
	skip := false
	for _, info := range originals {
		data := info.(string)
		svc := new(types.ServerInfo)
		if err := json.Unmarshal([]byte(data), svc); err != nil {
			blog.Errorf("handle module %s json %s unmarshal failed, %s", module, data, err.Error())
			continue
		}
		if len(svc.Cluster) == 0 {
			blog.Errorf("find driver %s node lost cluster information. detail: %s", module, data)
			continue
		}
		//! compatible code here, when mesos driver already start etcd registry feature
		//! discovery stop adopt zookeeper registry information
		if s.isClusterRestriction(svc.Cluster) {
			blog.Warnf("discovery check that cluster %s[%s] mesosdriver change to etcd registry, skip", svc.Cluster, svc.IP)
			skip = true
			continue
		}
		key := fmt.Sprintf("%s/%s", module, svc.Cluster)
		svcs[key] = append(svcs[key], svc)
	}
	if len(svcs) == 0 && !skip {
		blog.Errorf("unmarshal module %s json all failed!", module)
		return nil, fmt.Errorf("driver %s all clusters json err", module)
	}
	var localSvcs []*register.Service
	for k, v := range svcs {
		//data structure conversion
		rSvcs, err := s.adapter.GetService(k, v)
		if err != nil {
			blog.Errorf("converts module %s ServerInfo to api-gateway info failed in synchronization, %s", k, err.Error())
			continue
		}
		localSvcs = append(localSvcs, rSvcs)
	}
	if len(localSvcs) == 0 {
		blog.Errorf("convert module %s to api-gateway info failed!", module)
		return nil, fmt.Errorf("convert %s to api-gateway err", module)
	}
	return localSvcs, nil
}

func (s *DiscoveryServer) formatKubeAPIServerInfo(module string) ([]*register.Service, error) {
	var userMgrInst string
	_, umStrategy := defaultHTTPModules["usermanager"]
	if s.option.Etcd.Feature && umStrategy {
		//get api-server information from bcs-user-manager etcd registry
		node, err := s.microDiscovery.GetRandomServerInstance(types.BCS_MODULE_USERMANAGER)
		if err != nil {
			blog.Errorf("get user-manager module from etcd registry failed, %s", err.Error())
			return nil, err
		}
		userMgrInst = node.Address
		blog.Infof("get random user-manager instance [%s] from etcd registry for query kube-apiserver", userMgrInst)
	} else {
		//we get all kubernetes api-server from bcs-user-manager, zookeeper registry
		original, err := s.discovery.GetRandModuleServer(types.BCS_MODULE_USERMANAGER)
		if err != nil {
			blog.Errorf("get module %s info for list all kube-apiserver failed, %s", types.BCS_MODULE_USERMANAGER, err.Error())
			return nil, err
		}
		blog.V(5).Infof("get module %s string detail: %+v", types.BCS_MODULE_USERMANAGER, original)
		data := original.(string)
		svc := new(types.ServerInfo)
		if err := json.Unmarshal([]byte(data), svc); err != nil {
			blog.Errorf("in [kubernetes api-server] handle module %s json %s unmarshal failed, %s", types.BCS_MODULE_USERMANAGER, data, err.Error())
			return nil, err
		}
		userMgrInst = fmt.Sprintf("%s:%d", svc.IP, svc.Port)
		blog.Infof("get random user-manager instance [%s] from zookeeper registry for query kube-apiserver", userMgrInst)
	}
	//ready to get kube-apiserver list from bcs-user-manager
	config := &bcsapi.Config{
		Hosts:     []string{userMgrInst},
		AuthToken: s.option.AuthToken,
		Gateway:   false,
	}
	config.TLSConfig, _ = s.option.GetClientTLS()
	apiCli := bcsapi.NewClient(config)
	clusters, err := apiCli.UserManager().ListAllClusters()
	if err != nil {
		blog.Errorf("request all kube-apiserver cluster info from bcs-user-manager %+v failed, %s", config.Hosts, err.Error())
		return nil, err
	}
	if len(clusters) == 0 {
		blog.Warnf("No kube-apiserver registed, skip kube-apiserver proxy rules")
		return nil, nil
	}
	//construct inner Service definition
	var localSvcs []*register.Service
	for _, cluster := range clusters {
		k := fmt.Sprintf("%s/%s", module, cluster.ClusterID)
		//one clustercredential converts to ServerInfo
		var svcs []*types.ServerInfo
		clusterAddress := strings.Split(cluster.ServerAddresses, ",")
		for _, address := range clusterAddress {
			u, err := url.Parse(address)
			if err != nil {
				blog.Errorf("kube-apiserver[%s] cluster_address %s parse failed, %s", cluster.ClusterID, cluster.ServerAddresses, err.Error())
				continue
			}
			svc := &types.ServerInfo{
				Cluster: cluster.ClusterID,
				Scheme:  u.Scheme,
				Port:    443,
				//! trick here
				HostName: cluster.UserToken,
			}
			hostport := strings.Split(u.Host, ":")
			if len(hostport) == 2 {
				svc.IP = hostport[0]
				port, _ := strconv.Atoi(hostport[1])
				svc.Port = uint(port)
			} else {
				if svc.Scheme == "http" {
					svc.Port = 80
				}
			}
			svcs = append(svcs, svc)
		}
		blog.V(5).Infof("kube-apiserver cluster info [%s] ServerInfo %+v", k, svcs)
		rSvcs, err := s.adapter.GetService(k, svcs)
		if err != nil {
			blog.Errorf("converts module %s ServerInfo to api-gateway info failed in synchronization, %s", k, err.Error())
			continue
		}
		localSvcs = append(localSvcs, rSvcs)
	}
	if len(localSvcs) == 0 {
		blog.Errorf("convert kube-apiserver [%s] to api-gateway info failed!", module)
		return nil, fmt.Errorf("convert [%s] to api-gateway err", module)
	}
	return localSvcs, nil
}

func (s *DiscoveryServer) formatMultiServerInfo(modules []string) ([]*register.Service, error) {
	var regSvcs []*register.Service
	for _, m := range modules {
		if m == types.BCS_MODULE_MESOSDRIVER || m == types.BCS_MODULE_KUBERNETEDRIVER {
			svcs, err := s.formatDriverServerInfo(m)
			if err != nil {
				blog.Errorf("gateway-discovery format DriverModule %s to inner register Service failed %s, continue", m, err.Error())
				continue
			}
			regSvcs = append(regSvcs, svcs...)
		} else if m == types.BCS_MODULE_KUBEAGENT {
			svc, err := s.formatKubeAPIServerInfo(m)
			if err != nil {
				blog.Errorf("gateway-discovery format kubernetes api-server to inner register Service failed %s, continue", err.Error())
				continue
			}
			regSvcs = append(regSvcs, svc...)
		} else {
			svc, err := s.formatBCSServerInfo(m)
			if err != nil {
				blog.Errorf("gateway-discovery format BCSServerInfo Module %s from cache failed: %s, continue", m, err.Error())
				continue
			}
			regSvcs = append(regSvcs, svc)
		}
	}
	return regSvcs, nil
}
