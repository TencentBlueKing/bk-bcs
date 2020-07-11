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

package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/storage/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/rdiscover"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// Param storage parameter definition
type Param struct {
	ClusterID   string
	ClusterType types.ClusterType
	Type        string
	Namespace   string
	Name        string
	Parameters  map[string]string
	Field       []string
	Extra       map[string]string
	Data        interface{}
}

// Storage the interface definition of a storage
type Storage interface {
	QueryDynamic(param *Param) ([]byte, error)
	GetClusters() ([]string, error)
	GetDynamicNs(param *Param) ([]byte, error)
	SaveMetric(param *Param) error
	QueryMetric(param *Param) ([]byte, error)
	DeleteMetric(param *Param) error
	GetDynamicWatcher(param *Param) (watcher *watch.Watcher, err error)
	GetMetricWatcher(param *Param) (watcher *watch.Watcher, err error)
}

func New(cfg *config.Config, rd *rdiscover.RDiscover) (s Storage, err error) {
	c := httpclient.NewHttpClient()
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("Accept", "application/json")
	if cfg.RouteClientCert.IsSSL {
		if err = c.SetTlsVerity(cfg.RouteClientCert.CAFile, cfg.RouteClientCert.CertFile, cfg.RouteClientCert.KeyFile, cfg.RouteClientCert.CertPasswd); err != nil {
			return
		}
	}

	s = &storage{
		rd:     rd,
		client: c,
	}
	return
}

type storage struct {
	rd     *rdiscover.RDiscover
	client *httpclient.HttpClient
}

var (
	storagePrefix      = "%s/bcsstorage/v1/"
	storageGetClusters = storagePrefix + "metric/clusters"
	storageQueryMetric = storagePrefix + "metric/clusters/%s"
	storageWatchMetric = storagePrefix + "metric/watch/%s/%s"
	storageDealMetric  = storagePrefix + "metric/clusters/%s/namespaces/%s/%s/%s"

	storageQueryDynamic         = storagePrefix + "query/%s/dynamic/clusters/%s/%s"
	storageWatchDynamic         = storagePrefix + "dynamic/watch/containers/%s"
	storageGetNsDynamic         = storagePrefix + "%s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s"
	storageGetNsDynamicWithName = storagePrefix + "%s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s/%s"
)

func (cli *storage) fetchDynamic(address string) (interface{}, error) {

	blog.Infof("fetchDynamic: %s", address)
	rsp, rspErr := cli.client.Get(address, nil, nil)
	if nil != rspErr {
		return nil, rspErr
	}

	if http.StatusOK == rsp.StatusCode {

		var brsp bhttp.APIRespone
		jserr := json.Unmarshal(rsp.Reply, &brsp)
		if nil != jserr {
			blog.Error("can not unmarshal the storage response, %s", string(rsp.Reply))
			return nil, jserr
		}

		if common.BcsSuccess != brsp.Code {
			blog.Error("failed to query the metric, error %s", brsp.Message)
			return nil, fmt.Errorf("%s", brsp.Message)
		}

		return brsp.Data, nil
	}

	return nil, fmt.Errorf("failed to query the metric, error code is %d, errorinfo is %s", rsp.StatusCode, rsp.Status)
}

func (cli *storage) query(uri string) (resp *bhttp.APIRespone, raw []byte, err error) {
	var r *httpclient.HttpRespone

	if r, err = cli.client.Get(uri, nil, nil); err != nil {
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to query, http(%d)%s: %s", r.StatusCode, r.Status, uri)
		return
	}

	if err = codec.DecJson(r.Reply, &resp); err != nil {
		err = fmt.Errorf("%v: %s", err, string(r.Reply))
		return
	}

	if resp.Code != common.BcsSuccess {
		err = fmt.Errorf("failed to query, resp(%d)%s: %s", resp.Code, resp.Message, uri)
		return
	}

	if err = codec.EncJson(resp.Data, &raw); err != nil {
		return
	}
	return
}

func (cli *storage) GetClusters() ([]string, error) {
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return nil, err
	}

	uri := fmt.Sprintf(storageGetClusters, address)
	blog.V(3).Infof("get clusters: %s", uri)
	resp, _, err := cli.query(uri)
	if err != nil {
		blog.Errorf("%v", err)
		return nil, err
	}

	dataArr, dataOk := resp.Data.([]interface{})
	if !dataOk {
		err = fmt.Errorf("the storage data struct error")
		blog.Error("%v", err)
		return nil, err
	}

	clusters := make([]string, 0)
	for _, itemArr := range dataArr {
		itemVal, valOk := itemArr.(string)
		if !valOk {
			blog.Error("the storage data type error: %v", reflect.TypeOf(itemArr).Kind())
			continue
		}
		clusters = append(clusters, itemVal)
	}
	return clusters, nil
}

func (cli *storage) QueryDynamic(param *Param) (b []byte, err error) {
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return
	}

	parameters := ""
	if param.Namespace != "" {
		parameters = "?namespace=" + param.Namespace
	}

	uri := fmt.Sprintf(storageQueryDynamic, address, param.ClusterType.String(), param.ClusterID, param.Type) + parameters
	blog.V(3).Infof("query dynamic: %s", uri)

	if _, b, err = cli.query(uri); err != nil {
		blog.Errorf("%v", err)
		return
	}
	return
}

