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

// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File I/O for logs.

package glog

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// logNoScrolling boolean flag. Whether the restart service log is appended to the latest
// log file, or whether a new log file is created.
var logNoScrolling bool

// IsRestartLogScrolling return logNoScrolling.
func IsRestartLogScrolling() bool {
	return logNoScrolling
}

// logMaxSize is the maximum size of a log file in bytes.
var logMaxSize uint32 = 500 * 1024 * 1024

// MaxSize returns logMaxSize that is the maximum size of a log file in bytes.
func MaxSize() uint32 {
	return logMaxSize
}

// lineMaxSize is the maximum size of a line log in bytes.
var lineMaxSize uint32 = 10 * 1024

// LineMaxSize returns lineMaxSize that is the maximum size of a line log in bytes.
func LineMaxSize() uint32 {
	return lineMaxSize
}

// logMaxNum is the maximum of log files for one thread.
var logMaxNum = 10

// MaxNum returns logMaxNum that is the maximum of log files for one thread.
func MaxNum() int {
	return logMaxNum
}

// fileInfo contains log filename and its timestamp.
type fileInfo struct {
	name      string
	timestamp string
}

// fileInfoList implements Interface interface in sort. For
// sorting a list of fileInfo
type fileInfoList []fileInfo

// Len 用于排序
func (b fileInfoList) Len() int { return len(b) }

// Swap 用于排序
func (b fileInfoList) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Less 用于排序
func (b fileInfoList) Less(i, j int) bool { return b[i].timestamp < b[j].timestamp }

// fileBlock is a block of chain in logKeeper.
type fileBlock struct {
	fileInfo
	next *fileBlock
}

// logKeeper maintains a chain of each level log file. Its head
// is the earliest file while its tail is the oldest. It remains
// up to MaxNum() files, and the extra added will lead to delete
// the oldest. At first it load from logDir and take existing files
// into the chain. And remove the part over MaxNum().
type logKeeper struct {
	dir      string
	onceLoad sync.Once
	header   *fileBlock
	tail     *fileBlock
	total    int
}

// nolint result `ok` is always `false`
func (lk *logKeeper) add(newBlock *fileBlock) (ok bool) {
	block := lk.tail
	if block == nil {
		lk.header = newBlock
	} else {
		if block.name == newBlock.name {
			return false
		}
		block.next = newBlock
	}
	lk.tail = newBlock
	lk.total++
	for lk.total > MaxNum() {
		lk.remove()
	}
	return ok
}

// nolint result `ok` is always `false`
func (lk *logKeeper) remove() (ok bool) {
	if lk.header == nil || lk.total == 0 {
		return
	}
	block := lk.header
	if err := lk.removeFile(block.name); err != nil {
		panic(fmt.Sprintf("remove file is fail, error: %v", err))
	}
	lk.header = block.next
	block = nil // for GC
	lk.total--
	return ok
}

func (lk *logKeeper) removeFile(name string) error {
	return os.Remove(filepath.Join(lk.dir, name))
}

func (lk *logKeeper) load() {
	_dir, err := os.ReadDir(lk.dir)
	if err != nil {
		return
	}

	reg := logNameReg()
	blockList := make(fileInfoList, 0, len(_dir))
	for _, fi := range _dir {
		if fi.IsDir() {
			continue
		}

		result := reg.FindStringSubmatch(fi.Name())
		if result == nil {
			continue
		}

		name, timestamp := result[0], result[2]
		blockList = append(blockList, fileInfo{name: name, timestamp: timestamp})
	}

	sort.Sort(blockList)
	for i, block := range blockList {
		if i < MaxNum() {
			fb := &fileBlock{
				fileInfo: fileInfo{name: block.name, timestamp: block.timestamp},
				next:     nil,
			}
			if i == 0 {
				lk.header = fb
			} else {
				lk.tail.next = fb
			}
			lk.tail = fb
			lk.total++
		} else {
			if err := lk.removeFile(block.name); err != nil {
				panic(fmt.Sprintf("remove file is fail, error: %v", err))
			}
		}
	}
}

// logDirs lists the candidate directories for new log files.
var logDirs []*logKeeper

// If non-empty, overrides the choice of directory in which to write logs.
// See createLogDirs for the full list of possible destinations.
var logDir = "./logs"

func createLogDirs() {
	var dirs []string
	if logDir != "" {
		dirs = append(dirs, logDir)
	}
	dirs = append(dirs, os.TempDir())

	for _, dir := range dirs {
		head := new(fileBlock)
		tail := new(fileBlock)
		total := 0

		logDirs = append(logDirs, &logKeeper{dir: dir, header: head, tail: tail, total: total})
	}
}

var (
	pid      = os.Getpid()
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.ReplaceAll(userName, `\`, "_")
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(tag string, t time.Time) (name, link string) {
	name = fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
		program,
		host,
		userName,
		tag,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		pid)
	return name, program + "." + tag
}

// logNameReg returns a regexp object for match log file name.
func logNameReg() *regexp.Regexp {
	reg, _ := regexp.Compile(fmt.Sprintf(`^%s\..+\..+\.log.(%s)\.(\d{8}-\d{6})\.\d+$`,
		program,
		strings.Join(severityName, "|")))
	return reg
}

var onceLogDirs sync.Once

// create creates a new log file and returns the file and its filename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the file is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func create(t time.Time) (f *os.File, filename string, filesize uint32, err error) {
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, "", 0, errors.New("log: no log dirs")
	}

	name, link := logName(logFileTag, t)
	for _, lk := range logDirs {
		// when the system starts, you need to determine whether the last log file is full,
		// and if not, continue to write.
		needCreate := true
		var onceErr error
		lk.onceLoad.Do(func() {
			// load log dir all log file.
			lk.load()

			// judge whether log restart append is enabled.
			if !IsRestartLogScrolling() {
				return
			}

			filename = filepath.Join(lk.dir, lk.tail.name)

			// judge latest log file write full.
			var fInfo os.FileInfo
			fInfo, onceErr = os.Stat(filename)
			if onceErr != nil {
				return
			}

			// if log file write full, need to create new log file.
			if uint32(fInfo.Size()) >= MaxSize() {
				return
			}
			filesize = uint32(fInfo.Size())

			// if log file not write full, need to open latest log file.
			f, onceErr = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0600)
			if onceErr != nil {
				return
			}
			needCreate = false
		})

		if onceErr != nil {
			return nil, "", 0, onceErr
		}

		if !needCreate {
			return // nolint naked return
		}

		fname := filepath.Join(lk.dir, name)
		f, err = os.Create(fname)
		if err == nil {
			symlink := filepath.Join(lk.dir, link)
			if e := os.Remove(symlink); e != nil {
				fmt.Fprintf(os.Stderr, "log: create remove: %v\n", e)
			}

			if e := os.Symlink(name, symlink); e != nil {
				fmt.Fprintf(os.Stderr, "log: create symlink: %v\n", e)
			}

			lk.add(&fileBlock{fileInfo: fileInfo{name: name, timestamp: ""}, next: nil})
			return f, fname, 0, nil
		}
	}
	return nil, "", 0, fmt.Errorf("log: cannot create log: %v", err)
}
