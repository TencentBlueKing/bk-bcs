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
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// EmailConfig email config
type EmailConfig struct {
	// 发件人邮箱地址。不使用别名时请直接填写发件人邮箱地址，例如：noreply@mail.qcloud.com如需填写发件人别名时，
	// 请按照如下方式（注意别名与邮箱地址之间必须使用一个空格隔开）：别名+一个空格+<邮箱地址>，别名中不能带有冒号(:)。
	FromEmailAddress *string `json:"FromEmailAddress,omitempty" name:"FromEmailAddress"`

	// 收信人邮箱地址，最多支持群发50人。注意：邮件内容会显示所有收件人地址，非群发邮件请多次调用API发送。
	Destination []*string `json:"Destination,omitempty" name:"Destination"`

	// 邮件主题
	Subject *string `json:"Subject,omitempty" name:"Subject"`

	// 邮件的“回复”电子邮件地址。可以填写您能收到邮件的邮箱地址，可以是个人邮箱。如果不填，收件人的回复邮件将会发送失败。
	ReplyToAddresses *string `json:"ReplyToAddresses,omitempty" name:"ReplyToAddresses"`

	// 抄送人邮箱地址，最多支持抄送20人。
	Cc []*string `json:"Cc,omitempty" name:"Cc"`

	// 密送人邮箱地址，最多支持抄送20人,Bcc和Destination不能重复。
	Bcc []*string `json:"Bcc,omitempty" name:"Bcc"`

	// 使用模板发送时，填写模板相关参数。
	// <dx-alert infotype="notice" title="注意"> 如您未申请过特殊配置，则该字段为必填 </dx-alert>
	Template *ses.Template `json:"Template,omitempty" name:"Template"`

	// 需要发送附件时，填写附件相关参数。腾讯云接口请求最大支持 8M 的请求包，附件内容经过 Base64 预期扩大1.5倍，
	// 应该控制所有附件的总大小最大在 4M 以内，整体请求超出 8M 时接口会返回错误
	Attachments []*ses.Attachment `json:"Attachments,omitempty" name:"Attachments"`
}

// NewEmailClient new email client
func NewEmailClient() (*ses.Client, error) {
	credential := common.NewCredential(
		config.G.Sign.SecretId,
		config.G.Sign.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.G.Sign.EmailEndpoint
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, err := ses.NewClient(credential, config.G.Sign.Region, cpf)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// SendEmail send email, 参考地址： https://cloud.tencent.com/document/product/1288/51034
func SendEmail(emailConfig *EmailConfig) error {
	client, err := NewEmailClient()
	if err != nil {
		return err
	}
	request := ses.NewSendEmailRequest()
	request.FromEmailAddress = emailConfig.FromEmailAddress
	request.Destination = emailConfig.Destination
	request.Subject = emailConfig.Subject
	request.ReplyToAddresses = emailConfig.ReplyToAddresses
	request.Cc = emailConfig.Cc
	request.Bcc = emailConfig.Bcc
	request.Template = emailConfig.Template
	request.Attachments = emailConfig.Attachments
	emailResp, err := client.SendEmail(request)
	if err != nil {
		return err
	}
	if emailResp == nil || emailResp.Response == nil {
		return fmt.Errorf("SendEmail resp is nil")
	}
	blog.Infof("SendEmail resp request id %s", emailResp.Response.RequestId)

	return nil
}
