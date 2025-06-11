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

// Package helm helm
package helm

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpclient"
)

// UploadChart upload chart
func UploadChart(ctx context.Context, file *chart.Chart, projectCode, version string, force bool) error {
	// 创建临时目录
	tmp, err := os.MkdirTemp("", "helm-push-")
	if err != nil {
		return fmt.Errorf("create temporary directory error, %s", err.Error())
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			log.Error(ctx, "failed to remove temporary directory, %s: %s", path, err.Error())
		}
	}(tmp)

	// 生成 chart
	filename, err := chartutil.Save(file, tmp)
	if err != nil {
		return fmt.Errorf("failed to save chart, %s", err.Error())
	}

	// 上传
	url := fmt.Sprintf("%s/bcsapi/v4/helmmanager/api/v1/projects/%s/repos/%s/charts/upload?version=%s&force=%t",
		config.G.BCSAPIGW.Host, projectCode, projectCode, version, force)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeaders(contextx.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		SetFile("chart", filename).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to upload chart, %s", err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("http code %d != 200", resp.StatusCode())
	}
	return nil
}
