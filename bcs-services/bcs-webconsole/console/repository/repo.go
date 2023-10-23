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

// Package repository xxx
package repository

import (
	"context"
	"fmt"
	"io"
)

// Provider 对象存储接口
type Provider interface {
	UploadFile(ctx context.Context, localFile, filePath string) error
	ListFile(ctx context.Context, folderName string) ([]string, error)
	ListFolders(ctx context.Context, folderName string) ([]string, error)
	IsExist(ctx context.Context, filePath string) (bool, error)
	DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error)
}

// NewProvider init provider factory by storage type
func NewProvider(providerType string) (Provider, error) {
	switch providerType {
	case "cos":
		return newCosStorage()
	case "bkrepo":
		return newBkRepoStorage()
	default:
		return nil, fmt.Errorf("%s is not supported", providerType)
	}
}
