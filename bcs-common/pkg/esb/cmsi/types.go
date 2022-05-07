/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmsi

// BaseReq base request for esb
type BaseReq struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	AccessToken string `json:"access_token,omitempty"`
	BkTicket    string `json:"bk_ticket,omitempty"`
	BkUsername  string `json:"bk_username,omitempty"`
}

// BaseResp base resp from esb
type BaseResp struct {
	Code    int64  `json:"code"`
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// SendRtxReq request for sending rtx
type SendRtxReq struct {
	BaseReq          `json:",inline"`
	Title            string `json:"title"`
	ReceiverUsername string `json:"receiver__username"`
	Content          string `json:"content"`
	IsContentBase64  bool   `json:"is_content_base64,omitempty"`
}

// SendRtxResp response for sending rtx
type SendRtxResp struct {
	BaseResp `json:",inline"`
}

// UserInfomation user information in request for sending voice message
type UserInfomation struct {
	Username    string `json:"username"`
	MobilePhone string `json:"mobile_phone,omitempty"`
}

// SendVoiceMsgReq request for sending voice message
type SendVoiceMsgReq struct {
	BaseReq             `json:",inline"`
	AutoReadMessage     string           `json:"auto_read_message"`
	UserListInformation []UserInfomation `json:"user_list_information,omitempty"`
	ReceiverUsername    string           `json:"receiver__username,omitempty"`
}

// SendVoiceMsgRespData response data field for sending voice message
type SendVoiceMsgRespData struct {
	InstanceID string `json:"instance_id"`
}

// SendVoiceMsgResp response for sending voice message
type SendVoiceMsgResp struct {
	BaseResp `json:",inline"`
	Data     *SendVoiceMsgRespData `json:"data,omitempty"`
}

// SendWeixinReqData request data for sending weixin
type SendWeixinReqData struct {
	BaseReq         `json:",inline"`
	Heading         string `json:"heading"`
	Message         string `json:"message"`
	Date            string `json:"date,omitempty"`
	Remark          string `json:"remark,omitempty"`
	IsContentBase64 bool   `json:"is_content_base64,omitempty"`
}

// SendWeixinReq request for sending weixin
type SendWeixinReq struct {
	BaseReq          `json:",inline"`
	Receiver         string            `json:"receiver,omitempty"`
	ReceiverUserName string            `json:"receiver__username,omitempty"`
	Data             SendWeixinReqData `json:"data"`
	WxQyAgentID      string            `json:"wx_qy_agentid,omitempty"`
	WxQyCorpsecret   string            `json:"wx_qy_corpsecret,omitempty"`
}

// SendWeixinResp response for sending weixin
type SendWeixinResp struct {
	BaseResp `json:",inline"`
}

// MailAttachment attachment for mail
type MailAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Type        string `json:"type,omitempty"`
	Disposition string `json:"disposition,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
}

// SendMailReq request for sending mail
type SendMailReq struct {
	BaseReq          `json:",inline"`
	Receiver         string           `json:"receiver,omitempty"`
	ReceiverUsername string           `json:"receiver__username,omitempty"`
	Sender           string           `json:"sender,omitempty"`
	Title            string           `json:"title"`
	Content          string           `json:"content"`
	CC               string           `json:"cc,omitempty"`
	CCUsername       string           `json:"cc_username,omitempty"`
	IsContentBase64  bool             `json:"is_content_base64,omitempty"`
	Attachments      []MailAttachment `json:"attachments"`
}

// SendMailResp response for sending mail
type SendMailResp struct {
	BaseResp `json:",inline"`
}
