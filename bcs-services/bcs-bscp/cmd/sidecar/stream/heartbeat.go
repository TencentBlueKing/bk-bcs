/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package stream

import (
	"fmt"
	"time"

	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"
)

const (
	// TODO: these config can set by config file.
	// defaultHeartbeatIntervalSec defines heartbeat default interval.
	defaultHeartbeatInterval = 15 * time.Second
	// defaultHeartbeatTimeout defines default heartbeat request timeout.
	defaultHeartbeatTimeout = 5 * time.Second
	// maxHeartbeatRetryCount defines heartbeat max retry count.
	maxHeartbeatRetryCount = 3
)

func (s *stream) loopHeartbeat() error {

	heartbeatPayload := sfs.HeartbeatPayload{}
	payload, err := heartbeatPayload.Encode()
	if err != nil {
		logs.Errorf("stream start loop heartbeat failed, encode heartbeat payload err, %v", err)
		return fmt.Errorf("encode heartbeat payload err, %v", err)
	}

	logs.Infof("stream start loop heartbeat, heartbeat interval: %v", defaultHeartbeatInterval)

	go func() {
		for {
			time.Sleep(defaultHeartbeatInterval)

			vas, _ := s.vasBuilder()

			logs.V(1).Infof("stream will heartbeat, rid: %s", vas.Rid)

			if err := s.heartbeatOnce(vas, heartbeatPayload.MessagingType(), payload); err != nil {
				logs.Warnf("stream heartbeat failed, notify reconnect upstream, err: %v, rid: %s", err, vas.Rid)

				s.NotifyReconnect(types.ReconnectSignal{Reason: "stream heartbeat failed"})
				continue
			}

			logs.V(1).Infof("stream heartbeat successfully, rid: %s", vas.Rid)
		}
	}()

	return nil
}

// heartbeatOnce send heartbeat to upstream server, if failed maxHeartbeatRetryCount count, return error.
func (s *stream) heartbeatOnce(vas *kit.Vas, msgType sfs.MessagingType, payload []byte) error {
	retry := tools.NewRetryPolicy(maxHeartbeatRetryCount, [2]uint{1000, 3000})

	var lastErr error
	for {
		if retry.RetryCount() == maxHeartbeatRetryCount {
			return lastErr
		}

		if err := s.sendHeartbeatMessaging(vas, msgType, payload); err != nil {
			logs.Errorf("send heartbeat message failed, retry count: %d, err: %v, rid: %s",
				retry.RetryCount(), err, vas.Rid)
			lastErr = err
			retry.Sleep()
			continue
		}

		return nil
	}

	return nil
}

// sendHeartbeatMessaging send heartbeat message to upstream server.
func (s *stream) sendHeartbeatMessaging(vas *kit.Vas, msgType sfs.MessagingType, payload []byte) error {
	timeoutVas, cancel := vas.WithTimeout(defaultHeartbeatTimeout)
	defer cancel()

	if _, err := s.client.Messaging(timeoutVas, msgType, payload); err != nil {
		return err
	}

	return nil
}
