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

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"errors"
	regd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bresp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	cli "github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func newStatusController(zkServers string, tls TLSConfig) (*Status, error) {
	blog.Infof("staring status controller")
	disc := regd.NewRegDiscoverEx(zkServers, time.Second*5)
	if err := disc.Start(); nil != err {
		return nil, fmt.Errorf("start get ccapi zk service failed. Error:%v", err)
	}

	watchPath := fmt.Sprintf("%s/%s/master", types.BCS_SERV_BASEPATH, types.BCS_MODULE_HEALTH)
	eventChan, eventErr := disc.DiscoverService(watchPath)
	if nil != eventErr {
		return nil, fmt.Errorf("start running discover service failed. Error:%v", eventErr)
	}
	status := &Status{
		discoverChan: eventChan,
		tlsConfig:    tls,
		Servers: Servers{
			healthServers: make(map[string]*cli.HttpClient),
		},
	}

	blog.Infof("start status controller success.")
	return status, nil
}

//Servers health http api server
type Servers struct {
	locker        sync.RWMutex
	healthServers map[string]*cli.HttpClient
}

//Status health http api server status
type Status struct {
	tlsConfig    TLSConfig
	Servers      Servers
	discoverChan <-chan *regd.DiscoverEvent
}

func (s *Status) run() {
	go func() {
		blog.Infof("start to sync bcs-health address from zk.")
		for svr := range s.discoverChan {
			blog.Info("received one zk event which may contains bcs-health address.")
			if svr.Err != nil {
				blog.Errorf("get bcs-health addr failed. but will continue watch. err: %v", svr.Err)
				continue
			}
			if len(svr.Server) <= 0 {
				s.resetServers()
				blog.Warnf("get 0 bcs-health addr from zk, reset health servers.")
				continue
			}
			s.updateServers(svr.Server[0])
		}
	}()
}

func (s *Status) resetServers() {
	s.Servers.locker.Lock()
	defer s.Servers.locker.Unlock()
	s.Servers.healthServers = make(map[string]*cli.HttpClient)
}

func (s *Status) updateServers(svr string) {
	s.Servers.locker.Lock()
	defer s.Servers.locker.Unlock()

	info := types.ServerInfo{}
	if err := json.Unmarshal([]byte(svr), &info); nil != err {
		blog.Errorf("unmashal health server info failed. reason: %v", err)
		return
	}
	if len(info.IP) == 0 || info.Port == 0 || len(info.Scheme) == 0 {
		blog.Errorf("get invalid health info: %s", svr)
		return
	}
	addr := fmt.Sprintf("%s://%s:%d", info.Scheme, info.IP, info.Port)
	if _, exist := s.Servers.healthServers[addr]; exist {
		return
	}
	client := cli.NewHttpClient()
	client.SetHeader("Content-Type", "application/json")

	if info.Scheme == "https" {
		if err := client.SetTlsVerity(s.tlsConfig.CaFile,
			s.tlsConfig.CertFile,
			s.tlsConfig.KeyFile,
			static.ClientCertPwd); err != nil {
			blog.Errorf("setting tls configuration failed, %s", err.Error())
		}
	}
	s.Servers.healthServers[addr] = client
	blog.Infof("*** get new bcs-health client, addr: %s ***", addr)

	for svr := range s.Servers.healthServers {
		if svr != addr {
			delete(s.Servers.healthServers, svr)
			blog.Infof("*** remove old bcs-health server, addr: %s ***.", svr)
		}
	}
}

// TryDoRequest try to request with speified data
func (s *Status) TryDoRequest(data string) error {
	url := "/bcshealth/v1/create"
	return s.doOneRequest(url, data)
}

// TryDoAlarmRequest  alarm request with specified data
func (s *Status) TryDoAlarmRequest(data string) error {
	url := "/bcshealth/v1/sendalarm"
	return s.doOneRequest(url, data)
}

//func (s *Status) doOneRequest(client *cli.HttpClient, req *http.Request) error {
func (s *Status) doOneRequest(url, data string) error {
	server, client := s.getServers()
	if nil == client {
		return errors.New("oops, no health server can be used")
	}
	req, err := http.NewRequest("POST", server+url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("request health server: %s failed, err: %+v", server, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, rerr := client.GetClient().Do(req)
	if rerr != nil {
		return rerr
	}
	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	var response bresp.APIRespone
	if err := json.Unmarshal(body, &response); nil != err {
		return fmt.Errorf("unmarshal response body failed. err: %v", err)
	}
	if !response.Result {
		return fmt.Errorf("response failed. err: %s", response.Message)
	}
	blog.V(4).Infof("do request to url: %s success, response: %s", req.URL.String(), string(body))
	return nil
}

func (s *Status) getServers() (string, *cli.HttpClient) {
	s.Servers.locker.Lock()
	defer s.Servers.locker.Unlock()
	for svr, c := range s.Servers.healthServers {
		return svr, c
	}
	return "", nil
}
