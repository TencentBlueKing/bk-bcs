package pbkv

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// Kv convert pb Kv to table Kv
func (k *Kv) Kv() (*table.Kv, error) {
	if k == nil {
		return nil, nil
	}

	return &table.Kv{
		ID:         k.Id,
		Spec:       k.Spec.KvSpec(),
		Attachment: k.Attachment.KvAttachment(),
	}, nil
}

// KvSpec convert pb kv to table KvSpec
func (k *KvSpec) KvSpec() *table.KvSpec {
	if k == nil {
		return nil
	}

	return &table.KvSpec{
		Name: k.Name,
	}
}

// KvAttachment convert pb KvAttachment to table KvAttachment
func (k *KvAttachment) KvAttachment() *table.KvAttachment {
	if k == nil {
		return nil
	}

	return &table.KvAttachment{
		BizID: k.BizId,
		AppID: k.AppId,
	}
}

// PbKvs convert table kv to pb kv
func PbKvs(s []*table.Kv) []*Kv {
	if s == nil {
		return make([]*Kv, 0)
	}

	result := make([]*Kv, 0)
	for _, one := range s {
		result = append(result, PbKv(one))
	}

	return result
}

// PbKv convert table kv to pb kv
func PbKv(k *table.Kv) *Kv {
	if k == nil {
		return nil
	}

	return &Kv{
		Id:         k.ID,
		Spec:       PbKvSpec(k.Spec),
		Attachment: PbKvAttachment(k.Attachment),
		Revision:   pbbase.PbRevision(k.Revision),
	}
}

// PbKvSpec convert table KvSpec to pb KvSpec
func PbKvSpec(spec *table.KvSpec) *KvSpec {
	if spec == nil {
		return nil
	}

	return &KvSpec{
		Name: spec.Name,
	}
}

// PbKvAttachment convert table KvAttachment to pb KvAttachment
func PbKvAttachment(ka *table.KvAttachment) *KvAttachment {
	if ka == nil {
		return nil
	}

	return &KvAttachment{
		BizId: ka.BizID,
		AppId: ka.AppID,
	}
}
