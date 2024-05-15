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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
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

	newFiles, err := scanFiles(unpackTempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	// 获取某个空间下的所有配置文件
	items, err := c.cfgClient.ListTemplates(kt.RpcCtx(), &pbcs.ListTemplatesReq{
		BizId:           kt.BizID,
		TemplateSpaceId: uint32(tmplSpaceID),
		All:             true,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	filesToCompare := []tools.CIUniqueKey{}
	for _, v := range items.GetDetails() {
		filesToCompare = append(filesToCompare, tools.CIUniqueKey{
			Name: v.Spec.Name,
			Path: v.Spec.Path,
		})
	}

	// 检测同级下不能出现同名的文件和文件夹
	if err = tools.DetectFilePathConflicts(newFiles, filesToCompare); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	folder, uploadErr, err := c.scanFolderAndUploadFiles(kt, unpackTempDir, unpackTempDir)
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
		newItem := &types.TemplateItem{
			Name:     item.Name,
			Path:     item.Path,
			FileType: item.FileType,
			Sign:     item.Sign,
			ByteSize: item.ByteSize,
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
	msg := "上传完成"
	if len(uploadErr) > 0 {
		msg = fmt.Sprintf("上传完成，失败 %d 个", len(uploadErr))
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
func (c *configImport) ConfigFileImport(w http.ResponseWriter, r *http.Request) { // nolint
	kt := kit.MustGetKit(r.Context())
	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	kt.AppID = uint32(appId)

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
	defer func() { _ = os.RemoveAll(unpackTempDir) }()

	fileItems, err := scanFiles(unpackTempDir)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	if err = c.getAllConfigFileByApp(kt, fileItems); err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	folder, uploadErr, err := c.scanFolderAndUploadFiles(kt, unpackTempDir, unpackTempDir)
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
	for _, item := range tuple.GetDetails() {
		configItem[path.Join(item.GetSpec().GetPath(), item.GetSpec().GetName())] = item
	}
	exist, nonExist := []*types.TemplateItem{}, []*types.TemplateItem{}
	for _, item := range folder {
		pathName := path.Join(item.Path, item.Name)
		config, ok := configItem[pathName]
		newItem := &types.TemplateItem{
			Name:     item.Name,
			Path:     item.Path,
			FileType: item.FileType,
			Sign:     item.Sign,
			ByteSize: item.ByteSize,
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
	msg := "上传完成"
	if len(uploadErr) > 0 {
		msg = fmt.Sprintf("上传完成，失败 %d 个", len(uploadErr))
	}
	sortByPathName(exist)
	sortByPathName(nonExist)
	_ = render.Render(w, r, rest.OKRender(&types.TemplatesImportResp{
		Exist:    exist,
		NonExist: nonExist,
		Msg:      msg,
	}))
}

// scanFiles 遍历指定的文件夹，并返回包含所有文件的CIUniqueKey切片
func scanFiles(fileDir string) ([]tools.CIUniqueKey, error) {
	var files []tools.CIUniqueKey
	rootDir := fileDir
	err := filepath.WalkDir(fileDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			fileDir, err = filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			fileDir = filepath.Dir(fileDir)
			// 默认根目录
			if fileDir == "." {
				fileDir = "/"
			} else {
				fileDir = "/" + fileDir
			}
			files = append(files, tools.CIUniqueKey{Path: fileDir, Name: fileInfo.Name()})
		}
		return nil
	})

	return files, err
}

// scanFolderAndUploadFiles 遍历指定的文件夹，并上传文件
func (c *configImport) scanFolderAndUploadFiles(kt *kit.Kit, fileDir, rootDir string) ([]*types.FileInfo,
	[]error, error) {
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

// uploadFile 上传文件
func (c *configImport) uploadFile(kt *kit.Kit, filePath, rootDir string) (*types.FileInfo, error) {
	var (
		fileType = string(table.Text)
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

	fileDir, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return nil, err
	}
	fileDir = filepath.Dir(fileDir)
	// 默认根目录
	if fileDir == "." {
		fileDir = "/"
	} else {
		fileDir = "/" + fileDir
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

// getAllConfigFileByApp 获取某个服务下的所有非模板配置文件
func (c *configImport) getAllConfigFileByApp(kt *kit.Kit, newFiles []tools.CIUniqueKey) error {
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

	return tools.DetectFilePathConflicts(newFiles, filesToCompare)
}
