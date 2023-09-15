/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// getAllAppCIs get all template and non-template config items for the app which can be rendered
func (s *Service) getAllAppCIs(kt *kit.Kit) ([]*table.TemplateRevision, []*pbci.ConfigItem, error) {
	tmplRevisions, err := s.getAppTmplRevisions(kt)
	if err != nil {
		logs.Errorf("extract app template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	tmplRevisions = filterSizeForTmplRevisions(tmplRevisions)

	cis, err := s.getAppConfigItems(kt)
	if err != nil {
		logs.Errorf("get app's all config items failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	cis = filterSizeForConfigItems(cis)

	return tmplRevisions, cis, nil
}

// getAppTmplRevisions get app template revision details
func (s *Service) getAppTmplRevisions(kt *kit.Kit) ([]*table.TemplateRevision, error) {
	opt := &types.BasePage{All: true}
	details, _, err := s.dao.AppTemplateBinding().List(kt, kt.BizID, kt.AppID, opt)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// so far, no any template config item exists for the app
	if len(details) == 0 {
		return []*table.TemplateRevision{}, nil
	}

	// get template revision details
	tmplRevisions, err := s.dao.TemplateRevision().
		ListByIDs(kt, details[0].Spec.TemplateRevisionIDs)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return tmplRevisions, nil
}

// getAppConfigItems get app config item details
func (s *Service) getAppConfigItems(kt *kit.Kit) ([]*pbci.ConfigItem, error) {
	req := &pbds.ListConfigItemsReq{
		BizId:      kt.BizID,
		AppId:      kt.AppID,
		All:        true,
		WithStatus: false,
	}
	resp, err := s.ListConfigItems(kt.RpcCtx(), req)
	if err != nil {
		logs.Errorf("list all config items failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// so far, no any config item exists for the app
	if len(resp.Details) == 0 {
		return []*pbci.ConfigItem{}, nil
	}

	return resp.Details, nil
}

// filterSizeForTmplRevisions get template config items which can be rendered
func filterSizeForTmplRevisions(tmplRevisions []*table.TemplateRevision) []*table.TemplateRevision {
	rs := make([]*table.TemplateRevision, 0)
	for _, r := range tmplRevisions {
		if r.Spec.ContentSpec.ByteSize <= constant.MaxRenderBytes {
			rs = append(rs, r)
		}
	}
	return rs
}

// filterSizeForConfigItems get non-template config items which can be rendered
func filterSizeForConfigItems(cis []*pbci.ConfigItem) []*pbci.ConfigItem {
	rs := make([]*pbci.ConfigItem, 0)
	for _, ci := range cis {
		if ci.CommitSpec.Content.ByteSize <= constant.MaxRenderBytes {
			rs = append(rs, ci)
		}
	}
	return rs
}

// downloadTmplContent download template config item content from repo.
// the order of elements in slice contents and slice tmplRevisions is consistent
func (s *Service) downloadTmplContent(kt *kit.Kit, tmplRevisions []*table.TemplateRevision) ([][]byte, error) {
	if len(tmplRevisions) == 0 {
		return [][]byte{}, nil
	}

	contents := make([][]byte, len(tmplRevisions))
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, r := range tmplRevisions {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, r *table.TemplateRevision) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			k := kt.GetKitForRepoTmpl(r.Attachment.TemplateSpaceID)
			body, _, err := s.repo.Download(k, r.Spec.ContentSpec.Signature)
			if err != nil {
				hitError = fmt.Errorf("download template config content from repo failed, "+
					"template id: %d, name: %s, path: %s, error: %v",
					r.Attachment.TemplateID, r.Spec.Name, r.Spec.Path, err)
				return
			}
			content, err := io.ReadAll(body)
			if err != nil {
				hitError = fmt.Errorf("read template config content from body failed, "+
					"template id: %d, name: %s, path: %s, error: %v",
					r.Attachment.TemplateID, r.Spec.Name, r.Spec.Path, err)
				return
			}

			contents[idx] = content
		}(idx, r)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("download template config content failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	return contents, nil
}

// downloadCIContent download normal config item content from repo.
// the order of elements in slice contents and slice tmplRevisions is consistent
func (s *Service) downloadCIContent(kt *kit.Kit, cis []*pbci.ConfigItem) ([][]byte, error) {
	if len(cis) == 0 {
		return [][]byte{}, nil
	}

	contents := make([][]byte, len(cis))
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for idx, c := range cis {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, c *pbci.ConfigItem) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			k := kt.GetKitForRepoCfg()
			body, _, err := s.repo.Download(k, c.CommitSpec.Content.Signature)
			if err != nil {
				hitError = fmt.Errorf("download config item content from repo failed, "+
					"config item id: %d, name: %s, path: %s, error: %v",
					c.Id, c.Spec.Name, c.Spec.Path, err)
				return
			}
			content, err := io.ReadAll(body)
			if err != nil {
				hitError = fmt.Errorf("read config item content from body failed, "+
					"config item id: %d, name: %s, path: %s, error: %v",
					c.Id, c.Spec.Name, c.Spec.Path, err)
				return
			}

			contents[idx] = content
		}(idx, c)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("download config item content failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	return contents, nil
}

// uploadRenderedTmplContent upload rendered template config item content to repo.
func (s *Service) uploadRenderedTmplContent(kt *kit.Kit, renderedContentMap map[uint32][]byte,
	signatureMap map[uint32]string, revisionMap map[uint32]*table.TemplateRevision) error {
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for revisionID := range renderedContentMap {
		wg.Add(1)

		pipe <- struct{}{}
		go func(revisionID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			r := revisionMap[revisionID]
			k := kt.GetKitForRepoCfg()
			_, err := s.repo.Upload(k, signatureMap[revisionID], bytes.NewReader(renderedContentMap[revisionID]))
			if err != nil {
				hitError = fmt.Errorf("upload rendered template config content to repo failed, "+
					"template id: %d, name: %s, path: %s, error: %v",
					r.Attachment.TemplateID, r.Spec.Name, r.Spec.Path, err)
				return
			}
		}(revisionID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("upload rendered template config content failed, err: %v, rid: %s", hitError, kt.Rid)
		return hitError
	}

	return nil
}

// uploadRenderedCIContent upload rendered config item content to repo.
func (s *Service) uploadRenderedCIContent(kt *kit.Kit, renderedContentMap map[uint32][]byte,
	signatureMap map[uint32]string, ciMap map[uint32]*pbci.ConfigItem) error {
	var hitError error
	pipe := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for configItemID := range renderedContentMap {
		wg.Add(1)

		pipe <- struct{}{}
		go func(configItemID uint32) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			ci := ciMap[configItemID]
			k := kt.GetKitForRepoCfg()
			_, err := s.repo.Upload(k, signatureMap[configItemID], bytes.NewReader(renderedContentMap[configItemID]))
			if err != nil {
				hitError = fmt.Errorf("upload rendered config item content to repo failed, "+
					"config item id: %d, name: %s, path: %s, error: %v",
					ci.Id, ci.Spec.Name, ci.Spec.Path, err)
				return
			}
		}(configItemID)
	}
	wg.Wait()

	if hitError != nil {
		logs.Errorf("upload rendered config item content failed, err: %v, rid: %s", hitError, kt.Rid)
		return hitError
	}

	return nil
}
