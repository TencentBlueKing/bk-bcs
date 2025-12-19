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

// Package cmd http 接口实现
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
	httpUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/http"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpx"
)

// SpaceNames model template space name
type SpaceNames struct {
	TemplateSpaceNames []string
}

// TemplateContent model template content
type TemplateContent struct {
	TemplateName string
	Content      string
}

const (
	// MaxFileSize 压缩文件解压之后不能大于10M
	MaxFileSize = 10 << 20
)

// NewAPIRouter http handler
func NewAPIRouter(crs *clusterResourcesService) *mux.Router {
	r := mux.NewRouter()
	// add middleware
	r.Use(httpx.LoggingMiddleware)
	r.Use(httpx.AuthenticationMiddleware)
	r.Use(httpx.ParseProjectIDMiddleware)
	r.Use(httpx.AuthorizationMiddleware)
	r.Use(httpx.AuditMiddleware)
	r.Use(httpx.TenantMiddleware)

	// events 接口代理
	r.Methods("GET").Path("/clusterresources/api/v1/projects/{projectCode}/clusters/{clusterID}/events").
		Handler(httpx.ParseClusterIDMiddleware(http.HandlerFunc(StorageEvents(crs)), false))
	r.Methods("POST").Path("/clusterresources/api/v1/projects/{projectCode}/clusters/{clusterID}/events").
		Handler(httpx.ParseClusterIDMiddleware(http.HandlerFunc(StorageEvents(crs)), true))
	// import template
	r.Methods("POST").Path("/clusterresources/api/v1/projects/{projectCode}/import/template").
		HandlerFunc(ImportTemplate(crs))
	// export template
	r.Methods("POST").Path("/clusterresources/api/v1/projects/{projectCode}/export/template").
		HandlerFunc(ExportTemplate(crs))
	return r
}

// StorageEvents reverse proxy events
func StorageEvents(crs *clusterResourcesService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURLPath := fmt.Sprintf("%s/bcsstorage/v1/events", config.G.Component.BCSStorageHost)

		targetURL, err := url.Parse(targetURLPath)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		clusterID := contextx.GetClusterIDFromCtx(r.Context())
		query := r.URL.Query()
		query.Set("clusterId", clusterID)
		targetURL.RawQuery = query.Encode()

		proxy := httpUtil.NewHTTPReverseProxy(crs.clientTLSConfig, func(request *http.Request) {
			request.URL = targetURL
			request.Method = r.Method
			request.Body = r.Body
		})
		proxy.ServeHTTP(w, r)
	}
}

// ImportTemplate import template
func ImportTemplate(crs *clusterResourcesService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		projectCode := contextx.GetProjectCodeFromCtx(r.Context())

		// 文件夹名称, 存放tar解压后的文件夹名称、文件名称、文件内容
		templateSpaceNames, tarContent, err := parseTarContent(r)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		// 空压缩文件的时候直接返回
		if len(tarContent) == 0 {
			httpx.ResponseOK(w, r, nil)
			return
		}

		condTemplateSpace := operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectCode: projectCode,
			entity.FieldKeyName: operator.M{
				"$in": templateSpaceNames,
			},
		})

		// 获取已存在的文件夹列表
		templateSpaces, err := crs.model.ListTemplateSpace(r.Context(), condTemplateSpace)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		now := time.Now().Unix()
		// 已存在的文件夹用 imported + 时间戳替换
		for _, v := range templateSpaces {
			if _, ok := tarContent[v.Name]; ok {
				tarContent[v.Name+"-imported-"+fmt.Sprintf("%d", now)] = tarContent[v.Name]
			}
			delete(tarContent, v.Name)
		}

		// 将要创建的模板文件夹、模板文件、模板版本准备好
		createTemplateSpaces, createTemplates, createTemplateVersions :=
			parseTemplate(ctxkey.GetUsernameFromCtx(r.Context()), projectCode, tarContent)

		// 创建顺序：templateVersion -> template -> templateSpace
		err = crs.model.CreateTemplateVersionBatch(r.Context(), createTemplateVersions)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		err = crs.model.CreateTemplateBatch(r.Context(), createTemplates)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		err = crs.model.CreateTemplateSpaceBatch(r.Context(), createTemplateSpaces)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		httpx.ResponseOK(w, r, nil)
	}

}

