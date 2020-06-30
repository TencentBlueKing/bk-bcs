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

package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"net/http"
	"time"
)

type output struct {
	cfg    *config.Config
	client *http.Client
	input  chan *InputMessage
}

func (cli *output) resetClient() error {

	return nil
}
func (cli *output) run(ctx context.Context) {

	// 接收数据
	for {
		select {
		case msg := <-cli.input:

			msg.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
			js, err := json.Marshal(msg)
			if err != nil {
				blog.Errorf("marshal input message failed, skip send data. dataid[%d]err: %v", msg.DataID, err)
				continue
			}
			blog.V(5).Infof("send data: %s", string(js))
			var req *http.Request
			var errReq error

			address, err := cli.cfg.Rd.GetExporterServer()
			if nil != err {
				blog.Error("failed to get exporter server, error is %s", err.Error())
				continue
			}

			url := fmt.Sprintf("%s/api/v1/export/type/%d/data/%d", address, cli.cfg.ExporterType, msg.DataID)
			req, errReq = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(js))

			if errReq != nil {
				blog.Error("request error:%s", errReq.Error())
				continue
			}

			rsp, err := cli.client.Do(req)
			if err != nil {
				blog.Error("failed to do request, error info is %s", err.Error())
				continue
			}

			// TODO: need to clsoe response
			rsp.Body.Close()

		case <-ctx.Done():
			blog.Error("exit from loop")
			return
		}
	}
}

func (cli *output) Input(msg *InputMessage) error {

	select {
	case cli.input <- msg:
	default:
		blog.Warn("the chan is full ")
	}
	return nil
}
