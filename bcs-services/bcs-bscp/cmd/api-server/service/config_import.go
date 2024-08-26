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

package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/panjf2000/ants/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/archive"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type configImport struct {
	authorizer auth.Authorizer
	provider   repository.Provider
	cfgClient  pbcs.ConfigClient
	mc         *metric
}

// TemplateConfigFileImport Import template config file
//
//nolint:funlen
func (c *configImport) TemplateConfigFileImport(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	unzipStr := r.Header.Get("X-Bscp-Unzip")
	unzip, _ := strconv.ParseBool(unzipStr)

	tmplSpaceIdStr := chi.URLParam(r, "template_space_id")
	tmplSpaceID, _ := strconv.Atoi(tmplSpaceIdStr)
	if tmplSpaceID == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "id is required"))))
		return
	}

	fileName := chi.URLParam(r, "filename")

	// Validation size
	totleContentLength, singleContentLength := getUploadConfig(kt.BizID)

	var maxSize int64
	if unzip {
		maxSize = totleContentLength
	} else {
		maxSize = singleContentLength
	}

	if err := checkUploadSize(kt, r, maxSize); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	// Ensure r.Body is closed after reading
	defer r.Body.Close()
	buffer := make([]byte, 512)
	// 读取前512字节以检测文件类型
	n, errR := r.Body.Read(buffer[:512])
	if errR != nil && errR != io.EOF {
		_ = render.Render(w, r,
			rest.BadRequest(errors.New(i18n.T(kt, "read file failed, err: %v", errR))))
		return
	}

	// 组合上一次读取
	combinedReader := io.MultiReader(bytes.NewReader(buffer[:n]), r.Body)

	// 默认当文件处理
	identifyFileType := archive.Unknown
	// 自动解压
	if unzip {
		identifyFileType = archive.IdentifyFileType(buffer[:n])
	}

	// 创建目录
	dirPath := path.Join(os.TempDir(), constant.UploadTemporaryDirectory)
	if err := createTemporaryDirectory(dirPath); err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "create directory failed, err: %v", err))))
		return
	}
	// 随机生成临时目录
	tempDir, err := os.MkdirTemp(dirPath, "templateConfigItem-")
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "create temporary directory failed, err: %v", err))))
		return
	}

	if identifyFileType != archive.Unknown {
		if err = archive.Unpack(combinedReader, identifyFileType, tempDir, singleContentLength*constant.MB); err != nil {
			compare := new(errf.ErrorF)
			if errors.As(err, &compare) {
				if compare.Code == int32(archive.FileTooLarge) {
					_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt,
						"decompress the file. The size of file %s exceeds the maximum limit of %s", compare.Message,
						tools.BytesToHumanReadable(uint64(singleContentLength*constant.MB))))))
					return
				}
			}
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "decompression failed, err: %v", err))))
			return
		}
	} else {
		if err = saveFile(combinedReader, tempDir, fileName); err != nil {
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
			return
		}
	}

	defer func() { _ = os.RemoveAll(tempDir) }()

	// 先扫描一遍文件夹，获取路径和名称，
	fileItems, err := getFilePathsAndNames(tempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
		return
	}

	// 扫描某个目录大小
	// 暴露 metrics
	totalSize, _ := getDirSize(dirPath)
	c.uploadFileMetrics(kt.BizID, tmplSpaceIdStr, dirPath, totalSize)

	if err = c.checkFileConfictsWithTemplates(kt, uint32(tmplSpaceID), fileItems); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	folder, err := c.processAndUploadDirectoryFiles(kt, tempDir, len(fileItems))
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
		return
	}

	templateItems := []*pbcs.ListTemplateByTupleReq_Item{}
	uploadErrCount := 0
	for _, item := range folder {
		if item.Err != nil {
			uploadErrCount++
		}
		templateItems = append(templateItems, &pbcs.ListTemplateByTupleReq_Item{
			Name: item.File.Name,
			Path: item.File.Path,
		})
	}

	// 压缩包里面文件过多会导致sql占位符不够, 需要分批处理
	batchSize := constant.UploadBatchSize
	templateItem := map[string]*pbcs.ListTemplateByTupleResp_Item{}
	for i := 0; i < len(templateItems); i += batchSize {
		end := i + batchSize
		if end > len(templateItems) {
			end = len(templateItems)
		}
		batch := templateItems[i:end]
		// 批量验证biz_id、template_space_id、name、path是否存在
		tuple, err := c.cfgClient.ListTemplateByTuple(kt.RpcCtx(), &pbcs.ListTemplateByTupleReq{
			BizId:           kt.BizID,
			TemplateSpaceId: uint32(tmplSpaceID),
			Items:           batch,
		})
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "list template config failed, err: %v", err))))
			return
		}
		if len(tuple.Items) > 0 {
			for _, item := range tuple.Items {
				templateItem[path.Join(item.GetTemplateRevision().GetSpec().Path,
					item.GetTemplateRevision().GetSpec().GetName())] = item
			}
		}
	}

	exist := make([]*types.TemplateItem, 0)
	nonExist := make([]*types.TemplateItem, 0)
	for _, item := range folder {
		pathName := path.Join(item.File.Path, item.File.Name)
		data, ok := templateItem[pathName]
		newItem := &types.TemplateItem{
			Name:     item.File.Name,
			Path:     item.File.Path,
			FileType: item.File.FileType,
			Sign:     item.File.Sign,
			ByteSize: item.File.ByteSize,
			Md5:      item.File.Md5,
		}
		if !ok {
			newItem.FileMode = string(table.Unix)
			nonExist = append(nonExist, newItem)
		} else {
			newItem.Id = data.GetTemplate().GetId()
			newItem.FileMode = data.GetTemplateRevision().GetSpec().GetFileMode()
			newItem.Memo = data.GetTemplate().GetSpec().GetMemo()
			newItem.Privilege = data.GetTemplateRevision().GetSpec().GetPermission().GetPrivilege()
			newItem.User = data.GetTemplateRevision().GetSpec().GetPermission().GetUser()
			newItem.UserGroup = data.GetTemplateRevision().GetSpec().GetPermission().GetUserGroup()
			exist = append(exist, newItem)
		}
	}

	msg := i18n.T(kt, "upload completed")
	if uploadErrCount > 0 {
		msg = i18n.T(kt, "upload completed, %d failed", uploadErrCount)
	}
	sortByPathName(exist)
	sortByPathName(nonExist)
	_ = render.Render(w, r, rest.OKRender(&types.TemplatesImportResp{
		Exist:    exist,
		NonExist: nonExist,
		Msg:      msg,
	}))
}

