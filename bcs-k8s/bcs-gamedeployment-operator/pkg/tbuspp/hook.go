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
package tbuspp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"net/http"
)
//
const (
	TbusppControllerService  = "tbuspp-controller.tbuspp-system"
	DeleteFailedRetrySeconds = 10

)
// CheckCanDelete Todo 异常要告警
func CheckCanDelete(podName string, podNameSpace string) bool {
	// Resp struct
	type Resp struct {
		Code        int    `json:"code"`
		ErrMsg      string `json:"err_msg"`
		//AccessToken string `json:"access_token"`
	}
	res := new(Resp)

	_, err := net.ResolveIPAddr("ip", TbusppControllerService)
	if err != nil {
		klog.Errorf("can not resolve %s , please check", TbusppControllerService)
		// no need check tbuspp before scale
		return false
	}
	TbusppControllerServiceUrl := "http://" + TbusppControllerService + ":10086/hpa-reduction/can-reduce"
	values := map[string]string{"pod_name": podName, "namespace":podNameSpace}

	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(TbusppControllerServiceUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		klog.Infof(" http.PostForm error %s", err.Error())
		return false
	}
	defer resp.Body.Close()
	// 验证结果回包
	if resp.StatusCode == 200 {
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Infof("ReadAll error %s", err.Error())
			return false
		}

		err = json.Unmarshal(respBytes, res)
		if err != nil {
			klog.Infof("Unmarshal error %s", err.Error())
			return false
		}
		if res.Code != 0 {
			klog.Warningf("check scale success,but cannot delete now, code %d err %s, please try later.",res.Code, res.ErrMsg)
			return false
		}
		fmt.Printf("check scale success, delete now.")
		return true

	}
	fmt.Printf(" check scale failed, code %d != 200. ", resp.StatusCode)
	return false
}
// 预退出接口
func PreDelete(podName string, podNameSpace string) bool {
	// Resp struct
	type Resp struct {
		Code        int    `json:"code"`
		ErrMsg      string `json:"err_msg"`
		//AccessToken string `json:"access_token"`
	}
	res := new(Resp)

	_, err := net.ResolveIPAddr("ip", TbusppControllerService)
	if err != nil {
		klog.Errorf("can not resolve %s , please check", TbusppControllerService)
		// no need check tbuspp before scale
		return false
	}
	TbusppControllerServiceUrl := "http://" + TbusppControllerService + ":10086/hpa-reduction/pre-reduce"
	values := map[string]string{"pod_name": podName, "namespace":podNameSpace}

	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(TbusppControllerServiceUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		klog.Infof(" http.PostForm error %s", err.Error())
		return false
	}
	defer resp.Body.Close()
	// 验证结果回包
	if resp.StatusCode == 200 {
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Infof("ReadAll error %s", err.Error())
			return false
		}

		err = json.Unmarshal(respBytes, res)
		if err != nil {
			klog.Infof("Unmarshal error %s", err.Error())
			return false
		}
		if res.Code != 0 {
			klog.Warningf("send pre delete success,but code %d not equal 0 , err %s, please try later.",res.Code, res.ErrMsg)
			return false
		}
		fmt.Printf("check scale success, delete now.")
		return true

	}
	fmt.Printf(" check scale failed, code %d != 200. ", resp.StatusCode)
	return false


}