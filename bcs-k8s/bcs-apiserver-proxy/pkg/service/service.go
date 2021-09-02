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

package service

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/ipvs"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/utils"
)

//LvsProxy is lvs virtualServer and realServer operation interface
type LvsProxy interface {
	// CreateVirtualServer create the specified VirtualServer by vs
	CreateVirtualServer(vs string) error
	// IsVirtualServerAvailable check vs available or not
	IsVirtualServerAvailable(vs string) bool
	// DeleteVirtualServer delete vs form host
	DeleteVirtualServer(vs string) error
	// CreateRealServer create real server
	CreateRealServer(rs string) error
	// ListRealServer list real servers by lvs
	ListRealServer() ([]string, error)
	// DeleteRealServer delete real server
	DeleteRealServer(rs string) error
}

// NewLvsProxy init LvsProxy interface
func NewLvsProxy() LvsProxy {
	l := &lvsProxy{}
	l.handle = ipvs.New()

	return l
}

type lvsProxy struct {
	vs     *utils.EndPoint
	lock   sync.Mutex
	rs     []*utils.EndPoint
	handle ipvs.Interface
}

// CreateVirtualServer create virtual server and set lvsProxy.vs by vs, return err when create fails
func (l *lvsProxy) CreateVirtualServer(vs string) error {
	virIP, virPort := utils.SplitServer(vs)
	if virIP == "" || virPort == 0 {
		blog.Error("CreateVirtualServer error: virtual server ip and port is empty")
		return fmt.Errorf("virtual server ip and port is empty")
	}
	// set virtual server
	l.vs = &utils.EndPoint{IP: virIP, Port: virPort}

	vServer := utils.BuildVirtualServer(vs)
	err := l.handle.AddVirtualServer(vServer)
	if errors.Is(err, syscall.EEXIST) {
		blog.Debug("CreateRealServer exist: ", err)
		return nil
	}
	if err != nil {
		blog.Warn("CreateVirtualServer error: ", err)
		return fmt.Errorf("new virtual server failed: %s", err)
	}

	return nil
}

// DeleteVirtualServer delete virtual server if exist
func (l *lvsProxy) DeleteVirtualServer(vs string) error {
	vIP, vPort := utils.SplitServer(vs)
	if vIP == "" || vPort == 0 {
		blog.Error("DeleteVirtualServer error: real server ip and port is empty ")
		return fmt.Errorf("virtual server ip and port is null")
	}
	virServer := utils.BuildVirtualServer(vs)
	err := l.handle.DeleteVirtualServer(virServer)
	if err != nil {
		blog.Warn("DeleteVirtualServer error: ", err)
		return err
	}

	l.vs = nil
	return nil
}

// IsVirtualServerAvailable check vs available or not, return true when vs is available
func (l *lvsProxy) IsVirtualServerAvailable(vs string) bool {
	isExist := false

	virIP, virPort := utils.SplitServer(vs)
	if virIP == "" || virPort == 0 {
		blog.Error("splitServer error: virtual server ip and port is empty")
		return false
	}

	// list all vs on the host
	virArray, err := l.handle.GetVirtualServers()
	if err != nil {
		blog.Warn("IsVirtualServerAvailable warn: vir servers is empty: ", err)
		return isExist
	}

	resultVirServer := utils.BuildVirtualServer(vs)
	for _, vir := range virArray {
		blog.Infof("IsVirtualServerAvailable debug: check vir ip: %s, port %v ", vir.Address.String(), vir.Port)
		if vir.String() == resultVirServer.String() {
			isExist = true
		}
	}

	if isExist {
		l.vs = &utils.EndPoint{
			IP:   virIP,
			Port: virPort,
		}
	}

	return isExist
}