// ConfigFileImport Import config file
//
//nolint:funlen
func (c *configImport) ConfigFileImport(w http.ResponseWriter, r *http.Request) {

	kt := kit.MustGetKit(r.Context())
	// Ensure r.Body is closed after reading
	defer r.Body.Close()

	unzipStr := r.Header.Get("X-Bscp-Unzip")
	unzip, _ := strconv.ParseBool(unzipStr)

	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "id is required"))))
		return
	}
	kt.AppID = uint32(appId)

	fileName := chi.URLParam(r, "filename")

	// Validation size
	totleContentLength, singleContentLength := getUploadConfig(kt.BizID)
	var maxSize int64
	if unzip {
		maxSize = totleContentLength
	} else {
		maxSize = singleContentLength
	}
	if err := checkUploadSize(kt, r, maxSize); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	buffer := make([]byte, 512)
	// 读取前512字节以检测文件类型
	n, errR := r.Body.Read(buffer[:512])
	if errR != nil && errR != io.EOF {
		_ = render.Render(w, r,
			rest.BadRequest(errors.New(i18n.T(kt, "read file failed, err: %v", errR))))
		return
	}

	// 组合上一次读取
	combinedReader := io.MultiReader(bytes.NewReader(buffer[:n]), r.Body)

	// 默认当文件处理
	identifyFileType := archive.Unknown
	// 自动解压
	if unzip {
		identifyFileType = archive.IdentifyFileType(buffer[:n])
	}

	// 创建目录
	dirPath := path.Join(os.TempDir(), constant.UploadTemporaryDirectory)
	if err := createTemporaryDirectory(dirPath); err != nil {
		_ = render.Render(w, r,
			rest.BadRequest(errors.New(i18n.T(kt, "create directory failed, err: %v", err))))
		return
	}

	// 随机生成临时目录
	tempDir, err := os.MkdirTemp(dirPath, "configItem-")
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "create temporary directory failed, err: %v", err))))
		return
	}

	if identifyFileType != archive.Unknown {
		if err = archive.Unpack(combinedReader, identifyFileType, tempDir, singleContentLength*constant.MB); err != nil {
			compare := new(errf.ErrorF)
			if errors.As(err, &compare) {
				if compare.Code == int32(archive.FileTooLarge) {
					_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt,
						"decompress the file. The size of file %s exceeds the maximum limit of %s", compare.Message,
						tools.BytesToHumanReadable(uint64(singleContentLength*constant.MB))))))
					return
				}
			}
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "decompression failed, err: %v", err))))
			return
		}
	} else {
		if err = saveFile(combinedReader, tempDir, fileName); err != nil {
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
			return
		}
	}

	defer func() { _ = os.RemoveAll(tempDir) }()

	// 先扫描一遍文件夹，获取路径和名称，
	fileItems, err := getFilePathsAndNames(tempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
		return
	}

	// 扫描某个目录大小
	// 暴露 metrics
	totalSize, _ := getDirSize(dirPath)
	c.uploadFileMetrics(kt.BizID, appIdStr, dirPath, totalSize)

	if err = c.checkFileConfictsWithNonTemplates(kt, fileItems); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	// 批量验证biz_id、app_id、name、path是否存在
	configItems := []*pbcs.ListConfigItemByTupleReq_Item{}
	for _, item := range fileItems {
		configItems = append(configItems, &pbcs.ListConfigItemByTupleReq_Item{
			Name: item.Name,
			Path: item.Path,
		})
	}

	// 压缩包里面文件过多会导致sql占位符不够, 需要分批处理
	batchSize := constant.UploadBatchSize
	configItem := map[string]*pbci.ConfigItem{}
	for i := 0; i < len(configItems); i += batchSize {
		end := i + batchSize
		if end > len(configItems) {
			end = len(configItems)
		}
		batch := configItems[i:end]
		tuple, errC := c.cfgClient.ListConfigItemByTuple(kt.RpcCtx(), &pbcs.ListConfigItemByTupleReq{
			BizId: kt.BizID,
			AppId: kt.AppID,
			Items: batch,
		})
		if errC != nil {
			_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "list config item failed, err: %v", err))))
			return
		}
		for _, item := range tuple.GetDetails() {
			configItem[path.Join(item.GetSpec().GetPath(), item.GetSpec().GetName())] = item
		}
	}

	// 配置项总数 + 引入的套餐配置项数量 - 已存在的数量 + 新增的数量
	result, err := c.cfgClient.GetTemplateAndNonTemplateCICount(kt.RpcCtx(), &pbcs.GetTemplateAndNonTemplateCICountReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(
			i18n.T(kt, "get the current number of service config items failed, err: %v", err))))
		return
	}

	addCount := len(fileItems) - len(configItem)
	total := int(result.GetConfigItemCount()+result.GetTemplateConfigItemCount()) + len(configItem) + addCount
	limit := getAppConfigCnt(kt.BizID)
	if total > limit {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt,
			"decompress file failed, exceeding the maximum file limit threshold of %d", limit))))
		return
	}

	folder, err := c.processAndUploadDirectoryFiles(kt, tempDir, len(fileItems))
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(errors.New(i18n.T(kt, "upload file failed, err: %v", err))))
		return
	}

	uploadErrCount := 0
	for _, item := range folder {
		if item.Err != nil {
			uploadErrCount++
		}
	}

	exist := make([]*types.TemplateItem, 0)
	nonExist := make([]*types.TemplateItem, 0)
	for _, item := range folder {
		pathName := path.Join(item.File.Path, item.File.Name)
		config, ok := configItem[pathName]
		newItem := &types.TemplateItem{
			Name:     item.File.Name,
			Path:     item.File.Path,
			FileType: item.File.FileType,
			Sign:     item.File.Sign,
			ByteSize: item.File.ByteSize,
			Md5:      item.File.Md5,
		}
		if !ok {
			newItem.FileMode = string(table.Unix)
			nonExist = append(nonExist, newItem)
		} else {
			newItem.Id = config.GetId()
			newItem.FileMode = config.GetSpec().GetFileMode()
			newItem.Memo = config.GetSpec().GetMemo()
			newItem.Privilege = config.GetSpec().GetPermission().GetPrivilege()
			newItem.User = config.GetSpec().GetPermission().GetUser()
			newItem.UserGroup = config.GetSpec().GetPermission().GetUserGroup()
			exist = append(exist, newItem)
		}
	}

	msg := i18n.T(kt, "upload completed")
	if uploadErrCount > 0 {
		msg = i18n.T(kt, "upload completed, %d failed", uploadErrCount)
	}
	sortByPathName(exist)
	sortByPathName(nonExist)
	_ = render.Render(w, r, rest.OKRender(&types.TemplatesImportResp{
		Exist:    exist,
		NonExist: nonExist,
		Msg:      msg,
	}))
}

