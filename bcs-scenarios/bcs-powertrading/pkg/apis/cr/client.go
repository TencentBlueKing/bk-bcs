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
 */

// Package cr xxx
package cr

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

// Client interface
type Client interface {
	GetPerfDetail(req *GetPerfDetailReq) (*GetPerfDetailRsp, error)
}

// New client
func New(opt *apis.ClientOptions, r requester.Requester) Client {
	cli := &client{
		requestClient: r,
		opt:           opt,
	}
	err := cli.SetDefaultHeader()
	if err != nil {
		return nil
	}
	return cli
}

type client struct {
	defaultHeader map[string]string
	requestClient requester.Requester
	opt           *apis.ClientOptions
}

// SetDefaultHeader set default header
func (c *client) SetDefaultHeader() error {
	header := make(map[string]string)
	auth := &apis.BkAuthOpts{
		BkAppCode:   c.opt.AppCode,
		BkAppSecret: c.opt.AppSecret,
		AccessToken: c.opt.AccessToken,
	}
	str, err := json.Marshal(auth)
	if err != nil {
		blog.Errorf("marshal header error:%s", err.Error())
		return err
	}
	header["X-Bkapi-Authorization"] = string(str)
	header["user"] = c.opt.UserName
	header["Content-Type"] = "application/json"
	c.defaultHeader = header
	return nil
}

// GetPerfDetail get perf detail
func (c *client) GetPerfDetail(req *GetPerfDetailReq) (*GetPerfDetailRsp, error) {
	url := fmt.Sprintf(getPerfDetailUrl, c.opt.Endpoint)
	data, err := json.Marshal(req)
	if err != nil {
		blog.Errorf("Error encoding JSON: %v", err)
	}
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, data)
	if requestErr != nil {
		return nil, fmt.Errorf("do GetPerfDetail request error:%s", requestErr.Error())
	}
	result := &GetPerfDetailRsp{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("GetPerfDetail request failed, code:%d, url: %s, message:%s",
			result.Code, url, result.Message)
	}
	return result, nil
}
