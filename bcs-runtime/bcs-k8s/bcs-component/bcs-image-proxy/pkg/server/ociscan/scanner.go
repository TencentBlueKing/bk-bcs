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

package ociscan

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/docker/docker/distribution/metadata"
	"github.com/moby/moby/client"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// ScanHandler defines the handler for scan oci
type ScanHandler struct {
	op         *options.ImageProxyOption
	cacheStore store.CacheStore

	dc               *dockerdChecker
	cc               *containerdChecker
	dockerLayers     map[string]string
	containerdLayers map[string]string
}

// NewScanHandler create scan handler instance
func NewScanHandler() *ScanHandler {
	op := options.GlobalOptions()
	return &ScanHandler{
		op:         op,
		cacheStore: store.GlobalRedisStore(),
	}
}

// Init the scan handler
func (s *ScanHandler) Init() error {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		s.dc = s.initDockerdChecker()
	}()
	go func() {
		defer wg.Done()
		s.cc = s.initContainerdChecker()
	}()
	wg.Wait()
	s.reportOCILayers(context.Background())
	return nil
}

// TickerReport ticker report oci layers
func (s *ScanHandler) TickerReport(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.reportOCILayers(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// reportOCILayers report docker and containerd oci-layers
func (s *ScanHandler) reportOCILayers(ctx context.Context) {
	wg := &sync.WaitGroup{}
	if s.dc != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			layers := s.dc.Parse()
			s.dockerLayers = layers
			for k, v := range layers {
				if err := s.cacheStore.SaveOCILayer(ctx, store.DOCKERD, k, v); err != nil {
					blog.Errorf(err.Error())
				}
			}
		}()
	}
	if s.cc != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			layers := s.cc.Parse(ctx)
			s.containerdLayers = layers
			for k, v := range layers {
				if err := s.cacheStore.SaveOCILayer(ctx, store.CONTAINERD, k, v); err != nil {
					blog.Errorf(err.Error())
				}
			}
		}()
	}
	wg.Wait()
}

// GenerateLayer generate layers to target file with oci api
func (s *ScanHandler) GenerateLayer(ctx context.Context, ociType string, layer string) (string, error) {
	var result string
	var err error
	switch store.LayerType(ociType) {
	case store.DOCKERD:
		if s.dc == nil || s.dc.RootDir == "" {
			return "", errors.Errorf("copy docker layer no handler")
		}
		result, err = s.handleDockerCopy(ctx, layer)
	case store.CONTAINERD:
		if s.cc == nil {
			return "", errors.Errorf("copy containerd layer no handler")
		}
		result, err = s.handleContainerdCopy(ctx, layer)
	default:
		return "", errors.Errorf("layer path 'type(%s), file(%s)' is unknown", ociType, layer)
	}
	if err != nil {
		return "", errors.Wrapf(err, "generate '%s' oci-layer failed", ociType)
	}
	return result, nil
}

// handleDockerCopy handle docker copy
func (s *ScanHandler) handleDockerCopy(ctx context.Context, layer string) (string, error) {
	if s.dockerLayers == nil {
		return "", errors.Errorf("dockerd not have digests")
	}
	layerFile, ok := s.dockerLayers[layer]
	if !ok {
		return "", errors.Errorf("dockerd not have digest '%s'", layer)
	}
	targetFile := path.Join(s.op.StoragePath, layer+"tar.gzip")
	_ = os.RemoveAll(targetFile)
	if err := utils.CreateTarGz(layerFile, targetFile); err != nil {
		return "", errors.Wrapf(err, "dockerd save digest '%s' failed", layer)
	}
	logctx.Infof(ctx, "layer-docker create tar.gz success: %s", targetFile)
	result := path.Join(s.op.OCIPath, layer+"tar.gzip")
	if err := os.Rename(targetFile, result); err != nil {
		return "", errors.Wrapf(err, "rename '%s' to '%s' failed", targetFile, result)
	}
	return result, nil
}

