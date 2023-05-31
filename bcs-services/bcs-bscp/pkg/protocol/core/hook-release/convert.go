package pbhr

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/runtime/selector"
	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

// HookReleaseSpace convert pb HookReleaseSpace to table HookReleaseSpace
func (m *HookRelease) HookReleaseSpace() (*table.HookRelease, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.HookReleaseSpec()
	if err != nil {
		return nil, err
	}

	return &table.HookRelease{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.HookReleaseAttachment(),
	}, nil
}

// HookReleaseSpec convert pb HookReleaseSpace to table HookReleaseSpace
func (m *HookReleaseSpec) HookReleaseSpec() (*table.HookReleaseSpec, error) {
	if m == nil {
		return nil, nil
	}

	return &table.HookReleaseSpec{
		Name:       m.Name,
		PublishNum: 0,
		PubState:   table.NotReleased,
		Content:    m.Content,
		Memo:       m.Memo,
	}, nil
}

// PbHookReleaseSpec convert table HookReleaseSpec to pb HookReleaseSpec
func PbHookReleaseSpec(spec *table.HookReleaseSpec) (*HookReleaseSpec, error) {
	if spec == nil {
		return nil, nil
	}

	return &HookReleaseSpec{
		Name:       spec.Name,
		Content:    spec.Content,
		PublishNum: spec.PublishNum,
		PubState:   spec.PubState.String(),
		Memo:       spec.Memo,
	}, nil
}

// HookReleaseAttachment convert pb HookReleaseAttachment to table HookReleaseAttachment
func (m *HookReleaseAttachment) HookReleaseAttachment() *table.HookReleaseAttachment {
	if m == nil {
		return nil
	}

	return &table.HookReleaseAttachment{
		BizID:  m.BizId,
		HookID: m.HookId,
	}
}

// HookReleaseSpaceAttachment convert table HookReleaseAttachment to pb HookReleaseAttachment
func HookReleaseSpaceAttachment(at *table.HookReleaseAttachment) *HookReleaseAttachment {
	if at == nil {
		return nil
	}

	return &HookReleaseAttachment{
		BizId:  at.BizID,
		HookId: at.HookID,
	}
}

// PbHookReleaseSpaces convert table HookRelease to pb HookRelease
func PbHookReleaseSpaces(s []*table.HookRelease) ([]*HookRelease, error) {
	if s == nil {
		return make([]*HookRelease, 0), nil
	}

	result := make([]*HookRelease, 0)
	for _, one := range s {
		hook, err := PbHookRelease(one)
		if err != nil {
			return nil, err
		}
		result = append(result, hook)
	}

	return result, nil
}

// PbHookRelease convert table HookRelease to pb HookRelease
func PbHookRelease(s *table.HookRelease) (*HookRelease, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbHookReleaseSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &HookRelease{
		Id:         s.ID,
		Spec:       spec,
		Attachment: HookReleaseSpaceAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}, nil
}

// UnmarshalSelector unmarshal pb struct to selector.
func UnmarshalSelector(pb *pbstruct.Struct) (*selector.Selector, error) {
	json, err := pb.MarshalJSON()
	if err != nil {
		return nil, err
	}

	s := new(selector.Selector)
	if err = s.Unmarshal(json); err != nil {
		return nil, err
	}

	return s, nil
}
