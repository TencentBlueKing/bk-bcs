/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package errf

import "errors"

var (
	// ErrCPSInconsistent is error when the number of cps to be queried is inconsistent with the
	// number of cps found. because the application cpsID is obtained through db, then the cps
	// details are queried, data inconsistencies may occur.
	ErrCPSInconsistent = errors.New("current published strategies are inconsistent")

	// ErrCPSNotFound is error when the current published strategies not found in db.
	ErrCPSNotFound = errors.New("current published strategies not found")

	// ErrPermissionDenied
	ErrPermissionDenied = errors.New("no permission")

	// ErrCredentialInvalid is error when the credential not found in db.
	ErrCredentialInvalid = errors.New("invalid credential")

	// ErrAppInstanceNotMatchedRelease
	ErrAppInstanceNotMatchedRelease = errors.New("this app instance can not match any release")
)
