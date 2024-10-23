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

package repository

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// haClient is client for high availability
// write to master repo only, read from master or slave repo
// there is synchronization mechanism which sync data from master to slave
type haClient struct {
	master  Provider
	slave   Provider
	syncMgr *SyncManager
}

// SyncManager is sync manager
type SyncManager struct {
	queue  *syncQueue
	master Provider
	slave  Provider
}

type syncQueue struct {
	client bedis.Client
}

// SyncManager implements HAEnhancer interface
func (c *haClient) SyncManager() *SyncManager {
	return c.syncMgr
}

// Upload Uploads file to ha repo master
// push file metadata to queue for sync
func (c *haClient) Upload(kt *kit.Kit, sign string, body io.Reader) (*ObjectMetadata, error) {
	md, err := c.master.Upload(kt, sign, body)
	if err != nil {
		return nil, err
	}
	_ = c.syncMgr.PushToQueue(kt, sign)
	return md, err
}

func haErr(masterErr, slaveErr error) error {
	return fmt.Errorf("master error: %v, slave error: %v", masterErr, slaveErr)
}

// Download downloads file from ha repo, read priority: master > slave
func (c *haClient) Download(kt *kit.Kit, sign string) (io.ReadCloser, int64, error) {
	masterBody, masterSize, masterErr := c.master.Download(kt, sign)
	if masterErr == nil {
		return masterBody, masterSize, nil
	}

	slaveBody, slaveSize, slaveErr := c.slave.Download(kt, sign)
	if slaveErr == nil {
		return slaveBody, slaveSize, nil
	}

	return nil, 0, haErr(masterErr, slaveErr)
}

// Metadata only get metadata from master repo
// !Note: slave repo only used for download
func (c *haClient) Metadata(kt *kit.Kit, sign string) (*ObjectMetadata, error) {
	return c.master.Metadata(kt, sign)
}

// InitMultipartUpload init multipart upload file for ha repo master
// push only after the completion of multipart upload
func (c *haClient) InitMultipartUpload(kt *kit.Kit, sign string) (string, error) {
	return c.master.InitMultipartUpload(kt, sign)
}

// MultipartUpload upload one part of the file to ha repo master
// push only after the completion of multipart upload
func (c *haClient) MultipartUpload(kt *kit.Kit, sign string, uploadID string, partNum uint32,
	body io.Reader) error {
	return c.master.MultipartUpload(kt, sign, uploadID, partNum, body)
}

// CompleteMultipartUpload complete multipart upload and return metadata for ha repo master
// push only after the completion of multipart upload
func (c *haClient) CompleteMultipartUpload(kt *kit.Kit, sign string, uploadID string) (*ObjectMetadata, error) {
	md, err := c.master.CompleteMultipartUpload(kt, sign, uploadID)
	if err != nil {
		return nil, err
	}
	_ = c.syncMgr.PushToQueue(kt, sign)
	return md, err
}

// URIDecorator ..
func (c *haClient) URIDecorator(bizID uint32) DecoratorInter {
	return newUriDecoratorInter(bizID)
}

// DownloadLink ha repo file download link, get download url from master and slave
func (c *haClient) DownloadLink(kt *kit.Kit, sign string, fetchLimit uint32) ([]string, error) {
	var urls []string
	masterUrl, masterErr := c.master.DownloadLink(kt, sign, fetchLimit)
	if masterErr == nil {
		urls = append(urls, masterUrl...)
	}

	slaveUrl, slaveErr := c.slave.DownloadLink(kt, sign, fetchLimit)
	if slaveErr == nil {
		urls = append(urls, slaveUrl...)
	}

	if masterErr != nil && slaveErr != nil {
		return nil, haErr(masterErr, slaveErr)
	}

	return urls, nil
}

// AsyncDownload ha repo
func (c *haClient) AsyncDownload(kt *kit.Kit, sign string) (string, error) {
	return "", nil
}

// AsyncDownloadStatus ha repo
func (c *haClient) AsyncDownloadStatus(kt *kit.Kit, sign string, taskID string) (bool, error) {
	return false, nil
}

// newHAClient new ha client
func newHAClient(settings cc.Repository) (BaseProvider, error) {
	var master, slave Provider
	var syncMgr *SyncManager
	var err error
	if master, err = newMasterProvider(settings); err != nil {
		return nil, err
	}
	if slave, err = newSlaveProvider(settings); err != nil {
		return nil, err
	}
	if syncMgr, err = newSyncManager(settings.RedisCluster, master, slave); err != nil {
		return nil, err
	}

	return &haClient{
		master:  master,
		slave:   slave,
		syncMgr: syncMgr,
	}, nil
}

