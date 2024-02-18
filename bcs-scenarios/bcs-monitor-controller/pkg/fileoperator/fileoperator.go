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
 *
 */

// Package fileoperator xxx
package fileoperator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

// FileOperator do file operation
type FileOperator struct {
	cli    client.Client
	loader Loader
}

// NewFileOperator return new file operator
func NewFileOperator(client client.Client) *FileOperator {
	return &FileOperator{
		cli:    client,
		loader: Loader{client: client},
	}
}

// Compress turns obj to yaml and compress it, return outputPath and error
func (f *FileOperator) Compress(objList ...interface{}) (string, error) {
	for _, obj := range objList {
		if err := f.validateType(obj); err != nil {
			blog.Errorf("obj validate failed, err: %s", err.Error())
			return "", err
		}
	}

	basePath, err := os.MkdirTemp("", "bcsmonitorctrl")
	if err != nil {
		blog.Errorf("mkdir temp dir failed, err: %s", err.Error())
		return "", err
	}
	defer os.RemoveAll(basePath)

	for _, obj := range objList {
		if err = f.createDirectoriesAndFile(obj, basePath); err != nil {
			blog.Errorf("create directory or yaml filed failed, err: %s", err.Error())
			return "", err
		}
	}

	outputPath := basePath + ".tar.gz"
	if err = f.compressFolder(basePath, outputPath); err != nil {
		blog.Errorf("compress folder failed, basePath: %s, outputPath: %s, err: %s", basePath, outputPath, err.Error())
		return "", err
	}
	return outputPath, nil
}

func (f *FileOperator) createDirectoriesAndFile(obj interface{}, basePath string) error {
	subPath := f.getNameAndSubPath(obj)
	dest := filepath.Join(basePath, subPath)
	if err := os.MkdirAll(dest, 0700); err != nil {
		blog.Errorf("mkdir failed, err: %s", err.Error())
		return err
	}
	switch obj.(type) {
	case *v1.NoticeGroup:
		ng := obj.(*v1.NoticeGroup)
		for _, group := range ng.Spec.Groups {
			yamlData, err := yaml.Marshal(group)
			if err != nil {
				blog.Errorf("transfer yaml failed, notice group: %s/%s, err: %s", ng.Namespace, ng.Name, err.Error())
				return err
			}

			fileName := group.Name + ".yaml"
			filePath := filepath.Join(dest, fileName)
			if inErr := ioutil.WriteFile(filePath, yamlData, 0644); inErr != nil {
				blog.Errorf("write file '%s' failed, err: %s", fileName, inErr.Error())
				return inErr
			}
		}
		return nil
	case *v1.MonitorRule:
		mr := obj.(*v1.MonitorRule)
		for _, rule := range mr.Spec.Rules {
			yamlData, err := yaml.Marshal(rule)
			if err != nil {
				blog.Errorf("transfer yaml failed, monitor rule: %s/%s, err: %s", mr.GetNamespace(), mr.GetName(),
					err.Error())
				return err
			}

			fileName := rule.Name + ".yaml"
			filePath := filepath.Join(dest, fileName)
			if inErr := ioutil.WriteFile(filePath, yamlData, 0644); inErr != nil {
				blog.Errorf("write file '%s' failed, err: %s", fileName, inErr.Error())
				return inErr
			}
		}
		return nil
	case *v1.Panel:
		panel := obj.(*v1.Panel)
		for _, board := range panel.Spec.DashBoard {
			var data []byte
			var err error

			ns := panel.Namespace
			if board.ConfigMapNs != "" {
				ns = board.ConfigMapNs
			}
			data, err = f.loader.LoadFileFromConfigMap(ns, board.ConfigMap)
			if err != nil {
				blog.Errorf("load data from board failed, board: %+v, err: %s", board, err.Error())
				return err
			}

			if err = os.MkdirAll(filepath.Join(dest, panel.Spec.Scenario), 0700); err != nil {
				blog.Errorf("mkdir failed, err: %s", err.Error())
				return err
			}

			fileName := board.Board + ".json"
			filePath := filepath.Join(dest, panel.Spec.Scenario, fileName)

			if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
				blog.Errorf("write file to path: %s failed, board %+v, err: %s", filePath, board, err.Error())
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("invalid type, obj: %+v", obj)
}

func (f *FileOperator) compressFolder(basePath, output string) error {
	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true

	items, err := ioutil.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", basePath, err)
	}

	var pathsToArchive []string
	for _, item := range items {
		itemPath := filepath.Join(basePath, item.Name())
		pathsToArchive = append(pathsToArchive, itemPath)
	}

	return tar.Archive(pathsToArchive, output)
}

func (f *FileOperator) validateType(i interface{}) error {
	switch i.(type) {
	case *v1.NoticeGroup, *v1.MonitorRule, *v1.Panel:
		return nil
	default:
		return fmt.Errorf("internal error: unknown type struct")
	}
}

// getNameAndSubPath 根据类型决定生成文件的目录，需要匹配蓝鲸监控的目录要求
// return fileName, filePath
// configs
// - config/notice
// - config/rule
// - config/grafana
// - config/action
func (f *FileOperator) getNameAndSubPath(i interface{}) string {
	if _, ok := i.(*v1.NoticeGroup); ok {
		return "configs/notice"
	}

	if _, ok := i.(*v1.MonitorRule); ok {
		return "configs/rule"
	}

	if _, ok := i.(*v1.Panel); ok {
		return "configs/grafana"
	}

	return ""
}
