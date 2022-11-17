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

package pbinstance

import (
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/protocol/core/release"
)

// ReleasedInstanceSpec convert pb ReleasedInstanceSpec to table ReleasedInstanceSpec
func (m *ReleasedInstanceSpec) ReleasedInstanceSpec() *table.ReleasedInstanceSpec {
	if m == nil {
		return nil
	}

	return &table.ReleasedInstanceSpec{
		Uid:       m.Uid,
		ReleaseID: m.ReleaseId,
		Memo:      m.Memo,
	}
}

// PbReleasedInstanceSpec convert table ReleasedInstanceSpec to pb ReleasedInstanceSpec
func PbReleasedInstanceSpec(spec *table.ReleasedInstanceSpec) *ReleasedInstanceSpec {
	if spec == nil {
		return nil
	}

	return &ReleasedInstanceSpec{
		Uid:       spec.Uid,
		ReleaseId: spec.ReleaseID,
		Memo:      spec.Memo,
	}
}

// PbCRInstances convert table CurrentReleasedInstance to pb CurrentReleasedInstance
func PbCRInstances(cris []*table.CurrentReleasedInstance) []*CurrentReleasedInstance {
	if cris == nil {
		return make([]*CurrentReleasedInstance, 0)
	}

	result := make([]*CurrentReleasedInstance, 0)
	for _, c := range cris {
		result = append(result, PbCRInstance(c))
	}

	return result
}

// PbCRInstance convert table CurrentReleasedInstance to pb CurrentReleasedInstance
func PbCRInstance(cri *table.CurrentReleasedInstance) *CurrentReleasedInstance {
	if cri == nil {
		return nil
	}

	return &CurrentReleasedInstance{
		Id:         cri.ID,
		Spec:       PbReleasedInstanceSpec(cri.Spec),
		Attachment: pbrelease.PbReleaseAttachment(cri.Attachment),
		Revision:   pbbase.PbCreatedRevision(cri.Revision),
	}
}
