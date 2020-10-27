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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/context"
)

// plugin must implement
// func GetHostAttributes([]string) (map[string]*types.HostAttributes,error)
// func input: ip list, example: []string{"127.0.0.1"}
// func ouput: map key = ip, example: map["127.0.0.1"] = &types.HostAttributes{}
// implement func Init(para *types.InitPluginParameter) error
// func input: *types.InitPluginParameter
// func output: error

//for example

//CertConfig is configuration of Cert
type certConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

type config struct {
	BcsZk          string `json:"bcsZk"`
	CAFile         string `json:"ca_file" value:"" usage:"CA file. If server_cert_file/server_key_file/ca_file are all set, it will set up an HTTPS server required and verified client cert" mapstructure:"ca_file"`
	ClientCertFile string `json:"client_cert_file" value:"" usage:"Client public key file(*.crt)" mapstructure:"client_cert_file"`
	ClientKeyFile  string `json:"client_key_file" value:"" usage:"Client private key file(*.key)" mapstructure:"client_key_file"`
	cert           *certConfig
}

const (
	defaultIPResouces = 1
)

var (
	initPara       *plugin.InitPluginParameter
	conf           *config
	currNetservice string
	cli            *httpclient.HttpClient

	isReady bool
	cxt     context.Context
	cancel  context.CancelFunc
)

// Init init ip-resource plugin
func Init(para *plugin.InitPluginParameter) error { //nolint
	initPara = para

	return initPlugin(initPara.ConfPath)
}

func initPlugin(p string) error {
	blog.Infof("plugin ip-resources init...")
	//go watchConfig()

	err := initConfig(p)
	if err != nil {
		blog.Errorf("plugin ip-resources init error %s", err.Error())
		return err
	}

	initCli()
	cxt, cancel = context.WithCancel(context.Background())

	go discvNetservice()
	isReady = true

	blog.Infof("plugin ip-resources init done")
	return nil
}

// Uninit release ip-resource plugin
func Uninit() error { //nolint
	stop()
	return nil
}

func stop() {
	isReady = false
	currNetservice = ""
	cancel()
}

func watchConfig() {
	blog.Infof("plugin ip-resources watchConfig")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		blog.Errorf("plugin ip-resources NewWatcher error %s", err.Error())
		time.Sleep(time.Second)
		go watchConfig()
		return
	}
	defer watcher.Close()

	err = watcher.Add(initPara.ConfPath)
	if err != nil {
		blog.Errorf("plugin ip-resources watch %s error %s", initPara.ConfPath, err.Error())
		time.Sleep(time.Second)
		go watchConfig()
		return
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-watcher.Events:
				blog.Infof("plugin ip-resources watch %s event", initPara.ConfPath)
				stop()
				initPlugin(initPara.ConfPath)
				done <- true
				return
			case err := <-watcher.Errors:
				blog.Errorf("plugin ip-resources watch %s error %s", initPara.ConfPath, err.Error())
			}
		}
	}()

	<-done
	blog.Infof("plugin ip-resources watch %s done", initPara.ConfPath)
}

// GetHostAttributes plugin interface implementation
func GetHostAttributes(para *plugin.HostPluginParameter) (map[string]*plugin.HostAttributes, error) { //nolint
	atrrs := make(map[string]*plugin.HostAttributes)

	if !isReady {
		return nil, fmt.Errorf("plugin ip-resources is not ready")
	}

	resp, err := getHostIPResources(para)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 || len(resp.Resource) == 0 {
		return nil, fmt.Errorf("request netservice error %s", resp.Message)
	}

	for _, ip := range para.Ips {
		hostAttr := &plugin.HostAttributes{
			Ip:         ip,
			Attributes: make([]*plugin.Attribute, 0),
		}

		var number int
		var ok bool

		if number, ok = resp.Resource[ip]; !ok || number < 0 {
			number = defaultIPResouces
		}

		atrri := &plugin.Attribute{
			Name:   plugin.SlaveAttributeIpResources,
			Type:   plugin.ValueScalar,
			Scalar: plugin.Value_Scalar{Value: float64(number)},
		}

		hostAttr.Attributes = append(hostAttr.Attributes, atrri)
		atrrs[ip] = hostAttr
	}

	return atrrs, nil
}

