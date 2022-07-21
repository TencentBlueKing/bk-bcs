/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errorx

import "fmt"

// ProjectError 项目中需要的Error
type ProjectError struct {
	code uint32
	msg  string
}

// Error ...
func (e *ProjectError) Error() string {
	return e.msg
}

// Code ...
func (e *ProjectError) Code() uint32 {
	return e.code
}

// NewProjectError 初始化 ProjectError
func NewProjectError(code uint32, msg string, extra ...interface{}) *ProjectError {
	return &ProjectError{code: code, msg: fmt.Sprintf(msg, extra...)}
}

// NewParamErr ...
func NewParamErr(msg ...interface{}) *ProjectError {
	return NewProjectError(ParamErr, ParamErrMsg, msg...)
}

// NewInnerErr ...
func NewInnerErr(msg ...interface{}) *ProjectError {
	return NewProjectError(InnerErr, InnerErrMsg, msg...)
}

// NewDBErr ...
func NewDBErr(msg ...interface{}) *ProjectError {
	return NewProjectError(DBErr, DBErrMsg, msg...)
}

// NewAuthErr ...
func NewAuthErr(msg ...interface{}) *ProjectError {
	return NewProjectError(UnauthErr, UnauthErrMsg, msg...)
}

// NewIAMClientErr ...
func NewIAMClientErr(msg ...interface{}) *ProjectError {
	return NewProjectError(IAMClientErr, IAMClientErrMsg, msg...)
}

// NewIAMOPErr ...
func NewIAMOPErr(msg ...interface{}) *ProjectError {
	return NewProjectError(IAMOPErr, IAMOPErrMsg, msg...)
}

// NewRequestIAMErr ...
func NewRequestIAMErr(msg ...interface{}) *ProjectError {
	return NewProjectError(RequestIAMErr, RequestIAMErrMsg, msg...)
}

// NotFoundHeaderUserErr ...
func NewNotFoundHeaderUserErr(msg ...interface{}) *ProjectError {
	return NewProjectError(NotFoundHeaderUserErr, NotFoundHeaderUserErrMsg, msg...)
}

// NewRequestCMDBErr ...
func NewRequestCMDBErr(msg ...interface{}) *ProjectError {
	return NewProjectError(RequestCMDBErr, RequestCMDBErrMsg, msg...)
}

// NewNoMaintainerRoleErr ...
func NewNoMaintainerRoleErr(msg ...interface{}) *ProjectError {
	return NewProjectError(NoMaintainerRoleErr, NoMaintainerRoleErrMsg, msg...)
}

// NewRequestBCSCCErr ...
func NewRequestBCSCCErr(msg ...interface{}) *ProjectError {
	return NewProjectError(RequestCMDBErr, RequestBCSCCErrMsg, msg...)
}
