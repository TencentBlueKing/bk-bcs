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

package sfs

import (
	"fmt"
)

// FailedReason defines the reasons for failure
type FailedReason uint32

const (
	// PreHookFailed represents failure in pre-hook execution
	PreHookFailed FailedReason = 1 << iota
	// PostHookFailed represents failure in post-hook execution
	PostHookFailed
	// DownloadFailed represents failure in downloading
	DownloadFailed
	// TokenFailed represents failure due to token issues
	TokenFailed
	// SdkVersionIsTooLowFailed represents failure due to SDK version being too low
	SdkVersionIsTooLowFailed
	// AppMetaFailed represents failure in app metadata
	AppMetaFailed
	// DeleteOldFilesFailed represents failure in deleting old files
	DeleteOldFilesFailed
	// UpdateMetadataFailed represents failure in updating metadata
	UpdateMetadataFailed
)

// Validate the failed reason is valid or not
func (fr FailedReason) Validate() error {
	switch fr {
	case PreHookFailed, PostHookFailed, DownloadFailed, TokenFailed,
		SdkVersionIsTooLowFailed, AppMetaFailed, DeleteOldFilesFailed,
		UpdateMetadataFailed:
		return nil
	default:
		return fmt.Errorf("unknown %d sidecar failed reason", fr)
	}
}

// String return the corresponding string type
func (fr FailedReason) String() string {
	switch fr {
	case PreHookFailed:
		return "PreHookFailed"
	case PostHookFailed:
		return "PostHookFailed"
	case DownloadFailed:
		return "DownloadFailed"
	case TokenFailed:
		return "TokenFailed"
	case SdkVersionIsTooLowFailed:
		return "SdkVersionIsTooLowFailed"
	case AppMetaFailed:
		return "AppMetaFailed"
	case DeleteOldFilesFailed:
		return "DeleteOldFilesFailed"
	case UpdateMetadataFailed:
		return "UpdateMetadataFailed"
	default:
		return "UnknownFailed"
	}
}

// SpecificFailedReason defines the sub-reasons for failure
type SpecificFailedReason uint32

const (
	// NewFolderFailed represents failure in creating a new folder
	NewFolderFailed SpecificFailedReason = 1 << iota
	// TraverseFolderFailed represents failure in traversing folder
	TraverseFolderFailed
	// DeleteFolderFailed represents failure in deleting folder
	DeleteFolderFailed

	// WriteFileFailed represents failure in writing file
	WriteFileFailed
	// ReadFileFailed represents failure in reading file
	ReadFileFailed
	// OpenFileFailed represents failure in opening file
	OpenFileFailed
	// StatFileFailed indicates failure to obtain file information
	StatFileFailed
	// WriteEnvFileFailed represents failure in writing environment file
	WriteEnvFileFailed
	// CheckFileExistsFailed represents failure in checking if file exists
	CheckFileExistsFailed
	// FilePathNotFound represents failure due to file path not found
	FilePathNotFound

	// ScriptTypeNotSupported represents failure due to unsupported script type
	ScriptTypeNotSupported
	// ScriptExecutionFailed represents failure in script execution
	ScriptExecutionFailed

	// NoDownloadPermission represents failure due to lack of download permission
	NoDownloadPermission
	// GenerateDownloadLinkFailed represents failure in generating download link
	GenerateDownloadLinkFailed
	// ValidateDownloadFailed represents failure in validating downloaded file
	ValidateDownloadFailed
	// DownloadChunkFailed represents failure in downloading file chunks
	DownloadChunkFailed
	// RetryDownloadFailed represents failure in retrying to download file
	RetryDownloadFailed
	// DataEmpty represents failure due to empty data
	DataEmpty
	// SerializationFailed represents failure in serialization
	SerializationFailed
	// FormattingFailed represents failure in formatting
	FormattingFailed
	// TokenPermissionFailed represents failure due to token permission issues
	TokenPermissionFailed
)

// Validate the specific failed reason is valid or not
func (s SpecificFailedReason) Validate() error {
	switch s {
	case NewFolderFailed, WriteFileFailed, WriteEnvFileFailed, TokenPermissionFailed,
		ScriptTypeNotSupported, ScriptExecutionFailed,
		CheckFileExistsFailed, FilePathNotFound, OpenFileFailed, ReadFileFailed,
		NoDownloadPermission, GenerateDownloadLinkFailed, ValidateDownloadFailed,
		DownloadChunkFailed, RetryDownloadFailed,
		DataEmpty, SerializationFailed,
		FormattingFailed, TraverseFolderFailed, StatFileFailed,
		DeleteFolderFailed:
		return nil
	default:
		return fmt.Errorf("unknown %d sidecar specific failed reason", s)
	}
}

// String returns the corresponding string representation of the specific failed reason
func (s SpecificFailedReason) String() string {
	switch s {
	case NewFolderFailed:
		return "NewFolderFailed"
	case WriteFileFailed:
		return "WriteFileFailed"
	case WriteEnvFileFailed:
		return "WriteScriptEnvFileFailed"
	case ScriptTypeNotSupported:
		return "ScriptTypeNotSupported"
	case ScriptExecutionFailed:
		return "ScriptExecutionFailed"
	case CheckFileExistsFailed:
		return "CheckFileExistsFailed"
	case FilePathNotFound:
		return "FilePathNotFound"
	case OpenFileFailed:
		return "OpenFileFailed"
	case NoDownloadPermission:
		return "NoDownloadPermission"
	case GenerateDownloadLinkFailed:
		return "GenerateDownloadLinkFailed"
	case ValidateDownloadFailed:
		return "ValidateDownloadFailed"
	case DownloadChunkFailed:
		return "DownloadChunkFailed"
	case RetryDownloadFailed:
		return "RetryDownloadFailed"
	case DataEmpty:
		return "DataEmpty"
	case SerializationFailed:
		return "SerializationFailed"
	case FormattingFailed:
		return "FormattingFailed"
	case TraverseFolderFailed:
		return "TraverseFolderFailed"
	case DeleteFolderFailed:
		return "DeleteFolderFailed"
	case StatFileFailed:
		return "StatFileFailed"
	case TokenPermissionFailed:
		return "TokenPermissionFailed"
	case ReadFileFailed:
		return "ReadFileFailed"
	default:
		return "UnknownFailed"
	}
}

// PrimaryError is a primary error struct
type PrimaryError struct {
	FailedReason
	SecondaryError
}

// SecondaryError is a secondary error struct
type SecondaryError struct {
	SpecificFailedReason
	Err error
}

// WrapPrimaryError is a function that wraps a secondary error with a primary error
func WrapPrimaryError(primaryError FailedReason, secondaryError SecondaryError) error {
	return PrimaryError{
		FailedReason:   primaryError,
		SecondaryError: secondaryError,
	}
}

// Error method implements the error interface, returning a formatted error message
// indicating the primary error category, secondary error category, and the underlying error.
func (e PrimaryError) Error() string {
	return fmt.Sprintf("primary error: %s, secondary error: %s, err: %s", e.FailedReason.String(),
		e.SpecificFailedReason.String(), e.Err)
}

// WrapSecondaryError is a function that wraps a error with a WrapPrimaryError error
func WrapSecondaryError(specificError SpecificFailedReason, err error) error {
	return SecondaryError{
		SpecificFailedReason: specificError,
		Err:                  err,
	}
}

// Error method implements the error interface, returning a formatted error message
// including the specific error category and the underlying error message.
func (e SecondaryError) Error() string {
	return fmt.Sprintf("secondary error: %s, err: %s", e.SpecificFailedReason.String(), e.Err)
}
