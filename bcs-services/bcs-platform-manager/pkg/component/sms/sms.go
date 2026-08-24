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

// Package sms client
package sms

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// NewSmsClient new sms client
func NewSmsClient() (*sms.Client, error) {
	credential := common.NewCredential(
		config.G.Sign.SecretId,
		config.G.Sign.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.G.Sign.SmsEndpoint
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, err := sms.NewClient(credential, config.G.Sign.Region, cpf)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SendSms send sms, 参考地址： https://cloud.tencent.com/document/product/382/55981
func SendSms(
	phoneNumbers []*string, templateID *string, templateParams []*string) ([]*sms.SendStatus, error) {
	client, err := NewSmsClient()
	if err != nil {
		return nil, err
	}
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = &config.G.Sign.SmsSdkAppId
	request.SignName = &config.G.Sign.SmsSignName
	request.PhoneNumberSet = phoneNumbers
	request.TemplateId = templateID
	request.TemplateParamSet = templateParams
	smsResp, err := client.SendSms(request)
	if err != nil {
		return nil, err
	}
	if smsResp == nil || smsResp.Response == nil {
		return nil, fmt.Errorf("SendSms resp is nil")
	}
	blog.Infof("SendSms resp request id %s", smsResp.Response.RequestId)

	return smsResp.Response.SendStatusSet, nil
}
