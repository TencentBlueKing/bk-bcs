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

// Package thirdparty provides a client for interacting with bcs-thirdparty-service.
package thirdparty

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	third "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/pkg/bcsapi/thirdparty-service"
)

// SendRtx sends a message via WeCom using the bcs-thirdparty-service.
func (t *thirdpartyClient) SendRtx(req *third.SendRtxRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := t.thirdpartySvc.SendRtx(ctx, req)
	if err != nil {
		blog.Errorf("sendRtx failed err: %s", err.Error())
		return fmt.Errorf("sendRtx failed: %v", err)
	}
	blog.Infof("sendRtx resp: %+v", resp)
	if !resp.Result {
		return fmt.Errorf("sendRtx failed, code: %s, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}

// SendMail sends an email using the bcs-thirdparty-service.
func (t *thirdpartyClient) SendMail(req *third.SendMailRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.SendMail(ctx, req)
	if err != nil {
		blog.Errorf("sendMail failed err: %s", err.Error())
		return fmt.Errorf("sendMail failed: %v", err)
	}

	blog.Infof("sendMail resp: %+v", resp)
	if !resp.Result {
		return fmt.Errorf("sendMail failed, code: %s, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}

// SendMsg sends an bkchat-msg using the bcs-thirdparty-service.
func (t *thirdpartyClient) SendMsg(req *third.SendMsgRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.SendMsg(ctx, req)
	if err != nil {
		blog.Errorf("sendMsg failed err: %s", err.Error())
		return fmt.Errorf("sendMsg failed: %v", err)
	}

	blog.Infof("sendMsg resp: %+v", resp)
	if !resp.Result {
		return fmt.Errorf("sendMsg failed, code: %s, message: %s",
			resp.Code, resp.Message)
	}
	return nil
}
