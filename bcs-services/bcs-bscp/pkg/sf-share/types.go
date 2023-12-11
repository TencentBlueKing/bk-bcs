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
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/validator"
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	pbhook "bscp.io/pkg/protocol/core/hook"
	pbkv "bscp.io/pkg/protocol/core/kv"
	pbfs "bscp.io/pkg/protocol/feed-server"
	"bscp.io/pkg/runtime/jsoni"
)

const (
	// Bounce means the feed server is shutting down or something happens to tell
	// sidecar to close the current connection and re-connect to the other feed
	// servers.
	Bounce FeedMessageType = 1
	// PublishRelease means this app instance matched release has been changed because
	// of new publish has been fired.
	PublishRelease FeedMessageType = 2

	// Unknown is Unknown operator
	Unknown = "unknown"
)

// FeedMessageType defines message types to sidecar delivered form feed server.
type FeedMessageType uint32

// String return the corresponding string type
func (sm FeedMessageType) String() string {
	switch sm {
	case Bounce:
		return "Bounce"
	case PublishRelease:
		return "PublishRelease"
	default:
		return Unknown
	}
}

// MessagingType defines the message type delivered from sidecar to feed server.
type MessagingType uint32

const (
	// SidecarOffline means the sidecar is shutting down or something happens, to tell feed server
	// this sidecar is offline.
	SidecarOffline MessagingType = 1
	// Heartbeat means the sidecar is online, to tell feed server this sidecar is live.
	Heartbeat MessagingType = 2
	// VersionChangeMessage the version change event was reported. Procedure
	VersionChangeMessage MessagingType = 3
	// PullStatus the current pull status is reported
	PullStatus MessagingType = 4
	// ClientInfo report basic information about the client when the client first connects to the client
	ClientInfo MessagingType = 5
)

// Validate the messaging type is valid or not.
func (sm MessagingType) Validate() error {
	switch sm {
	case SidecarOffline:
	case Heartbeat:
	case VersionChangeMessage:
	case PullStatus:
	case ClientInfo:
	default:
		return fmt.Errorf("unknown %d sidecar message type", sm)
	}

	return nil
}

// String return the corresponding string type
func (sm MessagingType) String() string {
	switch sm {
	case SidecarOffline:
		return "SidecarOffline"
	case Heartbeat:
		return "Heartbeat"
	case VersionChangeMessage:
		return "VersionChange"
	case PullStatus:
		return "PullStatus"
	case ClientInfo:
		return "ClientInfo"
	default:
		return Unknown
	}
}

// SideWatchPayload defines the payload information for sidecar to watch feed server.
type SideWatchPayload struct {
	BizID        uint32        `json:"bizID"`
	Applications []SideAppMeta `json:"apps"`
}

