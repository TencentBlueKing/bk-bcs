package table

import (
	"errors"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/runtime/credential"
)

// CredentialColumns defines credential's columns
var CredentialScopeColumns = mergeColumns(CredentialScopeColumnDescriptor)

// CredentialScopeColumnDescriptor is mergeColumnDescriptors
var CredentialScopeColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", CredentialScopeSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CredentialScopeAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor),
)

// CredentialScope defines CredentialScope's columns
type CredentialScope struct {
	// ID is an auto-increased value, which is a unique identity of a Credential.
	ID         uint32                     `db:"id" json:"id"`
	Spec       *CredentialScopeSpec       `db:"spec" json:"spec"`
	Attachment *CredentialScopeAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision                  `db:"revision" json:"revision"`
}

// TableName is the CredentialScope's database table name.
func (s CredentialScope) TableName() Name {
	return CredentialScopeTable
}

// ValidateCreate validate Credential is valid or not when create it.
func (s CredentialScope) ValidateCreate() error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.CredentialScope.Validate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// CredentialScopeSpecColumns defines credential scope's columns
var CredentialScopeSpecColumns = mergeColumns(CredentialScopeSpecColumnDescriptor)

// CredentialScopeSpecColumnDescriptor defines credential scope's descriptor
var CredentialScopeSpecColumnDescriptor = ColumnDescriptors{
	{Column: "credential_scope", NamedC: "credential_scope", Type: enumor.String},
	{Column: "expired_at", NamedC: "expired_at", Type: enumor.Time},
}

// CredentialScopeSpec defines credential scope's Spec
type CredentialScopeSpec struct {
	CredentialScope credential.CredentialScope `db:"credential_scope" json:"credential_scope"`
	ExpiredAt       time.Time                  `db:"expired_at" json:"expired_at"`
}

// CredentialScopeAttachmentColumnDescriptor defines credential scope's ColumnDescriptors
var CredentialScopeAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "credential_id", NamedC: "credential_id", Type: enumor.Numeric},
}

// CredentialScopeAttachment defines the credential scope attachments.
type CredentialScopeAttachment struct {
	BizID        uint32 `db:"biz_id" json:"biz_id"`
	CredentialId uint32 `db:"credential_id" json:"credential_id"`
}

// ValidateDelete credential scope validate
func (s CredentialScope) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("credential scope id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// ValidateUpdate validate Credential is valid or not when update it.
func (s CredentialScope) ValidateUpdate() error {

	if s.ID <= 0 {
		return errors.New("credential scope id should be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.CredentialScope.Validate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
