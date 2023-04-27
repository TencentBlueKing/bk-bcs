package table

import (
	"errors"
	"fmt"
	"time"

	"bscp.io/pkg/criteria/enumor"
)

// CredentialColumns defines credential's columns
var CredentialColumns = mergeColumns(CredentialColumnDescriptor)

// CredentialColumnDescriptor is Credential's column descriptors.
var CredentialColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", CredentialSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CredentialAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor),
)

// Credential defines a hook for an app to publish.
// it contains the selector to define the scope of the matched instances.
type Credential struct {
	// ID is an auto-increased value, which is a unique identity of a Credential.
	ID         uint32                `db:"id" json:"id"`
	Spec       *CredentialSpec       `db:"spec" json:"spec"`
	Attachment *CredentialAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision             `db:"revision" json:"revision"`
}

// TableName  is the Credential's database table name.
func (s Credential) TableName() Name {
	return CredentialTable
}

// ValidateCreate validate Credential is valid or not when create it.
func (s Credential) ValidateCreate() error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// CredentialSpecColumns defines CredentialSpec's columns
var CredentialSpecColumns = mergeColumns(CredentialSpecColumnDescriptor)

// CredentialSpecColumnDescriptor is CredentialSpec's column descriptors.
var CredentialSpecColumnDescriptor = ColumnDescriptors{
	{Column: "credential_type", NamedC: "credential_type", Type: enumor.String},
	{Column: "enc_credential", NamedC: "enc_credential", Type: enumor.String},
	{Column: "enc_algorithm", NamedC: "enc_algorithm", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "enable", NamedC: "enable", Type: enumor.Boolean},
	{Column: "expired_at", NamedC: "expired_at", Type: enumor.Time},
}

// CredentialSpec defines all the specifics for credential set by user.
type CredentialSpec struct {
	CredentialType CredentialType `db:"credential_type" json:"credential_type"`
	EncCredential  string         `db:"enc_credential" json:"enc_credential"`
	EncAlgorithm   string         `db:"enc_algorithm" json:"enc_algorithm"`
	Memo           string         `db:"memo" json:"memo"`
	Enable         bool           `db:"enable"  json:"enable"`
	ExpiredAt      time.Time      `db:"expired_at" json:"expired_at"`
}

const (
	// BearToken is the type default
	BearToken CredentialType = "bearToken"
)

// CredentialType is the type of credential
type CredentialType string

// Validate validate the credential type
func (s CredentialType) Validate() error {
	if s == "" {
		return nil
	}
	switch s {
	case BearToken:
	default:
		return fmt.Errorf("unsupported credential type: %s", s)
	}

	return nil
}

// String credential to string
func (s CredentialType) String() string {
	return string(s)
}

// ValidateCreate validate credential spec when it is created.
func (c CredentialSpec) ValidateCreate() error {
	if err := c.CredentialType.Validate(); err != nil {
		return err
	}
	return nil
}

// CredentialAttachment defines the credential attachments.
type CredentialAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id"`
}

// CredentialAttachmentColumnDescriptor is CredentialAttachment's column descriptors.
var CredentialAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
}

// IsEmpty test whether credential attachment is empty or not.
func (c CredentialAttachment) IsEmpty() bool {
	return c.BizID == 0
}

// Validate whether credential attachment is valid or not.
func (c CredentialAttachment) Validate() error {
	if c.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}

// CredentialRevisionColumns defines all the Revision table's columns.
var CredentialRevisionColumns = mergeColumns(CredentialRevisionColumnDescriptor)

// CredentialRevisionColumnDescriptor is Revision's column descriptors.
var CredentialRevisionColumnDescriptor = ColumnDescriptors{
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	{Column: "expired_at", NamedC: "expired_at", Type: enumor.Time},
}

// ValidateDelete validate the credential's info when delete it.
func (s Credential) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("credential id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// ValidateUpdate validate Credential is valid or not when update it.
func (s Credential) ValidateUpdate() error {

	if s.ID <= 0 {
		return errors.New("id should be set")
	}

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}