// Validate the sidecar's watch payload is valid or not.
func (s SideWatchPayload) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid sidecar watch payload biz id")
	}

	if len(s.Applications) == 0 {
		return errors.New("invalid sidecar watch payload, no apps are set")
	}

	if len(s.Applications) > validator.MaxAppMetas {
		return fmt.Errorf("at most %d apps is allowed for one sidecar", validator.MaxAppMetas)
	}

	for _, one := range s.Applications {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SideAppMeta defines an app's metadata within the sidecar.
type SideAppMeta struct {
	AppID     uint32            `json:"appID"`
	App       string            `json:"app"`
	Namespace string            `json:"namespace"`
	Uid       string            `json:"uid"`
	Labels    map[string]string `json:"labels"`
	// CurrentReleaseID is sidecar's current effected release id.
	CurrentReleaseID uint32 `json:"currentReleaseID"`
	// sidecar's current cursor id
	CurrentCursorID uint32 `json:"currentCursorID"`
	// TargetReleaseID is sidecar's target release id
	TargetReleaseID uint32 `json:"targetReleaseID"`
}

// Validate the sidecar's app meta is valid or not.
func (s SideAppMeta) Validate() error {
	if s.AppID <= 0 {
		return errors.New("invalid sidecar's app id")
	}

	if len(s.Namespace) != 0 {
		if err := validator.ValidateNamespace(s.Namespace); err != nil {
			return fmt.Errorf("invalid sidecar's app namespace, err: %v", err)
		}
	}

	if err := validator.ValidateUid(s.Uid); err != nil {
		return fmt.Errorf("invalid sidecar's app uid, err: %v", err)
	}

	return nil
}

// ConfigItemMetaV1 defines the released configure item's metadata.
type ConfigItemMetaV1 struct {
	// ID is released configuration item identity id.
	ID                   uint32                     `json:"id"`
	CommitID             uint32                     `json:"commentID"`
	ContentSpec          *pbcontent.ContentSpec     `json:"contentSpec"`
	ConfigItemSpec       *pbci.ConfigItemSpec       `json:"configItemSpec"`
	ConfigItemAttachment *pbci.ConfigItemAttachment `json:"configItemAttachment"`
	RepositoryPath       string                     `json:"repositoryPath"`
}

// PbFileMeta returns the pb file meta.
func (cim *ConfigItemMetaV1) PbFileMeta() *pbfs.FileMeta {
	return &pbfs.FileMeta{
		Id:       cim.ID,
		CommitId: cim.CommitID,
		CommitSpec: &pbcommit.CommitSpec{
			Content: &pbcontent.ContentSpec{
				Signature: cim.ContentSpec.Signature,
				ByteSize:  cim.ContentSpec.ByteSize,
			},
		},
		ConfigItemSpec:       cim.ConfigItemSpec,
		ConfigItemAttachment: cim.ConfigItemAttachment,
	}
}

// ReleaseEventMetaV1 defines the event details when the sidecar watch the feed server to
// get the latest release.
type ReleaseEventMetaV1 struct {
	AppID      uint32              `json:"appID"`
	App        string              `json:"app"`
	ReleaseID  uint32              `json:"releaseID"`
	CIMetas    []*ConfigItemMetaV1 `json:"ciMetas"`
	KvMetas    []*KvMetaV1         `json:"kv_metas"`
	Repository *RepositoryV1       `json:"repository"`
	PreHook    *pbhook.HookSpec    `json:"preHook"`
	PostHook   *pbhook.HookSpec    `json:"postHook"`
}

// InstanceSpec defines the specifics for an app instance to watch the event.
type InstanceSpec struct {
	BizID      uint32            `json:"bizID"`
	AppID      uint32            `json:"appID"`
	App        string            `json:"app"`
	Uid        string            `json:"uid"`
	Labels     map[string]string `json:"labels"`
	ConfigType table.ConfigType  `json:"config_type"`
}

// Validate the instance spec is valid or not
func (is InstanceSpec) Validate() error {
	if is.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if is.App == "" {
		return errors.New("invalid app")
	}

	if len(is.Uid) == 0 {
		return errors.New("invalid uid")
	}

	if err := validator.ValidateLabel(is.Labels); err != nil {
		return err
	}

	return nil
}

// Format the instance spec's basic info to string.
func (is *InstanceSpec) Format() string {
	return fmt.Sprintf("biz: %d, app: %s, uid: %s", is.BizID, is.App, is.Uid)
}

// RepositoryV1 defines repository related metas.
type RepositoryV1 struct {
	// Root is the root path to download the configuration items from repository.
	Root            string    `json:"root"`
	TLS             *TLSBytes `json:"tls,omitempty"`
	AccessKeyID     string    `json:"accessKeyId"`
	SecretAccessKey string    `json:"secretAccessKey"`
	Url             string    `json:"url"`
}

// DownloadUri generate the fully qualified URI to download the config item from repository.
func (r RepositoryV1) DownloadUri(rs *RepositorySpecV1) string {
	subPath := strings.TrimRight(rs.Path, " ")
	subPath = strings.Trim(subPath, "/")
	return fmt.Sprintf("%s/%s", r.Root, subPath)
}

// TLSBytes defines the repository's TLS file's body bytes.
// Note: each file's content byte is encoded with BASE64 when it is marshaled
// with json and decoded it from BASE64 when unmarshal it from json.
type TLSBytes struct {
	InsecureSkipVerify bool   `json:"insecure"`
	CaFileBytes        string `json:"ca"`
	CertFileBytes      string `json:"cert"`
	KeyFileBytes       string `json:"key"`
}

// broker used to marshal/unmarshal TLSBytes.
type broker struct {
	InsecureSkipVerify bool   `json:"insecure"`
	CaFileBase64       string `json:"ca"`
	CertFileBase64     string `json:"cert"`
	KeyFileBase64      string `json:"key"`
}

// MarshalJSON marshal the TLSBytes with its each field's value being encoded with BASE64.
func (tls TLSBytes) MarshalJSON() ([]byte, error) {
	tlsBase64 := &broker{
		InsecureSkipVerify: tls.InsecureSkipVerify,
		CaFileBase64:       base64.StdEncoding.EncodeToString([]byte(tls.CaFileBytes)),
		CertFileBase64:     base64.StdEncoding.EncodeToString([]byte(tls.CertFileBytes)),
		KeyFileBase64:      base64.StdEncoding.EncodeToString([]byte(tls.KeyFileBytes)),
	}

	return jsoni.Marshal(tlsBase64)
}

// UnmarshalJSON unmarshal the bytes to TLSBytes with its each field's value being decoded with BASE64.
func (tls *TLSBytes) UnmarshalJSON(bytes []byte) error {
	tlsBase64 := new(broker)
	if err := jsoni.Unmarshal(bytes, &tlsBase64); err != nil {
		return err
	}

	tls.InsecureSkipVerify = tlsBase64.InsecureSkipVerify

	ca, err := base64.StdEncoding.DecodeString(tlsBase64.CaFileBase64)
	if err != nil {
		return fmt.Errorf("decode ca file bytes from base64 failed, err: %v", err)
	}

	tls.CaFileBytes = string(ca)

	cert, err := base64.StdEncoding.DecodeString(tlsBase64.CertFileBase64)
	if err != nil {
		return fmt.Errorf("decode cert file bytes from base64 failed, err: %v", err)
	}

	tls.CertFileBytes = string(cert)

	key, err := base64.StdEncoding.DecodeString(tlsBase64.KeyFileBase64)
	if err != nil {
		return fmt.Errorf("decode key file bytes from base64 failed, err: %v", err)
	}

	tls.KeyFileBytes = string(key)

	return nil
}

// LoadTLSBytes load tls bytes. if tls is disabled, return nil.
func LoadTLSBytes(tls cc.Repository) (*TLSBytes, error) {
	if tls.StorageType == cc.BkRepo {

		if !tls.BkRepo.TLS.Enable() {
			return new(TLSBytes), nil
		}

		ca, err := os.ReadFile(tls.BkRepo.TLS.CAFile)
		if err != nil {
			return nil, err
		}

		cert, err := os.ReadFile(tls.BkRepo.TLS.CertFile)
		if err != nil {
			return nil, err
		}

		key, err := os.ReadFile(tls.BkRepo.TLS.KeyFile)
		if err != nil {
			return nil, err
		}

		tlsBytes := &TLSBytes{
			InsecureSkipVerify: tls.BkRepo.TLS.InsecureSkipVerify,
			CaFileBytes:        string(ca),
			CertFileBytes:      string(cert),
			KeyFileBytes:       string(key),
		}

		return tlsBytes, nil
	}
	return nil, nil
}

// RepositorySpecV1 defines the sub path of the related configuration item stored
// in the repository.
type RepositorySpecV1 struct {
	// Path is the configuration item's relative path according to the root path.
	Path string `json:"path"`
}

// ReleaseChangeEvent defines the release change event's detail information.
type ReleaseChangeEvent struct {
	Rid        string
	APIVersion *pbbase.Versioning
	Payload    []byte
}

// ReleaseChangePayload defines the details when the sidecar's app instance's related
// release has been changed.
type ReleaseChangePayload struct {
	ReleaseMeta *ReleaseEventMetaV1 `json:"releaseMeta"`
	Instance    *InstanceSpec       `json:"instance"`
	CursorID    uint32              `json:"cursorID"`
}

// PayloadName return this payload's name.
func (rc *ReleaseChangePayload) PayloadName() string {
	return "ReleaseChangePayload"
}

// MessageType return the payload related message type.
func (rc *ReleaseChangePayload) MessageType() FeedMessageType {
	return PublishRelease
}

// Encode the ReleaseChangePayload to bytes.
func (rc *ReleaseChangePayload) Encode() ([]byte, error) {
	if rc == nil {
		return nil, errors.New("ReleaseChangePayload is nil, can not be encoded")
	}

	return jsoni.Marshal(rc)
}

// SidecarHandshakePayload defines the options which is returned by feed server
type SidecarHandshakePayload struct {
	ServiceInfo   *ServiceInfo          `json:"serviceInfo"`
	RuntimeOption *SidecarRuntimeOption `json:"runtimeOption"`
}

// SidecarRuntimeOption defines the sidecar's runtime options delivered from the
// upstream server with handshake.
type SidecarRuntimeOption struct {
	// BounceIntervalHour sidecar connect bounce interval, if reach this bounce interval, sidecar will
	// reconnect stream server instance.
	BounceIntervalHour uint                          `json:"bounceInterval"`
	RepositoryTLS      *TLSBytes                     `json:"repositoryTLS"`
	Repository         *RepositoryV1                 `json:"repository"`
	AppReloads         map[ /*appID*/ uint32]*Reload `json:"reload"`
}

// Reload defines the sidecar's notify app to reload config file options delivered from the
// upstream server with handshake.
type Reload struct {
	ReloadType     table.AppReloadType `json:"reload_type"`
	FileReloadSpec *FileReloadSpec     `json:"file_reload_spec"`
}

// FileReloadSpec defines sidecar file reload need info.
type FileReloadSpec struct {
	ReloadFilePath string `json:"reload_file_path"`
}

// ServiceInfo defines the sidecar's need info from the upstream server with handshake.
type ServiceInfo struct {
	// Name feed server instance name, it is used to determine which service instance sidecar is connected to.
	Name string `json:"name"`
}

// OfflinePayload defines sidecar offline to send payload to feed server.
type OfflinePayload struct {
	Applications []AppMeta `json:"applications"`
}

// AppMeta start sidecar bind app meta info.
type AppMeta struct {
	App       string            `json:"app"`
	Namespace string            `json:"namespace"`
	Uid       string            `json:"uid"`
	Labels    map[string]string `json:"labels"`
}

// PayloadName return this payload's name.
func (o *OfflinePayload) PayloadName() string {
	return "OfflinePayload"
}

// MessagingType return the payload related sidecar message type.
func (o *OfflinePayload) MessagingType() MessagingType {
	return SidecarOffline
}

// Encode the OfflinePayload to bytes.
func (o *OfflinePayload) Encode() ([]byte, error) {
	if o == nil {
		return nil, errors.New("OfflinePayload is nil, can not be encoded")
	}

	return jsoni.Marshal(o)
}

// HeartbeatPayload defines sdk heartbeat to send payload to feed server.
type HeartbeatPayload struct {
	BasicData BasicData `json:"basicData"`
	// Applications sdk instance bind app meta info,include app,namespace,uid,labels and app current release id.
	Applications []SideAppMeta `json:"applications"`
	// ResourceUsage 资源相关信息：例如 cpu、内存等
	ResourceUsage
}

// MessagingType return the payload related sidecar message type.
func (h *HeartbeatPayload) MessagingType() MessagingType {
	return Heartbeat
}

// Encode the HeartbeatPayload to bytes.
func (h *HeartbeatPayload) Encode() ([]byte, error) {
	if h == nil {
		return nil, errors.New("HeartbeatPayload is nil, can not be encoded")
	}

	return jsoni.Marshal(h)
}

// KvMetaV1 defines the released kv metadata.
type KvMetaV1 struct {
	// ID is released configuration item identity id.
	ID           uint32             `json:"id"`
	Key          string             `json:"key"`
	KvAttachment *pbkv.KvAttachment `json:"kv_attachment"`
}

// ClientMode define the client mode structure
type ClientMode uint32

const (
	// Pull xxx
	Pull ClientMode = 1
	// Watch xxx
	Watch ClientMode = 2
)

// Validate the client mode is valid or not.
func (cm ClientMode) Validate() error {
	switch cm {
	case Pull:
	case Watch:
	default:
		return fmt.Errorf("unknown %d sidecar client mode", cm)
	}

	return nil
}

// String return the corresponding string type
func (cm ClientMode) String() string {
	switch cm {
	case Pull:
		return "Pull"
	case Watch:
		return "Watch"
	default:
		return Unknown
	}
}

// Labels 标签
type Labels map[string]string

// String return the corresponding string type
func (l Labels) String() string {
	marshal, err := jsoni.Marshal(l)
	if err != nil {
		return ""
	}
	return string(marshal)
}

// BasicData 上报时基础数据
type BasicData struct {
	FingerPrint string `json:"fingerprint"`
	// BizID xxx
	BizID uint32 `json:"bizID"`
	// Labels xxx
	Labels Labels `json:"labels"`
	// ClientMode 客户端模式 pull、watch
	ClientMode ClientMode `json:"clientMode"`
}

// Validate the instance spec is valid or not
func (bd BasicData) Validate() error {
	if bd.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if len(bd.FingerPrint) == 0 {
		return errors.New("invalid fingerPrint")
	}

	return nil
}

// ResourceUsage Resource utilization rate
type ResourceUsage struct {
	MaxCPUUsage     float64 `json:"maxCPUUsage"`
	CurrentCPUUsage float64 `json:"currentCPUUsage"`
	MaxMemUsage     uint64  `json:"maxMemUsage"`
	CurrentMemUsage uint64  `json:"currentMemUsage"`
}

// FailedReason define the failure cause structure
type FailedReason uint32

const (
	// PreHookFailed pre hook failed
	PreHookFailed FailedReason = 1
	// PostHookFailed post hook failed
	PostHookFailed FailedReason = 2
	// DownloadFailed download failed
	DownloadFailed FailedReason = 3
	// AlreadyExistFailed already exist failed
	AlreadyExistFailed FailedReason = 4
)

// Validate the failed reason is valid or not.
func (fr FailedReason) Validate() error {
	switch fr {
	case PreHookFailed:
	case PostHookFailed:
	case DownloadFailed:
	case AlreadyExistFailed:
	default:
		return fmt.Errorf("unknown %d sidecar failed reason", fr)
	}

	return nil
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
	case AlreadyExistFailed:
		return "AlreadyExistFailed"
	default:
		return Unknown
	}
}

