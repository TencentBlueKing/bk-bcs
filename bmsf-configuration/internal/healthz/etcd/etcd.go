/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package etcdz

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"bk-bscp/pkg/ssl"
)

// HealthInfo is etcd health info, e.g. '{"health":"true"}'.
type HealthInfo struct {
	// Health is state flag, it's string not boolean.
	Health string `json:"health"`
}

// Healthz checks the etcd health state.
func Healthz(host, caFile, certFile, keyFile, passwd string) (bool, error) {
	var err error
	var tlsConf *tls.Config

	scheme := "http"

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		if tlsConf, err = ssl.ClientTLSConfVerify(caFile, certFile, keyFile, passwd); err != nil {
			return false, err
		}
		scheme = "https"
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}

	resp, err := client.Get(fmt.Sprintf("%s://%s/health", scheme, host))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("response status[%+v]", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	info := &HealthInfo{}
	if err := json.Unmarshal(body, info); err != nil {
		return false, err
	}

	isHealth, err := strconv.ParseBool(info.Health)
	if err != nil {
		return false, err
	}
	return isHealth, nil
}