// ExportTemplate Export template
func ExportTemplate(crs *clusterResourcesService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		projectCode := contextx.GetProjectCodeFromCtx(r.Context())

		// 获取req内容
		templateSpaceNames, err := parseTemplateSpaceNames(r)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		// 根据templateSpaceNames查询列表
		condTemplateSpace := operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectCode: projectCode,
			entity.FieldKeyName: operator.M{
				"$in": templateSpaceNames,
			},
		})

		templateSpaces, err := crs.model.ListTemplateSpace(r.Context(), condTemplateSpace)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		if len(templateSpaces) != len(templateSpaceNames) {
			httpx.ResponseSystemError(w, r,
				errors.New(i18n.GetMsg(r.Context(), "templateSpaceNames没找到或者无权限访问")))
			return
		}

		// 通过项目编码、文件夹名称检索
		cond := operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectCode: projectCode,
			entity.FieldKeyTemplateSpace: operator.M{
				"$in": templateSpaceNames,
			},
		})
		templates, err := crs.model.ListTemplate(r.Context(), cond)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		// 通过文件夹名称、模板名称、版本检索
		var templateIDs []entity.TemplateID
		for _, v := range templates {
			templateIDs = append(templateIDs, entity.TemplateID{
				TemplateSpace:   v.TemplateSpace,
				TemplateName:    v.Name,
				TemplateVersion: v.Version,
			})
		}

		templateVersions := crs.model.ListTemplateVersionFromTemplateIDs(r.Context(), projectCode, templateIDs)

		templateContents := getTemplateSpaceContent(templateSpaceNames, templateVersions)
		timeName := fmt.Sprintf("%d", time.Now().Unix())
		// 加上时间戳
		fileName := projectCode + "_template_" + timeName + ".tgz"
		// 对模板文件文件夹的版本内容进行压缩
		err = packTemplateSpace(fileName, templateContents)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		// 返回数据处理
		err = getFileData(fileName, w)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		defer func(filePath string) {
			err = os.Remove(filePath)
			if err != nil {
				fmt.Printf("Failed to remove tgz file %s: %s", filePath, err)
			}
		}(fileName)
	}
}

// parse template、template version、template space
func parseTemplate(username, projectCode string,
	tarContent map[string][]TemplateContent) ([]*entity.TemplateSpace, []*entity.Template, []*entity.TemplateVersion) {

	var createTemplateSpaces []*entity.TemplateSpace
	var createTemplates []*entity.Template
	var createTemplateVersions []*entity.TemplateVersion
	for k, v := range tarContent {
		createTemplateSpaces = append(createTemplateSpaces, &entity.TemplateSpace{
			Name:        k,
			ProjectCode: projectCode,
			Description: "import template",
		})
		// 模板文件内容及版本
		for _, vv := range v {
			createTemplateVersions = append(createTemplateVersions, &entity.TemplateVersion{
				ProjectCode:   projectCode,
				Description:   "import template",
				TemplateName:  vv.TemplateName,
				TemplateSpace: k,
				Version:       "1.0.0",
				EditFormat:    "yaml",
				Content:       vv.Content,
				Creator:       username,
			})

			createTemplates = append(createTemplates, &entity.Template{
				Name:          vv.TemplateName,
				ProjectCode:   projectCode,
				Description:   "import template",
				TemplateSpace: k,
				ResourceType:  parser.GetResourceTypesFromManifest(vv.Content),
				Creator:       username,
				Updator:       username,
				VersionMode:   0,
				Version:       "1.0.0",
				IsDraft:       false,
			})
		}
	}
	return createTemplateSpaces, createTemplates, createTemplateVersions
}

// parse template space names
func parseTemplateSpaceNames(r *http.Request) ([]string, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var spaceNames SpaceNames
	err = json.Unmarshal(b, &spaceNames)
	if err != nil {
		return nil, err
	}

	if len(spaceNames.TemplateSpaceNames) == 0 {
		return nil, errors.New("templateSpaceNames is null")
	}

	return spaceNames.TemplateSpaceNames, nil
}

