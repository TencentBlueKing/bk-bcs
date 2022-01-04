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
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	filename = "test.log"
	size     = "1"
	backups  = "2"
	age      = "3"
)

func TestGetWriter(t *testing.T) {
	// test default writer
	writer, err := getWriter("", map[string]string{})
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	assert.Equal(t, writer, os.Stdout)

	// test stdout writer
	writer, err = getWriter("os", map[string]string{})
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	assert.Equal(t, writer, os.Stdout)

	// test stderr writer
	writer, err = getWriter("os", map[string]string{"name": "stderr"})
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	assert.Equal(t, writer, os.Stderr)

	// test file writer
	writer, err = getWriter("file", map[string]string{"name": filename})
	if err != nil {
		t.Errorf("get writer error: %v", err)
	}
	// size、backup、age is default value
	expectedWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500,
		MaxBackups: 10,
		MaxAge:     7,
		LocalTime:  true,
	}
	assert.Equal(t, writer, expectedWriter)
}

func TestGetFileWriter(t *testing.T) {
	// log with default value
	writer, err := getFileWriter(map[string]string{})
	if err != nil {
		t.Errorf("get file writer error: %v", err)
	}
	expectedWriter, ok := writer.(*lumberjack.Logger)
	if !ok {
		t.Errorf("the expected writer is not lumberjack.Logger")
	}
	assert.Equal(t, expectedWriter.Filename, defaultFileName)
	assert.Equal(t, expectedWriter.MaxSize, maxFileSize)
	assert.Equal(t, expectedWriter.MaxAge, maxAge)
	assert.Equal(t, expectedWriter.MaxBackups, maxBackups)

	// set log settings
	writer, err = getFileWriter(map[string]string{"name": filename, "size": size, "backups": backups, "age": age})
	if err != nil {
		t.Errorf("get file writer error: %v", err)
	}
	expectedWriter, ok = writer.(*lumberjack.Logger)
	if !ok {
		t.Errorf("the expected writer is not lumberjack.Logger")
	}
	assert.Equal(t, expectedWriter.Filename, filename)
	assert.Equal(t, strconv.Itoa(expectedWriter.MaxSize), size)
	assert.Equal(t, strconv.Itoa(expectedWriter.MaxAge), age)
	assert.Equal(t, strconv.Itoa(expectedWriter.MaxBackups), backups)
}
