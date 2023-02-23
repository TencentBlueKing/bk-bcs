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
	"strings"
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

func listBKCMDB(ctx context.Context, client esbcli.Client, username string, bizIdList []int) ([]*Space, error) {
	var params *cmdb.SearchBizParams
	if username != "" {
		params = &cmdb.SearchBizParams{
			Condition: map[string]string{
				"bk_biz_maintainer": username,
			},
		}
	} else {
		params = &cmdb.SearchBizParams{
			BizPropertyFilter: &cmdb.QueryFilter{
				Rule: cmdb.CombinedRule{
					Condition: cmdb.ConditionAnd,
					Rules: []cmdb.Rule{
						cmdb.AtomRule{
							Field:    "bk_biz_id",
							Operator: cmdb.OperatorIn,
							Value:    bizIdList,
						},
					},
				},
			},
		}
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
			SpaceUid:      BuildSpaceUid(BK_CMDB, strconv.FormatInt(biz.BizID, 10)),
		})
	}

	return spaceList, nil
}

func listBCSProject(ctx context.Context, username string, projectCodeList []string) ([]*Space, error) {
	var (
		projects []*bcs.Project
		err      error
	)

	if username != "" {
		projects, err = bcs.ListAuthorizedProjects(ctx, username)
	} else {
		projects, err = bcs.ListProjects(ctx, projectCodeList)
	}

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
			SpaceUid:      BuildSpaceUid(BCS, project.Code),
		})
	}
	return spaceList, nil
}

// ListUserSpace 并发获取cmdb的业务, bcs的项目列表
func ListUserSpace(ctx context.Context, client esbcli.Client, username string) ([]*Space, error) {
	var (
		spaceList = []*Space{}
		mtx       sync.Mutex
		wg        sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()

		spaces, err := listBCSProject(ctx, username, []string{})
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

		spaces, err := listBKCMDB(ctx, client, username, []int{})
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

func QuerySpace(ctx context.Context, client esbcli.Client, spaceUidList []string) ([]*Space, error) {
	var (
		spaceList = []*Space{}
		mtx       sync.Mutex
		wg        sync.WaitGroup
	)

	spaceMap, err := BuildSpaceMap(spaceUidList)
	if err != nil {
		return nil, err
	}

	wg.Add(2)
	go func() {
		defer wg.Done()

		idList := spaceMap[BCS.ID]
		if len(idList) == 0 {
			return
		}

		spaces, err := listBCSProject(ctx, "", idList)
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

		idList := spaceMap[BK_CMDB.ID]
		if len(idList) == 0 {
			return
		}

		// cmdb bk_biz_id 需要 int 类型
		idIntList := make([]int, 0, len(idList))
		for _, id := range idList {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				klog.Warningf("%s not integer", id)
				continue
			}
			idIntList = append(idIntList, idInt)
		}

		spaces, err := listBKCMDB(ctx, client, "", idIntList)
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

// spaceUidList 分解
func BuildSpaceMap(spaceUidList []string) (map[string][]string, error) {
	s := map[string][]string{}
	for _, uid := range spaceUidList {
		patterns := strings.Split(uid, "__")
		if len(patterns) != 2 {
			return nil, fmt.Errorf("space_uid not valid, %s", uid)
		}
		s[patterns[0]] = append(s[patterns[0]], patterns[1])
	}
	return s, nil
}

// BuildSpaceUid 组装 space_uid
func BuildSpaceUid(t SpaceType, id string) string {
	return fmt.Sprintf("%s__%s", t.ID, id)
}
