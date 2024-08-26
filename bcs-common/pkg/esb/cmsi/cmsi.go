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

// Package cmsi xxx
package cmsi

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// ClientInterface client interface for cmsi
type ClientInterface interface {
	SendRtx(req *SendRtxReq) (*SendRtxResp, error)
	SendVoiceMsg(req *SendVoiceMsgReq) (*SendVoiceMsgResp, error)
	SendWeixin(req *SendWeixinReq) (*SendWeixinResp, error)
	SendMail(req *SendMailReq) (*SendMailResp, error)
}

// NewClientInterface create client interface
func NewClientInterface(host string, tlsConf *tls.Config) *Client {
	var cli *paasclient.RESTClient
	if tlsConf != nil {
		cli = paasclient.NewRESTClientWithTLS(tlsConf).
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	} else {
		cli = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(1000, 1000))
	}
	return &Client{
		host:   host,
		client: cli,
	}
}

// Client paas cmdb client
type Client struct {
	host          string
	defaultHeader http.Header
	client        *paasclient.RESTClient
	credential    Credential
}

// Credential credential to be filled in post body
type Credential struct {
	BKAppCode   string `json:"bk_app_code"`
	BKAppSecret string `json:"bk_app_secret"`
	BKUsername  string `json:"bk_username,omitempty"`
}

// SetDefaultHeader set default headers
func (c *Client) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// GetHeader get headers
func (c *Client) GetHeader() http.Header {
	authBytes, _ := json.Marshal(c.credential)
	c.defaultHeader.Add("X-Bkapi-Authorization", string(authBytes))
	return c.defaultHeader
}

// WithCredential set credential
func (c *Client) WithCredential(appCode, appSecret, username string) {
	c.credential = Credential{
		BKAppCode:   appCode,
		BKAppSecret: appSecret,
		BKUsername:  username,
	}
}

// SendRtx send rtx message
func (c *Client) SendRtx(req *SendRtxReq) (*SendRtxResp, error) {
	resp := new(SendRtxResp)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cmsi/").
		SubPathf("send_rtx").
		WithHeaders(c.GetHeader()).
		WithJSON(req).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SendVoiceMsg send voice message
func (c *Client) SendVoiceMsg(req *SendVoiceMsgReq) (*SendVoiceMsgResp, error) {
	resp := new(SendVoiceMsgResp)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cmsi/").
		SubPathf("send_voice_msg").
		WithHeaders(c.GetHeader()).
		WithJSON(req).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SendMail send mail
func (c *Client) SendMail(req *SendMailReq) (*SendMailResp, error) {
	resp := new(SendMailResp)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/api/c/compapi/v2/cmsi/").
		SubPathf("send_mail").
		WithHeaders(c.GetHeader()).
		WithJSON(req).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
