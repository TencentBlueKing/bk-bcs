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

package api

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// UploadHandler 上传文件
// NOCC:golint/fnsize(设计如此:)
// nolint
func (s *service) UploadHandler(c *gin.Context) {
	uploadPath := c.PostForm("upload_path")
	sessionId := c.Param("sessionId")

	if uploadPath == "" {
		rest.APIError(c, i18n.GetMessage(c, "请先输入上传路径"))
		return
	}
	err := checkFileExists(uploadPath, sessionId)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "目标路径不存在"))
		return
	}
	err = checkPathIsDir(uploadPath, sessionId)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "目标路径不存在"))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		logger.Errorf("get file from request failed, err: %s", err.Error())
		rest.APIError(c, i18n.GetMessage(c, "解析上传文件失败"))
		return
	}

	opened, err := file.Open()
	if err != nil {
		logger.Errorf("open file from request failed, err: %s", err.Error())
		rest.APIError(c, i18n.GetMessage(c, "解析上传文件失败"))
		return
	}
	defer opened.Close()

	podCtx, err := sessions.NewStore().WebSocketScope().Get(c.Request.Context(), sessionId)
	if err != nil {
		logger.Errorf("get pod context by session %s failed, err: %s", sessionId, err.Error())
		rest.APIError(c, i18n.GetMessage(c, "获取pod信息失败"))
		return
	}
	reader, writer := io.Pipe()
	pe, err := podCtx.NewPodExec()
	if err != nil {
		logger.Errorf("new pod exec failed, err: %s", err.Error())
		rest.APIError(c, i18n.GetMessage(c, "执行上传命令失败"))
		return
	}
	errChan := make(chan error, 1)
	// nolint
	go func(r io.Reader, pw *io.PipeWriter) {
		tarWriter := tar.NewWriter(writer)
		defer func() {
			tarWriter.Close() // nolint
			writer.Close()    // nolint
			close(errChan)
		}()
		e := tarWriter.WriteHeader(&tar.Header{
			Name: file.Filename,
			Size: file.Size,
			Mode: 0644,
		})
		if e != nil {
			logger.Errorf("writer tar header failed, err: %s", e.Error())
			errChan <- e
			return
		}
		_, e = io.Copy(tarWriter, opened)
		if e != nil {
			logger.Errorf("writer tar from opened file failed, err: %s", e.Error())
			errChan <- e
			return
		}
		errChan <- nil
	}(opened, writer)

	pe.Stdin = reader
	// 需要同时读取 stdout/stderr, 否则可能会 block 住
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}

	pe.Command = []string{"tar", "-xmf", "-", "-C", uploadPath}
	pe.Tty = false

	if err = pe.Exec(); err != nil {
		logger.Errorf("pod exec failed, err: %s", err.Error())
		rest.APIError(c, i18n.GetMessage(c, "执行上传命令失败"))
		return
	}

	err, ok := <-errChan
	if ok && err != nil {
		logger.Errorf("writer to tar failed, err: %s", err.Error())
		rest.APIError(c, i18n.GetMessage(c, "文件上传失败"))
		return
	}

	rest.APIOK(c, i18n.GetMessage(c, "文件上传成功"), gin.H{})
}

// DownloadHandler 下载文件
func (s *service) DownloadHandler(c *gin.Context) {
	downloadPath := c.Query("download_path")
	sessionId := c.Param("sessionId")
	reader, writer := io.Pipe()
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			reader.Close() // nolint
			writer.Close() // nolint
			close(errChan)
		}()
		podCtx, err := sessions.NewStore().WebSocketScope().Get(c.Request.Context(), sessionId)
		if err != nil {
			errChan <- err
			return
		}

		pe, err := podCtx.NewPodExec()
		if err != nil {
			errChan <- err
			return
		}
		pe.Stdout = writer

		pe.Command = append([]string{"tar", "cf", "-"}, downloadPath)
		pe.Stderr = &bytes.Buffer{}
		pe.Tty = false
		err = pe.Exec()
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()
	tarReader := tar.NewReader(reader)
	_, err := tarReader.Next()
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "复制文件流失败"))
		return
	}
	fileName := downloadPath[strings.LastIndex(downloadPath, "/")+1:]
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("X-File-Name", fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	io.Copy(c.Writer, tarReader)
}

// CheckDownloadHandler 下载文件预检查
func (s *service) CheckDownloadHandler(c *gin.Context) {
	downloadPath := c.Query("download_path")
	sessionId := c.Param("sessionId")

	// 检查都返回200, 具体错误在 CheckPassed 中处理
	msg := "check done"

	if err := checkFileExists(downloadPath, sessionId); err != nil {
		rest.APIOK(c, msg, types.CheckPassed{
			Passed: false,
			Detail: err.Error(),
			Reason: i18n.GetMessage(c, "目标文件不存在"),
		})
		return
	}

	if err := checkPathIsDir(downloadPath, sessionId); err == nil {
		rest.APIOK(c, msg, types.CheckPassed{
			Passed: false,
			Detail: err.Error(),
			Reason: i18n.GetMessage(c, "暂不支持文件夹下载"),
		})
		return
	}

	if err := checkFileSize(downloadPath, sessionId, FileSizeLimits*FileSizeUnitMb); err != nil {
		rest.APIOK(c, msg, types.CheckPassed{
			Passed: false,
			Detail: err.Error(),
			Reason: i18n.GetMessage(c, "文件不能超过{}MB", map[string]int{"fileLimit": FileSizeLimits}),
		})
		return
	}

	data := types.CheckPassed{Passed: true}
	rest.APIOK(c, msg, data)
}

func checkPathIsDir(path, sessionID string) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = append([]string{"test", "-d"}, path)
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func checkFileExists(path, sessionID string) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = append([]string{"test", "-e"}, path)
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func checkFileSize(path, sessionID string, sizeLimit int) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = []string{"stat", "-c", "%s", path}
	stdout := &bytes.Buffer{}
	pe.Stdout = stdout
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	// 解析文件大小, stdout 会返回 \r\n 或者 \n
	sizeText := strings.TrimSuffix(stdout.String(), "\n")
	sizeText = strings.TrimSuffix(sizeText, "\r")
	size, err := strconv.Atoi(sizeText)
	if err != nil {
		return err
	}
	if size > sizeLimit {
		return errors.Errorf("file size %d > %d", size, sizeLimit)
	}
	return nil
}
