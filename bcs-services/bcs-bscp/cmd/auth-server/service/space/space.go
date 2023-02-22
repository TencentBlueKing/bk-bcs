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

package space

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"bscp.io/pkg/components/bcs"
	"bscp.io/pkg/components/bkcmdb"
	esbcli "bscp.io/pkg/thirdparty/esb/client"
	"bscp.io/pkg/thirdparty/esb/cmdb"
	"k8s.io/klog/v2"
)

// SpaceType 空间类型
type SpaceType struct {
	ID   string
	Name string
}

var (
	// BCS 项目类型
	BCS = SpaceType{ID: "bcs", Name: "容器项目"}
	// BK_CMDB cmdb 业务类型
	BK_CMDB = SpaceType{ID: "bkcmdb", Name: "业务"}
)

// SpaceStatus 空间状态, 预留
type SpaceStatus string

const (
	// SpaceNormal 正常状态
	SpaceNormal SpaceStatus = "normal"
)

// Space 空间
type Space struct {
	SpaceId       string
	SpaceName     string
	SpaceTypeID   string
	SpaceTypeName string
	SpaceUid      string
}

func listBKCMDB(ctx context.Context, client esbcli.Client, username string) ([]*Space, error) {
	params := &cmdb.SearchBizParams{
		Condition: map[string]string{
			"bk_biz_maintainer": username,
		},
	}

	bizList, err := bkcmdb.SearchBusiness(ctx, params)
	if err != nil {
		return nil, err
	}

	spaceList := make([]*Space, 0, len(bizList.Info))

	for _, biz := range bizList.Info {
		spaceList = append(spaceList, &Space{
			SpaceId:       strconv.FormatInt(biz.BizID, 10),
			SpaceName:     biz.BizName,
			SpaceTypeID:   BK_CMDB.ID,
			SpaceTypeName: BK_CMDB.Name,
			SpaceUid:      fmt.Sprintf("%s__%d", BK_CMDB.ID, biz.BizID),
		})
	}

	return spaceList, nil
}

func listBCSProject(ctx context.Context, username string) ([]*Space, error) {
	projects, err := bcs.ListAuthorizedProjects(ctx, username)
	if err != nil {
		return nil, err
	}
	spaceList := make([]*Space, 0, len(projects))
	for _, project := range projects {
		spaceList = append(spaceList, &Space{
			SpaceId:       project.Code,
			SpaceName:     project.Name,
			SpaceTypeID:   BCS.ID,
			SpaceTypeName: BCS.Name,
			SpaceUid:      fmt.Sprintf("%s__%s", BCS.ID, project.Code),
		})
	}
	return spaceList, nil
}

// ListSpace 并发获取cmdb的业务, bcs的项目列表
func ListSpace(ctx context.Context, client esbcli.Client, username string) ([]*Space, error) {
	var (
		spaceList = []*Space{}
		mtx       sync.Mutex
		wg        sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()

		spaces, err := listBCSProject(ctx, username)
		if err != nil {
			klog.Warningf("list bcs space failed. err: %s", err)
			return
		}
		mtx.Lock()
		defer mtx.Unlock()

		spaceList = append(spaceList, spaces...)
	}()

	go func() {
		defer wg.Done()

		spaces, err := listBKCMDB(ctx, client, username)
		if err != nil {
			klog.Warningf("list bk_cmdb space failed. err: %s", err)
			return
		}
		mtx.Lock()
		defer mtx.Unlock()

		spaceList = append(spaceList, spaces...)
	}()

	wg.Wait()
	return spaceList, nil
}
