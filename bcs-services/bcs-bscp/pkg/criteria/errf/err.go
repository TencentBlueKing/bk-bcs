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

package errf

import (
	"errors"

	"bscp.io/pkg/kit"
)

var (
	// ErrCPSInconsistent is error when the number of cps to be queried is inconsistent with the
	// number of cps found. because the application cpsID is obtained through db, then the cps
	// details are queried, data inconsistencies may occur.
	ErrCPSInconsistent = errors.New("current published strategies are inconsistent")

	// ErrCPSNotFound is error when the current published strategies not found in db.
	ErrCPSNotFound = errors.New("current published strategies not found")

	// ErrPermissionDenied is error when the user has no permission to do this operation.
	ErrPermissionDenied = errors.New("no permission")

	// ErrCredentialInvalid is error when the credential not found in db.
	ErrCredentialInvalid = errors.New("invalid credential")

	// ErrAppInstanceNotMatchedRelease is error when the app instance can not match any release.
	ErrAppInstanceNotMatchedRelease = errors.New("this app instance can not match any release")

	// ErrFileContentNotFound is error when the file content not found in file provider.
	ErrFileContentNotFound = errors.New("file content not found")
)

var (
	// ErrDBOpsFailedF is for db operation failed
	ErrDBOpsFailedF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, Internal, "db operation failed")
	}
	// ErrInvalidArgF is for invalid argument
	ErrInvalidArgF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "invalid argument")
	}
	// ErrWithIDF is for id should not be set
	ErrWithIDF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "id should not be set")
	}
	// ErrNoSpecF is for spec not set
	ErrNoSpecF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "spec not set")
	}
	// ErrNoAttachmentF is for attachment not set
	ErrNoAttachmentF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "attachment not set")
	}
	// ErrNoRevisionF is for revision not set
	ErrNoRevisionF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "revision not set")
	}
)
