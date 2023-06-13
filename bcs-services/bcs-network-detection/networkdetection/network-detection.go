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

package networkdetection

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	mesosdriver "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/networkdetection/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network-detection/config"

	"github.com/emicklei/go-restful"
)

// NetworkDetection implementation
type NetworkDetection struct {
	sync.RWMutex
	// network detection config
	conf *config.Config
	// network detection clusters
	// example BCS-MESOS-10000
	clusters []string
	// deploy detection node infos
	// key=clusterid/idc, example BCS-MESOS-10000/上海-周浦
	deploys map[string]*types.DeployDetection
	// ContainerPlatform include mesos、k8s
	platform ContainerPlatform
	// cmdb client
	cmdb *CmdbClient
	// deployment template json
	deployTemplate commtypes.BcsDeployment
	// http server
	httpServ *httpserver.HttpServer
	// http actions
	acts []*httpserver.Action
}

// ContainerPlatform definition for fetch container info
type ContainerPlatform interface {
	// GetNodes xxx
	// get cluster all nodes
	GetNodes(clusterid string) ([]*types.NodeInfo, error)
	// CeateDeployment xxx
	// deploy application
	// deploy is definition json
	CeateDeployment(clusterid string, deploy []byte) error
	// FetchDeployment xxx
	// fetch deployed deployment info
	FetchDeployment(deploy *types.DeployDetection) (interface{}, error)
	// FetchPods xxx
	// fetch deloyment't pods
	FetchPods(clusterid, ns, name string) ([]byte, error)
}

// NewNetworkDetection new NetworkDetection object
func NewNetworkDetection(conf *config.Config) *NetworkDetection {
	n := &NetworkDetection{
		conf:     conf,
		clusters: strings.Split(conf.Clusters, ","),
		deploys:  make(map[string]*types.DeployDetection),
		httpServ: httpserver.NewHttpServer(conf.Port, conf.Address, ""),
	}
	if conf.ServerCert.IsSSL {
		n.httpServ.SetSsl(conf.ServerCert.CAFile, conf.ServerCert.CertFile, conf.ServerCert.KeyFile,
			conf.ServerCert.CertPasswd)
	}
	return n
}

// Start networkdetection work
func (n *NetworkDetection) Start() error {
	var err error
	// init deployment template
	by, err := ioutil.ReadFile(n.conf.Template)
	if err != nil {
		return err
	}
	err = json.Unmarshal(by, &n.deployTemplate)
	if err != nil {
		return err
	}
	// new mesos platform
	conf := &mesosdriver.MesosDriverClientConfig{
		ZkAddr:     n.conf.BcsZk,
		ClientCert: n.conf.ClientCert,
	}
	n.platform, err = mesosdriver.NewMesosDriverClient(conf)
	if err != nil {
		return err
	}
	// new cmdb client
	n.cmdb, err = NewCmdbClient(n.conf)
	if err != nil {
		return err
	}
	// create DeployInfo
	n.createDeployInfo()
	// ticker deploy detection nodes
	go n.tickerDeployNetworkDetectionNode()
	// register endpoints in bcs zk
	go n.regDiscover()
	// init http server
	n.initActions()
	n.httpServ.RegisterWebServer("/detection/v4", nil, n.acts)
	go func() {
		err := n.httpServ.ListenAndServe()
		blog.Errorf("http listen and service end! err:%s", err.Error())
		os.Exit(1)
	}()
	return nil
}

func (n *NetworkDetection) createDeployInfo() {
	blog.Infof("NetworkDetection create deployinfo")
	var err error
	for _, clusterid := range n.clusters {
		err = n.createClusterDeployInfo(clusterid)
		if err != nil {
			blog.Errorf("create cluster %s deployInfo failed: %s", clusterid, err.Error())
		}
		blog.Infof("create cluster %s deployInfo success", clusterid)
	}
}

