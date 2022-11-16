/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package types NOTES
package types

import (
	pbbase "bscp.io/pkg/protocol/core/base"
)

// SidecarSpec defines sidecar's specifics
type SidecarSpec struct {
	BizID uint32
	Metas []*SidecarMeta
}

// SidecarMeta defines sidecar's metadata.
type SidecarMeta struct {
	AppID uint32
	Uid   string
}

// ReconnectSignal defines the signal information to tell the
// stream to reconnect the remote upstream server.
type ReconnectSignal struct {
	Reason string
}

// String format the reconnect signal to a string.
func (rs ReconnectSignal) String() string {
	return rs.Reason
}

// ReleaseChangeEvent defines the release change event's
// detail information.
type ReleaseChangeEvent struct {
	Rid        string
	APIVersion *pbbase.Versioning
	Payload    []byte
}

// OnChange defines the callback handlers for stream to notify the
// related events.
type OnChange struct {
	// OnAppReleaseChange is used to receive app release change event from the upstream.
	OnReleaseChange func(event *ReleaseChangeEvent)
	// CurrentRelease get the current release metadata if it exists for an app.
	CurrentRelease func(appID uint32) (releaseID uint32, cursorID uint32, exist bool)
}

// CurrentReleaseMeta defines the app current release metadata.
type CurrentReleaseMeta struct {
	ReleaseID uint32
	CursorID  uint32
}
