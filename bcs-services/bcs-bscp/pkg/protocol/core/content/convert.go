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

// Package pbcontent provides content core protocol struct and convert functions.
package pbcontent

import (
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// ContentSpec convert pb ContentSpec to table ContentSpec
func (m *ContentSpec) ContentSpec() *table.ContentSpec {
	if m == nil {
		return nil
	}

	return &table.ContentSpec{
		Signature: m.Signature,
		ByteSize:  m.ByteSize,
	}
}

// PbContentSpec convert table ContentSpec to pb ContentSpec
func PbContentSpec(spec *table.ContentSpec) *ContentSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &ContentSpec{
		Signature: spec.Signature,
		ByteSize:  spec.ByteSize,
	}
}

// ReleasedContentSpec convert pb ReleasedContentSpec to table ReleasedContentSpec
func (m *ReleasedContentSpec) ReleasedContentSpec() *table.ReleasedContentSpec {
	if m == nil {
		return nil
	}

	return &table.ReleasedContentSpec{
		Signature:       m.Signature,
		ByteSize:        m.ByteSize,
		OriginSignature: m.OriginSignature,
		OriginByteSize:  m.OriginByteSize,
	}
}

// PbReleasedContentSpec convert table ReleasedContentSpec to pb ReleasedContentSpec
func PbReleasedContentSpec(spec *table.ReleasedContentSpec) *ReleasedContentSpec {
	if spec == nil {
		return nil
	}

	return &ReleasedContentSpec{
		Signature:       spec.Signature,
		ByteSize:        spec.ByteSize,
		OriginSignature: spec.OriginSignature,
		OriginByteSize:  spec.OriginByteSize,
	}
}

// ContentAttachment convert pb ContentAttachment to table ContentAttachment
func (m *ContentAttachment) ContentAttachment() *table.ContentAttachment {
	if m == nil {
		return nil
	}

	return &table.ContentAttachment{
		BizID:        m.BizId,
		AppID:        m.AppId,
		ConfigItemID: m.ConfigItemId,
	}
}

// PbContentAttachment convert table ContentAttachment to pb ContentAttachment
func PbContentAttachment(at *table.ContentAttachment) *ContentAttachment { //nolint:revive
	if at == nil {
		return nil
	}

	return &ContentAttachment{
		BizId:        at.BizID,
		AppId:        at.AppID,
		ConfigItemId: at.ConfigItemID,
	}
}

// PbContents convert table Content to pb Content
func PbContents(cs []*table.Content) []*Content {
	if cs == nil {
		return make([]*Content, 0)
	}

	result := make([]*Content, 0)
	for _, c := range cs {
		result = append(result, PbContent(c))
	}

	return result
}

// PbContent convert table Content to pb Content
func PbContent(c *table.Content) *Content {
	if c == nil {
		return nil
	}

	return &Content{
		Id:         c.ID,
		Spec:       PbContentSpec(c.Spec),
		Attachment: PbContentAttachment(c.Attachment),
		Revision:   pbbase.PbCreatedRevision(c.Revision),
	}
}

// MarshalJSONPB ContentSpec to json.
func (m *ContentSpec) MarshalJSONPB(mars *jsonpb.Marshaler) ([]byte, error) {
	return jsoni.Marshal(m)
}

// UnmarshalJSONPB json to ContentSpec.
func (m *ContentSpec) UnmarshalJSONPB(um *jsonpb.Unmarshaler, data []byte) error {
	return jsoni.Unmarshal(data, &m)
}
