/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxBackups  = 10  // the maximum number of old log files to retain
	maxFileSize = 500 // the maximum size in megabytes of the log file, megabytes
	maxAge      = 7   // the maximum number of days to retain old log files
)

// get log writer from zap or os
func getWriter(writerType string, settings map[string]string) (io.Writer, error) {
	switch writerType {
	case "os":
		return getOSWriter(settings)
	case "file":
		return getFileWriter(settings)
	default:
		return getOSWriter(map[string]string{"name": "stdout"})
	}
}

// get os writer
func getOSWriter(settings map[string]string) (io.Writer, error) {
	switch settings["name"] {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return os.Stdout, nil
	}
}

// get file log writer
func getFileWriter(settings map[string]string) (io.Writer, error) {
	path, ok := settings["path"]
	if ok {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("file path %s not exists", path)
		}
	} else {
		return nil, errors.New("log file path should not be empty")
	}

	// file path, default is <processname>-lumberjack.log
	filename := settings["name"]
	logPath := filename
	if path != "" {
		rawPath := strings.TrimSuffix(path, "/")
		logPath = filepath.Join(rawPath, filename)
	}

	// backup file
	backups := maxBackups
	backupsStr, ok := settings["backups"]
	if ok {
		backupsInt, err := strconv.Atoi(backupsStr)
		if err != nil {
			return nil, errors.New("backups should be integer")
		}
		backups = backupsInt
	}

	// file size
	size := maxFileSize
	sizeStr, ok := settings["size"]
	if ok {
		sizeInt, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, errors.New("size should be integer")
		}
		size = sizeInt
	}

	// retain file time
	age := maxAge
	ageStr, ok := settings["age"]
	if ok {
		ageInt, err := strconv.Atoi(ageStr)
		if err != nil {
			return nil, errors.New("age should be integer")
		}
		age = ageInt
	}

	// 使用lumberjack实现日志切割归档
	writer := &lumberjack.Logger{
		Filename: logPath,
		// megabytes
		MaxSize:    size,
		MaxBackups: backups,
		// days
		MaxAge:    age,
		LocalTime: true,
	}

	return writer, nil
}