func initConfig(p string) error {
	f, err := os.Open(fmt.Sprintf("%s/ip-resources.conf", p))
	if err != nil {
		return err
	}

	by, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(by, &conf)
	if err != nil {
		return err
	}

	conf.cert = &certConfig{}

	if conf.CAFile != "" {
		conf.cert.CertFile = conf.ClientCertFile
		conf.cert.KeyFile = conf.ClientKeyFile
		conf.cert.CAFile = conf.CAFile
		conf.cert.IsSSL = true
		conf.cert.CertPasswd = static.ClientCertPwd
	}

	return nil
}

func initCli() {
	cli = httpclient.NewHttpClient()

	if conf.cert.IsSSL {
		cli.SetTlsVerity(conf.cert.CAFile, conf.cert.CertFile, conf.cert.KeyFile,
			conf.cert.CertPasswd)
	}

	cli.SetHeader("Content-Type", "application/json")
	cli.SetHeader("Accept", "application/json")
}

func discvNetservice() {
	blog.Infof("plugin ipResources begin to discover netservice from (%s)", conf.BcsZk)

	regDiscv := rd.NewRegDiscover(conf.BcsZk)
	if regDiscv == nil {
		blog.Errorf("plugin ipResources netservice discover(%s) return nil", conf.BcsZk)
		time.Sleep(3 * time.Second)
		go discvNetservice()
		return
	}
	blog.Infof("plugin ipResources new netservice discover(%s) succ", conf.BcsZk)

	err := regDiscv.Start()
	if err != nil {
		blog.Errorf("plugin ipResources netservice discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go discvNetservice()
		return
	}

	blog.Infof("plugin ipResources netservice discover start succ")

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_NETSERVICE
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Errorf("plugin ipResources watch netservice under (%s: %s) error(%s)", conf.BcsZk, discvPath, err.Error())
		regDiscv.Stop()
		time.Sleep(3 * time.Second)
		go discvNetservice()
		return
	}

	blog.Infof("plugin ipResources watch netservice under (%s: %s)", conf.BcsZk, discvPath)

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-cxt.Done():
			blog.Infof("plugin ipResources done")
			return

		case <-tick.C:
			blog.Infof("plugin ipResources netservice discover(%s:%s), curr netservice:%s", conf.BcsZk, discvPath, currNetservice)

		case event := <-discvEvent:
			blog.Infof("plugin ipResources discover event for netservice")

			if event.Err != nil {
				blog.Errorf("plugin ipResources get netservice discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				time.Sleep(3 * time.Second)
				go discvNetservice()
				return
			}

			currNet := ""
			blog.Infof("plugin ipResources get netservice node num(%d)", len(event.Server))

			for i, server := range event.Server {
				blog.Infof("plugin ipResources get netservice: server[%d]: %s %s", i, event.Key, server)

				var serverInfo commtype.NetServiceInfo

				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("plugin ipResources fail to unmarshal netservice(%s), err:%s", string(server), err.Error())
				}

				if i == 0 {
					currNet = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))
				}
			}

			if currNet != currNetservice {
				blog.Infof("plugin ipResources netservice changed(%s-->%s)", currNetservice, currNet)
				currNetservice = currNet
			}
		} // select
	} // for
}

func getHostIPResources(para *plugin.HostPluginParameter) (*responseBody, error) {
	req := requestPara{
		Cluster: para.ClusterId,
		Hosts:   para.Ips,
	}

	by, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	by, err = requestNetservice("POST", "resource", by)
	if err != nil {
		return nil, err
	}

	var resp *responseBody
	err = json.Unmarshal(by, &resp)

	return resp, err
}

type requestPara struct {
	Cluster string   `json:"cluster"`
	Hosts   []string `json:"hosts"`
}

type responseBody struct {
	Code     int            `json:"code"`
	Message  string         `json:"message"`
	Resource map[string]int `json:"resource"`
}

func requestNetservice(method, uri string, data []byte) ([]byte, error) {
	if currNetservice == "" {
		return nil, fmt.Errorf("there is no netservice")
	}

	uri = fmt.Sprintf("%s/v1/%s", currNetservice, uri)

	blog.V(3).Infof("plugin ipResources request uri %s data %s", uri, string(data))

	var by []byte
	var err error

	switch method {
	case "GET":
		by, err = cli.GET(uri, nil, data)

	case "POST":
		by, err = cli.POST(uri, nil, data)

	case "DELETE":
		by, err = cli.DELETE(uri, nil, data)

	case "PUT":
		by, err = cli.PUT(uri, nil, data)

	default:
		err = fmt.Errorf("uri %s method %s is invalid", uri, method)
	}

	blog.V(3).Infof("plugin ipResources reqeust uri %s response data %s", uri, string(by))

	return by, err
}