// handleContainerdCopy handle containerd copy
func (s *ScanHandler) handleContainerdCopy(ctx context.Context, layer string) (string, error) {
	layer = "sha256:" + layer
	layerDigest := digest.Digest(layer)
	nsCtx := namespaces.WithNamespace(ctx, "k8s.io")
	if _, err := s.cc.Client.ContentStore().Info(nsCtx, layerDigest); err != nil {
		if errdefs.IsNotFound(err) {
			return "", errors.Wrapf(err, "containerd get layer '%s' not found", layerDigest)
		}
		return "", errors.Wrapf(err, "containerd get layer info failed")
	}

	ra, err := s.cc.Client.ContentStore().ReaderAt(nsCtx, ocispec.Descriptor{Digest: digest.Digest(layer)})
	if err != nil {
		return "", errors.Wrapf(err, "containerd read digest failed")
	}
	defer ra.Close()
	logctx.Infof(ctx, "layer-containerd read layer '%s' sucess", layer)

	reader := content.NewReader(ra)
	targetFile := path.Join(s.op.StoragePath, layer+"tar.gzip")
	_ = os.RemoveAll(targetFile)
	dstFile, err := os.Create(targetFile)
	if err != nil {
		return "", errors.Wrapf(err, "containerd create layer '%s' failed", targetFile)
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, reader); err != nil {
		return "", errors.Wrapf(err, "containerd copy layer '%s' failed", targetFile)
	}
	result := path.Join(s.op.OCIPath, layer+"tar.gzip")
	if err = os.Rename(targetFile, result); err != nil {
		return "", errors.Wrapf(err, "rename '%s' to '%s' failed", targetFile, result)
	}
	return result, nil
}

// CopyLayerToDestPath copy layer to dest path
func (s *ScanHandler) CopyLayerToDestPath(ctx context.Context, ociType string, layerPath string,
	destFile string) error {
	// handle downloaded layer
	if strings.HasPrefix(layerPath, s.op.StoragePath) {
		if err := os.Rename(layerPath, destFile); err != nil {
			return errors.Wrapf(err, "rename layer from '%s' to '%s' failed", layerPath, destFile)
		}
		return nil
	}
	if strings.HasPrefix(layerPath, s.op.TransferPath) {
		return nil
	}
	switch store.LayerType(ociType) {
	case store.DOCKERD:
		if s.dc != nil && s.dc.RootDir != "" {
			if err := s.handleCopyDockerLayer(ctx, layerPath, destFile); err != nil {
				return errors.Wrapf(err, "copy docker layer from '%s' to '%s' failed", layerPath, destFile)
			}
			return nil
		}
		return errors.Errorf("copy docker layer no handler")
	case store.CONTAINERD:
		if s.cc != nil {
			if err := s.handleCopyContainerdLayer(ctx, layerPath, destFile); err != nil {
				return errors.Wrapf(err, "copy containerd file from '%s to '%s' failed'", layerPath, destFile)
			}
			return nil
		}
		return errors.Errorf("copy containerd layer no handler")
	default:
		return errors.Errorf("layer path 'type(%s), file(%s)' is unknown", ociType, layerPath)
	}
}

// handleCopyDockerLayer handle docker layer copy
func (s *ScanHandler) handleCopyDockerLayer(ctx context.Context, layerPath string, destFile string) error {
	tmpDest := destFile + ".tmp"
	if err := utils.CreateTarGz(layerPath, tmpDest); err != nil {
		return errors.Wrapf(err, "create tar.gz '%s' from '%s' failed", tmpDest, layerPath)
	}
	logctx.Infof(ctx, "layer-docker create tar.gz success: %s", tmpDest)
	if err := os.Rename(tmpDest, destFile); err != nil {
		return errors.Wrapf(err, "rename '%s' to '%s' failed", tmpDest, destFile)
	}
	return nil
}

// handleCopyContainerdLayer handle containerd layer copy
func (s *ScanHandler) handleCopyContainerdLayer(ctx context.Context, layer string, destFile string) error {
	layer = "sha256:" + layer
	layerDigest := digest.Digest(layer)
	nsCtx := namespaces.WithNamespace(ctx, "k8s.io")
	_, err := s.cc.Client.ContentStore().Info(nsCtx, layerDigest)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return errors.Wrapf(err, "containerd get layer '%s' not found", layerDigest)
		}
		return errors.Wrapf(err, "containerd get layer info failed")
	}

	ra, err := s.cc.Client.ContentStore().ReaderAt(nsCtx,
		ocispec.Descriptor{Digest: digest.Digest(layer)})
	if err != nil {
		return errors.Wrapf(err, "containerd read digest failed")
	}
	defer ra.Close()
	logctx.Infof(ctx, "layer-containerd read digest sucess")

	reader := content.NewReader(ra)
	_ = os.RemoveAll(destFile)
	dstFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, reader)
	return err
}

// dockerdChecker defines the docker checker instance
type dockerdChecker struct {
	RootDir string
	Client  *client.Client
}

// initDockerdChecker init the dockerd checker
func (s *ScanHandler) initDockerdChecker() *dockerdChecker {
	if !s.op.EnableDockerd {
		return nil
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		blog.Errorf("ignore docker. init docker client failed: %s", err.Error())
		return nil
	}
	blog.Infof("init docker client success")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dockerInfo, err := cli.Info(ctx)
	if err != nil {
		blog.Warnf("ignore docker. get docker info failed: %s", err.Error())
		return nil
	} else {
		blog.Infof("init docker get info success")
	}

	checker := &dockerdChecker{
		RootDir: dockerInfo.DockerRootDir,
		Client:  cli,
	}
	return checker
}