// 返回文件夹及里面的内容
func getTemplateSpaceContent(
	templateSpaceNames []string, templateVersions []*entity.TemplateVersion) map[string][]TemplateContent {

	result := make(map[string][]TemplateContent, 0)
	// 空的文件夹也要进行压缩
	for _, v := range templateSpaceNames {
		result[v] = []TemplateContent{}
	}

	for _, v := range templateVersions {
		result[v.TemplateSpace] = append(result[v.TemplateSpace], TemplateContent{
			TemplateName: v.TemplateName,
			Content:      v.Content,
		})
	}
	return result
}

// 将文件夹及内容进行压缩
func packTemplateSpace(fileName string, templateContents map[string][]TemplateContent) error {

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %s", err.Error())
	}
	defer file.Close() // nolint

	// 创建 gzip.Writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close() // nolint

	// 创建 tar.Writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close() // nolint

	for templateSpaceName, contents := range templateContents {
		// 创建 tar.Header
		header := &tar.Header{
			Name: transName(templateSpaceName) + "/",
			Mode: 0700,
		}

		// 写入 header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("failed to write tar header: %s", err.Error())
		}

		for _, v := range contents {
			// 写入 header
			header.Name = filepath.Join(transName(templateSpaceName), transName(v.TemplateName)+".yaml")
			header.Size = int64(len(v.Content))
			err = tarWriter.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("failed to write tar header: %s", err.Error())
			}

			tarWriter.Write([]byte(v.Content)) // nolint
		}
	}

	return nil
}

// 生成文件流
func getFileData(filePath string, w http.ResponseWriter) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filePath)
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	return nil
}

// 转换文件夹/文件名称
func transName(s string) string {
	if strings.Contains(s, "/") {
		s = strings.ReplaceAll(s, "/", "_")
	}
	return s
}

// 解析tgz内容
func parseTarContent(r *http.Request) ([]string, map[string][]TemplateContent, error) {
	// 存放tar解压后的文件夹名称、文件名称、文件内容
	tarContent := make(map[string][]TemplateContent, 0)
	// 文件夹名称
	var templateSpaceNames []string

	f, _, err := r.FormFile("templateFile")
	if err != nil {
		return templateSpaceNames, tarContent, fmt.Errorf("failed to get file 'attachment': %s", err.Error())
	}
	defer f.Close()
	// 创建 gzip.Reader
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return templateSpaceNames, tarContent, fmt.Errorf("Error creating gzip reader: %s", err)
	}
	defer gzipReader.Close() // nolint

	// 创建 tar.Reader
	tarReader := tar.NewReader(gzipReader)

	var maxSize int64
	// 遍历 tar 文件中的每个文件和目录, tar文件中仅允许一层文件夹
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return templateSpaceNames, tarContent, err
		}

		// 处理压缩文件
		switch header.Typeflag {
		case tar.TypeDir:
			// 校验文件夹格式，仅允许有一个文件夹层： folder/
			folders := strings.Split(header.Name, "/")
			if len(folders) != 2 || folders[1] != "" {
				return templateSpaceNames, tarContent, fmt.Errorf("invalid tar folder: %s", header.Name)
			}
			tarContent[folders[0]] = []TemplateContent{}
		case tar.TypeReg:
			if maxSize+header.Size > MaxFileSize {
				return templateSpaceNames, tarContent,
					errors.New("the extracted file is larger than 10M")
			}
			maxSize = header.Size + maxSize
			// 处理文件
			filePath := strings.Split(header.Name, "/")
			if len(filePath) != 2 {
				return templateSpaceNames, tarContent, fmt.Errorf("invalid tar filepath: %s", header.Name)
			}

			templateSpaceName := filePath[0]
			templateName := filePath[1]

			buf := make([]byte, header.Size)
			n, err := tarReader.Read(buf)
			if err != nil && err != io.EOF {
				return templateSpaceNames, tarContent, err
			}
			// 文件没有内容,不创建模板文件
			if n == 0 {
				continue
			}
			tarContent[templateSpaceName] = append(tarContent[templateSpaceName], TemplateContent{
				TemplateName: templateName,
				Content:      string(buf),
			})
		}
	}

	for k := range tarContent {
		templateSpaceNames = append(templateSpaceNames, k)
	}
	return templateSpaceNames, tarContent, nil
}
