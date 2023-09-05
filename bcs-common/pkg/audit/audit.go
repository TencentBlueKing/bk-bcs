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

package audit

import (
	"encoding/json"
	"time"

	"github.com/TencentBlueKing/bk-audit-go-sdk/bkaudit"
	"github.com/go-playground/validator/v10"
	"k8s.io/klog"
)

const (
	// ForwardedForHeader is the header name of X-Forwarded-For.
	ForwardedForHeader = "X-Forwarded-For"
	// UserAgentHeader is the header name of User-Agent.
	UserAgentHeader = "Grpcgateway-User-Agent"
)

// Client is a client for audit and activity.
type Client struct {
	bcsHost  string
	bcsToken string
	logger   Logger
}

// NewClient returns a new audit client.
func NewClient(host, token string, logger Logger) *Client {
	if logger == nil {
		logger = klog.V(0)
	}
	// init formatter
	var formatter = &bkaudit.EventFormatter{}
	var exporters = []bkaudit.Exporter{&bkaudit.LoggerExporter{Logger: logger}}
	// init client
	var err error
	auditClient, err = bkaudit.InitEventClient(bkAppCode, "", formatter, exporters, 0, nil)
	if err != nil {
		logger.Info("init auditClient client failed, %s", err.Error())
		return nil
	}
	return &Client{
		logger:   logger,
		bcsHost:  host,
		bcsToken: token,
	}
}

// R returns a new recorder.
func (c *Client) R() *Recorder {
	bcsHost = c.bcsHost
	token = c.bcsToken
	return &Recorder{
		logger:         c.logger,
		enableAudit:    true,
		enableActivity: true,
	}
}

// Recorder is a recorder for audit and activity.
type Recorder struct {
	enableAudit    bool
	enableActivity bool
	logger         Logger
	ctx            RecorderContext
	resource       Resource
	action         Action
	result         ActionResult
}

// RecorderContext is the context of recorder.
type RecorderContext struct {
	Username  string `validate:"required"`
	SourceIP  string
	UserAgent string
	RequestID string
	StartTime time.Time
	EndTime   time.Time
}

// DisableAudit disables audit.
func (r *Recorder) DisableAudit() *Recorder {
	r.enableAudit = false
	return r
}

// DisableActivity disables activity.
func (r *Recorder) DisableActivity() *Recorder {
	r.enableActivity = false
	return r
}

// SetContext sets the context of recorder.
func (r *Recorder) SetContext(ctx RecorderContext) *Recorder {
	r.ctx = ctx
	return r
}

// SetResource sets the resource to be recorded.
func (r *Recorder) SetResource(resource Resource) *Recorder {
	r.resource = resource
	return r
}

// SetAction sets the action to be recorded.
func (r *Recorder) SetAction(action Action) *Recorder {
	r.action = action
	return r
}

// SetResult sets the result of action.
func (r *Recorder) SetResult(result ActionResult) *Recorder {
	r.result = result
	return r
}

func (r *Recorder) validate() error {
	validate := validator.New()
	var err error
	if err = validate.Struct(r.ctx); err != nil {
		return err
	}
	if err = validate.Struct(r.resource); err != nil {
		return err
	}
	if err = validate.Struct(r.action); err != nil {
		return err
	}
	if err = validate.Struct(r.result); err != nil {
		return err
	}
	return nil
}

// Do records audit and activity.
func (r *Recorder) Do() error {
	if err := r.validate(); err != nil {
		r.logger.Info("audit validate failed, %s", err.Error())
		return err
	}
	if r.enableAudit {
		auditData := AuditData{
			EventContent:  r.result.ResultContent,
			ActionID:      r.action.ActionID,
			RequestID:     r.ctx.RequestID,
			Username:      r.ctx.Username,
			StartTime:     r.ctx.StartTime,
			EndTime:       r.ctx.EndTime,
			ResourceType:  r.resource.ResourceType,
			InstanceID:    r.resource.ResourceID,
			InstanceName:  r.resource.ResourceName,
			InstanceData:  r.resource.ResourceData,
			ResultCode:    r.result.ResultCode,
			ResultContent: r.result.ResultContent,
			SourceIP:      r.ctx.SourceIP,
			UserAgent:     r.ctx.UserAgent,
		}
		if auditData.StartTime.Unix() == 0 {
			auditData.StartTime = time.Now()
		}
		if auditData.EndTime.Unix() == 0 || auditData.EndTime.Before(auditData.StartTime) {
			auditData.EndTime = auditData.StartTime
		}
		AddEvent(auditData)
	}

	if r.enableActivity {
		var extra string
		if r.result.ExtraData != nil {
			extraData, _ := json.Marshal(r.result.ExtraData)
			extra = string(extraData)
		}
		PushActivity(Activity{
			ProjectCode:  r.resource.ProjectCode,
			ResourceType: r.resource.ResourceType,
			ResourceName: r.resource.ResourceName,
			ResourceID:   r.resource.ResourceID,
			ActivityType: r.action.ActivityType,
			Status:       r.result.Status,
			Username:     r.ctx.Username,
			Description:  r.result.ResultContent,
			Extra:        extra,
		})
	}
	return nil
}

// Resource is the resource to be recorded.
type Resource struct {
	// 项目 Code，如果获取不到，可以填写 projectID
	ProjectCode string `validate:"required"`
	// 资源类型，如: project, cluster
	ResourceType ResourceType `validate:"required"`
	// 资源 ID，如: project id, cluster id
	ResourceID string `validate:"required"`
	// 资源名称
	ResourceName string `validate:"required"`
	// 资源相关数据
	ResourceData map[string]any
}

// Action is the action to be recorded.
type Action struct {
	// 操作 ID，可以是权限中心 action，也可以是业务自定义 action，如: cluster_create
	ActionID string `validate:"required"`
	// 操作类型，如: create, update, delete
	ActivityType ActivityType `validate:"required"`
}

// ActionResult is the result of action.
type ActionResult struct {
	// 操作结果状态，如: success, failed
	Status ActivityStatus `validate:"required"`
	// 操作结果代码，为 http 返回码，如: 200, 400, 500
	ResultCode int `validate:"required"`
	// 操作结果描述，如: 创建了 xxx 应用
	ResultContent string
	// 扩展字段，可以自定义一些信息，供操作记录和其他功能联动，如 cluster-manager 任务
	ExtraData map[string]any
}

// AuditData is the data to be recorded.
type AuditData struct {
	EventContent  string
	ActionID      string
	RequestID     string
	Username      string
	StartTime     time.Time
	EndTime       time.Time
	ResourceType  ResourceType
	InstanceID    string
	InstanceName  string
	InstanceData  map[string]any
	ResultCode    int
	ResultContent string
	ExtendData    map[string]any

	// Extend
	SourceIP  string
	UserAgent string
}
