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
	"io/ioutil"
	"sync"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/types"
)

// getAppTmplRevisions get app template revision details
func (s *Service) getAppTmplRevisions(kt *kit.Kit, bizID, appID uint32) ([]*table.TemplateRevision, error) {
	opt := &types.BasePage{All: true}
	details, _, err := s.dao.AppTemplateBinding().List(kt, bizID, appID, opt)
	if err != nil {
		logs.Errorf("get app template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// so far, no any template config item exists for the app
	if len(details) == 0 {
		return nil, nil
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

// downloadTmplContent download template config item content from repo.
// the order of elements in slice contents and slice tmplRevisions is consistent
func (s *Service) downloadTmplContent(kt *kit.Kit, tmplRevisions []*table.TemplateRevision) ([][]byte, error) {
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
			content, err := ioutil.ReadAll(body)
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
