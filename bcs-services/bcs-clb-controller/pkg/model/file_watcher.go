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
 *
 */

package model

import (
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/fsnotify/fsnotify"
)

// FileEventType event type for file watch
type FileEventType int

const (
	// FileEventCreate file created
	FileEventCreate = iota
	// FileEventUpdate file updated
	FileEventUpdate
	// FileEventDelete file deleted
	FileEventDelete
)

// FileEvent event returned by file watcher
type FileEvent struct {
	Filename string
	Type     FileEventType
}

// NewFileEvent new a file event
func NewFileEvent(filename string, t FileEventType) FileEvent {
	return FileEvent{
		Filename: filename,
		Type:     t,
	}
}

// FileWatcher watcher for a dir or a file
type FileWatcher struct {
	CurrentDir string
	Period     time.Duration
}

// NewFileWatcher create filewatcher
func NewFileWatcher(period time.Duration) *FileWatcher {
	return &FileWatcher{
		Period: period,
	}
}

// WatchFile watch file change
func (fw *FileWatcher) WatchFile(path string) (chan FileEvent, error) {
	ech := make(chan FileEvent, 1)
	_, err := os.Stat(path)
	if err != nil {
		blog.V(5).Infof("failed to stat %s", path)
		return nil, err
	}
	pathMd5Str, err := FileMd5(path)
	if err != nil {
		return nil, err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create inotify, err %s", err.Error())
	}
	err = w.Add(path)
	if err != nil {
		return nil, fmt.Errorf("failed to add fsnotify watch for path %s, err %s", path, err.Error())
	}
	go func() {
		defer w.Close()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				blog.V(5).Infof("fsnotify watch event %s", event.String())
				filename := event.Name //filepath.Join(fw.CurrentDir, event.Name)
				if event.Op&fsnotify.Write == fsnotify.Write {
					md5Str, err := FileMd5(filename)
					if err != nil {
						blog.Errorf("falied to get md5 for file %s", filename)
						continue
					}
					if pathMd5Str == md5Str {
						continue
					}
					pathMd5Str = md5Str
					ech <- NewFileEvent(filename, FileEventUpdate)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					ech <- NewFileEvent(filename, FileEventDelete)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					ech <- NewFileEvent(filename, FileEventDelete)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					ech <- NewFileEvent(filename, FileEventCreate)
				} else {
					blog.Warnf("silent event %v for path %s", event, filename)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				blog.Errorf("fsnotify watcher get error %s", err.Error())
			}
		}
	}()
	if err != nil {
		blog.V(5).Infof("doWatch path %s failed, err %s", path, err.Error())
		return nil, fmt.Errorf("doWatch path %s failed, err %s", path, err.Error())
	}
	return ech, nil
}

// WatchDir watch dir change
// when fsnotify.Write event is received, we check the file md5sum to see if there is a file content update
// because on different OS, there may be 1 to 2 fsnotify.Write events for a single file write operation
func (fw *FileWatcher) WatchDir(path string) (chan FileEvent, error) {
	ech := make(chan FileEvent, 1)
	// files map to record file, (filename, md5) pair
	filesMap := make(map[string]string)
	fstat, err := os.Stat(path)
	if err != nil {
		blog.V(5).Infof("failed to stat %s", path)
		return nil, err
	}
	if !fstat.IsDir() {
		blog.V(5).Infof("%s is not dir", path)
		return nil, fmt.Errorf("%s is not dir", path)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create inotify, err %s", err.Error())
	}
	err = w.Add(path)
	if err != nil {
		return nil, fmt.Errorf("failed to add fsnotify watch for path %s, err %s", path, err.Error())
	}
	go func() {
		defer w.Close()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					blog.Warnf("end watch %s loop", path)
					return
				}
				blog.V(5).Infof("fsnotify watch event %s", event.String())
				filename := event.Name //filepath.Join(fw.CurrentDir, event.Name)
				if event.Op&fsnotify.Write == fsnotify.Write {
					md5Str, err := FileMd5(filename)
					if err != nil {
						blog.Errorf("falied to get md5 for file %s", filename)
						continue
					}
					oldMd5Str, ok := filesMap[filename]
					if !ok {
						filesMap[filename] = md5Str
					} else {
						if oldMd5Str == md5Str {
							continue
						}
						filesMap[filename] = md5Str
					}
					ech <- NewFileEvent(filename, FileEventUpdate)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					if filename == path {
						blog.Warnf("watched path %s is deleted, end watch loop", path)
						break
					}
					if _, ok := filesMap[filename]; ok {
						delete(filesMap, filename)
					}
					ech <- NewFileEvent(filename, FileEventDelete)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					if _, ok := filesMap[filename]; ok {
						delete(filesMap, filename)
					}
					ech <- NewFileEvent(filename, FileEventDelete)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					md5Str, err := FileMd5(filename)
					if err != nil {
						blog.Errorf("falied to get md5 for file %s", filename)
						continue
					}
					filesMap[filename] = md5Str
					ech <- NewFileEvent(filename, FileEventCreate)
				} else {
					blog.Warnf("silent event %v for path %s", event, filename)
				}
			case err, ok := <-w.Errors:
				if !ok {
					blog.Warnf("end watch %s loop", path)
					return
				}
				blog.Errorf("fsnotify watcher get error %s", err.Error())
			}
		}
	}()

	if err != nil {
		blog.V(5).Infof("doWatch path %s failed, err %s", path, err.Error())
		return nil, fmt.Errorf("doWatch path %s failed, err %s", path, err.Error())
	}
	return ech, nil
}
