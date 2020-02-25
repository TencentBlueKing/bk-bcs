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

package mesos

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	bhttp "bk-bcs/bcs-common/common/http"
	"bk-bcs/bcs-common/common/http/httpclient"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-api/metric"
	"bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"bk-bcs/bcs-services/bcs-api/regdiscv"

	"github.com/emicklei/go-restful"
	"github.com/ghodss/yaml"
)

const (
	//BcsApiPrefix prefix for mesos container scheduler
	BcsApiPrefix = "/bcsapi/v4/scheduler/mesos/"

	//mediaHeader key for http media content type
	medieTypeHeader = "Content-Type"
	//mediaTypeApplicationJSON json payload for http body
	mediaTypeApplicationJSON = "application/json"
	//mediaTypeApplicationYaml yaml payload for http body
	mediaTypeApplicationYaml = "application/x-yaml"
)

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: "/bcsapi/v4/scheduler/mesos/{uri:*}", Params: nil, Handler: handlerPostActions})
	actions.RegisterAction(actions.Action{Verb: "PUT", Path: "/bcsapi/v4/scheduler/mesos/{uri:*}", Params: nil, Handler: handlerPutActions})
	actions.RegisterAction(actions.Action{Verb: "GET", Path: "/bcsapi/v4/scheduler/mesos/{uri:*}", Params: nil, Handler: handlerGetActions})
	actions.RegisterAction(actions.Action{Verb: "DELETE", Path: "/bcsapi/v4/scheduler/mesos/{uri:*}", Params: nil, Handler: handlerDeleteActions})
}

func request2mesosapi(req *restful.Request, uri, method string) (string, error) {
	start := time.Now()

	data, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Error("handler url %s read request body failed, error: %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrCommHttpReadBodyFail, common.BcsErrCommHttpReadBodyFailStr)
		return err1.Error(), nil
	}
	//check application media type
	if mediaTypeApplicationYaml == req.Request.Header.Get(medieTypeHeader) {
		data, err = yamlTOJSON(data)
		if err != nil {
			blog.Errorf("bcs-api handle url %s yaml to json failed, %s", uri, err.Error())
			mediaErr := bhttp.InternalError(common.BcsErrApiMediaTypeError, common.BcsErrApiMediaTypeErrorStr)
			return mediaErr.Error(), nil
		} else {
			blog.V(3).Infof("bcs-api handle url %s converting yaml to json successfully", uri)
		}
	}

	cluster := req.Request.Header.Get("BCS-ClusterID")
	if cluster == "" {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Error("handler url %s read header BCS-ClusterID is empty", uri)
		err1 := bhttp.InternalError(common.BcsErrCommHttpParametersFailed, "http header BCS-ClusterID can't be empty")
		return err1.Error(), nil
	}

	rd, err := regdiscv.GetRDiscover()
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Error("hander url %s get RDiscover error %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiInternalFail, common.BcsErrApiInternalFailStr)
		return err1.Error(), nil
	}

	serv, err := rd.GetModuleServers(fmt.Sprintf("%s/%s", types.BCS_MODULE_MESOSAPISERVER, cluster))
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Error("get cluster %s servers %s error %s", cluster, types.BCS_MODULE_MESOSAPISERVER, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiGetMesosApiFail, fmt.Sprintf("mesos cluster %s not found", cluster))
		return err1.Error(), nil
	}

	ser, ok := serv.(*types.BcsMesosApiserverInfo)
	if !ok {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Errorf("servers convert to BcsMesosApiserverInfo")
		err1 := bhttp.InternalError(common.BcsErrApiGetMesosApiFail, common.BcsErrApiGetMesosApiFailStr)
		return err1.Error(), nil
	}

	//host := servInfo.Scheme + "://" + servInfo.IP + ":" + strconv.Itoa(int(servInfo.Port))
	var host string
	if ser.ExternalIp != "" && ser.ExternalPort != 0 {
		host = fmt.Sprintf("%s://%s:%d", ser.Scheme, ser.ExternalIp, ser.ExternalPort)
	} else {
		host = fmt.Sprintf("%s://%s:%d", ser.Scheme, ser.IP, ser.Port)
	}
	//url := routeHost + "/api/v1/" + uri //a.Conf.BcsRoute
	url := fmt.Sprintf("%s/mesosdriver/v4/%s", host, uri)
	blog.V(3).Infof("do request to url(%s), method(%s)", url, method)

	httpcli := httpclient.NewHttpClient()
	httpcli.SetHeader(medieTypeHeader, "application/json")
	httpcli.SetHeader("Accept", "application/json")
	if strings.ToLower(ser.Scheme) == "https" {
		cliTls, err := rd.GetClientTls()
		if err != nil {
			blog.Errorf("get client tls error %s", err.Error())
		}
		httpcli.SetTlsVerityConfig(cliTls)
	}

	reply, err := httpcli.Request(url, method, req.Request.Header, data)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("mesos", method).Inc()
		metric.RequestErrorLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
		blog.Error("request url %s error %s", url, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiRequestMesosApiFail, common.BcsErrApiRequestMesosApiFailStr)
		return err1.Error(), nil
	}

	metric.RequestCount.WithLabelValues("mesos", method).Inc()
	metric.RequestLatency.WithLabelValues("mesos", method).Observe(time.Since(start).Seconds())
	return string(reply), err
}

func handlerPostActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)

	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2mesosapi(req, url, "POST")
	resp.Write([]byte(data))
}

func handlerGetActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2mesosapi(req, url, "GET")
	resp.Write([]byte(data))
}

func handlerDeleteActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2mesosapi(req, url, "DELETE")
	resp.Write([]byte(data))
}

func handlerPutActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2mesosapi(req, url, "PUT")
	resp.Write([]byte(data))
}

//yamlTOJSON check if mesos request body is yaml,
// then convert yaml to json
func yamlTOJSON(rawData []byte) ([]byte, error) {
	if len(rawData) == 0 {
		return nil, nil
	}
	return yaml.YAMLToJSON(rawData)
}