// SaveMetric  save metric
func (cli *storage) SaveMetric(param *Param) (err error) {
	blog.Info("storage save metric(%s), %+v", param.Type, param)
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return
	}

	uri := fmt.Sprintf(storageDealMetric, address, param.ClusterID, param.Namespace, param.Type, param.Name)
	blog.V(3).Infof("save metric: %s", uri)

	var data []byte
	if err = codec.EncJson(map[string]interface{}{"data": param.Data}, &data); err != nil {
		blog.Errorf("failed to encode metric data: %v", err)
		return
	}

	r, err := cli.client.Put(uri, nil, data)
	if err != nil {
		blog.Errorf("save metric to storage failed: %v", err)
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("save metric to storage failed: (%d)%s", r.StatusCode, r.Status)
		return
	}

	var resp *bhttp.APIRespone
	if err = codec.DecJson(r.Reply, &resp); err != nil {
		err = fmt.Errorf("%v: %s", err, string(r.Reply))
		return
	}

	if resp.Code != common.BcsSuccess {
		err = fmt.Errorf("save metric to storage failed: (%d)%s", resp.Code, resp.Message)
		return
	}
	return
}

func (cli *storage) DeleteMetric(param *Param) (err error) {
	blog.Info("storage delete metric(%s), %+v", param.Type, param)
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return
	}

	uri := fmt.Sprintf(storageDealMetric, address, param.ClusterID, param.Namespace, param.Type, param.Name)
	blog.V(3).Infof("delete metric: %s", uri)

	r, err := cli.client.Delete(uri, nil, nil)
	if err != nil {
		blog.Errorf("delete metric from storage failed: %v", err)
		return
	}

	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("delete metric from storage failed: (%d)%s", r.StatusCode, r.Status)
		return
	}

	var resp *bhttp.APIRespone
	if err = codec.DecJson(r.Reply, &resp); err != nil {
		err = fmt.Errorf("%v: %s", err, string(r.Reply))
		return
	}

	if resp.Code != common.BcsSuccess && resp.Code != common.BcsErrStorageResourceNotExist {
		err = fmt.Errorf("delete metric from storage failed: (%d)%s", resp.Code, resp.Message)
		return
	}
	return
}

func (cli *storage) QueryMetric(param *Param) ([]byte, error) {
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return nil, err
	}

	parameters := url.Values{}
	for k, v := range param.Parameters {
		parameters.Set(k, v)
	}
	if param.Type != "" {
		parameters.Set("type", param.Type)
	}
	if param.Namespace != "" {
		parameters.Set("namespace", param.Namespace)
	}
	if param.Name != "" {
		parameters.Set("name", param.Name)
	}

	uri := fmt.Sprintf(storageQueryMetric, address, param.ClusterID)
	if len(parameters) > 0 {
		uri += "?" + parameters.Encode()
	}

	blog.V(3).Infof("query metric uri: %s", uri)
	_, raw, err := cli.query(uri)
	if err != nil {
		blog.Errorf("%v", err)
		return nil, err
	}
	blog.V(3).Infof("query metric result: %s", string(raw))
	return raw, nil
}

func (cli *storage) GetDynamicNs(param *Param) ([]byte, error) {
	address, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("failed to get storage url: %v", err)
		return nil, err
	}

	parameters := url.Values{}
	if len(param.Field) > 0 {
		parameters.Set("field", strings.Join(param.Field, ","))
	}
	if param.Extra != nil && len(param.Extra) > 0 {
		var tmp []byte
		if err = codec.EncJson(param.Extra, &tmp); err != nil {
			return nil, err
		}
		parameters.Set("extra", base64.StdEncoding.EncodeToString(tmp))
	}

	uri := fmt.Sprintf(storageGetNsDynamic, address, param.ClusterType.String(), param.ClusterID, param.Namespace, param.Type)
	if param.Name != "" {
		uri = fmt.Sprintf(storageGetNsDynamicWithName, address, param.ClusterType.String(), param.ClusterID, param.Namespace, param.Type, param.Name)
	}

	if len(parameters) > 0 {
		uri += "?" + parameters.Encode()
	}

	blog.V(3).Infof("query ns dynamic uri: %s", uri)
	_, raw, err := cli.query(uri)
	if err != nil {
		blog.Errorf("%v", err)
		return nil, err
	}
	blog.V(3).Infof("query ns dynamic result: %s", string(raw))
	return raw, nil
}

func (cli *storage) GetDynamicWatcher(param *Param) (watcher *watch.Watcher, err error) {
	blog.Infof("storage: start watch dynamic: %s %s", param.ClusterID, param.Type)

	watcher = watch.NewWithOption(&operator.WatchOptions{MustDiff: "data"}, cli.client)
	addr, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("get storage server failed: %v", err)
		return
	}
	uri := fmt.Sprintf(storageWatchDynamic, addr, param.ClusterID)
	blog.V(3).Infof("try to watch: %s", uri)
	if err = watcher.Connect([]string{uri}); err != nil {
		blog.Errorf("connect to storage watcher failed: %v", err)
		return
	}
	return
}

func (cli *storage) GetMetricWatcher(param *Param) (watcher *watch.Watcher, err error) {
	blog.Infof("storage: start watch metric: %s %s", param.ClusterID, param.Type)

	watcher = watch.NewWithOption(&operator.WatchOptions{MustDiff: "data.version"}, cli.client)
	addr, err := cli.rd.GetStorageServer()
	if err != nil {
		blog.Errorf("get storage server failed: %v", err)
		return
	}
	uri := fmt.Sprintf(storageWatchMetric, addr, param.ClusterID, param.Type)
	blog.V(3).Infof("try to watch: %s", uri)
	if err = watcher.Connect([]string{uri}); err != nil {
		blog.Errorf("connect to storage watcher failed: %v", err)
		return
	}
	return
}
