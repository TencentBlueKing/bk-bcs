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
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// ExtractTemplateVariables extract template variables.
func (s *Service) ExtractTemplateVariables(ctx context.Context, req *pbds.ExtractTemplateVariablesReq) (
	*pbds.ExtractTemplateVariablesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{All: true}
	details, _, err := s.dao.AppTemplateBinding().List(kt, req.BizId, req.AppId, opt)
	if err != nil {
		logs.Errorf("extract template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// so far, no any template config item exists for the app
	if len(details) == 0 {
		return &pbds.ExtractTemplateVariablesResp{
			Details: []string{},
		}, nil
	}

	// get template revision details
	tmplRevisions, err := s.dao.TemplateRevision().
		ListByIDs(kt, details[0].Spec.TemplateRevisionIDs)
	if err != nil {
		logs.Errorf("extract template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplRevisions) == 0 {
		return &pbds.ExtractTemplateVariablesResp{
			Details: []string{},
		}, nil
	}

	tmplRevisionMap := make(map[uint32]*table.TemplateRevision, len(tmplRevisions))
	for _, r := range tmplRevisions {
		tmplRevisionMap[r.ID] = r
	}

	// download template config item content
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
		logs.Errorf("extract template variables failed, err: %v, rid: %s", hitError, kt.Rid)
		return nil, hitError
	}

	// merge all template content
	allContent := bytes.Join(contents, []byte(" "))
	// extract all template variables
	variables := s.tmplProc.ExtractVariables(allContent)

	return &pbds.ExtractTemplateVariablesResp{
		Details: variables,
	}, nil
}
