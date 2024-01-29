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

package repo

import (
	"errors"
)

const (
	// EnvNameGitRepoURL env name for default git URL
	EnvNameGitRepoURL = "GIT_URL"
	// EnvNameGitUserName env name for default git username
	EnvNameGitUserName = "GIT_USERNAME"
	// EnvNameGitSecret env name for default git secret
	EnvNameGitSecret = "GIT_SECRET"

	// RepoKeyDefault default key
	RepoKeyDefault = "DEFAULT"
)

var errStop = errors.New("stop")

// Repo repo interface
type Repo interface {
	Pull() error
	Clone() error
	Reload() ([]string, error)

	GetURL() string
	GetDirectory() string
	GetRepoKey() string
}
