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
	"fmt"
	"io/ioutil"
	"strings"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	bhttp "bk-bcs/bcs-common/common/http"
	"bk-bcs/bcs-common/common/http/httpclient"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"bk-bcs/bcs-services/bcs-api/regdiscv"

	"github.com/emicklei/go-restful"
)

const (
	BcsApiPrefix = "/bcsapi/v4/storage/"
)

func init() {
	actions.RegisterAction(actions.Action{"POST", "/bcsapi/v4/storage/{uri:*}", nil, handlerPostActions})
	actions.RegisterAction(actions.Action{"PUT", "/bcsapi/v4/storage/{uri:*}", nil, handlerPutActions})
	actions.RegisterAction(actions.Action{"GET", "/bcsapi/v4/storage/{uri:*}", nil, handlerGetActions})
	actions.RegisterAction(actions.Action{"DELETE", "/bcsapi/v4/storage/{uri:*}", nil, handlerDeleteActions})
}

func request2storage(req *restful.Request, uri, method string) (string, error) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("handler url %s read request body failed, error: %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrCommHttpReadBodyFail, common.BcsErrCommHttpReadBodyFailStr)
		return err1.Error(), nil
	}

	rd, err := regdiscv.GetRDiscover()
	if err != nil {
		blog.Error("hander url %s get RDiscover error %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiInternalFail, common.BcsErrApiInternalFailStr)
		return err1.Error(), nil
	}

	serv, err := rd.GetModuleServers(types.BCS_MODULE_STORAGE)
	if err != nil {
		blog.Error("get servers %s error %s", types.BCS_MODULE_STORAGE, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiGetStorageFail, common.BcsErrApiGetStorageFailStr)
		return err1.Error(), nil
	}

	ser, ok := serv.(*types.BcsStorageInfo)
	if !ok {
		blog.Errorf("servers convert to BcsStorageInfo")
		err1 := bhttp.InternalError(common.BcsErrApiGetStorageFail, common.BcsErrApiGetStorageFailStr)
		return err1.Error(), nil
	}

	host := fmt.Sprintf("%s://%s:%d", ser.Scheme, ser.IP, ser.Port)
	url := fmt.Sprintf("%s/bcsstorage/v1/%s", host, uri)
	blog.V(3).Infof("do request to url(%s), method(%s)", url, method)

	httpcli := httpclient.NewHttpClient()
	httpcli.SetHeader("Content-Type", "application/json")
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
		blog.Error("request url %s error %s", url, err.Error())
		err1 := bhttp.InternalError(common.BcsErrApiRequestMesosApiFail, common.BcsErrApiRequestMesosApiFailStr)
		return err1.Error(), nil
	}

	return string(reply), err
}

func handlerPostActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)

	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2storage(req, url, "POST")
	resp.Write([]byte(data))
}

func handlerGetActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2storage(req, url, "GET")
	resp.Write([]byte(data))
}

func handlerDeleteActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2storage(req, url, "DELETE")
	resp.Write([]byte(data))
}

func handlerPutActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := strings.Replace(req.Request.URL.Path, BcsApiPrefix, "", 1)

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := request2storage(req, url, "PUT")
	resp.Write([]byte(data))
}
