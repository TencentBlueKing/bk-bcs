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

package table

import (
	"errors"
	"fmt"
	"time"
)

// Client is a client
type Client struct {
	ID         uint32            `gorm:"column:id" json:"id"`
	Attachment *ClientAttachment `json:"attachment" gorm:"embedded"`
	Spec       *ClientSpec       `json:"spec" gorm:"embedded"`
}

// ClientSpec is a client spec
type ClientSpec struct {
	ClientVersion             string       `gorm:"column:client_version" json:"client_version"`
	ClientType                ClientType   `gorm:"column:client_type" json:"client_type"`
	Ip                        string       `gorm:"column:ip" json:"ip"`
	Labels                    string       `gorm:"column:labels" json:"labels"`
	Annotations               string       `gorm:"column:annotations" json:"annotations"`
	FirstConnectTime          time.Time    `gorm:"column:first_connect_time" json:"first_connect_time"`
	LastHeartbeatTime         time.Time    `gorm:"column:last_heartbeat_time" json:"last_heartbeat_time"`
	OnlineStatus              OnlineStatus `gorm:"column:online_status" json:"online_status"`
	Resource                  Resource     `json:"resource" gorm:"embedded"`
	CurrentReleaseID          uint32       `gorm:"column:current_release_id" json:"current_release_id"`
	TargetReleaseID           uint32       `gorm:"column:target_release_id" json:"target_release_id"`
	ReleaseChangeStatus       Status       `gorm:"column:release_change_status" json:"release_change_status"`
	ReleaseChangeFailedReason FailedReason `gorm:"column:release_change_failed_reason" json:"release_change_failed_reason"`
	FailedDetailReason        string       `gorm:"column:failed_detail_reason" json:"failed_detail_reason"`
}

// ClientAttachment is a client attachment
type ClientAttachment struct {
	UID   string `gorm:"column:uid" json:"uid"`
	BizID uint32 `db:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `db:"app_id" gorm:"column:app_id"`
}

// Resource resource information
type Resource struct {
	CpuUsage       float64 `gorm:"column:cpu_usage" json:"cpu_usage"`
	CpuMaxUsage    float64 `gorm:"column:cpu_max_usage" json:"cpu_max_usage"`
	MemoryUsage    uint64  `gorm:"column:memory_usage" json:"memory_usage"`
	MemoryMaxUsage uint64  `gorm:"column:memory_max_usage" json:"memory_max_usage"`
}

// TableName is the app's database table name.
func (c *Client) TableName() string {
	return "clients"
}

// AppID AuditRes interface
func (c *Client) AppID() uint32 {
	return c.ID
}

// ResID AuditRes interface
func (c *Client) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *Client) ResType() string {
	return "client"
}

// ValidateCreate validate app's info when created.
func (c *Client) ValidateCreate() error {
	if c.ID != 0 {
		return errors.New("id can not be set")
	}

	if c.Spec == nil {
		return errors.New("spec not set")
	}

	if err := c.Spec.ValidateCreate(); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := c.Attachment.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate client spec when it is created.
func (c *ClientSpec) ValidateCreate() error {
	if c.ClientVersion == "" {
		return errors.New("client version not set")
	}

	if c.Ip == "" {
		return errors.New("ip not set")
	}

	if c.LastHeartbeatTime.String() == "" {
		return errors.New("last heartbeat time not set")
	}

	if err := c.OnlineStatus.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate client attachment when it is created.
func (c *ClientAttachment) ValidateCreate() error {
	if c.BizID <= 0 {
		return errors.New("biz id not set")
	}

	if c.UID == "" {
		return errors.New("uid not set")
	}

	if c.AppID <= 0 {
		return errors.New("app id not set")
	}

	return nil
}

// Status define the status structure
type Status string

const (
	// Success xxx
	Success Status = "Success"
	// Failed xxx
	Failed Status = "Failed"
	// Processing xxx
	Processing Status = "Processing"
	// Skip xxx
	Skip Status = "Skip"
)

// Validate the version change status is valid or not.
func (rs Status) Validate() error {
	switch rs {
	case Success:
	case Failed:
	case Processing:
	case Skip:
	}

	return nil
}

// OnlineStatus define the online status structure
type OnlineStatus string

const (
	// Online xxx
	Online OnlineStatus = "online"
	// Offline xxx
	Offline OnlineStatus = "offline"
)

// Validate the online status is valid or not.
func (os OnlineStatus) Validate() error {
	switch os {
	case Online:
	case Offline:
	default:
		return fmt.Errorf("unknown %s sidecar online status", os)
	}

	return nil
}

// FailedReason define the failure cause structure
type FailedReason string

const (
	// PreHookFailed pre hook failed
	PreHookFailed FailedReason = "PreHookFailed"
	// PostHookFailed post hook failed
	PostHookFailed FailedReason = "PostHookFailed"
	// DownloadFailed download failed
	DownloadFailed FailedReason = "DownloadFailed"
	// SkipFailed Skip failed
	SkipFailed FailedReason = "SkipFailed"
)

// Validate the failed reason is valid or not.
func (fr FailedReason) Validate() error {
	switch fr {
	case PreHookFailed:
	case PostHookFailed:
	case DownloadFailed:
	case SkipFailed:
	}

	return nil
}

// ClientType client type (agent、sidecar、sdk、command).
type ClientType string

const (
	// SDK xxx
	SDK ClientType = "sdk"
	// Sidecar xxx
	Sidecar ClientType = "sidecar"
	// Agent xxx
	Agent ClientType = "agent"
	// Command xxx
	Command ClientType = "command"
	// Unknown xxx
	Unknown ClientType = "unknown"
)

// Validate the client type is valid or not.
func (ct ClientType) Validate() error {
	switch ct {
	case SDK:
	case Sidecar:
	case Agent:
	case Command:
	case Unknown:
	default:
		return fmt.Errorf("unknown %s client type", ct)
	}

	return nil
}
