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

package errorx

import "fmt"

// ProjectError 项目中需要的Error
type ProjectError struct {
	code uint32
	msg  string
}

// Error xxx
func (e *ProjectError) Error() string {
	return e.msg
}

// Code xxx
func (e *ProjectError) Code() uint32 {
	return e.code
}

// NewProjectError 初始化 ProjectError
func NewProjectError(code uint32, msg string, extra string) *ProjectError {
	err := &ProjectError{code: code}
	if extra != "" {
		err.msg = fmt.Sprint(msg, ", ", extra)
	} else {
		err.msg = msg
	}
	return err
}

// NewParamErr xxx
func NewParamErr(msg string) *ProjectError {
	return NewProjectError(ParamErr, ParamErrMsg, fmt.Sprint(msg))
}

// NewInnerErr xxx
func NewInnerErr(msg string) *ProjectError {
	return NewProjectError(InnerErr, InnerErrMsg, fmt.Sprint(msg))
}

// NewDBErr xxx
func NewDBErr(msg string) *ProjectError {
	return NewProjectError(DBErr, DBErrMsg, fmt.Sprint(msg))
}

// NewClusterErr xxx
func NewClusterErr(msg string) *ProjectError {
	return NewProjectError(DBErr, ClusterMsg, fmt.Sprint(msg))
}

// NewAuthErr xxx
func NewAuthErr(msg string) *ProjectError {
	return NewProjectError(UnauthErr, UnauthErrMsg, fmt.Sprint(msg))
}

// NewIAMClientErr xxx
func NewIAMClientErr(msg string) *ProjectError {
	return NewProjectError(IAMClientErr, IAMClientErrMsg, fmt.Sprint(msg))
}

// NewIAMOPErr xxx
func NewIAMOPErr(msg string) *ProjectError {
	return NewProjectError(IAMOPErr, IAMOPErrMsg, fmt.Sprint(msg))
}

// NewRequestIAMErr xxx
func NewRequestIAMErr(msg string) *ProjectError {
	return NewProjectError(RequestIAMErr, RequestIAMErrMsg, fmt.Sprint(msg))
}

// NewNotFoundHeaderUserErr xxx
func NewNotFoundHeaderUserErr(msg string) *ProjectError {
	return NewProjectError(NotFoundHeaderUserErr, NotFoundHeaderUserErrMsg, fmt.Sprint(msg))
}

// NewRequestCMDBErr xxx
func NewRequestCMDBErr(msg string) *ProjectError {
	return NewProjectError(RequestCMDBErr, RequestCMDBErrMsg, fmt.Sprint(msg))
}

// NewNoMaintainerRoleErr xxx
func NewNoMaintainerRoleErr() *ProjectError {
	return NewProjectError(NoMaintainerRoleErr, NoMaintainerRoleErrMsg, "")
}

// NewRequestBCSCCErr xxx
func NewRequestBCSCCErr(msg string) *ProjectError {
	return NewProjectError(RequestBCSCCErr, RequestBCSCCErrMsg, fmt.Sprint(msg))
}

// NewRequestBKSSMErr xxx
func NewRequestBKSSMErr(msg string) *ProjectError {
	return NewProjectError(RequestBKSSMErr, RequestBKSSMMsg, fmt.Sprint(msg))
}

// NewRequestITSMErr xxx
func NewRequestITSMErr(msg string) *ProjectError {
	return NewProjectError(RequestITSMErr, RequestITSMErrMsg, fmt.Sprint(msg))
}

// NewRequestBkMonitorErr xxx
func NewRequestBkMonitorErr(msg string) *ProjectError {
	return NewProjectError(RequestBkMonitorErr, RequestBkMonitorErrMsg, fmt.Sprint(msg))
}

// NewReadableErr return user-friendly error
func NewReadableErr(code uint32, msg string) *ProjectError {
	return NewProjectError(code, msg, "")
}

// NewBuildTaskErr build task error
func NewBuildTaskErr(msg string) *ProjectError {
	return NewProjectError(RequestTaskErr, RequestTaskErrMsg, fmt.Sprint(msg))
}

// NewCheckQuotaStatusErr check quota status error
func NewCheckQuotaStatusErr(msg string) *ProjectError {
	return NewProjectError(RequestCheckQuotaStatusErr, RequestQuotaStatusErrMsg, fmt.Sprint(msg))
}