// 检测上传文件的大小
func checkUploadSize(kt *kit.Kit, r *http.Request, maxSizeMB int64) error {
	maxSize := maxSizeMB * constant.MB
	if r.ContentLength > maxSize {
		return errors.New(i18n.T(kt, "upload failed, please make sure the file size does not exceed %s",
			tools.BytesToHumanReadable(uint64(maxSize))))
	}
	return nil
}

// getFilePathsAndNames Retrieve all files in the specified directory and
// return the path, name, and directory size of the files
func getFilePathsAndNames(fileDir string) ([]tools.CIUniqueKey, error) {

	rootDir := fileDir
	files := make([]tools.CIUniqueKey, 0)

	err := filepath.WalkDir(fileDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			fileDir, err := handlerFilePath(rootDir, path)
			if err != nil {
				return err
			}
			files = append(files, tools.CIUniqueKey{Path: fileDir, Name: fileInfo.Name()})
		}
		return nil
	})
	return files, err
}

// 处理目录中的文件，包括检测文件类型、计算 SHA-256 哈希值以及上传文件
func (c *configImport) processAndUploadDirectoryFiles(kt *kit.Kit, fileDir string, numFiles int) (
	[]types.UploadTask, error) {

	rootDir := fileDir
	results := make([]types.UploadTask, 0, numFiles)

	// 创建一个并发池
	pool, err := ants.NewPool(constant.MaxConcurrentUpload)
	if err != nil {
		return nil, fmt.Errorf("generates an instance of ants pool fail %v", err)
	}
	defer pool.Release()

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	err = filepath.WalkDir(fileDir, func(path string, file os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !file.IsDir() {
			wg.Add(1)
			upload := func() {
				defer func() {
					wg.Done()
					mu.Unlock()
				}()
				result, err := c.fileScannerHasherUploader(kt, path, rootDir, file)
				mu.Lock()
				results = append(results, types.UploadTask{
					File: result,
					Err:  err,
				})
			}
			// 提交上传任务到并发池中执行
			if submitErr := pool.Submit(upload); submitErr != nil {
				return submitErr
			}
		}
		return nil
	})
	wg.Wait()

	return results, err
}

