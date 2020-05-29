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

package scaler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
)

type bcsMesosScaler struct {
	//hpa controller config
	config *config.Config

	//http bcs client
	cli *httpclient.HttpClient

	//mesos driver address:port
	currmesosdriver string
}

func NewBcsMesosScalerController(config *config.Config) ScalerProcess {
	scaler := &bcsMesosScaler{
		config: config,
	}

	//init http client
	scaler.initCli()
	//start discovery mesos driver
	go scaler.discvMesosdriver()
	return scaler
}

//scale deployment
func (r *bcsMesosScaler) ScaleDeployment(namespace, name string, instance uint) error {
	uri := fmt.Sprintf("/namespaces/%s/deployments/%s/scale/%d", namespace, name, instance)
	_, err := r.requestMesosdriverV4("PUT", uri, nil)
	return err
}

//scale application
func (r *bcsMesosScaler) ScaleApplication(namespace, name string, instance uint) error {
	uri := fmt.Sprintf("/namespaces/%s/applications/%s/scale/%d", namespace, name, instance)
	_, err := r.requestMesosdriverV4("PUT", uri, nil)
	return err
}

func (r *bcsMesosScaler) registerCrdAutoscaler() {
	autoscaler := &commtype.Crr{
		TypeMeta: commtype.TypeMeta{
			APIVersion: "v4",
			Kind:       "crr",
		},
		Spec: commtype.CrrSpec{
			Names: commtype.CrrSpecName{
				Kind: "autoscaler",
			},
		},
	}

	by, _ := json.Marshal(autoscaler)
	_, err := r.requestMesosdriverV4("POST", "/crr/register", by)
	if err != nil {
		blog.Errorf("register crr autoscaler error %s", err)
	} else {
		blog.Infof("register crr autoscaler success")
	}
}

func (r *bcsMesosScaler) initCli() {
	r.cli = httpclient.NewHttpClient()

	if r.config.ClientCert != nil && r.config.ClientCert.IsSSL {
		r.cli.SetTlsVerity(r.config.ClientCert.CAFile, r.config.ClientCert.CertFile, r.config.ClientCert.KeyFile,
			r.config.ClientCert.CertPasswd)
	}

	r.cli.SetHeader("Content-Type", "application/json")
	r.cli.SetHeader("Accept", "application/json")
	r.cli.SetHeader("BCS-ClusterID", r.config.ClusterID)
}

func (r *bcsMesosScaler) discvMesosdriver() {
	blog.Infof("bcsMesosScaler begin to discover mesosdriver from (%s)", r.config.ClusterZkAddr)

	MesosDiscv := r.config.BcsZkAddr
	ClusterID := r.config.ClusterID
	regDiscv := rd.NewRegDiscover(MesosDiscv)
	if regDiscv == nil {
		blog.Errorf("new mesosdriver discover(%s) return nil", MesosDiscv)
		time.Sleep(3 * time.Second)
		go r.discvMesosdriver()
		return
	}

	blog.Infof("new mesosdriver discover(%s) succ", MesosDiscv)
	err := regDiscv.Start()
	if err != nil {
		blog.Errorf("mesosdriver discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go r.discvMesosdriver()
		return
	}

	blog.Infof("mesosdriver discover start succ")
	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSAPISERVER + "/" + ClusterID
	discvMesosEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Errorf("watch mesosdriver under (%s: %s) error(%s)", MesosDiscv, discvPath, err.Error())
		regDiscv.Stop()
		time.Sleep(3 * time.Second)
		go r.discvMesosdriver()
		return
	}

	blog.Infof("watch mesosdriver under (%s: %s)", MesosDiscv, discvPath)
	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Infof("mesosdriver discover(%s:%s), curr mesosdriver:%s", MesosDiscv, discvPath, r.currmesosdriver)

		case event := <-discvMesosEvent:
			blog.Infof("discover event for mesosdriver")
			if event.Err != nil {
				blog.Errorf("get mesosdriver discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				time.Sleep(3 * time.Second)
				go r.discvMesosdriver()
				return
			}

			currMesosdriver := ""
			blog.Infof("get mesosdriver node num(%d)", len(event.Server))
			for i, server := range event.Server {
				blog.Infof("get mesosdriver: server[%d]: %s %s", i, event.Key, server)

				var serverInfo commtype.MesosDriverServInfo
				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("fail to unmarshal mesosdriver(%s), err:%s", string(server), err.Error())
				}

				if i == 0 {
					currMesosdriver = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))
				}
			}

			if currMesosdriver != r.currmesosdriver {
				blog.Infof("mesosdriver changed(%s-->%s)", r.currmesosdriver, currMesosdriver)
				r.currmesosdriver = currMesosdriver
				r.registerCrdAutoscaler()
			}
		} // select
	} // for
}

func (r *bcsMesosScaler) requestMesosdriverV4(method, uri string, data []byte) ([]byte, error) {
	if r.currmesosdriver == "" {
		return nil, fmt.Errorf("there is no mesosdriver")
	}

	uri = fmt.Sprintf("%s/mesosdriver/v4%s", r.currmesosdriver, uri)
	blog.V(3).Infof("request uri %s data %s", uri, string(data))

	var resp *httpclient.HttpRespone
	var err error

	switch method {
	case "GET":
		resp, err = r.cli.Get(uri, nil, data)

	case "POST":
		resp, err = r.cli.Post(uri, nil, data)

	case "DELETE":
		resp, err = r.cli.Delete(uri, nil, data)

	case "PUT":
		resp, err = r.cli.Put(uri, nil, data)

	default:
		err = fmt.Errorf("uri %s method %s is invalid", uri, method)
	}

	if err != nil {
		return nil, fmt.Errorf("request %s error %s", uri, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request %s response http code %d data %s", uri, resp.StatusCode, string(resp.Reply))
	}

	return resp.Reply, nil
}
