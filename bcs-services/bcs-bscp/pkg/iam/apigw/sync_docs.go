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

// Package apigw document sync gateway
package apigw

import (
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/docs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

const (
	// Name 网关名
	Name        = "bk-bscp"
	env         = "prod"
	description = "服务配置平台（bk_bscp）API 网关，包含了服务、配置项/模板、版本、分组、发布等相关资源的查询和操作接口"
)

// ReleaseSwagger 导入swagge 文档
// nolint:funlen
func ReleaseSwagger(esbOpt cc.Esb, apiGwOpt cc.ApiGateway, language, version string) error {

	// 获取需要导入的文档
	swaggerData, err := docs.Assets.ReadFile("swagger/bkapigw.swagger.json")
	if err != nil {
		return fmt.Errorf("reads and returns the content of the named file failed, err: %s", err.Error())
	}
	// 初始化网关
	gw, err := NewApiGw(esbOpt, apiGwOpt)
	if err != nil {
		return fmt.Errorf("init api gateway failed, err: %s", err.Error())
	}

	// 创建或者更新网关
	syncApiResp, err := gw.SyncApi(Name, &SyncApiReq{
		Description: description,
		Maintainers: []string{"admin"},
		IsPublic:    true,
	})
	if err != nil {
		return fmt.Errorf("create or update gateway failed, err: %s", err.Error())
	}
	if syncApiResp.Code != 0 && syncApiResp.Data.ID == 0 {
		return fmt.Errorf("create or update gateway failed, err: %s", syncApiResp.Message)
	}

	// 同步环境
	syncStageResp, err := gw.SyncStage(syncApiResp.Data.Name, &SyncStageReq{
		Name: env,
		Vars: map[string]string{},
		ProxyHttp: ProxyHttp{
			Timeout: 30,
			Upstreams: Upstreams{
				Loadbalance: "roundrobin",
				Hosts: []Host{{
					Host:   esbOpt.BscpHost,
					Weight: 100,
				}},
			},
			TransformHeaders: TransformHeaders{
				Set:    map[string]string{},
				Delete: []string{},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("sync stage failed, err: %s", err.Error())
	}
	if syncStageResp.Code != 0 && syncStageResp.Data.ID == 0 {
		return fmt.Errorf("sync stage failed, err: %s", syncStageResp.Message)
	}

	// 同步资源
	syncResourcesResp, err := gw.SyncResources(syncApiResp.Data.Name, &SyncResourcesReq{
		Content: string(swaggerData),
		Delete:  false,
	})
	if err != nil {
		return fmt.Errorf("sync resource failed, err: %s", err.Error())
	}
	if syncResourcesResp.Code != 0 {
		return fmt.Errorf("sync resource failed, err: %s", syncResourcesResp.Message)
	}

	// 导入swagger文档
	importResp, err := gw.ImportResourceDocsBySwagger(syncApiResp.Data.Name, &ImportResourceDocsBySwaggerReq{
		Language: language,
		Swagger:  string(swaggerData),
	})

	if err != nil {
		return fmt.Errorf("import resource docs by swagger failed, err: %s", err.Error())
	}
	if importResp.Code != 0 {
		return fmt.Errorf("import resource docs by swagger failed, err: %s", importResp.Message)
	}

	// 获取资源最新版本
	lrvResp, err := gw.GetLatestResourceVersion(syncApiResp.Data.Name)
	if err != nil {
		return fmt.Errorf("get latest resource version failed, err: %s", err.Error())
	}
	if lrvResp.Code != 0 {
		return fmt.Errorf("get latest resource version failed, err: %s", lrvResp.Message)
	}

	// 如果版本为空或者自定义版本和当前版本不一致时创建版本
	if version == "" || version != lrvResp.Data.Version {
		// 创建资源版本
		createResourceVersionResp, cErr := gw.CreateResourceVersion(syncApiResp.Data.Name, &CreateResourceVersionReq{
			Version: version,
			Title:   fmt.Sprintf("%s 版本", version),
			Comment: "正式环境",
		})
		if cErr != nil {
			return fmt.Errorf("create resource version failed, err: %s", cErr.Error())
		}
		if createResourceVersionResp.Code != 0 && createResourceVersionResp.Data.ID == 0 {
			return fmt.Errorf("create resource version failed, err: %s", createResourceVersionResp.Message)
		}
		version = createResourceVersionResp.Data.Version
	}

	// 发布版本
	releaseResp, err := gw.Release(syncApiResp.Data.Name, &ReleaseReq{
		Version:    version,
		StageNames: []string{env},
		Comment:    fmt.Sprintf("发布 %s 版本", version),
	})
	if err != nil {
		return fmt.Errorf("release failed, err: %s", err.Error())
	}
	if releaseResp.Code != 0 {
		return fmt.Errorf("release failed, err: %s", releaseResp.Message)
	}

	fmt.Println("swagger sync successful")

	return nil
}
