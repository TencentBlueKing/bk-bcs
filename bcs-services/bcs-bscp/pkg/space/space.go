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

// Package space provides bscp space manager.
package space

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"

	esbcli "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/client"
)

// Type 空间类型
type Type struct {
	ID   string
	Name string
}

var (
	// BCS 项目类型
	BCS = Type{ID: "bcs", Name: "容器项目"}
	// BK_CMDB cmdb 业务类型
	BK_CMDB = Type{ID: "bkcmdb", Name: "业务"}
)

// Status 空间状态, 预留
type Status string

const (
	// SpaceNormal 正常状态
	SpaceNormal Status = "normal"
)

// Space 空间
type Space struct {
	SpaceId       string
	SpaceName     string
	SpaceTypeID   string
	SpaceTypeName string
	SpaceUid      string
}

// Manager Space定时拉取
type Manager struct {
	mtx         sync.Mutex
	client      esbcli.Client
	cachedSpace []*Space
}

// NewSpaceMgr 新增Space定时拉取, 注: 每个实例一个 goroutine
func NewSpaceMgr(ctx context.Context, client esbcli.Client) (*Manager, error) {
	mgr := &Manager{client: client}

	initCtx, initCancel := context.WithTimeout(ctx, time.Second*10)
	defer initCancel()

	// 启动初始化拉一次
	if err := mgr.fetchAllSpace(initCtx); err != nil {
		return nil, err
	}

	// 定期拉取
	mgr.run(ctx)

	return mgr, nil
}

// run 定时刷新全量业务信息
func (s *Manager) run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.fetchAllSpace(ctx); err != nil {
					klog.ErrorS(err, "fetch all space failed")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// AllSpaces 返回全量业务
func (s *Manager) AllSpaces() []*Space {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.cachedSpace
}

// GetSpaceByUID 按id查询业务
func (s *Manager) GetSpaceByUID(uid string) (*Space, error) {
	for _, v := range s.AllSpaces() {
		if v.SpaceId == uid {
			return v, nil
		}
	}
	return nil, fmt.Errorf("space %s not found", uid)
}

// QuerySpace 按uid批量查询业务
func (s *Manager) QuerySpace(spaceUidList []string) ([]*Space, error) {
	spaceList := []*Space{}
	spaceUidMap := map[string]struct{}{}

	for _, uid := range spaceUidList {
		spaceUidMap[uid] = struct{}{}
	}
	for _, v := range s.AllSpaces() {
		if _, ok := spaceUidMap[v.SpaceId]; ok {
			spaceList = append(spaceList, v)
		}
	}
	return spaceList, nil
}

// fetchAllSpace 获取全量业务列表
func (s *Manager) fetchAllSpace(ctx context.Context) error {
	bizList, err := s.client.Cmdb().ListAllBusiness(ctx)
	if err != nil {
		return err
	}

	if len(bizList.Info) == 0 {
		return fmt.Errorf("biz list is empty")
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

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.cachedSpace = spaceList

	klog.InfoS("fetch all space done", "biz_count", len(s.cachedSpace))
	return nil
}

// buildSpaceMap 分解
func buildSpaceMap(spaceUidList []string) (map[string][]string, error) {
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
func BuildSpaceUid(t Type, id string) string {
	return fmt.Sprintf("%s__%s", t.ID, id)
}
