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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultFileName = "bcs-project-manager.log"
	// 默认文件大小，单位 MB
	maxFileSize = 500
	// 日志保留时间，单位 天
	maxAge = 7
	// 历史文件保留数量
	maxBackups = 10
)

// getWriter 获取 writer
func getWriter(conf *config.LogConfig) (io.Writer, error) {
	if _, err := os.Stat(conf.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file path %s is not exists", conf.Path)
	}
	// 文件名称，默认为 bcs-project-manager.log
	name := conf.Name
	if name == "" {
		name = defaultFileName
	}
	rawPath := strings.TrimSuffix(conf.Path, "/")
	fileName := filepath.Join(rawPath, name)

	// 文件大小
	size := conf.Size
	if size == 0 {
		size = maxFileSize
	}

	// 日志保存时间
	age := conf.Age
	if age == 0 {
		age = maxAge
	}

	// 历史日志文件数量
	backups := conf.Backups
	if backups == 0 {
		backups = maxBackups
	}

	// 使用lumberjack实现日志切割归档
	writer := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    size,
		MaxBackups: backups,
		MaxAge:     age,
		LocalTime:  true,
	}

	return writer, nil
}
