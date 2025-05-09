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

// Package util xxx
package util

import (
	"bufio"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"syscall"
	"time"
)

// LogFile xxx
type LogFile struct {
	filename      string
	file          *os.File
	pos           int64
	ino           uint64
	LogChann      chan string
	searchKeyList []string
}

func openLogFile(filename string) (*LogFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		file.Close()
		return nil, fmt.Errorf("failed to get inode number for %s", filename)
	}

	return &LogFile{filename: filename, file: file, pos: info.Size(), ino: stat.Ino, LogChann: make(chan string)}, nil
	//return &LogFile{filename: filename, file: file, pos: 0, ino: stat.Ino, LogChann: make(chan string)}, nil
}

// SetSearchKey xxx
func (f *LogFile) SetSearchKey(searchKeyList []string) {
	f.searchKeyList = searchKeyList
}

// Start xxx
func (f *LogFile) Start() {
	go func() {
		for {
			err := f.checkNewEntries()
			if err != nil {
				fmt.Println("Error checking file:", err)
				close(f.LogChann)
				break
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

// CheckNewEntriesOnce xxx
func (f *LogFile) CheckNewEntriesOnce() ([]string, error) {
	info, err := f.file.Stat()
	if err != nil {
		return nil, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("failed to get inode number for %s", f.filename)
	}

	if stat.Ino != f.ino {
		// 文件已经被轮转，关闭旧的文件并打开新的文件
		klog.Info("%s file already changed, reopen it.", f.filename)
		f.file.Close()

		newFile, err := openLogFile(f.filename)
		if err != nil {
			return nil, err
		}

		*f = *newFile
		f.pos = 0

		info, err = f.file.Stat()
		if err != nil {
			return nil, err
		}
	}

	result := make([]string, 0, 0)
	if info.Size() > f.pos {
		if _, err := f.file.Seek(f.pos, io.SeekStart); err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(f.file)
		for scanner.Scan() {
			line := scanner.Text()

			if f.searchKeyList != nil && len(f.searchKeyList) > 0 {
				for _, key := range f.searchKeyList {
					if strings.Contains(line, key) {
						result = append(result, line)
						break
					}
				}
			} else {
				result = append(result, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		f.pos = info.Size()
	}

	return result, nil
}

func (f *LogFile) checkNewEntries() error {
	info, err := f.file.Stat()
	if err != nil {
		return err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get inode number for %s", f.filename)
	}

	if stat.Ino != f.ino {
		// 文件已经被轮转，关闭旧的文件并打开新的文件
		f.file.Close()

		newFile, err := openLogFile(f.filename)
		if err != nil {
			return err
		}

		*f = *newFile

		// 不读取历史内容
		f.pos = info.Size()
		return nil
	}

	if info.Size() > f.pos {
		if _, err := f.file.Seek(f.pos, io.SeekStart); err != nil {
			return err
		}
		scanner := bufio.NewScanner(f.file)
		for scanner.Scan() {
			line := scanner.Text()

			// mce: [Hardware Error]
			if f.searchKeyList != nil && len(f.searchKeyList) > 0 {
				for _, key := range f.searchKeyList {
					if strings.Contains(line, key) {
						f.LogChann <- line
					}
				}
			} else {
				f.LogChann <- line
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		f.pos = info.Size()
	}

	return nil
}

// NewLogFile xxx
func NewLogFile(path string) *LogFile {
	logFile, err := openLogFile(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}

	return logFile
}
