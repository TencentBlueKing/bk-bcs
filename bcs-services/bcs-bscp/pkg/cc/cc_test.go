/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cc

import (
	"encoding/json"
	"log"
	"net"
	"testing"
)

func TestCC(t *testing.T) {
	InitService(APIServerName)

	sys := &SysOption{
		ConfigFiles: []string{"../../cmd/api-server/etc/api_server.yaml"},
		BindIP:      net.IPv4(127, 0, 0, 1),
	}

	if err := LoadSettings(sys); err != nil {
		log.Println(err)
	}

	server := ApiServer()
	marshal, err := json.Marshal(server)
	if err != nil {
		log.Println(err)
	}

	if string(marshal) != "{\"Network\":{\"BindIP\":\"127.0.0.1\",\"RpcPort\":0,\"HttpPort\":8080,\"TLS\":"+
		"{\"InsecureSkipVerify\":false,\"CertFile\":\"\",\"KeyFile\":\"\",\"CAFile\":\"\",\"Password\":\"\"}},"+
		"\"ServiceName\":{\"Etcd\":{\"Endpoints\":[\"127.0.0.1:2379\"],\"DialTimeoutMS\":200,\"Username\":\"\","+
		"\"Password\":\"\",\"TLS\":{\"InsecureSkipVerify\":false,\"CertFile\":\"\",\"KeyFile\":\"\",\"CAFile\""+
		":\"\",\"Password\":\"\"}}},\"Log\":{\"LogDir\":\"./log\",\"MaxPerFileSizeMB\":1024,\"MaxPerLineSizeKB\""+
		":2,\"MaxFileNum\":5,\"LogAppend\":false,\"ToStdErr\":false,\"AlsoToStdErr\":false,\"Verbosity\":0},\"Repo\""+
		":{\"Endpoints\":[\"http://127.0.0.1:2379\"],\"Token\":\"xxxxx\",\"Project\":\"bk_bscp\",\"User\":\"admin\","+
		"\"TLS\":{\"InsecureSkipVerify\":false,\"CertFile\":\"\",\"KeyFile\":\"\",\"CAFile\":\"\","+
		"\"Password\":\"\"}}}" {
		t.Errorf("cc is not expected, ApiServer: %+v", server)
		return
	}
}
