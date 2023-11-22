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
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/panjf2000/ants/v2"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/dal/repository"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbci "bscp.io/pkg/protocol/core/config-item"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/archive"
	"bscp.io/pkg/types"
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
}

// TemplateConfigFileImport Import template config file
func (c *configImport) TemplateConfigFileImport(w http.ResponseWriter, r *http.Request) { // nolint
	kt := kit.MustGetKit(r.Context())
	tmplSpaceIdStr := chi.URLParam(r, "template_space_id")
	tmplSpaceID, _ := strconv.Atoi(tmplSpaceIdStr)
	if tmplSpaceID == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	// Validation size
	if r.ContentLength > constant.MaxUploadContentLength {
		_ = render.Render(w, r, rest.BadRequest(errors.New("request body size exceeds 100MB")))
		return
	}
	// Unzip file
	unpackTempDir, err := archive.Unpack(r.Body)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	// 删除临时文件夹
	defer func() { _ = os.RemoveAll(unpackTempDir) }()
	folder, uploadErr, err := c.scanFolder(kt, unpackTempDir, unpackTempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	templateItems := []*pbcs.ListTemplateByTupleReq_Item{}
	for _, item := range folder {
		templateItems = append(templateItems, &pbcs.ListTemplateByTupleReq_Item{
			Name: item.Name,
			Path: item.Path,
		})
	}
	// 批量验证biz_id、template_space_id、name、path是否存在
	tuple, err := c.cfgClient.ListTemplateByTuple(kt.RpcCtx(), &pbcs.ListTemplateByTupleReq{
		BizId:           kt.BizID,
		TemplateSpaceId: uint32(tmplSpaceID),
		Items:           templateItems,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	templateItem := map[string]*pbcs.ListTemplateByTupleResp_Item{}
	if len(tuple.Items) > 0 {
		for _, item := range tuple.Items {
			templateItem[path.Join(item.GetTemplateRevision().GetSpec().Path,
				item.GetTemplateRevision().GetSpec().GetName())] = item
		}
	}
	exist := make([]*types.TemplateItem, 0)
	nonExist := make([]*types.TemplateItem, 0)
	for _, item := range folder {
		pathName := path.Join(item.Path, item.Name)
		data, ok := templateItem[pathName]
		if !ok {
			nonExist = append(nonExist, &types.TemplateItem{
				Name:     item.Name,
				Path:     item.Path,
				FileType: item.FileType,
				FileMode: string(table.Unix),
				Sign:     item.Sign,
				ByteSize: item.ByteSize,
			})
		} else {
			exist = append(exist, &types.TemplateItem{
				Id:        data.GetTemplate().GetId(),
				Name:      item.Name,
				Path:      item.Path,
				FileType:  item.FileType,
				FileMode:  data.GetTemplateRevision().GetSpec().GetFileMode(),
				Memo:      data.GetTemplate().GetSpec().GetMemo(),
				Privilege: data.GetTemplateRevision().GetSpec().GetPermission().GetPrivilege(),
				User:      data.GetTemplateRevision().GetSpec().GetPermission().GetUser(),
				UserGroup: data.GetTemplateRevision().GetSpec().GetPermission().GetUserGroup(),
				Sign:      item.Sign,
				ByteSize:  item.ByteSize,
			})
		}
	}
	msg := "上传完成"
	if len(uploadErr) > 0 {
		msg = fmt.Sprintf("上传完成，失败 %d 个", len(uploadErr))
	}
	sort.Slice(nonExist, func(i, j int) bool {
		return nonExist[i].Path < nonExist[j].Path
	})
	sort.Slice(exist, func(i, j int) bool {
		return exist[i].Path < exist[j].Path
	})
	_ = render.Render(w, r, rest.OKRender(&types.TemplatesImportResp{
		Exist:    exist,
		NonExist: nonExist,
		Msg:      msg,
	}))
}

// ConfigFileImport Import config file
func (c *configImport) ConfigFileImport(w http.ResponseWriter, r *http.Request) { // nolint
	kt := kit.MustGetKit(r.Context())
	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	// Validation size
	if r.ContentLength > constant.MaxUploadContentLength {
		_ = render.Render(w, r, rest.BadRequest(errors.New("request body size exceeds 100MB")))
		return
	}
	// Unzip file
	unpackTempDir, err := archive.Unpack(r.Body)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	// 删除临时文件夹
	defer func() { _ = os.RemoveAll(unpackTempDir) }()
	folder, uploadErr, err := c.scanFolder(kt, unpackTempDir, unpackTempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	configItems := []*pbcs.ListConfigItemByTupleReq_Item{}
	for _, item := range folder {
		configItems = append(configItems, &pbcs.ListConfigItemByTupleReq_Item{
			Name: item.Name,
			Path: item.Path,
		})
	}
	// 批量验证biz_id、app_id、name、path是否存在
	tuple, err := c.cfgClient.ListConfigItemByTuple(kt.RpcCtx(), &pbcs.ListConfigItemByTupleReq{
		BizId: kt.BizID,
		AppId: uint32(appId),
		Items: configItems,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	configItem := map[string]*pbci.ConfigItem{}
	if len(tuple.GetDetails()) > 0 {
		for _, item := range tuple.Details {
			configItem[path.Join(item.GetSpec().GetPath(), item.GetSpec().GetName())] = item
		}
	}
	exist := make([]*types.TemplateItem, 0)
	nonExist := make([]*types.TemplateItem, 0)
	for _, item := range folder {
		pathName := path.Join(item.Path, item.Name)
		config, ok := configItem[pathName]
		if !ok {
			nonExist = append(nonExist, &types.TemplateItem{
				Name:     item.Name,
				Path:     item.Path,
				FileType: item.FileType,
				FileMode: string(table.Unix),
				Sign:     item.Sign,
				ByteSize: item.ByteSize,
			})
		} else {
			exist = append(exist, &types.TemplateItem{
				Id:        config.GetId(),
				Name:      item.Name,
				Path:      item.Path,
				FileType:  item.FileType,
				FileMode:  config.GetSpec().GetFileMode(),
				Memo:      config.GetSpec().GetMemo(),
				Privilege: config.GetSpec().GetPermission().GetPrivilege(),
				User:      config.GetSpec().GetPermission().GetUser(),
				UserGroup: config.GetSpec().GetPermission().GetUserGroup(),
				Sign:      item.Sign,
				ByteSize:  item.ByteSize,
			})
		}
	}
	msg := "上传完成"
	if len(uploadErr) > 0 {
		msg = fmt.Sprintf("上传完成，失败 %d 个", len(uploadErr))
	}
	sort.Slice(nonExist, func(i, j int) bool {
		return nonExist[i].Path < nonExist[j].Path
	})
	sort.Slice(exist, func(i, j int) bool {
		return exist[i].Path < exist[j].Path
	})
	_ = render.Render(w, r, rest.OKRender(&types.TemplatesImportResp{
		Exist:    exist,
		NonExist: nonExist,
		Msg:      msg,
	}))
}

// 扫描文件夹
// fileDir 文件目录
// rootDir 根目录
func (c *configImport) scanFolder(kt *kit.Kit, fileDir, rootDir string) ([]*types.FileInfo, []error, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		files []*types.FileInfo
		errs  []error
	)

	// 启用协程池
	pool, err := ants.NewPool(constant.MaxConcurrentUpload)
	if err != nil {
		return nil, nil, fmt.Errorf("generates an instance of ants pool fail %s", err)
	}
	defer ants.Release()

	err = filepath.WalkDir(fileDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			wg.Add(1)
			upload := func() {
				defer wg.Done()
				file, uploadErr := c.uploadFile(kt, path, rootDir)
				if uploadErr != nil {
					mu.Lock()
					errs = append(errs, uploadErr)
					mu.Unlock()
				} else {
					mu.Lock()
					files = append(files, file)
					mu.Unlock()
				}
			}
			_ = pool.Submit(upload)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	wg.Wait()
	return files, errs, nil
}

// 上传文件
func (c *configImport) uploadFile(kt *kit.Kit, filePath, rootDir string) (*types.FileInfo, error) {
	var (
		fileType = string(table.Text)
		fileDir  = ""
	)
	fileContent, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file fail %s: %v", filePath, err)
	}
	defer func(fileContent *os.File) {
		_ = fileContent.Close()
	}(fileContent)

	// 计算文件的SHA-256散列值
	hash := sha256.New()
	if _, err = io.Copy(hash, fileContent); err != nil {
		return nil, err
	}
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	// 处理路径、名称、文件类型
	fileInfo, err := fileContent.Stat()
	if err != nil {
		return nil, err
	}

	fileDir = strings.Replace(filePath, rootDir, "", 1)
	fileDir = strings.ReplaceAll(fileDir, "\\", "/")
	lastSlashIndex := strings.LastIndex(fileDir, "/")
	if lastSlashIndex >= 0 {
		fileDir = fileDir[:lastSlashIndex]
	}
	// 默认根目录
	if len(fileDir) == 0 {
		fileDir = "/"
	}
	// 从池中获取一个缓冲区
	fileBuffer := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		// 清空缓冲区内容 以备复用
		fileBuffer.Reset()
		// 将缓冲区放回池中
		bufferPool.Put(fileBuffer)
	}()

	// Check if the file size is greater than 5MB (5 * 1024 * 1024 bytes)
	if fileInfo.Size() > constant.MaxUploadTextFileSize {
		fileType = string(table.Binary)
	} else {
		// 重置文件读取位置
		_, _ = fileContent.Seek(0, 0)
		_, _ = io.Copy(fileBuffer, fileContent)
		if !utf8.Valid(fileBuffer.Bytes()) {
			fileType = string(table.Binary)
		}
	}

	// 重置文件读取位置
	_, _ = fileContent.Seek(0, 0)
	upload, err := c.provider.Upload(kt, hashString, fileContent)
	if err != nil {
		return nil, fmt.Errorf("file upload fail %s: %v", filePath, err)
	}
	return &types.FileInfo{
		Name:     fileInfo.Name(),
		Path:     fileDir,
		FileType: fileType,
		Sign:     upload.Sha256,
		ByteSize: uint64(upload.ByteSize),
	}, nil
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
	}
	return config, nil
}
