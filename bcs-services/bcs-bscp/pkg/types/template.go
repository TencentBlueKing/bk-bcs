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

package types

// FileInfo File info
type FileInfo struct {
	Name     string
	Path     string
	FileType string
	Sign     string
	ByteSize uint64
}

// TemplateItem Template info
type TemplateItem struct {
	Id        uint32 `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	FileType  string `json:"file_type"`
	FileMode  string `json:"file_mode"`
	Memo      string `json:"memo"`
	Privilege string `json:"privilege"`
	User      string `json:"user"`
	UserGroup string `json:"user_group"`
	Sign      string `json:"sign"`
	ByteSize  uint64 `json:"byte_size"`
}

// TemplatesImportResp Import template return
type TemplatesImportResp struct {
	Exist    []*TemplateItem `json:"exist"`
	NonExist []*TemplateItem `json:"non_exist"`
	Msg      string          `json:"msg"`
}

// UploadTask 上传任务结构体
type UploadTask struct {
	File FileInfo
	Err  error
}
