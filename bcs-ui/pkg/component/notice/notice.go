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

// Package notice xxx
package notice

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

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

// RegisterSystem 注册系统
func RegisterSystem(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/register/", config.G.BKNotice.Host)

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		config.G.Base.AppCode, config.G.Base.AppSecret)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		Post(url)

	if err != nil {
		return err
	}

	if err := component.UnmarshalBKResult(resp, nil); err != nil {
		return err
	}
	return nil
}

// GetCurrentAnnouncements 获取系统当前通知
// 请求通知中心失败时，返回空数组
func GetCurrentAnnouncements(ctx context.Context, lang string) ([]Announcement, error) {

	announcements := []Announcement{}

	url := fmt.Sprintf("%s/v1/announcement/get_current_announcements/?platform=%s&language=%s",
		config.G.BKNotice.Host, config.G.Base.AppCode, lang)

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		config.G.Base.AppCode, config.G.Base.AppSecret)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		Get(url)

	if err != nil {
		return announcements, err
	}

	if err := component.UnmarshalBKResult(resp, &announcements); err != nil {
		return announcements, err
	}
	return announcements, nil
}
