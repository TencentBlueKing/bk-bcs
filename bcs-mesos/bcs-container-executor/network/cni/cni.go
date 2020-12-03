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

package cni

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network"

	"github.com/containernetworking/cni/libcni"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
)

const (
	//defaultCNIVersion = "0.3.0"
	defautNICName = "eth1"
)

//NewPlugin loading config for designated
//cni command line tool
func NewPlugin(binpath, confFile string) (string, network.NetworkPlugin) {
	var (
		conf        *libcni.NetworkConfig
		conflist    *libcni.NetworkConfigList
		err         error
		isCniList   = true
		networkName string
	)

	conflist, err = libcni.ConfListFromFile(confFile)
	if err != nil {
		logs.Errorf("Loading CNI config from %s err: %s\n", confFile, err.Error())
		isCniList = false

		conf, err = libcni.ConfFromFile(confFile)
		if err != nil {
			logs.Errorf("Loading CNI config from %s err: %s\n", confFile, err.Error())
			return "", nil
		}
		networkName = conf.Network.Name
	} else {
		networkName = conflist.Name
	}

	//if binpath is not set, use system executable path
	var path []string
	if binpath == "" {
		//Get executable bianry from system path
		syspathstr := os.Getenv("PATH")
		syspath := strings.Split(syspathstr, ":")
		path = append(path, syspath...)
	} else {
		path = append(path, binpath)
	}
	operator := &libcni.CNIConfig{
		Path: path,
	}
	plugin := &CNIPlugin{
		fileName:    filepath.Base(confFile),
		binDir:      binpath,
		netConf:     conf,
		netConfList: conflist,
		cniNet:      operator,
		isCniList:   isCniList,
		networkName: networkName,
	}
	return networkName, plugin
}

//CNIPlugin plugin for cni
type CNIPlugin struct {
	fileName    string                //config file name
	binDir      string                //binary file path
	netConf     *libcni.NetworkConfig //cni standard network configure
	netConfList *libcni.NetworkConfigList
	status      *network.NetStatus //status for network, only available after SetUpPod
	cniNet      *libcni.CNIConfig  //cni invoke depends on libcni
	networkName string             //network name
	isCniList   bool               // whether is plugin list
}

//Name Get plugin name
func (plugin *CNIPlugin) Name() string {
	return plugin.networkName
}

//Type Get plugin executable binary name
/*func (plugin *CNIPlugin) Type() string {
	return plugin.netConf.Network.Type
}*/

//Init init Plugin
func (plugin *CNIPlugin) Init(host string) error {
	return nil
}

//SetUpPod Setup Network info for pod
func (plugin *CNIPlugin) SetUpPod(podInfo container.Pod) error {
	//build runtime conf for libcni
	runConf := &libcni.RuntimeConf{
		ContainerID: podInfo.GetContainerID(),
		NetNS:       podInfo.GetNetns(),
		IfName:      defautNICName,
	}

	//setting network flow limit
	if len(podInfo.GetNetArgs()) > 0 {
		runConf.Args = append(runConf.Args, podInfo.GetNetArgs()...)
	}

	//setting ip address if needed
	if podInfo.Injection() {
		logs.Infof("CNI plugin %s ADD COMMAND with ip address %s\n", plugin.networkName, podInfo.GetIPAddr())
		runConf.Args = append(runConf.Args, [2]string{"IP", podInfo.GetIPAddr()})
	}

	by, _ := json.Marshal(runConf)
	logs.Infof("CNI plugin %s set conf %s", plugin.networkName, string(by))

	result, err := plugin.addNetworkV2(runConf)
	if err != nil {
		logs.Errorf("CNI plugin %s addNetwork error %s", plugin.networkName, err.Error())
		return err
	}

	if result.IPs == nil || len(result.IPs) == 0 {
		logs.Errorf("CNI plugin %s apply ip resource failed, lack of resource or netservice err, result: %s", plugin.networkName, result.String())
		return fmt.Errorf("lack of ip resource")
	}

	if !podInfo.Injection() {
		podInfo.SetIPAddr(result.IPs[0].Address.IP.String())
	}

	logs.Infof("plugin %s ADD network succ. plugin output: %s\n", plugin.networkName, result.String())
	return nil
}

