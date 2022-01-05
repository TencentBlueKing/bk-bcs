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
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

var logFileName = "cr.log"

type logContent struct {
	Level string `json:"level"`
	Ts    string `json:"ts"`
	Msg   string `json:"msg"`
}

func TestGetLogger(t *testing.T) {
	logConf := config.LogConf{
		Level:         "info",
		FlushInterval: 5,
		Path:          ".",
	}
	// 获取 logger
	log := newZapJSONLogger(&logConf)

	// 写入日志
	log.Error("this is a test")
	log.Sync()

	// 读取日志文件内容
	file, err := os.Open(logFileName)
	if err != nil {
		t.Errorf("log file not found")
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	var logC logContent

	// 转换为json，用以判断内容
	if err := json.Unmarshal(content, &logC); err != nil {
		t.Errorf("content to json error, %v", err)
	}
	assert.Equal(t, logC.Level, "ERROR")

	// 删除日志文件
	if err = os.Remove(logFileName); err != nil {
		t.Errorf("delete log file failed")
	}
}
