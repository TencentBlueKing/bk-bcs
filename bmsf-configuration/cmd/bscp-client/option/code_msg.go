/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package option

type MsgType string

const (
	// 客户端输入参数个数不对
	ErrMsg_PARAM_NUM MsgType = "The number of parameters does not meet the requirements"
	// 客户端输入缺少参数
	ErrMsg_PARAM_MISS MsgType = "Missing parameters"
	// 客户端输入参数值超出范围
	ErrMsg_PARAM_RANGE MsgType = "Input parameter value is out of range"
	// 客户端输入参数组合不符合条件
	ErrMsg_PARAM_Combin MsgType = "The input parameter combination does not meet the conditions"

	// 文件不存在
	ErrMsg_FILE_NOEXIST = "file does not exist"
	// 文件读取失败
	ErrMsg_FILE_READFAIL = "file read failed"
	// 文件写入失败
	ErrMsg_FILE_WRITEFAIL = "file write failed"
	// 目录不存在
	ErrMsg_DIR_NOEXIST = "directory does not exist"

	// 未发现查询到的资源
	SucMsg_DATA_NO_FOUNT MsgType = "No resources found, query by the parameters you entered"
	// 创建资源成功
	SucMsg_DATA_Create MsgType = "Create resources successfully"
	// 更新资源成功
	SucMsg_DATA_UPDATE MsgType = "Update resources successfully"
)
