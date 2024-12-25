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

// Package common xxx
package common

import "helm.sh/helm/v3/pkg/release"

const (
	// ServiceDomain domain name for service
	ServiceDomain = "helmmanager.bkbcs.tencent.com"

	// MicroMetaKeyHTTPPort http port in micro-service meta
	MicroMetaKeyHTTPPort = "httpport"

	// TimeFormat time format YYYY-mm-dd HH:MM:SS
	TimeFormat = "2006-01-02 15:04:05"

	// PublicRepoName public repo name
	PublicRepoName = "public-repo"
	// PublicRepoDisplayName public repo display name
	PublicRepoDisplayName = "公共仓库"
	// ProjectRepoDefaultDisplayName default display name
	ProjectRepoDefaultDisplayName = "项目仓库"
	// PersonalRepoDefaultDisplayName default display name
	PersonalRepoDefaultDisplayName = "个人仓库"
)

// ReleaseStatus
const (
	// ReleaseStatusInstallFailed xxx
	ReleaseStatusInstallFailed release.Status = "failed-install"
	// ReleaseStatusUpgradeFailed xxx
	ReleaseStatusUpgradeFailed release.Status = "failed-upgrade"
	// ReleaseStatusRollbackFailed xxx
	ReleaseStatusRollbackFailed release.Status = "failed-rollback"
	// ReleaseStatusUninstallFailed xxx
	ReleaseStatusUninstallFailed release.Status = "failed-uninstall"
)

const (
	// LangCookieName 语言版本 Cookie 名称
	LangCookieName = "blueking_language"
)

// shared cluster
const (
	// AnnotationKeyProjectCode namespace 的 projectcode 注解 key 默认值
	AnnotationKeyProjectCode = "io.tencent.bcs.projectcode"
)