// createClusterDeployInfo xxx
// create DeployDetection object
// include clusterid, idc, nodes
func (n *NetworkDetection) createClusterDeployInfo(clusterid string) error {
	// get cluster node list
	nodes, err := n.platform.GetNodes(clusterid)
	if err != nil {
		blog.Errorf("get cluster %s nodes error %s", clusterid, err.Error())
		return err
	}

	for _, node := range nodes {
		// update node cmdb info
		// include Idc、modulename
		err = n.cmdb.updateNodeInfo(node)
		if err != nil {
			blog.Errorf("update node %s cmdb info error %s", node.Ip, err.Error())
			continue
		}

		// n.deploys key, key=clusterid/idc, example BCS-MESOS-10000/上海-周浦
		key := fmt.Sprintf("%s/%s", node.Clusterid, node.Idc)
		if _, ok := n.deploys[key]; !ok {
			deploy := &types.DeployDetection{
				Clusterid: node.Clusterid,
				Idc:       node.Idc,
				IdcUnit:   node.IdcUnit,
				Nodes:     make([]*types.NodeInfo, 0),
				Template:  n.deployTemplate,
			}
			n.deploys[key] = deploy
		}
		n.deploys[key].Nodes = append(n.deploys[key].Nodes, node)
	}

	// list all nodes info
	for k, v := range n.deploys {
		for _, node := range v.Nodes {
			blog.Infof("%s %s", k, node.Ip)
		}
	}
	return nil
}

func (n *NetworkDetection) tickerDeployNetworkDetectionNode() {
	for {
		time.Sleep(time.Second * 10)
		n.deployNetworkDetectionNode()
	}
}

func (n *NetworkDetection) deployNetworkDetectionNode() {
	// check region whether deploy detection nodes
	for _, o := range n.deploys {
		// if application=nil, then deploy detection application
		if o.Application == nil {
			// if fetch deployment, then continue
			if !n.deployNodes(o) {
				continue
			}
		}

		// fetch taskgroup
		by, err := n.platform.FetchPods(o.Clusterid, o.Application.RunAs, o.Application.Name)
		if err != nil {
			blog.Errorf("region(%s:%s) fetch deployment(%s:%s) pods failed: %s",
				o.Clusterid, o.Idc, o.Application.RunAs, o.Application.Name, err.Error())
			continue
		}
		n.Lock()
		err = json.Unmarshal(by, &o.Pods)
		n.Unlock()
		if err != nil {
			blog.Errorf("region(%s:%s) Unmarshal Pods body(%s) failed: %s", o.Clusterid, o.Idc, string(by), err.Error())
			continue
		}
		blog.Infof("ticker sync region(%s:%s) deployed pods success", o.Clusterid, o.Idc)
		for _, pod := range o.Pods {
			blog.Infof("region(%s:%s) deployed pod %s status %s ip %s", o.Clusterid, o.Idc, pod.ID, pod.Status,
				schedtypes.GetTaskgroupIP(pod))
		}
	}
}

// deployNodes xxx
// if return true, show fetch deployment success, then list relevant pods
// if return false, show fetch deployment failed, then continue
func (n *NetworkDetection) deployNodes(o *types.DeployDetection) bool {
	// fetch deployment
	i, err := n.platform.FetchDeployment(o)
	if err != nil {
		if err.Error() == "Not found" {
			blog.Errorf("region(%s:%s) fetch deployment not found, then create it", o.Clusterid, o.Idc)
		} else {
			blog.Errorf("region(%s:%s) fetch deployment failed: %s", o.Clusterid, o.Idc, err.Error())
			return false
		}
	} else {
		o.Application, _ = i.(*schedtypes.Application)
		blog.Infof("region(%s:%s) fetch deployment(%s:%s) success",
			o.Clusterid, o.Idc, o.Application.RunAs, o.Application.Name)
		return true
	}

	// create deployment in container cluster
	deployJSON := o.Template
	// deepcopy Constraints
	by, _ := json.Marshal(o.Template.Constraints)
	deployJSON.Constraints = new(commtypes.Constraint)
	json.Unmarshal(by, &deployJSON.Constraints)
	deployJSON.Name = fmt.Sprintf("pinger-%d", time.Now().UnixNano())
	deployJSON.Annotations = map[string]string{
		"idc":     o.Idc,
		"cluster": o.Clusterid,
	}
	// parse deploy constraint
	for _, node := range o.Nodes {
		union := deployJSON.Constraints.IntersectionItem[0].UnionData[0]
		union.Set.Item = append(union.Set.Item, node.Ip)
	}

	// create deployment
	by, _ = json.Marshal(deployJSON)
	blog.Infof("region(%s:%s) deploy template json(%s)", o.Clusterid, o.Idc, string(by))
	err = n.platform.CeateDeployment(o.Clusterid, by)
	if err != nil {
		blog.Errorf("region(%s:%s) create deployment(%s:%s) failed: %s",
			o.Clusterid, o.Idc, deployJSON.NameSpace, deployJSON.Name, err.Error())
	} else {
		blog.Infof("region(%s:%s) create deployment(%s:%s) done",
			o.Clusterid, o.Idc, deployJSON.NameSpace, deployJSON.Name)
	}

	return false
}