//TearDownPod Teardown pod network info
func (plugin *CNIPlugin) TearDownPod(podInfo container.Pod) error {
	//build runtime conf for libcni
	runConf := &libcni.RuntimeConf{
		ContainerID: podInfo.GetContainerID(),
		NetNS:       podInfo.GetNetns(),
		IfName:      defautNICName,
	}

	//setting network flow limit
	if len(podInfo.GetNetArgs()) > 0 {
		runConf.Args = append(runConf.Args, podInfo.GetNetArgs()...)
	}

	//pod ip was injected by other, releasing it with ip address
	if podInfo.Injection() {
		logs.Infof("CNI plugin %s DEL command with ip address %s\n", plugin.networkName, podInfo.GetIPAddr())
		runConf.Args = append(runConf.Args, [2]string{"IP", podInfo.GetIPAddr()})
	}

	var err error

	if plugin.isCniList {
		err = plugin.cniNet.DelNetworkList(plugin.netConfList, runConf)
	} else {
		err = plugin.cniNet.DelNetwork(plugin.netConf, runConf)
	}
	if err != nil {
		logs.Errorf("CNI plugin %s DEL command for pod error: %s\n", plugin.networkName, err.Error())
		return err
	}

	logs.Infof("plugin %s DEL network succ.", plugin.networkName)
	return nil
}

func (plugin *CNIPlugin) addNetwork(runConf *libcni.RuntimeConf) (*current.Result, error) {
	result := &current.Result{
		Interfaces: make([]*current.Interface, 0),
		IPs:        make([]*current.IPConfig, 0),
		Routes:     make([]*cnitypes.Route, 0),
	}

	netConfList := make([]*libcni.NetworkConfig, 0)
	if plugin.isCniList {
		netConfList = append(netConfList, plugin.netConfList.Plugins...)

	} else {
		netConfList = append(netConfList, plugin.netConf)
	}

	for _, conf := range netConfList {
		by, _ := json.Marshal(conf)
		logs.Infof("cni net add network %s", string(by))

		r, err := plugin.cniNet.AddNetwork(conf, runConf)
		if err != nil {
			return nil, err
		}

		// Convert whatever the IPAM result was into the current Result type
		cr, err := current.NewResultFromResult(r)
		if err != nil {
			logs.Errorf("CNI plugin %s format json result err, %s, original result: %s\n", plugin.networkName, err.Error(), r.String())
			return nil, err
		}

		by, _ = json.Marshal(cr)
		logs.Infof("cni plugin %s result %s", plugin.networkName, string(by))

		result.Interfaces = append(result.Interfaces, cr.Interfaces...)
		result.IPs = append(result.IPs, cr.IPs...)
		result.Routes = append(result.Routes, cr.Routes...)
	}

	return result, nil
}

func (plugin *CNIPlugin) addNetworkV2(runConf *libcni.RuntimeConf) (*current.Result, error) {
	var err error
	var r cnitypes.Result
	if plugin.isCniList {
		r, err = plugin.cniNet.AddNetworkList(plugin.netConfList, runConf)
	} else {
		r, err = plugin.cniNet.AddNetwork(plugin.netConf, runConf)
	}
	if err != nil {
		return nil, err
	}

	// Convert whatever the IPAM result was into the current Result type
	cr, err := current.NewResultFromResult(r)
	if err != nil {
		logs.Errorf("CNI plugin %s format json result err, %s, original result: %s\n", plugin.networkName, err.Error(), r.String())
		return nil, err
	}

	by, _ := json.Marshal(cr)
	logs.Infof("cni plugin %s result %s", plugin.networkName, string(by))

	return cr, nil
}
