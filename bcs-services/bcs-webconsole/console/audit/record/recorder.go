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

// Package record 终端session record
package record

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit/asciinema"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

type state int

const (
	dateTimeFormat         = "2006-01-02"
	dayTimeFormat          = "150405"
	initState        state = iota // 初始化状态
	runningState                  // webconsole session 存活状态
	terminationState              // webconsole session 已终止
	uploadingState                // cast 上传状态
	uploadedState                 // cast 已上传状态
)

type castFile struct {
	dir      string
	name     string
	filePath string // /{dir}/{name}
	absPath  string // {dataDir}/{dir}/{name}
	fd       *os.File
}

func (c *castFile) clean() {
	if c.fd != nil {
		c.fd.Close() // nolint
	}

	if err := os.Remove(c.absPath); err != nil {
		blog.Error("remove file %s, err: %s", c.absPath, err)
	}
}

// ReplyInfo 回访记录初始信息
type ReplyInfo struct {
	Width     uint16
	Height    uint16
	TimeStamp time.Time
}

// ReplyRecorder 终端回放记录器
type ReplyRecorder struct {
	SessionID string
	Info      *ReplyInfo
	Writer    *asciinema.Writer
	uploader  *Uploader
	cast      *castFile
	once      sync.Once
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

// createCastFile 创建目录和文件
func createCastFile(podCtx *types.PodContext) (*castFile, error) {
	dir := time.Now().Format(dateTimeFormat)

	err := ensureDirExist(filepath.Join(config.G.Audit.DataDir, dir))
	if err != nil {
		return nil, fmt.Errorf("create dir %s, err: %s", dir, err)
	}

	namePrefix := time.Now().Format(dayTimeFormat)
	name := fmt.Sprintf("%s_%s_%s_%s_%s.cast",
		namePrefix, podCtx.ProjectCode, podCtx.ClusterId, podCtx.Username, podCtx.SessionId[:6])
	c := castFile{
		dir:      dir,
		name:     name,
		filePath: filepath.Join("/", dir, name),
	}
	c.absPath = filepath.Join(config.G.Audit.DataDir, c.filePath)

	GetGlobalUploader().setState(c.filePath, initState)
	fd, err := os.Create(c.absPath)
	if err != nil {
		return nil, fmt.Errorf("create replay file %s, err: %s", c.absPath, err)
	}
	c.fd = fd

	return &c, nil
}

// NewReplayRecord 初始化Recorder
// 确认是否开启终端记录 / 创建记录文件 / 初始记录信息
func NewReplayRecord(ctx context.Context, podCtx *types.PodContext, terminalSize *types.TerminalSize) (
	*ReplyRecorder, error) {

	// 不开启审计
	if !config.G.Audit.Enabled {
		return nil, nil
	}

	orgInfo := &ReplyInfo{TimeStamp: time.Now()}
	orgInfo.Width = terminalSize.Cols
	orgInfo.Height = terminalSize.Rows

	recorder := &ReplyRecorder{
		SessionID: podCtx.SessionId,
		Info:      orgInfo,
		uploader:  GetGlobalUploader(),
	}

	cast, err := createCastFile(podCtx)
	if err != nil {
		return nil, fmt.Errorf("init replay file err: %s", err)
	}
	recorder.cast = cast

	options := make([]asciinema.Option, 0, 3)
	options = append(options, asciinema.WithHeight(orgInfo.Height))
	options = append(options, asciinema.WithWidth(orgInfo.Width))
	options = append(options, asciinema.WithTimestamp(orgInfo.TimeStamp))
	recorder.Writer = asciinema.NewWriter(recorder.cast.fd, podCtx, options...)
	// 初始化时写入Header信息
	if err := recorder.Writer.WriteHeader(); err != nil {
		cast.clean()
		return nil, fmt.Errorf("session %s write replay header failed: %s", recorder.SessionID, err)
	}

	//  只有写入头部的 cast 文件才上传
	recorder.uploader.setState(cast.filePath, runningState)

	return recorder, nil
}

// Flush 缓存写入到local file
func (r *ReplyRecorder) Flush() {
	if r == nil || r.Writer == nil {
		return
	}

	r.Writer.WriteBuff.Flush() // nolint
}

// End 正常退出: 关闭缓存和文件
func (r *ReplyRecorder) End() {
	if r == nil {
		return
	}

	r.once.Do(func() {
		// 关闭前将剩余缓冲区数据写入
		r.Writer.WriteBuff.Flush() // nolint
		r.cast.fd.Close()          // nolint
		r.uploader.setState(r.cast.filePath, terminationState)

		r.Writer = nil
		blog.Infof("set file %s state to termination", r.cast.filePath)
	})
}

// RecordOutputEvent 记录终端输出信息
func RecordOutputEvent(r *ReplyRecorder, p []byte) { // nolint
	// 不开启terminal recorder时, ReplyRecorder返回nil
	if r == nil || r.Writer == nil {
		return
	}

	if len(p) == 0 {
		return
	}

	if err := r.Writer.WriteRow(p, asciinema.OutputEvent); err != nil {
		blog.Errorf("session %s write replay row failed: %s", r.SessionID, err)

		r.End()
	}
}

// RecordResizeEvent 记录终端变化
func RecordResizeEvent(r *ReplyRecorder, p []byte) { // nolint
	// 不开启terminal recorder时, ReplyRecorder返回nil
	if r == nil || r.Writer == nil {
		return
	}

	if len(p) == 0 {
		return
	}

	if err := r.Writer.WriteRow(p, asciinema.ResizeEvent); err != nil {
		blog.Errorf("Session %s write replay row failed: %s", r.SessionID, err)

		r.End()
	}
}
