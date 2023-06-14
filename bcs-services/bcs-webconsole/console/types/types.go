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

// Package types xxx
package types

import (
	"time"
)

// WebConsoleMode webconsole 类型
type WebConsoleMode string

const (
	// ClusterInternalMode xxx
	ClusterInternalMode WebConsoleMode = "cluster_internal" // 用户自己集群 inCluster 模式
	// ClusterExternalMode xxx
	ClusterExternalMode WebConsoleMode = "cluster_external" // 平台集群, 外部模式, 需要设置 AdminClusterId
	// ContainerDirectMode xxx
	ContainerDirectMode WebConsoleMode = "container_direct" // 直连容器
)

const (
	defaultSessionTimeout  = time.Minute * 30 // session 过期时间
	defaultConnIdleTimeout = time.Minute * 30 // 链接自动断开时间, 30分钟
	MaxSessionTimeout      = 24 * 60
	MaxConnIdleTimeout     = 24 * 60
)

// APIResponse xxx
type APIResponse struct {
	Data      interface{} `json:"data,omitempty"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
}

// AuditRecord 审计记录
type AuditRecord struct {
	InputRecord  string      `json:"input_record"`
	OutputRecord string      `json:"output_record"`
	SessionID    string      `json:"session_id"`
	Context      interface{} `json:"context"` // 这里使用户信息
	ProjectID    string      `json:"project_id"`
	ClusterID    string      `json:"cluster_id"`
	UserPodName  string      `json:"user_pod_name"`
	Username     string      `json:"username"`
}

// Container webconsole 连接三要素
type Container struct {
	Namespace     string
	PodName       string
	ContainerName string
}

// PodContext xxx
type PodContext struct {
	ProjectId       string         `json:"project_id"`
	Username        string         `json:"username"`
	Viewers         []string       `json:"viewers"`
	AdminClusterId  string         `json:"admin_cluster_id"` // kubectld pod 所在集群Id, kubectl api 连接的集群
	Namespace       string         `json:"namespace"`
	PodName         string         `json:"pod_name"`
	ClusterId       string         `json:"cluster_id"` // 目标集群Id
	ContainerName   string         `json:"container_name"`
	Commands        []string       `json:"commands"`
	Mode            WebConsoleMode `json:"mode"`
	Source          string         `json:"source"`
	SessionTimeout  int64          `json:"session_timeout"`   // session 过期时间, 单位分钟
	ConnIdleTimeout int64          `json:"conn_idle_timeout"` // 空闲时间, 单位分钟
	SessionId       string         `json:"session_id"`        // session id
}

// GetConnIdleTimeout 获取空闲过期时间
func (c *PodContext) GetConnIdleTimeout() time.Duration {
	if c.ConnIdleTimeout == 0 {
		return defaultConnIdleTimeout
	}
	return time.Minute * time.Duration(c.ConnIdleTimeout)
}

// GetSessionTimeout 获取 session 过期时间
func (c *PodContext) GetSessionTimeout() time.Duration {
	if c.SessionTimeout == 0 {
		return defaultSessionTimeout
	}
	return time.Minute * time.Duration(c.SessionTimeout)
}

// HasPerm 是否有权限
func (c *PodContext) HasPerm(username string) bool {
	if c.Username == username {
		return true
	}

	for _, viewer := range c.Viewers {
		if viewer == username {
			return true
		}
	}

	return false
}

// TimestampPodContext 带时间戳的 PodContext
type TimestampPodContext struct {
	PodContext
	Timestamp int64 `json:"timestamp"`
}

// IsExpired 是否过期
func (c *TimestampPodContext) IsExpired() bool {
	expireTimestamp := time.Now().Add(-c.GetSessionTimeout()).Unix()
	return c.Timestamp <= expireTimestamp
}