// ReleaseChangeStatus define the version change status structure
type ReleaseChangeStatus uint32

const (
	// Success xxx
	Success ReleaseChangeStatus = 1
	// Failed xxx
	Failed ReleaseChangeStatus = 2
	// Processing xxx
	Processing ReleaseChangeStatus = 3
)

// Validate the version change status is valid or not.
func (rs ReleaseChangeStatus) Validate() error {
	switch rs {
	case Success:
	case Failed:
	case Processing:
	default:
		return fmt.Errorf("unknown %d sidecar version change status", rs)
	}

	return nil
}

// String return the corresponding string type
func (rs ReleaseChangeStatus) String() string {
	switch rs {
	case Success:
		return "Success"
	case Failed:
		return "Failed"
	case Processing:
		return "Processing"
	default:
		return Unknown
	}
}

// VersionChangePayload defines sdk version change to send payload to feed server.
type VersionChangePayload struct {
	// BasicData 基础信息：例如客户端唯一标识、bizID、客户端模式
	BasicData BasicData `json:"basicData"`
	// SideAppMeta app相关信息：例如 appName、appID、currentReleaseID等
	SideAppMeta SideAppMeta `json:"sideAppMeta"`
	// ResourceUsage 资源相关信息：例如 cpu、内存等
	ResourceUsage ResourceUsage `json:"resourceUsage"`
	// ClientVersion client version/sdk version
	ClientVersion string `json:"clientVersion"`
	// IP client ip
	IP string `json:"ip"`
	// HeartbeatTime 心跳时间
	HeartbeatTime time.Time `json:"heartbeatTime"`
	// Annotations Additional info (Platform information such as cluster ID, agent ID, host ID, etc.)
	Annotations any `json:"annotations"`
	// TotalSeconds total time required for version changes
	TotalSeconds float64 `json:"totalSeconds"`
	// TotalFileNum pull the number of config files (example: 17/20)
	TotalFileNum int `json:"totalFileNum"`
	// TotalFileSize pull the total size of the config file
	TotalFileSize       uint64              `json:"totalFileSize"`
	StartTime           time.Time           `json:"startTime"`
	EndTime             time.Time           `json:"endTime"`
	FailedReason        FailedReason        `json:"failedReason"`
	ReleaseChangeStatus ReleaseChangeStatus `json:"releaseChangeStatus"`
	FailedDetailReason  string              `json:"failedDetailReason"`
}

