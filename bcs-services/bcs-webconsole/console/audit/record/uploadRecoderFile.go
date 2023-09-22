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

// Package record xxx
package record

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/repository"
)

// UploadRecordFile 定时上传录制文件
// 将实时上传失败的文件上传和被动关闭session前将文件上传
func UploadRecordFile(ctx context.Context, done chan struct{}) {
	timer := time.NewTicker(10 * time.Minute)
	defer timer.Stop()
	storage, err := repository.NewProvider(config.G.Repository.StorageType)
	if err != nil {
		klog.Errorf("Init storage err: %v\n", err)
		return
	}
	data_dir := config.G.Audit.DataDir
	remainFiles := make(map[string][]string)

	for {
		select {
		case <-ctx.Done():
			// deadline
			timeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// 关闭前将未上传文件上传
			shutDown(timeCtx, data_dir, remainFiles, storage)
			// Canceling this context releases resources associated with it
			cancel()
			done <- struct{}{}
			return
		case <-timer.C:
			// 1.查询遗留未上传文件
			// 2.已关闭的遗留文件上传, 未关闭待下次上传

			// 路径与文件映射
			pf := make(map[string][]string)
			err = filepath.Walk(data_dir, func(path string, info fs.FileInfo, err error) error {
				if !info.IsDir() {
					path = strings.TrimPrefix(path, data_dir)
					pattern := `/(\d{4}-\d{2}-\d{2})/*`
					re, err := regexp.Compile(pattern)
					if err != nil {
						return err
					}
					matches := re.FindAllStringSubmatch(path, -1)
					for _, match := range matches {
						if len(match) >= 2 {
							p := match[1]
							prefix := "/" + match[1] + "/"
							f := strings.TrimPrefix(path, prefix)
							pf[p] = append(pf[p], f)
						}
					}
				}
				return nil
			})
			if err != nil {
				klog.Errorf("Walk file path %s failed, %v\n", data_dir, err)
				return
			}

			for path, localFiles := range pf {
				remoteFiles, e := storage.ListFile(ctx, path)
				if e != nil {
					klog.Errorf("Get storage file list err: %v\n", e)
					return
				}
				remainFiles[path] = findRemain(localFiles, remoteFiles)
			}

			// 遗留文件上传
			err = uploadRemain(ctx, data_dir, remainFiles, storage)
			if err != nil {
				klog.Errorf("Upload remain files err: %v\n", err)
			}

		}
	}
}

func findRemain(local, remote []string) []string {
	remain := make([]string, 0)
	eleMap := make(map[string]bool)
	for _, r := range remote {
		eleMap[r] = true
	}
	for _, l := range local {
		if ok := eleMap[l]; !ok {
			remain = append(remain, l)
		}
	}
	return remain
}

func getRecordEnd(f string) (bool, error) {
	file, err := os.Open(f)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 定位到文件末尾
	_, err = file.Seek(-2, io.SeekEnd)
	if err != nil {
		return false, err
	}

	// 从末尾向前搜索换行符
	buffer := make([]byte, 1)
	var lastLine []byte
	for {
		_, err = file.Read(buffer)
		if err != nil {
			return false, fmt.Errorf("read file err: %v", err)
		}

		// 如果找到换行符，则停止读取
		if buffer[0] == '\n' {
			break
		}

		// 在最后一行前插入字符，以便构建完整的最后一行内容
		lastLine = append([]byte{buffer[0]}, lastLine...)
		_, err = file.Seek(-2, io.SeekCurrent)
		if err != nil {
			return false, fmt.Errorf("roll back file pointer err: %v", err)
		}

		// 如果已经达到文件开头，停止读取
		if _, offsetErr := file.Seek(0, io.SeekCurrent); offsetErr != nil {
			break
		}
	}
	if string(lastLine) == recordEnd {
		return true, nil
	}
	return false, nil
}

func writeRecordEnd(f string) error {
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open file error:%v", err)
	}
	defer file.Close()
	lastLine := append([]byte(recordEnd), []byte("\n")...)
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("move to file end err:%v", err)
	}
	_, err = file.Write(lastLine)
	if err != nil {
		return fmt.Errorf("write last line err:%v", err)
	}
	return nil
}

func uploadRemain(ctx context.Context, data_dir string,
	remainFiles map[string][]string, storage repository.Provider) error {
	for path, files := range remainFiles {
		for _, file := range files {
			localPath := filepath.Join(data_dir, path, file)
			end, err := getRecordEnd(localPath)
			if err != nil {
				return fmt.Errorf("read file last line err: %v", err)
			}
			if end {
				remotePath := filepath.Join(path, file)
				err = storage.UploadFile(ctx, localPath, remotePath)
				if err != nil {
					return fmt.Errorf("read file last line err: %v", err)
				}
				// 上传成功更新remainFiles
				remainFiles[path] = removeElement(remainFiles[path], file)
				if len(remainFiles[path]) == 0 {
					delete(remainFiles, path)
				}
			}
		}
	}
	return nil
}

// removeElement 移除element元素
func removeElement(slice []string, element string) []string {
	var result []string

	for _, value := range slice {
		if value != element {
			result = append(result, value)
		}
	}

	return result
}

// shutDown 关闭前上传遗留的文件
func shutDown(ctx context.Context, data_dir string, remainFiles map[string][]string, storage repository.Provider) {
	// 先写入recordEnd,超时时间内未完成上传留待下次上传
	for path, remainFile := range remainFiles {
		for _, file := range remainFile {
			localPath := filepath.Join(data_dir, path, file)
			err := writeRecordEnd(localPath)
			if err != nil {
				klog.Errorf("Write remain files recordEnd err: %v\n", err)
				return
			}
		}
	}
	// 上传
	err := uploadRemain(ctx, data_dir, remainFiles, storage)
	if err != nil {
		klog.Errorf("Upload remain files err: %v\n", err)
		return
	}

}