// 扫描文件夹中的文件、计算 SHA-256 哈希值、检测文件类型并上传文件
func (c *configImport) fileScannerHasherUploader(kt *kit.Kit, path, rootDir string, file fs.DirEntry) (
	types.FileInfo, error) {

	resp := types.FileInfo{}

	fileInfo, err := file.Info()
	if err != nil {
		return resp, err
	}

	fileContent, err := os.Open(path)
	if err != nil {
		return resp, fmt.Errorf("open file fail filepath: %s; err: %s", path, err.Error())
	}

	filePath, err := handlerFilePath(rootDir, path)
	if err != nil {
		return resp, fmt.Errorf("handler file path fail filepath: %s; err: %s", path, err.Error())
	}

	hashString, err := calculateSHA256(fileContent)
	if err != nil {
		return resp, fmt.Errorf("calculate SHA256 fail filename: %s; err: %s", fileInfo.Name(), err.Error())
	}

	fileType := ""
	// Check if the file size is greater than 5MB (5 * 1024 * 1024 bytes)
	if fileInfo.Size() > constant.MaxUploadTextFileSize {
		fileType = string(table.Binary)
	} else {
		// 从池中获取一个缓冲区
		fileBuffer := bufferPool.Get().(*bytes.Buffer)
		defer func() {
			// 清空缓冲区内容 以备复用
			fileBuffer.Reset()
			// 将缓冲区放回池中
			bufferPool.Put(fileBuffer)
		}()
		_, _ = fileContent.Seek(0, 0)
		_, _ = io.Copy(fileBuffer, fileContent)
		fileType = detectFileType(fileBuffer.Bytes())
	}
	_, _ = fileContent.Seek(0, 0)

	// if err is ErrFileContentNotFound, the file does not exist. Do not handle it
	existObjectMetadata, err := c.provider.Metadata(kt, hashString)
	if err != nil && !errors.Is(err, errf.ErrFileContentNotFound) {
		return resp, err
	}

	resp.Name = fileInfo.Name()
	resp.FileType = fileType
	resp.Path = filePath
	// it exists in repo/cos without any errors
	if err == nil {
		resp.ByteSize = uint64(existObjectMetadata.ByteSize)
		resp.Sign = existObjectMetadata.Sha256
		return resp, nil
	}

	result, err := c.provider.Upload(kt, hashString, fileContent)
	if err != nil {
		return resp, fmt.Errorf("fiel upload fail filename: %s; err: %s", fileInfo.Name(), err.Error())
	}

	// 验证文件是否上传成功
	repoRes, err := c.provider.Metadata(kt, hashString)
	if err != nil {
		return resp, fmt.Errorf("fiel upload fail filename: %s; err: %s", fileInfo.Name(), err.Error())
	}

	resp.ByteSize = uint64(result.ByteSize)
	resp.Sign = result.Sha256
	resp.Md5 = repoRes.Md5

	return resp, nil
}