// CreateRealServer create real server endpoint when virtual server exist, return err if vs not exist or create failed
func (l *lvsProxy) CreateRealServer(rs string) error {
	realIP, realPort := utils.SplitServer(rs)
	if realIP == "" || realPort == 0 {
		blog.Error("CreateRealServer error: real server ip and port is empty")
		return fmt.Errorf("real server ip and port is empty")
	}
	rsEp := &utils.EndPoint{IP: realIP, Port: realPort}

	l.lock.Lock()
	l.rs = append(l.rs, rsEp)
	l.lock.Unlock()

	realServer := utils.BuildRealServer(rs)

	if l.vs == nil {
		errMsg := fmt.Sprintf("CreateRealServer[%s] error: virtual server is empty.", rs)
		blog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	// virtual server build rs server
	vServer := utils.BuildVirtualServer(l.vs.String())
	err := l.handle.AddRealServer(vServer, realServer)
	if errors.Is(err, syscall.EEXIST) {
		blog.Debug("CreateRealServer exist: ", err)
		return nil
	}
	if err != nil {
		blog.Error("CreateRealServer error: ", err)
		return fmt.Errorf("new real server failed: %s", err)
	}

	return nil
}

// ListRealServer get backend lvs's rs servers
func (l *lvsProxy) ListRealServer() ([]string, error) {
	if l.vs == nil {
		return nil, fmt.Errorf("ListRealServer failed, lvsProxy l.vs is empty")
	}

	vs := utils.BuildVirtualServer(l.vs.String())
	dstArray, err := l.handle.GetRealServers(vs)
	if err != nil {
		blog.Errorf("GetRealServers failed: %s; %v ", vs, err)
		return nil, err
	}

	rsList := []string{}
	for _, rs := range dstArray {
		if rs != nil {
			rsList = append(rsList, rs.String())
		}
	}

	return rsList, nil
}

// GetRealServer get vip's real server by rsHost, return rs=nil when rsHost not exist and need to add real server
func (l *lvsProxy) GetRealServer(rsHost string) (*utils.EndPoint, int) {
	ip, port := utils.SplitServer(rsHost)

	// get virtual server backend rs
	vs := utils.BuildVirtualServer(l.vs.String())
	dstArray, err := l.handle.GetRealServers(vs)
	if err != nil {
		blog.Error("GetRealServer error[get real server failed]: %s;  %d; %v ", ip, port, err)
		return nil, 0
	}

	dip := net.ParseIP(ip)
	for _, dst := range dstArray {
		blog.Infof("GetRealServer debug[check real server ip]: %s;  %d; %v ", dst.Address.String(), dst.Port, err)
		if dst.Address.Equal(dip) && dst.Port == port {
			return &utils.EndPoint{IP: ip, Port: port}, dst.Weight
		}
	}
	return nil, 0
}

// DeleteRealServer delete real server
func (l *lvsProxy) DeleteRealServer(rs string) error {
	realIP, realPort := utils.SplitServer(rs)
	if realIP == "" || realPort == 0 {
		blog.Error("DeleteRealServer error: real server ip and port is empty ")
		return fmt.Errorf("real server ip and port is null")
	}

	if l.vs == nil {
		blog.Error("DeleteRealServer error: virtual service is empty.")
		return errors.New("virtual service is empty")
	}

	virServer := utils.BuildVirtualServer(l.vs.String())
	realServer := utils.BuildRealServer(rs)
	err := l.handle.DeleteRealServer(virServer, realServer)
	if err != nil {
		blog.Error("DeleteRealServer error[real server delete error]: ", err)
		return fmt.Errorf("real server delete error: %v", err)
	}

	// update rs server
	var resultRS []*utils.EndPoint
	for _, r := range l.rs {
		if r.IP == realIP && r.Port == realPort {
			continue
		}
		resultRS = append(resultRS, &utils.EndPoint{
			IP:   r.IP,
			Port: r.Port,
		})
	}

	l.lock.Lock()
	l.rs = resultRS
	l.lock.Unlock()

	return nil
}
