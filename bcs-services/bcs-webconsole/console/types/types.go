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

// Package types xxx
package types

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
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

// ContextValueKey is the key for context value
type ContextValueKey string

// LaneKey is the key for lane
const (
	// LaneKey is the key for lane
	LaneKey ContextValueKey = "X-Lane"
	// LaneIDPrefix 染色的header前缀
	LaneIDPrefix = "X-Lane-"
)

const (
	defaultSessionTimeout  = time.Minute * 30 // session 过期时间
	defaultConnIdleTimeout = time.Minute * 30 // 链接自动断开时间, 30分钟
	// MaxSessionTimeout session最多等待时间
	MaxSessionTimeout = 24 * 60
	// MaxConnIdleTimeout xx
	MaxConnIdleTimeout = 24 * 60
)

const (
	defaultTerminalCols = 211 // defaultTerminalCols DefaultRows 1080p页面测试得来
	defaultTerminalRows = 25  // defaultTerminalRows xxx

	// perfMeterKey 存Redis的延时命令Hash名称
	perfMeterKey = "bcs::webconsole::meter_key"
	// perfMeterData 用户延时统计数据列表的Redis key前缀
	perfMeterData = "bcs::webconsole::meter_data"
)

// APIResponse xxx
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// CheckPassed 检测是否OK，如下载文件大小等
type CheckPassed struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
	Detail string `json:"detail"`
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
	ProjectCode     string         `json:"project_code"`
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

// CommandDelay 用户延时命令设置
type CommandDelay struct {
	Enabled    bool   `json:"enabled"`
	ConsoleKey string `json:"console_key"`
}

// HashValue 使用 {enabled}:{key} 高效检索
func (c *CommandDelay) HashValue() string {
	e := "0"
	if c.Enabled {
		e = "1"
	}

	return fmt.Sprintf("%s:%s", e, c.ConsoleKey)
}

// CommandDelayMatch 是否开启匹配
func CommandDelayMatch(key string, msg byte) bool {
	return key == "1:"+string(msg)
}

// GetMeterKey meter 数据 key
func GetMeterKey() string {
	return fmt.Sprintf("%s::%s", perfMeterKey, config.G.Base.RunEnv)
}

// GetMeterDataKey meter 数据 key
func GetMeterDataKey(username string) string {
	return fmt.Sprintf("%s::%s::%s", perfMeterData, username, config.G.Base.RunEnv)
}

// MakeCommandDelay HashValue 转结构体
func MakeCommandDelay(v string) (*CommandDelay, error) {
	parts := strings.Split(v, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("not valid value: %s", v)
	}
	c := &CommandDelay{
		ConsoleKey: parts[1],
		Enabled:    false,
	}

	if parts[0] == "1" {
		c.Enabled = true
	}

	return c, nil
}

// DelayData 用户的延迟数据
type DelayData struct {
	ClusterId   string `json:"cluster_id"`
	TimeConsume string `json:"time_consume"`
	CreateTime  string `json:"create_time"`
	SessionId   string `json:"session_id"`
	PodName     string `json:"pod_name"`
	CommandKey  string `json:"command_key"`
	Username    string `json:"-"`
}

// UserMeterRsp 用户统计列表返回
type UserMeterRsp struct {
	ClusterId          string        `json:"cluster_id"`
	AverageTimeConsume string        `json:"average_time_consume"`
	MaxTimeConsume     string        `json:"max_time_consume"`
	MinTimeConsume     string        `json:"min_time_consume"`
	UserConsumes       []UserConsume `json:"user_consumes"`
}

// UserMeters 用户统计列表
type UserMeters struct {
	ClusterId          string        `json:"cluster_id"`
	AverageTimeConsume time.Duration `json:"average_time_consume"`
	MaxTimeConsume     time.Duration `json:"max_time_consume"`
	MinTimeConsume     time.Duration `json:"min_time_consume"`
	UserConsumes       []UserConsume `json:"user_consumes"`
}

// UserConsume 用户统计耗时单条数据
type UserConsume struct {
	TimeConsume string `json:"time_consume"`
	CreateTime  string `json:"create_time"`
	SessionId   string `json:"session_id"`
	PodName     string `json:"pod_name"`
	CommandKey  string `json:"command_key"`
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

// TerminalSize web终端发来的 resize 包
type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// DefaultTerminalSize Terminal 终端默认大小配置
func DefaultTerminalSize() *TerminalSize {
	return &TerminalSize{
		Rows: defaultTerminalRows,
		Cols: defaultTerminalCols,
	}
}

// GetLaneIDByCtx get lane id by ctx
func GetLaneIDByCtx(ctx context.Context) map[string]string {
	// http 格式的以key value方式存放，eg: key: X-Lane value: X-Lane-xxx:xxx
	v, ok := ctx.Value(LaneKey).(string)
	if ok || v != "" {
		result := strings.Split(v, ":")
		if len(result) != 2 {
			return nil
		}
		return map[string]string{result[0]: result[1]}
	}
	if !ok || v == "" {
		return grpcLaneIDValue(ctx)
	}
	return nil
}

// grpcLaneIDValue grpc lane id 处理
func grpcLaneIDValue(ctx context.Context) map[string]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			tmpKey := textproto.CanonicalMIMEHeaderKey(k)
			if strings.HasPrefix(tmpKey, LaneIDPrefix) && len(v) > 0 {
				return map[string]string{tmpKey: md.Get(k)[0]}
			}
		}
	}
	return nil
}

// WithLaneIdCtx ctx lane id
func WithLaneIdCtx(ctx context.Context, h http.Header) context.Context {
	for k, v := range h {
		if strings.HasPrefix(k, LaneIDPrefix) && len(v) > 0 {
			ctx = context.WithValue(ctx, LaneKey, k+":"+v[0])
		}
	}
	return ctx
}