const (
	defaultLayerDBDir   = "/image/overlay2/layerdb/sha256"
	defaultMetadataDiff = "/image/overlay2/distribution/v2metadata-by-diffid/sha256"
	dockerDiffFile      = "diff"
	dockerCacheIDFile   = "cache-id"
)

// Parse the layers from docker with oci api
func (c *dockerdChecker) Parse() map[string]string {
	result := make(map[string]string)
	layerDB := c.RootDir + defaultLayerDBDir
	layerDBEntries, err := os.ReadDir(layerDB)
	if err != nil {
		blog.Errorf("[docker] read layer dir '%s' failed: %s", layerDB, err.Error())
		return result
	}
	for i := range layerDBEntries {
		layerEntry := layerDBEntries[i]
		if !layerEntry.IsDir() {
			continue
		}
		layerDetailsDir := layerDB + "/" + layerEntry.Name()
		var layerDetails []os.DirEntry
		layerDetails, err = os.ReadDir(layerDetailsDir)
		if err != nil {
			blog.Errorf("[docker] read layer details '%s' failed: %s", layerDetailsDir, err.Error())
			continue
		}
		var key string
		var value string
		for j := range layerDetails {
			item := layerDetails[j]
			if item.Name() != dockerDiffFile && item.Name() != dockerCacheIDFile {
				continue
			}
			itemPath := layerDetailsDir + "/" + item.Name()
			var bs []byte
			bs, err = os.ReadFile(itemPath)
			if err != nil {
				blog.Errorf("[docker] read layer details item '%s' failed: %s", itemPath, err.Error())
				continue
			}
			if item.Name() == dockerDiffFile {
				key = strings.TrimSpace(string(bs))
			}
			if item.Name() == dockerCacheIDFile {
				value = strings.TrimSpace(string(bs))
			}
		}
		if key == "" || value == "" {
			blog.Errorf("[docker] read layer details '%s' diff='%s', cache-id='%s', either cannot be empty",
				layerDetailsDir, key, value)
			continue
		}
		metadataDiffDir := c.RootDir + defaultMetadataDiff
		metadataFile := metadataDiffDir + "/" + strings.TrimPrefix(key, "sha256:")
		var metadataBS []byte
		if metadataBS, err = os.ReadFile(metadataFile); err != nil {
			if !os.IsNotExist(err) {
				blog.Errorf("[docker] read layer metadata '%s' failed: %s", metadataFile, err.Error())
			}
			continue
		}
		metadataObjs := make([]metadata.V2Metadata, 0)
		if err = json.Unmarshal(metadataBS, &metadataObjs); err != nil {
			blog.Errorf("[docker] read layer metadata '%s' unmarshal failed: %s", metadataFile, err.Error())
			continue
		}
		if len(metadataObjs) == 0 {
			blog.Errorf("[docker] read layer metadata '%s' no body", metadataFile)
			continue
		}
		key = strings.TrimPrefix(metadataObjs[0].Digest.String(), "sha256:")
		if key == "" {
			continue
		}
		result[key] = c.RootDir + "/overlay2/" + value + "/diff"
	}
	return result
}

// containerdChecker defines the containerd checker
type containerdChecker struct {
	Client *containerd.Client
}

// initContainerdChecker init the containerd checker
func (s *ScanHandler) initContainerdChecker() *containerdChecker {
	if !s.op.EnableContainerd {
		return nil
	}
	cc, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		blog.Errorf("ignore containerd. init containerd client failed: %s", err.Error())
		return nil
	}
	blog.Infof("init containerd client success")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var vs containerd.Version
	if vs, err = cc.Version(ctx); err != nil {
		blog.Warnf("ignore containerd. get containerd version failed: %s", err.Error())
	} else {
		blog.Infof("init containerd get version sucees: %s", vs.Version)
	}
	return &containerdChecker{
		Client: cc,
	}
}

// Parse the layers from containerd
func (c *containerdChecker) Parse(ctx context.Context) map[string]string {
	nsCtx := namespaces.WithNamespace(ctx, "k8s.io")
	result := make(map[string]string)
	err := c.Client.ContentStore().Walk(nsCtx, func(info content.Info) error {
		digestStr := strings.TrimPrefix(info.Digest.String(), "sha256:")
		result[digestStr] = digestStr
		return nil
	})
	if err != nil {
		logctx.Errorf(ctx, "containerd walk get digests failed: %s", err.Error())
	}
	return result
}
