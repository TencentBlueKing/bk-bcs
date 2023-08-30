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

package terminalRecord

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit/asciinema"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

const (
	dateTimeFormat = "2006-01-02"

	replayFilenameSuffix = ".cast"
)

// ReplyInfo 回访记录初始信息
type ReplyInfo struct {
	Width     uint16
	Height    uint16
	TimeStamp time.Time
}

// ReplyRecorder 终端回放记录器
type ReplyRecorder struct {
	SessionID   string
	Info        *ReplyInfo
	absFilePath string
	//Target      string
	Writer *asciinema.Writer
	err    error

	file *os.File
	once sync.Once
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ensureDirExist(name string) error {
	if !fileExists(name) {
		return os.MkdirAll(name, os.ModePerm)
	}
	return nil
}

// NewReplayRecord 初始化Recorder
// 确认是否开启终端记录 / 创建记录文件 / 初始记录信息
func NewReplayRecord(podCtx *types.PodContext, originTerminalSize *ReplyInfo) *ReplyRecorder {
	if !config.G.TerminalRecord.Enable {
		return nil
	}
	recorder := &ReplyRecorder{
		SessionID: podCtx.SessionId,
		Info:      originTerminalSize,
	}
	path := config.G.TerminalRecord.FilePath
	err := ensureDirExist(path)
	if err != nil {
		klog.Errorf("Create dir %s error: %s\n", path, err)
		recorder.err = err
		return recorder
	}
	date := time.Now().Format(dateTimeFormat)
	f := fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s", date, podCtx.Username, podCtx.ClusterId, podCtx.Namespace, podCtx.PodName,
		podCtx.ContainerName, podCtx.SessionId[:6])
	filename := f + replayFilenameSuffix
	absFilePath := filepath.Join(path, filename)
	recorder.absFilePath = absFilePath
	fd, err := os.Create(recorder.absFilePath)
	if err != nil {
		klog.Errorf("Create replay file %s error: %s\n", recorder.absFilePath, err)
		recorder.err = err
		return recorder
	}
	recorder.file = fd
	options := make([]asciinema.Option, 0, 3)
	options = append(options, asciinema.WithHeight(originTerminalSize.Height))
	options = append(options, asciinema.WithWidth(originTerminalSize.Width))
	options = append(options, asciinema.WithTimestamp(originTerminalSize.TimeStamp))
	recorder.Writer = asciinema.NewWriter(recorder.file, options...)
	return recorder
}

func (r *ReplyRecorder) isNullError() bool {
	return r.err != nil
}

// Record 记录终端输出信息
func Record(r *ReplyRecorder, p []byte, event asciinema.EventType) {
	//不开启terminal recorder时, ReplyRecorder返回nil
	if r == nil {
		return
	}
	if r.isNullError() {
		return
	}
	if len(p) > 0 {
		r.once.Do(func() {
			if err := r.Writer.WriteHeader(); err != nil {
				klog.Errorf("Session %s write replay header failed: %s", r.SessionID, err)
			}
		})
		if err := r.Writer.WriteRow(p, event); err != nil {
			klog.Errorf("Session %s write replay row failed: %s", r.SessionID, err)
		}
	}
}

// End 会话结束前关闭缓存和文件
func (r *ReplyRecorder) End() {
	if r == nil {
		return
	} else {
		//关闭前将剩余缓冲区数据写入
		r.Writer.Write(r.Writer.WriteBuff)
		r.file.Close()
		return
	}
}
