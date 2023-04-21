package table

import (
	"errors"

	"bscp.io/pkg/criteria/enumor"
)

// CredentialColumns defines credential's columns
var CredentialScopeColumns = mergeColumns(CredentialScopeColumnDescriptor)

// CredentialScopeColumnDescriptor is mergeColumnDescriptors
var CredentialScopeColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", CredentialScopeSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CredentialScopeAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CredentialRevisionColumnDescriptor),
)

// CredentialScope defines CredentialScope's columns
type CredentialScope struct {
	// ID is an auto-increased value, which is a unique identity of a Credential.
	ID         uint32                     `db:"id" json:"id"`
	Spec       *CredentialScopeSpec       `db:"spec" json:"spec"`
	Attachment *CredentialScopeAttachment `db:"attachment" json:"attachment"`
	Revision   *CredentialRevision        `db:"revision" json:"revision"`
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
}

// CredentialScopeSpec defines credential scope's Spec
type CredentialScopeSpec struct {
	CredentialScope string `db:"credential_scope" json:"credential_scope"`
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

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
