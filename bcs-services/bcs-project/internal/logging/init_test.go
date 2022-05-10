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
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
)

var logFileName = "project.log"

type logContent struct {
	Level      string `json:"level"`
	Ts         string `json:"ts"`
	Msg        string `json:"msg"`
	Stacktrace string `json:"stacktrace"`
}

func TestGetLogger(t *testing.T) {
	logConf := config.LogConfig{
		Level:         "info",
		FlushInterval: 5,
		Path:          ".",
	}
	// 获取 logger
	log := newZapJSONLogger(&logConf)

	// 写入 info 和 error 级别日志
	log.Info("this is info test")
	log.Error("this is error test")
	log.Sync()

	file, err := os.Open(logFileName)
	if err != nil {
		t.Errorf("log file not found")
	}
	defer file.Close()

	// 读取日志文件内容，测试日志级别及堆栈存在
	var content []byte
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = scanner.Bytes()
		var logC logContent
		// 转换为json，用以判断内容
		if err := json.Unmarshal(content, &logC); err != nil {
			t.Errorf("content to json error, %v", err)
		}
		switch logC.Level {
		case "INFO":
			assert.Empty(t, logC.Stacktrace)
		case "ERROR":
			assert.NotEmpty(t, logC.Stacktrace)
		default:
			t.Errorf("log level is not in [INFO, ERROR]")
		}
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("scan file failed, %v", err)
	}

	// 删除日志文件
	if err := os.Remove(logFileName); err != nil {
		t.Errorf("delete log file failed")
	}
}