// 使用 filepath.Rel 函数计算给定文件路径 path 相对于根目录 rootDir 的相对路径
func handlerFilePath(rootDir, path string) (string, error) {
	// 计算文件相对于根目录的相对路径
	fileDir, err := filepath.Rel(rootDir, path)
	if err != nil {
		return "", err
	}
	// 获取相对路径的目录部分
	fileDir = filepath.Dir(fileDir)
	// 如果目录部分为"."，将其设置为根目录"/"
	if fileDir == "." {
		fileDir = "/"
	} else {
		// 否则在目录前加上"/"
		fileDir = "/" + fileDir
	}
	return fileDir, nil
}

// 计算文件的SHA-256散列值
func calculateSHA256(reader io.Reader) (string, error) {

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// 检测文件类型
func detectFileType(buf []byte) string {
	fileType := string(table.Text)
	// 通过前512字节判断
	filetype := http.DetectContentType(buf)
	if isTextType(filetype) {
		return fileType
	}

	// 通过是否能被utf8编码
	if !utf8.Valid(buf) {
		fileType = string(table.Binary)
	}

	return fileType
}

// 判断内容类型是否为文本类型
func isTextType(contentType string) bool {
	textTypes := []string{
		"text/plain",
		"text/html",
		"text/css",
		"text/javascript",
		"application/json",
		"application/xml",
		"application/xhtml+xml",
	}

	for _, t := range textTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func newConfigImportService(settings cc.Repository, authorizer auth.Authorizer,
	cfgClient pbcs.ConfigClient) (*configImport, error) {
	provider, err := repository.NewProvider(settings)
	if err != nil {
		return nil, err
	}
	config := &configImport{
		authorizer: authorizer,
		provider:   provider,
		cfgClient:  cfgClient,
		mc:         initMetric(),
	}
	return config, nil
}

// 使用sort包按照路径+名称排序
func sortByPathName(myStructs []*types.TemplateItem) {
	// 使用 sort 包按照路径+名称排序
	sort.Slice(myStructs, func(i, j int) bool {
		// 先按照路径排序
		if myStructs[i].Path != myStructs[j].Path {
			return myStructs[i].Path < myStructs[j].Path
		}
		// 如果路径相同，则按照名称排序
		return myStructs[i].Name < myStructs[j].Name
	})
}

// 检测与非配置文件冲突
func (c *configImport) checkFileConfictsWithNonTemplates(kt *kit.Kit, files []tools.CIUniqueKey) error {
	// 获取服务下的所有非模板配置文件
	items, err := c.cfgClient.ListConfigItems(kt.RpcCtx(), &pbcs.ListConfigItemsReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
		All:   true,
	})
	if err != nil {
		return err
	}

	filesToCompare := []tools.CIUniqueKey{}
	for _, v := range items.GetDetails() {
		filesToCompare = append(filesToCompare, tools.CIUniqueKey{
			Name: v.Spec.Name,
			Path: v.Spec.Path,
		})
	}

	return tools.DetectFilePathConflicts(kt, files, filesToCompare)
}

// 检测与模板套餐下的文件冲突
func (c *configImport) checkFileConfictsWithTemplates(kt *kit.Kit, templateSpaceId uint32,
	files []tools.CIUniqueKey) error {
	// 获取该空间下所有文件
	items, err := c.cfgClient.ListTemplates(kt.RpcCtx(), &pbcs.ListTemplatesReq{
		BizId:           kt.BizID,
		TemplateSpaceId: templateSpaceId,
		All:             true,
	})
	if err != nil {
		return errors.New(i18n.T(kt, "list templates failed, err: %v", err))
	}
	filesToCompare := []tools.CIUniqueKey{}
	for _, v := range items.GetDetails() {
		filesToCompare = append(filesToCompare, tools.CIUniqueKey{
			Name: v.Spec.Name,
			Path: v.Spec.Path,
		})
	}

	return tools.DetectFilePathConflicts(kt, files, filesToCompare)
}

// 临时保存文件
func saveFile(reader io.Reader, tempDir, fileName string) error {
	if filepath.Clean(tempDir) != tempDir {
		return fmt.Errorf("invalid temp dir: %s", tempDir)
	}
	if filepath.Clean(fileName) != fileName {
		return fmt.Errorf("invalid file name: %s", fileName)
	}
	// 创建文件
	file, err := os.Create(tempDir + "/" + fileName)
	if err != nil {
		return fmt.Errorf("create temp file failed, err: %v", err.Error())
	}
	defer file.Close()

	if _, err = io.Copy(file, reader); err != nil {
		return err
	}

	return nil
}

// Check if a directory exists and create it if it does not exist
func createTemporaryDirectory(dirPath string) error {
	// Check if the directory exists
	_, err := os.Stat(dirPath)
	if !os.IsNotExist(err) {
		return err
	}
	// If the directory does not exist, create it
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	return nil
}

// Function to calculate the size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// 上传文件夹时暴露 metrics
func (c *configImport) uploadFileMetrics(bizID uint32, resourceID, directory string,
	directorySize int64) {
	c.mc.currentUploadedFolderSize.WithLabelValues(strconv.Itoa(int(bizID)),
		resourceID, directory).Set(float64(directorySize))
}

func getUploadConfig(bizID uint32) (maxUploadContentLength, maxFileSize int64) {
	// 默认值
	resLimit := cc.ApiServer().FeatureFlags.ResourceLimit
	maxUploadContentLength, maxFileSize = int64(resLimit.Default.MaxUploadContentLength),
		int64(resLimit.Default.MaxFileSize)
	if resLimit, ok := cc.ApiServer().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.MaxUploadContentLength > 0 {
			maxUploadContentLength = int64(resLimit.MaxUploadContentLength)
		}
		if resLimit.MaxFileSize > 0 {
			maxFileSize = int64(resLimit.MaxFileSize)
		}
	}

	return
}

func getAppConfigCnt(bizID uint32) int {
	if resLimit, ok := cc.ApiServer().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.AppConfigCnt > 0 {
			return int(resLimit.AppConfigCnt)
		}
	}
	return int(cc.ApiServer().FeatureFlags.ResourceLimit.Default.AppConfigCnt)
}
