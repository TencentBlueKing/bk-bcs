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

package rbac

import (
	"encoding/base64"
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"time"
)

const (
	wsMessageTypeHi   = "hi"
	wsMessageTypePing = "ping"
	wsMessageTypeData = "data"
)

var WorkerId string

type WebsocketRespnse struct {
	AppCode  string `json:"app_code"`
	Type     string `json:"type"`
	WorkerId string `json:"worker_id"`
	DataId   string `json:"data_id"`
	Data     string `json:"data"`
}

type AuthRbacData struct {
	Operation        string     `json:"operation"`
	Principal        Principals `json:"principal"`
	ScopeType        string     `json:"scope_type"`
	ScopeInstance    string     `json:"scope_instance"`
	Service          string     `json:"service"`
	Action           string     `json:"action"`
	ResourceType     string     `json:"resource_type"`
	ResourceInstance Resource   `json:"resource_instance"`
	PolicyFrom       string     `json:"policy_from"`
}

type Resource struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
}

type Principals struct {
	PrincipalType string `json:"principal_type"`
	PrincipalId   string `json:"principal_id"`
}

type WebsocketReq struct {
	AppCode  string `json:"app_code"`
	Type     string `json:"type"`
	WorkerId string `json:"worker_id"`
	DataId   string `json:"data_id"`
}

// sync rbac data from paas_auth subserver
func SyncRbacFromAuth() {
	subServerHost := config.BKIamAuth.BKIamAuthSubServer
	appCode := config.BKIamAuth.BKIamAuthAppCode
	appSecret := config.BKIamAuth.BKIamAuthAppSecret
	wsUrl := url.URL{Scheme: "ws", Host: subServerHost, Path: "/subscribers/bcs-api/sub"}

	dialer := &websocket.Dialer{}
	header := http.Header{}
	header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(appCode+":"+appSecret)))

	var err error
	var conn *websocket.Conn
	// create websocket connection to paas_auth subserver
CONNECTION:
	for i := 0; i < 3; i++ {
		conn, _, err = dialer.Dial(wsUrl.String(), header)
		if err != nil {
			blog.Errorf("unable to connect to paas_auth subserver: %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}
	if err != nil {
		goto CONNECTION
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			blog.Errorf("error when reading message from websocet: %s", err.Error())
			conn.Close()
			goto CONNECTION
		}

		var wsResp WebsocketRespnse
		err = json.Unmarshal(message, &wsResp)
		if err != nil {
			blog.Errorf("error decode json from websocket response, %s", err.Error())
			continue
		}

		if wsResp.AppCode == appCode && wsResp.Type == wsMessageTypeHi {
			WorkerId = wsResp.WorkerId
		}
		if wsResp.AppCode == appCode && wsResp.Type == wsMessageTypePing {
			blog.Info("receive paas_auth subserver ping message")
		}
		if wsResp.AppCode == appCode && wsResp.Type == wsMessageTypeData {
			authRbacDataJson := wsResp.Data
			var authRbacData AuthRbacData
			err = json.Unmarshal([]byte(authRbacDataJson), &authRbacData)
			if err != nil {
				blog.Errorf("error decode json from websocket data: %s", err.Error())
				continue
			}
			blog.Info("data received")
			err := syncAuthRbacData(&authRbacData)
			if err != nil {
				blog.Errorf("error when sync data from paas_auth_subserver. dataid: %s, data: %v, err: %s", wsResp.DataId, wsResp.Data, err.Error())
				continue
			}

			wsReq := WebsocketReq{
				AppCode:  wsResp.AppCode,
				Type:     "ack",
				WorkerId: wsResp.WorkerId,
				DataId:   wsResp.DataId,
			}
			reqBytes, err := json.Marshal(wsReq)
			if err != nil {
				blog.Errorf("error when marshal json data: %s", err.Error())
				continue
			}
			if err := conn.WriteMessage(1, reqBytes); err != nil {
				blog.Errorf("error when writing ack message to subserver: %s", err.Error())
				continue
			}
		}
	}

}
