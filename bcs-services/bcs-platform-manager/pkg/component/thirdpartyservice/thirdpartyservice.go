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

// Package thirdpartyservice xxx
package thirdpartyservice

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsthirdpartyservice "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/thirdpartyservice"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	microRgt "go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

const (
	// ModuleThirdpartyServiceManager thirdparty service
	ModuleThirdpartyServiceManager = "bcsthirdpartyservice.bkbcs.tencent.com"
)

// NewClient create push manager service client
func NewClient(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ModuleThirdpartyServiceManager, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		bcsthirdpartyservice.SetClientConfig(tlsConfig, dis)
	} else {
		bcsthirdpartyservice.SetClientConfig(tlsConfig, nil)
	}
	return nil
}

// SendMail send mail to bcs thirdparty service
func SendMail(ctx context.Context, email *bcsthirdpartyservice.SendMailRequest) error {
	cli, close, err := bcsthirdpartyservice.GetClient(utils.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return err
	}
	resp, err := cli.SendMail(ctx, email)
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
