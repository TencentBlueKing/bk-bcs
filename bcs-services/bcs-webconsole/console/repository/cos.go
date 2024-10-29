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
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

type cosStorage struct {
	client *cos.Client
}

// UploadFile upload file to cos
func (c *cosStorage) UploadFile(ctx context.Context, localFile, filePath string) error {
	f, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("open local file %s failed: %v", localFile, err)
	}
	_, err = c.client.Object.Put(ctx, filePath, f, nil)
	if err != nil {
		return fmt.Errorf("upload file failed: %v", err)
	}

	return nil
}

// UploadFileByReader upload file to cos by Reader
func (c *cosStorage) UploadFileByReader(ctx context.Context, r io.Reader, filePath string) error {
	_, err := c.client.Object.Put(ctx, filePath, r, nil)
	if err != nil {
		return fmt.Errorf("upload file failed: %v", err)
	}

	return nil
}

// ListFile list current folder files
func (c *cosStorage) ListFile(ctx context.Context, folderName string) ([]string, error) {
	var marker string
	folderName = strings.Trim(folderName, "/")
	folderName += "/"
	opt := &cos.BucketGetOptions{
		Prefix:    folderName, // 表示要查询的文件夹
		Delimiter: "/",        // 表示分隔符,设置为/表示列出当前目录下的 object, 设置为空表示列出所有的 object(包括子目录文件)
		MaxKeys:   200,        // 设置最大遍历出多少个对象, 一次 listobject 最大支持1000
	}

	files := make([]string, 0)
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := c.client.Bucket.Get(ctx, opt)
		if err != nil {
			return files, fmt.Errorf("list file failed: %v", err)
		}
		if len(v.Contents) == 0 {
			return files, fmt.Errorf("folder %s is not exit", folderName)
		}
		for _, content := range v.Contents {
			fn := strings.TrimPrefix(content.Key, folderName)
			files = append(files, fn)
		}
		isTruncated = v.IsTruncated // 是否还有数据
		marker = v.NextMarker       // 设置下次请求的起始 key
	}
	return files, nil
}

// IsExist 是否存在
func (c *cosStorage) IsExist(ctx context.Context, filePath string) (bool, error) {
	return c.client.Object.IsExist(ctx, filePath)
}

// ListFolders list current folder folders
func (c *cosStorage) ListFolders(ctx context.Context, folderName string) ([]string, error) {
	var marker string
	folderName = strings.Trim(folderName, "/")
	folderName += "/"

	// cos 规范, 根目录需为空
	if folderName == "/" {
		folderName = ""
	}

	opt := &cos.BucketGetOptions{
		Prefix:    folderName, // 表示要查询的文件夹
		Delimiter: "/",        // 表示分隔符,设置为/表示列出当前目录下的 object, 设置为空表示列出所有的 object(包括子目录文件)
		MaxKeys:   200,        // 设置最大遍历出多少个对象, 一次 listobject 最大支持1000
	}

	folders := make([]string, 0)
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := c.client.Bucket.Get(ctx, opt)
		if err != nil {
			return folders, fmt.Errorf("list file failed: %v", err)
		}

		if len(v.Contents) == 0 && len(v.CommonPrefixes) == 0 {
			return folders, fmt.Errorf("folder %s is not exit", folderName)
		}
		// common prefix 表示表示被 delimiter 截断的路径, 如 delimter 设置为/, common prefix 则表示所有子目录的路径
		// nolint
		for _, commonPrefixe := range v.CommonPrefixes {
			folders = append(folders, commonPrefixe)
		}
		isTruncated = v.IsTruncated // 是否还有数据
		marker = v.NextMarker       // 设置下次请求的起始 key
	}
	return folders, nil
}

// DeleteFolders delete folder from cos
func (c *cosStorage) DeleteFolders(ctx context.Context, folderName string) error {
	folderName = strings.Trim(folderName, "/")
	folderName += "/"

	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:  folderName, // 表示要查询的文件夹
		MaxKeys: 1000,       // 设置最大遍历出多少个对象, 一次 listobject 最大支持1000
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := c.client.Bucket.Get(ctx, opt)
		if err != nil {
			return err
		}
		for _, content := range v.Contents {
			_, err = c.client.Object.Delete(ctx, content.Key)
			if err != nil {
				return err
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return nil
}

// DownloadFile download file from cos
func (c *cosStorage) DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	resp, err := c.client.Object.Get(ctx, filePath, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func newCosStorage() (Provider, error) {
	rawURL := fmt.Sprintf("https://%s.%s", config.G.Repository.Cos.BucketName,
		config.G.Repository.Cos.Endpoint)
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	bucketUrl := &cos.BaseURL{BucketURL: u}
	cli := cos.NewClient(bucketUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.G.Repository.Cos.SecretID,
			SecretKey: config.G.Repository.Cos.SecretKey,
		},
	})
	c := &cosStorage{
		client: cli,
	}
	return c, nil
}
