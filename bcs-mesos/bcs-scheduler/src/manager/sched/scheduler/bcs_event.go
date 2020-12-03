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

package scheduler

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

const (
	//MaxEventQueueLength event queue size
	MaxEventQueueLength = 10240
)

type bcsEventManager struct {
	sync.RWMutex

	bcsZk string
	//clientCertDir string
	clientCAFile   string
	clientCertFile string
	clientKeyFile  string

	currstorage string

	storageIsSsl bool
	clientCert   *commtypes.CertConfig
	cli          *httpclient.HttpClient

	eventQueue chan *commtypes.BcsStorageEventIf
}

// Create Event Manager
func newBcsEventManager(config util.Scheduler) *bcsEventManager {
	bcsEvent := &bcsEventManager{
		bcsZk: config.BcsZK,
		//clientCertDir:config.ClientCertDir,
		clientCAFile:   config.ClientCAFile,
		clientCertFile: config.ClientCertFile,
		clientKeyFile:  config.ClientKeyFile,
		eventQueue:     make(chan *commtypes.BcsStorageEventIf, MaxEventQueueLength),
	}

	//if bcsEvent.clientCertDir != "" {
	bcsEvent.clientCert = &commtypes.CertConfig{
		CertFile:   bcsEvent.clientCertFile,
		KeyFile:    bcsEvent.clientKeyFile,
		CAFile:     bcsEvent.clientCAFile,
		CertPasswd: static.ClientCertPwd,
	}
	//}

	return bcsEvent
}

// Run Event Manager
func (e *bcsEventManager) Run() {
	go e.discvstorage()
	go e.handleEventQueue()
}

func (e *bcsEventManager) initCli() {
	e.cli = httpclient.NewHttpClient()

	if e.storageIsSsl && e.clientCert != nil {
		e.cli.SetTlsVerity(e.clientCert.CAFile, e.clientCert.CertFile, e.clientCert.KeyFile,
			e.clientCert.CertPasswd)
	}

	e.cli.SetHeader("Content-Type", "application/json")
	e.cli.SetHeader("Accept", "application/json")
}

// Send Event
func (e *bcsEventManager) syncEvent(event *commtypes.BcsStorageEventIf) error {
	queue := len(e.eventQueue)
	if queue > 1024 {
		blog.Infof("bcsEventManager syncEvent %v queue(%d)", event, len(e.eventQueue))
	}
	e.eventQueue <- event
	return nil
}

func (e *bcsEventManager) discvstorage() {
	blog.Infof("bcsEventManager begin to discover storage from (%s), curr goroutine num(%d)", e.bcsZk, runtime.NumGoroutine())

	regDiscv := rd.NewRegDiscover(e.bcsZk)
	if regDiscv == nil {
		blog.Errorf("new storage discover(%s) return nil", e.bcsZk)
		time.Sleep(3 * time.Second)
		go e.discvstorage()
		return
	}

	blog.Infof("new storage discover(%s) succ, current goroutine num(%d)", e.bcsZk, runtime.NumGoroutine())

	err := regDiscv.Start()
	if err != nil {
		blog.Errorf("storage discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go e.discvstorage()
		return
	}

	blog.Infof("storage discover start succ, current goroutine num(%d)", runtime.NumGoroutine())

	discvPath := commtypes.BCS_SERV_BASEPATH + "/" + commtypes.BCS_MODULE_STORAGE
	discvStorageEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Errorf("watch storage under (%s: %s) error(%s)", e.bcsZk, discvPath, err.Error())
		regDiscv.Stop()
		time.Sleep(3 * time.Second)
		go e.discvstorage()
		return
	}

	blog.Infof("watch storage under (%s: %s), current goroutine num(%d)", e.bcsZk, discvPath, runtime.NumGoroutine())

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Infof("storage discover(%s:%s), curr storage:%s", e.bcsZk, discvPath, e.currstorage)

		case event := <-discvStorageEvent:
			blog.Infof("discover event for storage")
			if event.Err != nil {
				blog.Errorf("get storage discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				time.Sleep(3 * time.Second)
				go e.discvstorage()
				return
			}

			curr := ""
			blog.Infof("get storage node num(%d)", len(event.Server))

			for i, server := range event.Server {
				blog.Infof("get storage: server[%d]: %s %s", i, event.Key, server)

				var serverInfo commtypes.BcsStorageInfo

				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("fail to unmarshal storage(%s), err:%s", string(server), err.Error())
				}

				if i == 0 {
					curr = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))

					if serverInfo.ServerInfo.Scheme == "https" {
						e.storageIsSsl = true
					} else {
						e.storageIsSsl = false
					}
				}
			}

			if curr != e.currstorage {
				e.initCli()
				blog.Infof("storage changed(%s-->%s)", e.currstorage, curr)
				e.currstorage = curr
			}
		} // select
	} // for
}

func (e *bcsEventManager) handleEventQueue() {

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	var err error

	for {

		select {
		case <-tick.C:
			blog.V(3).Info("bcsEventManager handle event queue")

		case event := <-e.eventQueue:
			err = e.handleEvent(event)
			if err != nil {
				blog.Error("bcsEventManager handleEvent %v error %s", event, err.Error())
			} else {
				blog.V(3).Infof("bcsEventManager handleEvent %v success", event)
			}

		}

	}
}

func (e *bcsEventManager) handleEvent(event *commtypes.BcsStorageEventIf) error {
	by, _ := json.Marshal(event)

	uri := "events"
	begin := time.Now().UnixNano() / 1e6
	_, err := e.requestStorageV1("PUT", uri, by)
	end := time.Now().UnixNano() / 1e6
	useTime := end - begin
	if useTime > 100 {
		blog.Warnf("request storage event, %dms slow query", useTime)
	}
	return err
}

func (e *bcsEventManager) requestStorageV1(method, uri string, data []byte) ([]byte, error) {
	if e.currstorage == "" {
		return nil, fmt.Errorf("there is no storage")
	}

	uri = fmt.Sprintf("%s/bcsstorage/v1/%s", e.currstorage, uri)

	blog.V(3).Infof("request uri %s data %s", uri, string(data))

	var by []byte
	var err error

	switch method {
	case "GET":
		by, err = e.cli.GET(uri, nil, data)

	case "POST":
		by, err = e.cli.POST(uri, nil, data)

	case "DELETE":
		by, err = e.cli.DELETE(uri, nil, data)

	case "PUT":
		by, err = e.cli.PUT(uri, nil, data)

	default:
		err = fmt.Errorf("uri %s method %s is invalid", uri, method)
	}

	return by, err
}