// MessagingType return the payload related sidecar message type.
func (v *VersionChangePayload) MessagingType() MessagingType {
	return VersionChangeMessage
}

// Encode the VersionChangePayload to bytes.
func (v *VersionChangePayload) Encode() ([]byte, error) {
	if v == nil {
		return nil, errors.New("VersionChangePayload is nil, can not be encoded")
	}

	return jsoni.Marshal(v)
}

// PullStatusPayload defines sdk pull status to send payload to feed server.
type PullStatusPayload struct {
	// BasicData 例如：bizID 、 fingerprint等
	BasicData BasicData `json:"basicData"`
	// SideAppMeta sdk instance bind app meta info,include app,namespace,uid,labels and app current release id.
	SideAppMeta SideAppMeta `json:"sideAppMeta"`
	// ReleaseChangeStatus 版本变更状态
	ReleaseChangeStatus ReleaseChangeStatus `json:"releaseChangeStatus"`
}

// MessagingType return the payload related sidecar message type.
func (p *PullStatusPayload) MessagingType() MessagingType {
	return PullStatus
}

// Encode the PullStatusPayload to bytes.
func (p *PullStatusPayload) Encode() ([]byte, error) {
	if p == nil {
		return nil, errors.New("PullStatusPayload is nil, can not be encoded")
	}

	return jsoni.Marshal(p)
}

// ClientInfoPayload defines sdk client info to send payload to feed server.
type ClientInfoPayload struct {
	// BasicData 基础信息：例如客户端唯一标识、bizID、客户端模式
	BasicData
	// Applications app相关信息：例如 appName、appID、currentReleaseID等
	Applications []SideAppMeta `json:"applications"`
	// ClientVersion client version/sdk version
	ClientVersion string `json:"clientVersion"`
	// IP client ip
	IP string `json:"ip"`
	// HeartbeatTime 心跳时间
	HeartbeatTime time.Time `json:"heartbeatTime"`
	// Annotations Additional info (Platform information such as cluster ID, agent ID, host ID, etc.)
	Annotations any `json:"annotations"`
}

// MessagingType return the payload related sidecar message type.
func (c *ClientInfoPayload) MessagingType() MessagingType {
	return ClientInfo
}

// Encode the ClientInfoPayload to bytes.
func (c *ClientInfoPayload) Encode() ([]byte, error) {
	if c == nil {
		return nil, errors.New("ClientInfoPayload is nil, can not be encoded")
	}

	return jsoni.Marshal(c)
}
