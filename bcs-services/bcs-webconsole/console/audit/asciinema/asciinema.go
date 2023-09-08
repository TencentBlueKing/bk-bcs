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

package asciinema

import (
	"encoding/json"
	"errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"io"
	"time"
)

// EventType 事件类型: o:输出 / r:终端变化
type EventType string

const (
	maxBuffSize = 1024
	maxFileSize = 100 * 1000 * 1000 // 100M

	version      = 2
	defaultShell = "/bin/bash"
	defaultTerm  = "xterm"

	OutputEvent EventType = "o"
	ResizeEvent EventType = "r"
)

var (
	newLine = []byte{'\n'}
)

// NewWriter 初始化Writer
func NewWriter(w io.Writer, podCtx *types.PodContext, opts ...Option) *Writer {
	//m := Meta{
	//	UserName: podCtx.Username,
	//	ClusterID: podCtx.ClusterId,
	//	NameSpace: podCtx.Namespace,
	//	PodName: podCtx.PodName,
	//	ContainerName: podCtx.ContainerName,
	//	SessionID: podCtx.SessionId,
	//}
	conf := Config{
		Width:    80,
		Height:   40,
		EnvShell: defaultShell,
		EnvTerm:  defaultTerm,
		//Meta: m,
		podCtx: podCtx,
	}
	for _, setter := range opts {
		setter(&conf)
	}
	return &Writer{
		Config:        conf,
		TimestampNano: conf.Timestamp.UnixNano(),
		writer:        w,
		limit:         maxFileSize,
		WriteBuff:     make([]byte, 0, maxBuffSize),
	}
}

// Writer 自定义文件写入,文件格式遵循asciinema format
type Writer struct {
	Config
	TimestampNano int64
	writer        io.Writer
	limit         int
	written       int
	WriteBuff     []byte
}

// WriteHeader 写入头信息
func (w *Writer) WriteHeader() error {
	header := Header{
		Version:   version,
		Width:     w.Width,
		Height:    w.Height,
		Timestamp: w.Timestamp.Unix(),
		Title:     w.Title,
		Env: Env{
			Shell: w.EnvShell,
			Term:  w.EnvTerm,
		},
		Meta: Meta{
			UserName:      w.podCtx.Username,
			ClusterID:     w.podCtx.ClusterId,
			NameSpace:     w.podCtx.Namespace,
			PodName:       w.podCtx.PodName,
			ContainerName: w.podCtx.PodName,
			SessionID:     w.podCtx.SessionId,
		},
	}
	raw, err := json.Marshal(header)
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	if err != nil {
		return err
	}
	_, err = w.Write(newLine)
	return err
}

// WriteRow 记录terminal输出的流式信息
func (w *Writer) WriteRow(p []byte, event EventType) error {
	now := time.Now().UnixNano()
	ts := float64(now-w.TimestampNano) / 1000 / 1000 / 1000
	return w.WriteStdout(ts, p, event)
}

// WriteStdout 批量写入,减少文件io
func (w *Writer) WriteStdout(ts float64, data []byte, event EventType) error {
	row := []interface{}{ts, event, string(data)}
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}

	// buff 做批量写入文件
	if len(raw) > maxBuffSize { //读取的数据大于最大缓存
		//先缓存写入再将读取的数据写入文件
		_, err := w.Write(w.WriteBuff)
		if err != nil {
			return err
		}
		w.WriteBuff = w.WriteBuff[:0]
		_, err = w.Write(raw)
		if err != nil {
			return err
		}
		_, err = w.Write(newLine)
		return err
	} else if len(raw)+len(w.WriteBuff) > maxBuffSize { //检查是否达到最大容量
		//buff写入文件,清空buff,raw写入buff
		_, err := w.Write(w.WriteBuff)
		if err != nil {
			return err
		}
		w.WriteBuff = w.WriteBuff[:0]
		raw := append(raw, newLine...)
		w.WriteBuff = append(w.WriteBuff, raw...)
	} else {
		raw := append(raw, newLine...)
		w.WriteBuff = append(w.WriteBuff, raw...)
	}
	return nil
}

// WriteFileInterval 检查buff是否为空,不为空则写入文件
func (w *Writer) WriteFileInterval() error {
	if len(w.WriteBuff) > 0 {
		_, err := w.writer.Write(w.WriteBuff)
		w.WriteBuff = w.WriteBuff[:0]
		return err
	} else {
		return nil
	}
}

// Header 文件头部信息
type Header struct {
	Version   int    `json:"version"`
	Width     uint16 `json:"width"`
	Height    uint16 `json:"height"`
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title"`
	Env       Env    `json:"env"`
	Meta      Meta   `json:"meta"`
}

// Env 文件env信息
type Env struct {
	Shell string `json:"SHELL"`
	Term  string `json:"TERM"`
}

type Meta struct {
	UserName      string `json:"user_name"`
	ClusterID     string `json:"cluster_id"`
	NameSpace     string `json:"name_space"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
	SessionID     string `json:"session_id"`
}

// Write 限制文件大小
func (w *Writer) Write(p []byte) (n int, err error) {
	remainingSpace := w.limit - w.written
	if remainingSpace <= 0 {
		return 0, errors.New("Exceeds the file size")
	}

	if len(p) > remainingSpace {
		p = p[:remainingSpace]
	}

	n, err = w.writer.Write(p)
	w.written += n
	return n, err
}
