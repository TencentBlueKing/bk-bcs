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

package netservice

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/regdiscv"

	"github.com/emicklei/go-restful"
	"github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type APIResponse struct {
	Result  bool        `json:"result"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Data    interface{} `json:"data"`
}

func handlerActions(req *restful.Request, resp *restful.Response) {
	start := time.Now()

	uri := req.PathParameter("uri")
	data, err := request2netservice(req, uri)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("net_service", req.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("net_service", req.Request.Method).Observe(time.Since(start).Seconds())
		blog.Error("get netservice server failed! err: ", err.Error())
		resp.WriteHeaderAndEntity(
			http.StatusBadRequest,
			APIResponse{
				Result:  false,
				Message: err.Error(),
				Code:    "10",
				Data:    nil,
			})
		return
	}

	metric.RequestCount.WithLabelValues("net_service", req.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("net_service", req.Request.Method).Observe(time.Since(start).Seconds())

	resp.Write(data)
}

func request2netservice(req *restful.Request, uri string) (respBody []byte, err error) {
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("handler url %s read request body failed, error: %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrCommHttpReadBodyFail, common.BcsErrCommHttpReadBodyFailStr)
		return nil, err1
	}

	rd, err := regdiscv.GetRDiscover()
	if err != nil {
		blog.Error("hander url %s get RDiscover error %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiInternalFail, common.BcsErrApiInternalFailStr)
		return nil, err1
	}

	serv, err := rd.GetModuleServers(types.BCS_MODULE_NETSERVICE)
	if err != nil {
		blog.Error("get servers %s error %s", types.BCS_MODULE_NETSERVICE, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiGetNetserviceFail, common.BcsErrApiGetNetserviceFailStr)
		return nil, err1
	}

	ser, ok := serv.(*types.NetServiceInfo)
	if !ok {
		blog.Errorf("servers convert to NetServiceInfo")
		err1 := bhttp.InternalError(common.BcsErrApiGetNetserviceFail, common.BcsErrApiGetNetserviceFailStr)
		return nil, err1
	}

	host := fmt.Sprintf("%s://%s:%d", ser.Scheme, ser.IP, ser.Port)
	url := fmt.Sprintf("%s/v1/%s", host, uri)
	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	// Use client certs if given
	goReq := gorequest.New()
	if strings.ToLower(ser.Scheme) == "https" {
		cliTls, err := rd.GetClientTls()
		if err != nil {
			blog.Errorf("get client tls error %s", err.Error())
		} else {
			goReq = goReq.TLSClientConfig(cliTls)
		}

	}

	goReq.Set("Content-Type", "application/json")
	reflect.ValueOf(goReq).MethodByName(strings.Title(strings.ToLower(req.Request.Method))).Call([]reflect.Value{reflect.ValueOf(url)})
	if len(body) != 0 {
		goReq.SendString(json.Get(body).ToString())
	}
	resp, _, errs := goReq.End()
	blog.Info("request netservice, url: %s, method: %s", url, req.Request.Method)
	if len(errs) != 0 {
		blog.Error("http request failed. err: %v", errs[0])
	}

	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, err
}

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: "/bcsapi/v4/netservice/{uri:*}", Params: nil, Handler: handlerActions})
	actions.RegisterAction(actions.Action{Verb: "PUT", Path: "/bcsapi/v4/netservice/{uri:*}", Params: nil, Handler: handlerActions})
	actions.RegisterAction(actions.Action{Verb: "GET", Path: "/bcsapi/v4/netservice/{uri:*}", Params: nil, Handler: handlerActions})
	actions.RegisterAction(actions.Action{Verb: "DELETE", Path: "/bcsapi/v4/netservice/{uri:*}", Params: nil, Handler: handlerActions})
}
