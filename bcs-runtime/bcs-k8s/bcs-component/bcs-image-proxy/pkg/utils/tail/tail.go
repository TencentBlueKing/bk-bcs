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

// Package tail xxx
package tail

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
)

// OnceTailLines once tail the lines from file
func OnceTailLines(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return tailLines(file, limit)
}

// FollowTailLines follow the file to tail lines
func FollowTailLines(ctx context.Context, path string, limit int) (<-chan string, error) {
	offset := int64(limit) * 300 // 假定平均一行 300 哥字符
	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "stat file '%s' failed", path)
	}
	if offset > fi.Size() {
		offset = fi.Size()
	}
	t, err := tail.TailFile(path, tail.Config{
		ReOpen: true, Follow: true, Poll: false,
		Location: &tail.SeekInfo{Offset: -offset, Whence: io.SeekEnd},
		Logger:   tail.DiscardingLogger,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "tail file '%s' failed", path)
	}
	ch := make(chan string, limit)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
			case line, ok := <-t.Lines:
				if !ok {
					return
				}
				ch <- line.Text
			}
		}
	}()
	return ch, nil
}

func reverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func tailLines(file *os.File, limit int) ([]string, error) {
	stat, _ := file.Stat()
	size := stat.Size()
	bufSize := int64(4096) // 缓冲区大小
	lines := make([]string, 0, limit)
	remaining := size

	for len(lines) < limit && remaining > 0 {
		// 获取读取的起始位置
		readAt := max(0, remaining-bufSize)
		var buf []byte
		if readAt == 0 {
			buf = make([]byte, remaining)
		} else {
			buf = make([]byte, bufSize)
		}

		n, err := file.ReadAt(buf, readAt)
		if err != nil && err != io.EOF {
			break
		}
		chunk := buf[:n]
		lineBytes := bytes.Split(chunk, []byte("\n"))
		for i := len(lineBytes) - 1; i >= 1; i-- {
			line := string(bytes.TrimRight(lineBytes[i], "\r\n"))
			lines = append(lines, line)
			if len(lines) == limit {
				break
			}
		}

		// 如果已经读取完成了，将第 0 条数据插入
		if readAt == 0 && len(lineBytes) > 0 {
			line := string(bytes.TrimRight(lineBytes[0], "\r\n"))
			lines = append(lines, line)
		} else {
			// 如果读取未完成，将第 0 条数据的字节数量释放回去（因为第 0 条数据可能是不完整的）
			readAt += int64(len(lineBytes[0]))
		}
		remaining = readAt // 更新剩余字节数
	}
	return reverseSlice(lines), nil
}