// regDiscover networkdetection module in bcs zk
func (n *NetworkDetection) regDiscover() {
	blog.Infof("NetworkDetection to do register bcszk...")
	// register service
	regDiscv := rd.NewRegDiscoverEx(n.conf.BcsZk, time.Second*10)
	if regDiscv == nil {
		blog.Error("NewRegDiscover(%s) return nil, redo after 3 second ...", n.conf.BcsZk)
		time.Sleep(3 * time.Second)
		go n.regDiscover()
		return
	}
	blog.Info("NewRegDiscover(%s) success", n.conf.BcsZk)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDiscv start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)
		go n.regDiscover()
		return
	}
	blog.Info("RegDiscover start success")
	defer regDiscv.Stop()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("network-detection get hostname err: %s", err.Error())
		host = "UNKNOWN"
	}
	var regInfo commtypes.NetworkDetectionServInfo
	regInfo.ServerInfo.IP = n.conf.Address
	regInfo.ServerInfo.Port = n.conf.Port
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = "http"
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	if n.conf.ServerCert.IsSSL {
		regInfo.ServerInfo.Scheme = "https"
	}

	key := commtypes.BCS_SERV_BASEPATH + "/" + commtypes.BCS_MODULE_NETWORKDETECTION + "/" + n.conf.Address
	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("json Marshal error(%s)", err.Error())
		return
	}
	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("RegisterService(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go n.regDiscover()
		return
	}
	blog.Info("RegisterService(%s:%s) succ", key, data)

	discvPath := commtypes.BCS_SERV_BASEPATH + "/" + commtypes.BCS_MODULE_NETWORKDETECTION
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go n.regDiscover()
		return
	}
	blog.Info("DiscoverService(%s) succ", discvPath)

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("tick: driver(%s:%d) running, current goroutine num (%d)",
				n.conf.Address, n.conf.Port, runtime.NumGoroutine())

		case event := <-discvEvent:
			if event.Err != nil {
				blog.Error("get discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go n.regDiscover()
				return
			}

			isRegstered := false
			for i, server := range event.Server {
				blog.Info("discovered : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("discovered : server[%d] is myself", i)
					isRegstered = true
				}
			}

			if isRegstered == false {
				blog.Warn("drive is not regestered in zk, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				go n.regDiscover()
				return
			}
		} // end select
	} // end for
}

func (n *NetworkDetection) initActions() {
	n.acts = []*httpserver.Action{
		httpserver.NewAction("GET", "/detectionpods", nil, n.getAllDetectionPods),
	}
}

// getAllDetectionPods xxx
// http hander func
// response []types.DetectionPod
func (n *NetworkDetection) getAllDetectionPods(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("hander detection pods request")

	n.RLock()
	pods := make([]*types.DetectionPod, 0)
	for _, deploy := range n.deploys {
		for _, o := range deploy.Pods {
			if o.Status != schedtypes.TASKGROUP_STATUS_RUNNING {
				blog.V(3).Infof("region(%s:%s) Pod %s status %s, not ready",
					deploy.Clusterid, deploy.Idc, o.ID, o.Status)
				continue
			}

			pod := &types.DetectionPod{
				Ip:        schedtypes.GetTaskgroupIP(o),
				Idc:       deploy.Idc,
				IdcUnit:   deploy.IdcUnit,
				ClusterId: deploy.Clusterid,
			}
			if pod.Ip == "" || pod.Idc == "" {
				blog.Warnf("region(%s:%s) Pod %s Ip %s Idc %s, not ready",
					deploy.Clusterid, deploy.Idc, o.ID, pod.Ip, pod.Idc)
				continue
			}
			pods = append(pods, pod)
		}
	}
	n.RUnlock()

	data := createResponeData(nil, "success", pods)
	resp.Write([]byte(data))
}

func createResponeData(err error, msg string, data interface{}) string {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(common.BcsErrMesosSchedCommon, msg)
	} else {
		rpyErr = errors.New(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))
	}
	return rpyErr.Error()
}
