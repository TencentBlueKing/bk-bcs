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

package pod

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-contrib/sse"
	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

const (
	logMaxCount      = 200
	sseInterval      = time.Millisecond * 100
	shellYellowColor = "\033[0;33m"
)

// PodLogStream Server Sent Events Handler 连接处理函数
// @Summary SSE 实时日志流
// @Tags    Logs
// @Param   container_name query string true  "容器名称"
// @Param   started_at     query string false "开始时间"
// @Produce text/event-stream
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/:pod/logs/stream [get]
func PodLogStream(req *k8sclient.LogQuery, ss rest.StreamingServer) error { // nolint
	rctx, err := rest.GetRestContext(ss.Context())
	if err != nil {
		return err
	}
	clusterId := req.ClusterId
	namespace := req.Namespace
	pod := req.Pod

	// 重连时的Id
	lastEventId := rctx.Request.Header.Get("Last-Event-ID")
	if lastEventId != "" {
		sinceTime, errr := base64.StdEncoding.DecodeString(lastEventId)
		if errr == nil {
			blog.Infow("send log stream from Last-Event-ID", "Last-Event-ID", sinceTime)
			req.StartedAt = string(sinceTime)
		}
	}

	logChan, err := k8sclient.GetPodLogStream(ss.Context(), clusterId, namespace, pod, req)
	if err != nil {
		return err
	}

	var (
		logCount    int64
		lastLogTime string
	)
	tick := time.NewTicker(sseInterval)
	defer tick.Stop()

	logList := make([]*k8sclient.Log, 0, logMaxCount+1)

	for {
		select {
		case <-ss.Context().Done():
			return nil
		case <-tick.C:
			if len(logList) == 0 {
				continue
			}

			truncateLogCount := logCount - logMaxCount
			if truncateLogCount > 0 {
				logList = append(logList, &k8sclient.Log{
					Log:  fmt.Sprintf("%sWarning, already truncate %d logs...", shellYellowColor, truncateLogCount),
					Time: lastLogTime,
				})
			}

			// id 是最后一个日志时间
			id := base64.StdEncoding.EncodeToString([]byte(lastLogTime))
			_ = render.Render(ss, rctx.Request, &rest.Event{
				HTTPCode: -1,
				Event: sse.Event{
					Event: "message",
					Data:  logList,
					Id:    id,
					Retry: 5000, // 5 秒重试
				},
			})

			err := ss.Flush()
			if err != nil {
				return errors.New("flush error: " + err.Error())
			}

			// 清空列表
			logCount = 0
			logList = logList[:0]
		case log, ok := <-logChan:
			// 服务端主动关闭
			if !ok {
				return nil
			}

			logCount++
			if logCount <= logMaxCount {
				logList = append(logList, log)
			}
			lastLogTime = log.Time
		}
	}
}
