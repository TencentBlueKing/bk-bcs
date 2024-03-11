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

// ClientEvent is a client event
type ClientEvent struct {
	ID         uint32                 `gorm:"column:id" json:"id"`
	Attachment *ClientEventAttachment `json:"attachment" gorm:"embedded"`
	Spec       *ClientEventSpec       `json:"spec" gorm:"embedded"`
}

// ClientEventSpec is a client event spec
type ClientEventSpec struct {
	OriginalReleaseID         uint32       `gorm:"column:original_release_id" json:"original_release_id"`
	TargetReleaseID           uint32       `gorm:"column:target_release_id" json:"target_release_id"`
	StartTime                 time.Time    `gorm:"column:start_time" json:"start_time"`
	EndTime                   time.Time    `gorm:"column:end_time" json:"end_time"`
	TotalSeconds              float64      `gorm:"column:total_seconds" json:"total_seconds"`
	TotalFileSize             float64      `gorm:"column:total_file_size" json:"total_file_size"`
	DownloadFileSize          float64      `gorm:"column:download_file_size" json:"download_file_size"`
	TotalFileNum              uint32       `gorm:"column:total_file_num" json:"total_file_num"`
	DownloadFileNum           uint32       `gorm:"column:download_file_num" json:"download_file_num"`
	ReleaseChangeStatus       Status       `gorm:"column:release_change_status" json:"release_change_status"`
	ReleaseChangeFailedReason FailedReason `gorm:"column:release_change_failed_reason" json:"release_change_failed_reason"`
	FailedDetailReason        string       `gorm:"column:failed_detail_reason" json:"failed_detail_reason"`
}

// ClientEventAttachment is a client event attachment
type ClientEventAttachment struct {
	ClientID   uint32     `gorm:"column:client_id" json:"client_id"`
	CursorID   string     `gorm:"column:cursor_id" json:"cursor_id"`
	UID        string     `gorm:"column:uid" json:"uid"`
	BizID      uint32     `db:"biz_id" gorm:"column:biz_id"`
	AppID      uint32     `db:"app_id" gorm:"column:app_id"`
	ClientMode ClientMode `gorm:"column:client_mode" json:"client_mode"`
}

// ClientMode define the client mode structure
type ClientMode string

const (
	// Pull xxx
	Pull ClientMode = "Pull"
	// Watch xxx
	Watch ClientMode = "Watch"
)

// Validate the client mode is valid or not.
func (cm ClientMode) Validate() error {
	switch cm {
	case Pull:
	case Watch:
	default:
		return fmt.Errorf("unknown %s sidecar client mode", cm)
	}

	return nil
}

// TableName is the app's database table name.
func (c *ClientEvent) TableName() string {
	return "client_events"
}

// AppID AuditRes interface
func (c *ClientEvent) AppID() uint32 {
	return c.ID
}

// ResID AuditRes interface
func (c *ClientEvent) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *ClientEvent) ResType() string {
	return "client_event"
}

// ValidateCreate validate group is valid or not when create it.
func (c *ClientEvent) ValidateCreate() error {

	if c.ID > 0 {
		return errors.New("id should not be set")
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

// ValidateCreate validate client event spec when it is created.
func (c *ClientEventSpec) ValidateCreate() error {
	// if c.TargetReleaseID <= 0 {
	// 	return errors.New("target release id not set")
	// }

	if c.StartTime.String() == "" {
		return errors.New("start time not set")
	}

	if c.EndTime.String() == "" {
		return errors.New("end time not set")
	}

	return nil
}

// ValidateCreate validate client event attachment when it is created.
func (c *ClientEventAttachment) ValidateCreate() error {
	if c.BizID <= 0 {
		return errors.New("biz id not set")
	}

	if c.UID == "" {
		return errors.New("uid not set")
	}

	if c.AppID <= 0 {
		return errors.New("app id not set")
	}

	if c.ClientID <= 0 {
		return errors.New("client id not set")
	}

	if err := c.ClientMode.Validate(); err != nil {
		return err
	}

	return nil
}
