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

package route

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	bcsHttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/rdiscover"
)

const (
	routeMesosURLPrefix         = "/bcsapi/v4/scheduler/mesos/"
	routeMesosCreateApplication = routeMesosURLPrefix + "namespaces/%s/applications"
	routeMesosDeleteApplication = routeMesosURLPrefix + "namespaces/%s/applications/%s?enforce=1"

	routeK8SGetClusterID     = "/rest/clusters/bcs/query_by_cluster_id/?cluster_id=%s"
	routeK8SURLPrefix        = "/tunnels/clusters/%s/apis/extensions/v1beta1/"
	routeK8SCreateDeployment = routeK8SURLPrefix + "namespaces/%s/deployments"
	routeK8SDeleteResource   = routeK8SURLPrefix + "namespaces/%s/%s/%s"
)

// Route the interface definition of a route
type Route interface {
	CreateMesos(clusterID, namespace string, data []byte) error
	DeleteMesos(clusterID, namespace, name string, data []byte) error
	CreateK8S(clusterID, namespace string, data []byte) error
	DeleteK8S(clusterID, namespace, resource, name string, data []byte) error
}

func New(cfg *config.Config, rd *rdiscover.RDiscover) (r Route, err error) {
	c := httpclient.NewHttpClient()
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("Accept", "application/json")
	if cfg.RouteClientCert.IsSSL {
		if err = c.SetTlsVerity(cfg.RouteClientCert.CAFile, cfg.RouteClientCert.CertFile, cfg.RouteClientCert.KeyFile, cfg.RouteClientCert.CertPasswd); err != nil {
			return
		}
	}

	token := []byte("")
	if cfg.ApiToken != "" {
		token, err = encrypt.DesDecryptFromBase([]byte(cfg.ApiToken))
		if err != nil {
			return
		}
	}

	r = &route{
		rd:       rd,
		client:   c,
		apiToken: string(token),
	}
	return
}

type route struct {
	rd     *rdiscover.RDiscover
	client *httpclient.HttpClient

	apiToken string
}

func (cli *route) CreateMesos(clusterID, namespace string, data []byte) (err error) {
	uri := fmt.Sprintf(routeMesosCreateApplication, namespace)
	blog.V(3).Infof("create mesos application: %s", uri)

	address, err := cli.rd.GetApiServer()
	if err != nil {
		blog.Errorf("failed to get api url: %v", err)
		return
	}

	header := http.Header{}
	header.Set("BCS-ClusterID", clusterID)
	header.Set("X-Bcs-User-Token", cli.apiToken)

	blog.Infof("mesos api create: %s | %s", uri, string(data))

	r, err := cli.client.Post(address+uri, header, data)
	if err != nil {
		blog.Errorf("failed to create mesos through api: %v", err)
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to create mesos through api: (%d)%s", r.StatusCode, r.Status)
		return
	}

	var resp bcsHttp.APIRespone
	if err = codec.DecJson(r.Reply, &resp); err != nil {
		err = fmt.Errorf("%v: %s", err, string(r.Reply))
		return
	}

	if resp.Code != common.BcsSuccess {
		err = fmt.Errorf("failed to create mesos through api: (%d)%s", resp.Code, resp.Message)
		return
	}

	return
}

func (cli *route) DeleteMesos(clusterID, namespace, name string, data []byte) (err error) {
	uri := fmt.Sprintf(routeMesosDeleteApplication, namespace, name)
	blog.V(3).Infof("delete mesos application: %s", uri)

	address, err := cli.rd.GetApiServer()
	if err != nil {
		blog.Errorf("failed to get api url: %v", err)
		return
	}

	header := http.Header{}
	header.Set("BCS-ClusterID", clusterID)
	header.Set("X-Bcs-User-Token", cli.apiToken)

	blog.Infof("mesos api delete: %s", uri)

	r, err := cli.client.Delete(address+uri, header, data)
	if err != nil {
		blog.Errorf("failed to delete mesos through api: %v", err)
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to delete mesos through api: (%d)%s", r.StatusCode, r.Status)
		return
	}

	var resp bcsHttp.APIRespone
	if err = codec.DecJson(r.Reply, &resp); err != nil {
		err = fmt.Errorf("%v: %s", err, string(r.Reply))
		return
	}

	if resp.Code == common.BcsErrMesosSchedNotFound {
		err = types.DeleteCollectorNotExist
		return
	}
	if resp.Code != common.BcsSuccess {
		err = fmt.Errorf("failed to delete mesos through api: (%d)%s", resp.Code, resp.Message)
		return
	}

	return
}

