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

package tunnel

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestAuthorizeTunnel test the authorizeTunnel func
func TestAuthorizeTunnel(t *testing.T) {
	req := &http.Request{
		Header: make(http.Header),
	}
	_, _, err := authorizeTunnel(req)
	if err == nil || !strings.Contains(err.Error(), "module empty") {
		t.Error("failed to handle request whit empty module")
	}

	req.Header.Set(Module, "kube-agent")
	_, _, err = authorizeTunnel(req)
	if err == nil || !strings.Contains(err.Error(), "registerToken empty") {
		t.Error("failed to handle request whit empty registerToken")
	}

	req.Header.Set(RegisterToken, "abcdefg")
	_, _, err = authorizeTunnel(req)
	if err == nil || !strings.Contains(err.Error(), "clusterId empty") {
		t.Error("failed to handle request whit empty cluster")
	}

	req.Header.Set(Cluster, "k8s-001")
	_, _, err = authorizeTunnel(req)
	if err == nil {
		t.Error("failed to handle request whit empty Params")
	}

	params := map[string]interface{}{
		"address": "http:127.0.0.1:80",
	}
	bytes, err := json.Marshal(params)
	req.Header.Set(Params, base64.StdEncoding.EncodeToString(bytes))
	_, _, err = authorizeTunnel(req)
	if err == nil || !strings.Contains(err.Error(), "address or cacert or token empty") {
		t.Error("failed to handle request whit empty cacert or usertoken when module is bcs-kube-agent")
	}
}
