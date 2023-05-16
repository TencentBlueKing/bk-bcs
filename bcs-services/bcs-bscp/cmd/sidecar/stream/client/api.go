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

package client

import (
	"fmt"

	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/kit"
	pbfs "bscp.io/pkg/protocol/feed-server"
	sfs "bscp.io/pkg/sf-share"
)

// Handshake to the upstream server
func (rc *rollingClient) Handshake(vas *kit.Vas, spec *types.SidecarSpec) (*pbfs.HandshakeResp, error) {
	if err := rc.wait.WaitWithContext(vas.Ctx); err != nil {
		return nil, err
	}

	metas := make([]*pbfs.SidecarAppMeta, 0)
	for _, one := range spec.Metas {
		metas = append(metas, &pbfs.SidecarAppMeta{
			AppId: one.AppID,
			Uid:   one.Uid,
		})
	}

	msg := &pbfs.HandshakeMessage{
		ApiVersion: sfs.CurrentAPIVersion,
		Spec: &pbfs.SidecarSpec{
			BizId:   spec.BizID,
			Version: rc.sidecarVer,
		},
	}

	resp, err := rc.upstream.Handshake(vas.Ctx, msg)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Watch release related messages from upstream feed server.
func (rc *rollingClient) Watch(vas *kit.Vas, payload []byte) (pbfs.Upstream_WatchClient, error) {
	if err := rc.wait.WaitWithContext(vas.Ctx); err != nil {
		return nil, err
	}

	meta := &pbfs.SideWatchMeta{
		ApiVersion: sfs.CurrentAPIVersion,
		Payload:    payload,
	}

	return rc.upstream.Watch(vas.Ctx, meta)
}

// Messaging is a message pipeline to send message to the upstream feed server.
func (rc *rollingClient) Messaging(vas *kit.Vas, typ sfs.MessagingType, payload []byte) (*pbfs.MessagingResp,
	error) {

	if err := rc.wait.WaitWithContext(vas.Ctx); err != nil {
		return nil, err
	}

	if err := typ.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message type, %v", err)
	}

	msg := &pbfs.MessagingMeta{
		ApiVersion: sfs.CurrentAPIVersion,
		Rid:        vas.Rid,
		Type:       uint32(typ),
		Payload:    payload,
	}

	return rc.upstream.Messaging(vas.Ctx, msg)
}