type K8SClusterIDResp struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
}

func (cli *route) getClusterID(clusterID string) (trueClusterID string, err error) {
	uri := fmt.Sprintf(routeK8SGetClusterID, clusterID)
	blog.V(3).Infof("get k8s clusterID: %s", uri)

	address, err := cli.rd.GetApiServer()
	if err != nil {
		blog.Errorf("failed to get api url: %v", err)
		return
	}

	r, err := cli.client.Get(address+uri, nil, nil)
	if err != nil {
		blog.Errorf("failed to get k8s clusterID trough api: %v", err)
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get k8s clusterID through api: (%d)%s", r.StatusCode, r.Status)
		return
	}

	var clusterIDResp K8SClusterIDResp
	if err = codec.DecJson(r.Reply, &clusterIDResp); err != nil {
		err = fmt.Errorf("decode when get k8s clusterID failed: %v: %s", err, string(r.Reply))
		return
	}

	trueClusterID = clusterIDResp.Identifier
	return
}

func (cli *route) CreateK8S(clusterID, namespace string, data []byte) (err error) {
	trueClusterID, err := cli.getClusterID(clusterID)
	if err != nil {
		blog.Errorf("create k8s get clusterID failed: %v", err)
		return err
	}

	uri := fmt.Sprintf(routeK8SCreateDeployment, trueClusterID, namespace)
	blog.V(3).Infof("create k8s deployment: %s", uri)

	address, err := cli.rd.GetApiServer()
	if err != nil {
		blog.Errorf("failed to get api url: %v", err)
		return
	}

	header := http.Header{}
	header.Set("BCS-ClusterID", clusterID)
	header.Set("X-Bcs-User-Token", cli.apiToken)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", cli.apiToken))

	blog.Infof("k8s api create: %s | %s", uri, string(data))

	r, err := cli.client.Post(address+uri, header, data)
	if err != nil {
		blog.Errorf("failed to create k8s through api: %v", err)
		return
	}

	if r.StatusCode != http.StatusCreated {
		err = fmt.Errorf("failed to create k8s through api: (%d)%s, body: %s", r.StatusCode, r.Status, string(r.Reply))
		return
	}

	return
}

func (cli *route) DeleteK8S(clusterID, namespace, resource, name string, data []byte) (err error) {
	trueClusterID, err := cli.getClusterID(clusterID)
	if err != nil {
		blog.Errorf("delete k8s get clusterID failed: %v", err)
		return err
	}

	uri := fmt.Sprintf(routeK8SDeleteResource, trueClusterID, namespace, resource, name)
	blog.V(3).Infof("delete k8s deployment: %s", uri)

	address, err := cli.rd.GetApiServer()
	if err != nil {
		blog.Errorf("failed to get api url: %v", err)
		return
	}

	header := http.Header{}
	header.Set("BCS-ClusterID", clusterID)
	header.Set("X-Bcs-User-Token", cli.apiToken)
	header.Set("Authorization", fmt.Sprintf("Bearer %s", cli.apiToken))

	blog.Infof("k8s api delete: %s | %s", uri, string(data))

	r, err := cli.client.Delete(address+uri, header, data)
	if err != nil {
		blog.Errorf("failed to delete k8s through api: %v", err)
		return
	}

	if r.StatusCode == http.StatusNotFound {
		err = types.DeleteCollectorNotExist
		return
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to delete k8s through api: (%d)%s, body: %s", r.StatusCode, r.Status, string(r.Reply))
		return
	}

	return
}
