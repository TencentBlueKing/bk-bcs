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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
)

const (
	filename = "test.log"
	size     = 1
	backups  = 2
	age      = 3
)

func TestGetWriterByDefaultConf(t *testing.T) {
	// 测试默认配置
	logConf := config.LogConfig{
		Level:         "info",
		FlushInterval: 5,
		Path:          ".",
	}

	// 获取writer
	writer, err := getWriter(&logConf)
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	expectedWriter, ok := writer.(*lumberjack.Logger)
	if !ok {
		t.Errorf("the expected writer is not lumberjack.Logger")
	}
	assert.Equal(t, expectedWriter.Filename, defaultFileName)
	assert.Equal(t, expectedWriter.MaxSize, maxFileSize)
	assert.Equal(t, expectedWriter.MaxAge, maxAge)
	assert.Equal(t, expectedWriter.MaxBackups, maxBackups)
}

func TestGetWriter(t *testing.T) {
	// 测试默认配置
	logConf := config.LogConfig{
		Level:         "info",
		FlushInterval: 5,
		Path:          ".",
		Name:          filename,
		Size:          size,
		Backups:       backups,
		Age:           age,
	}

	// 获取writer
	writer, err := getWriter(&logConf)
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	expectedWriter, ok := writer.(*lumberjack.Logger)
	if !ok {
		t.Errorf("the expected writer is not lumberjack.Logger")
	}
	assert.Equal(t, expectedWriter.Filename, filename)
	assert.Equal(t, expectedWriter.MaxSize, size)
	assert.Equal(t, expectedWriter.MaxAge, age)
	assert.Equal(t, expectedWriter.MaxBackups, backups)
}
