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

// Package stream NOTES
package stream

import (
	"context"
	"errors"
	"fmt"

	"bscp.io/cmd/sidecar/stream/client"
	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"

	"go.uber.org/atomic"
)

// Interface defines the supported operations for stream.
type Interface interface {
	// Initialize the stream.
	Initialize() (*sfs.SidecarRuntimeOption, error)
	// StartWatch start watch events from upstream server.
	StartWatch(onChange *types.OnChange) error
	// FireEvent send the sidecar's event to the upstream server.
	FireEvent(payload sfs.MessagingPayloadBuilder) error
	// NotifyReconnect notify stream to reconnect upstream server.
	NotifyReconnect(signal types.ReconnectSignal)
}

// New create a stream instance.
func New(settings cc.SidecarSetting, fingerPrint sfs.FingerPrint) (Interface, error) {

	c, err := client.New(settings.Upstream)
	if err != nil {
		return nil, fmt.Errorf("new stream client failed, err: %v", err)
	}

	mh := sfs.SidecarMetaHeader{
		BizID:       settings.AppSpec.BizID,
		Fingerprint: fingerPrint.Encode(),
	}

	mhBytes, err := jsoni.Marshal(mh)
	if err != nil {
		return nil, fmt.Errorf("encode sidecar meta header failed, err: %v", err)
	}

	s := &stream{
		settings:        settings,
		client:          c,
		metaHeaderValue: string(mhBytes),
		reconnectChan:   make(chan types.ReconnectSignal, 5),
		reconnecting:    atomic.NewBool(false),
	}

	go s.waitForReconnectSignal()

	if err = s.loopHeartbeat(); err != nil {
		return nil, fmt.Errorf("start loop hearbeat failed, err: %v", err)
	}

	return s, nil
}

type stream struct {
	settings        cc.SidecarSetting
	runtimeOpt      *sfs.SidecarRuntimeOption
	metaHeaderValue string
	client          client.Interface

	reconnectChan chan types.ReconnectSignal
	reconnecting  *atomic.Bool

	watch *watch
}

// Initialize the stream including:
// 1. handshake with the upstream server.
// 2. start watch the release event from the upstream feed server.
func (s *stream) Initialize() (*sfs.SidecarRuntimeOption, error) {
	metas := make([]*types.SidecarMeta, 0)
	for _, one := range s.settings.AppSpec.Applications {
		metas = append(metas, &types.SidecarMeta{
			AppID: one.AppID,
			Uid:   one.Uid,
		})
	}

	spec := &types.SidecarSpec{
		BizID: s.settings.AppSpec.BizID,
		Metas: metas,
	}

	vas, cancel := s.vasBuilder()
	defer cancel()

	resp, err := s.client.Handshake(vas, spec)
	if err != nil {
		return nil, fmt.Errorf("handshake with upstream server failed, err: %v, rid: %s", err, vas.Rid)
	}

	if !sfs.IsAPIVersionMatch(resp.ApiVersion) {
		return nil, fmt.Errorf("sidecar's current api version[%s] is not match the upstream server's api version, "+
			"rid: %s", resp.ApiVersion.Format(), vas.Rid)
	}

	payload := new(sfs.SidecarHandshakePayload)
	if err := jsoni.Unmarshal(resp.Payload, payload); err != nil {
		return nil, fmt.Errorf("decode the handshake response payload failed, err: %v, rid: %s", err, vas.Rid)
	}

	if payload.RuntimeOption == nil {
		return nil, errors.New("runtime option is nil")
	}

	for _, meta := range spec.Metas {
		_, exist := payload.RuntimeOption.AppReloads[meta.AppID]
		if !exist {
			return nil, fmt.Errorf("app: %d reload not exist", meta.AppID)
		}
	}

	logs.Infof("sidecar handshake upstream server name: %s, rid: %s", payload.ServiceInfo.Name, vas.Rid)

	s.runtimeOpt = payload.RuntimeOption

	s.client.EnableBounce(s.runtimeOpt.BounceIntervalHour)

	return payload.RuntimeOption, nil
}

// FireEvent send the sidecar's event to the upstream server.
func (s *stream) FireEvent(payload sfs.MessagingPayloadBuilder) error {
	vas, cancel := s.vasBuilder()
	defer cancel()

	bytes, err := payload.Encode()
	if err != nil {
		return fmt.Errorf("payload encode failed, err: %v", err)
	}

	_, err = s.client.Messaging(vas, payload.MessagingType(), bytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *stream) vasBuilder() (*kit.Vas, context.CancelFunc) {
	pairs := make(map[string]string)
	// add user information
	pairs[constant.SideUserKey] = s.settings.Upstream.Authentication.User
	// add meta header
	pairs[constant.SidecarMetaKey] = s.metaHeaderValue

	vas := kit.OutgoingVas(pairs)
	ctx, cancel := context.WithCancel(vas.Ctx)
	vas.Ctx = ctx
	return vas, cancel
}
