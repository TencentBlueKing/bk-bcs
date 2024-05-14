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

// Package bknotice provides bknotice client.
package bknotice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
)

type registerSystemResp struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type getAnnouncementResp struct {
	Result  bool           `json:"result"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []Announcement `json:"data"`
}

// Announcement 通知中心公告
type Announcement struct {
	ID           int                   `json:"id"`
	Title        string                `json:"title"`
	Content      string                `json:"content"`
	ContentList  []AnnouncementContent `json:"content_list"`
	AnnounceType string                `json:"announce_type"`
	StartTime    string                `json:"start_time"`
	EndTime      string                `json:"end_time"`
}

// AnnouncementContent 通知中心公告内容，包括语言和内容
type AnnouncementContent struct {
	Content  string `json:"content"`
	Language string `json:"language"`
}

// RegisterSystem 注册系统到通知中心
func RegisterSystem(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/register/", cc.ApiServer().BKNotice.Host)

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.ApiServer().Esb.AppCode, cc.ApiServer().Esb.AppSecret)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		Post(url)

	if err != nil {
		return err
	}

	resigerResp := &registerSystemResp{}
	if err := json.Unmarshal(resp.Body(), resigerResp); err != nil {
		return err
	}

	if resigerResp.Code != 0 {
		return fmt.Errorf("register system to bknotice failed, code: %d, message: %s",
			resigerResp.Code, resigerResp.Message)
	}
	return nil
}

// GetCurrentAnnouncements 获取系统当前通知
func GetCurrentAnnouncements(ctx context.Context, lang string) ([]Announcement, error) {

	url := fmt.Sprintf("%s/v1/announcement/get_current_announcements/?platform=%s&language=%s",
		cc.ApiServer().BKNotice.Host, cc.ApiServer().Esb.AppCode, lang)

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.ApiServer().Esb.AppCode, cc.ApiServer().Esb.AppSecret)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		Get(url)

	if err != nil {
		return nil, err
	}

	getAnnouncementResp := &getAnnouncementResp{}

	if err := json.Unmarshal(resp.Body(), getAnnouncementResp); err != nil {
		return nil, err
	}
	return getAnnouncementResp.Data, nil
}