// newMasterProvider new master provider
func newMasterProvider(settings cc.Repository) (Provider, error) {
	switch strings.ToUpper(string(settings.StorageType)) {
	case string(cc.S3):
		return newCosProvider(settings.BaseRepo, settings.RedisCluster)
	case string(cc.BkRepo):
		return newBKRepoProvider(settings.BaseRepo, settings.RedisCluster)
	}
	return nil, fmt.Errorf("unsupported storage type: %s", settings.StorageType)
}

// newSlaveProvider new slave provider
func newSlaveProvider(settings cc.Repository) (Provider, error) {
	switch strings.ToUpper(string(settings.Slave.StorageType)) {
	case string(cc.S3):
		return newCosProvider(settings.Slave, settings.RedisCluster)
	case string(cc.BkRepo):
		return newBKRepoProvider(settings.Slave, settings.RedisCluster)
	}
	return nil, fmt.Errorf("unsupported storage type: %s", settings.StorageType)
}

// newHAProvider new ha provider
func newHAProvider(settings cc.Repository) (Provider, error) {
	p, err := newHAClient(settings)
	if err != nil {
		return nil, err
	}

	var c VariableCacher
	c, err = newVariableCacher(settings.RedisCluster, p)
	if err != nil {
		return nil, err
	}

	return &repoProvider{
		BaseProvider:   p,
		HAEnhancer:     p.(*haClient),
		VariableCacher: c,
	}, nil
}

// newSyncManager new a sync manager
func newSyncManager(redisConf cc.RedisCluster, master, slave Provider) (*SyncManager, error) {
	// init redis client
	client, err := bedis.NewRedisCache(redisConf)
	if err != nil {
		return nil, fmt.Errorf("new redis cluster failed, err: %v", err)
	}

	return &SyncManager{
		queue:  &syncQueue{client: client},
		master: master,
		slave:  slave,
	}, nil
}

// PushToQueue pushes the file metadata msg to queue so that the master's write operations can be received by slave
func (s *SyncManager) PushToQueue(kt *kit.Kit, sign string) error {
	if err := s.QueueClient().LPush(kt.Ctx, s.QueueName(), s.QueueMsg(kt.BizID, sign)); err != nil {
		logs.Errorf("push file metadata to queue failed, redis lpush err: %v, rid: %s", err, kt.Rid)
	}
	return nil
}

// QueueClient returns client for sync queue
func (s *SyncManager) QueueClient() bedis.Client {
	return s.queue.client
}

// QueueName returns name of sync queue
func (s *SyncManager) QueueName() string {
	return "sync_repo_queue"
}

// AckQueueName returns name of ack queue
func (s *SyncManager) AckQueueName() string {
	return "sync_repo_ack_queue"
}

// QueueMsg returns msg for sync queue
func (s *SyncManager) QueueMsg(bizID uint32, sign string) string {
	return fmt.Sprintf("%d_%s", bizID, sign)
}

// ParseQueueMsg parses msg of sync queue
func (s *SyncManager) ParseQueueMsg(msg string) (uint32, string, error) {
	elements := strings.Split(msg, "_")
	if len(elements) != 2 {
		return 0, "", fmt.Errorf("parse queue msg into two elements by '_' failed, msg: %s", msg)
	}

	bizID, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse queue msg's bizID failed, msg: %s, bizID: %s", msg, elements[0])
	}
	return uint32(bizID), elements[1], nil
}

// ErrNoFileInMaster is error of no file in master
var ErrNoFileInMaster = errors.New("file not found in master")

// Sync syncs file from master to slave
func (s *SyncManager) Sync(kt *kit.Kit, sign string) (skip bool, err error) {
	_, err = s.slave.Metadata(kt, sign)
	// the file already exists in slave repo, return directly
	if err == nil {
		return true, nil
	}
	// if the error is not 404 which means the slave repo service has some trouble, return directly
	if err != errf.ErrFileContentNotFound {
		return false, fmt.Errorf("sync file from master to slave failed, slave metadata err: %v", err)
	}

	reader, _, err := s.master.Download(kt, sign)
	if err != nil {
		if err == errf.ErrFileContentNotFound {
			return false, ErrNoFileInMaster
		}
		return false, fmt.Errorf("sync file from master to slave failed, master download err: %v", err)
	}
	defer reader.Close()

	// stream the downloaded content to slave repo for uploading
	_, err = s.slave.Upload(kt, sign, reader)
	if err != nil {
		return false, fmt.Errorf("sync file from master to slave failed, slave upload err: %v", err)
	}

	return false, nil
}
