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

package sidecar

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	sfs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// checkSidecarReleaseFile check sidecar release whether right, and download file dir and content is right.
func checkSidecarReleaseFile(bizID, appID, releaseID uint32, ciMeta []*sfs.ConfigItemMetaV1) error {

	workspace := os.Getenv(constant.EnvSuitTestSidecarWorkspace)
	if len(workspace) == 0 {
		return errors.New("sidecar workspace not set")
	}

	for _, one := range ciMeta {
		ciPath := filepath.Clean(fmt.Sprintf("%s/bk-bscp/fileReleaseV1/%d/%d/%d/configItems/%s/%s", workspace, bizID,
			appID, releaseID, one.ConfigItemSpec.Path, one.ConfigItemSpec.Name))

		sha256, err := tools.FileSHA256(ciPath)
		if err != nil {
			return fmt.Errorf("%s file sha256 failed, err: %v", ciPath, err)
		}

		if sha256 != one.ContentSpec.Signature {
			return fmt.Errorf("biz: %d app: %d release: %d ci: %v file sha256 not right, expect %s, but %s", bizID,
				appID, releaseID, one.ConfigItemSpec, one.ContentSpec.Signature, sha256)
		}
	}

	return nil
}
