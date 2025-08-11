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

package bittorrent

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// TorrentHandler defines the torrent handler
type TorrentHandler struct {
	sync.Mutex
	torrentLock lock.Interface
	op          *options.ImageProxyOption

	client       *torrent.Client
	pc           storage.PieceCompletion
	cacheStore   store.CacheStore
	torrentCache *sync.Map

	semaphore chan struct{}
}

func (th *TorrentHandler) storeTorrent(ctx context.Context, digest string, clientMagnet string) {
	th.Lock()
	defer th.Unlock()
	if err := th.cacheStore.SaveTorrent(ctx, digest, clientMagnet); err != nil {
		blog.Errorf("store torrent '%s' to cache-store failed: %s", digest, err.Error())
	}
}

func (th *TorrentHandler) delTorrent(digest string) {
	th.Lock()
	defer th.Unlock()
	if err := th.cacheStore.DeleteTorrent(context.Background(), digest); err != nil {
		blog.Warnf("delete torrent '%s' from cache-store failed: %s", digest, err.Error())
	}
}

// NewTorrentHandler create the torrent handler instance
func NewTorrentHandler() *TorrentHandler {
	return &TorrentHandler{
		op:           options.GlobalOptions(),
		cacheStore:   store.GlobalRedisStore(),
		torrentLock:  lock.NewLocalLock(),
		torrentCache: &sync.Map{},
		semaphore:    make(chan struct{}, 10),
	}
}

// Init the torrent handler
func (th *TorrentHandler) Init() error {
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = th.op.TorrentPath
	clientConfig.Seed = true
	clientConfig.ListenPort = int(th.op.TorrentPort)
	clientConfig.DefaultStorage = storage.NewMMap(th.op.TorrentPath)
	// clientConfig.DefaultStorage = storage.NewFileByInfoHash(th.op.TorrentPath)
	clientConfig.DisableUTP = true
	clientConfig.MaxUnverifiedBytes = 64 << 30
	clientConfig.NoDHT = true
	clientConfig.DisablePEX = false
	clientConfig.EstablishedConnsPerTorrent = 100
	clientConfig.HalfOpenConnsPerTorrent = 50
	clientConfig.TorrentPeersHighWater = 2000
	clientConfig.DisableAcceptRateLimiting = true
	clientConfig.AcceptPeerConnections = true
	if th.op.TorrentUploadLimit > 0 {
		clientConfig.UploadRateLimiter = rate.NewLimiter(rate.Limit(th.op.TorrentUploadLimit),
			int(th.op.TorrentUploadLimit))
	}
	if th.op.TorrentDownloadLimit > 0 {
		clientConfig.DownloadRateLimiter = rate.NewLimiter(rate.Limit(th.op.TorrentDownloadLimit),
			int(th.op.TorrentDownloadLimit))
	}
	tc, err := torrent.NewClient(clientConfig)
	if err != nil {
		return errors.Wrapf(err, "create torrent client failed")
	}
	th.client = tc
	th.pc, err = storage.NewDefaultPieceCompletionForDir(".")
	if err != nil {
		return errors.Wrapf(err, "new piece completion for dir '%s' failed", th.op.TorrentPath)
	}
	return nil
}

// GetClient get the torrent client
func (th *TorrentHandler) GetClient() *torrent.Client {
	return th.client
}

// TickReport tick report the torrent cache
func (th *TorrentHandler) TickReport(ctx context.Context) {
	ticker := time.NewTicker(90 * time.Second)
	defer ticker.Stop()
	defer th.pc.Close()
	defer func() {
		_, torrentStrings := th.returnLocalTorrents(ctx)
		for k := range torrentStrings {
			_ = th.cacheStore.DeleteTorrent(context.Background(), k)
		}
	}()
	for {
		select {
		case <-ticker.C:
			_, torrentStrings := th.returnLocalTorrents(ctx)
			th.torrentCache.Range(func(k, v interface{}) bool {
				digest := k.(string)
				torrentBase64, ok := torrentStrings[digest]
				if !ok {
					return true
				}
				if err := th.cacheStore.SaveTorrent(ctx, digest, torrentBase64); err != nil {
					blog.Errorf("torrent cache save to cache-store for digest '%s' failed: %s",
						digest, err.Error())
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

func (th *TorrentHandler) getLayerFiles(path string) ([]string, error) {
	layerFiles := make([]string, 0)
	if err := filepath.Walk(path, func(fp string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(fp, ".tar.gzip") {
			layerFiles = append(layerFiles, fp)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "read path '%s' failed", th.op.TorrentPath)
	}
	return layerFiles, nil
}

func (th *TorrentHandler) copySourceToTorrent(sourceFile, digest string) (string, error) {
	torrentFile := path.Join(th.op.TorrentPath, utils.LayerFileName(digest))
	_ = os.RemoveAll(torrentFile)
	torrentFi, err := os.Create(torrentFile)
	if err != nil {
		return torrentFile, errors.Wrapf(err, "create torrent file '%s' failed", torrentFile)
	}
	defer torrentFi.Close()
	source, err := os.Open(sourceFile)
	if err != nil {
		return torrentFile, errors.Wrapf(err, "open source file '%s' failed", sourceFile)
	}
	defer source.Close()
	if _, err = io.Copy(torrentFi, source); err != nil {
		return torrentFile, errors.Wrapf(err, "copy source file '%s' to '%s' failed", sourceFile, torrentFile)
	}
	return torrentFile, nil
}

// GenerateTorrent generate the file to torrent
func (th *TorrentHandler) GenerateTorrent(ctx context.Context, digest, sourceFile string) (string, error) {
	th.torrentLock.Lock(ctx, digest)
	defer th.torrentLock.UnLock(ctx, digest)
	to, torrentBase64 := th.CheckTorrentLocalExist(ctx, digest)
	if to != nil {
		logctx.Infof(ctx, "torrent already exist, return diectly")
		return torrentBase64, nil
	}

	// copy source-file to torrent path
	torrentFile, err := th.copySourceToTorrent(sourceFile, digest)
	if err != nil {
		return "", err
	}
	var serveTo *torrent.Torrent
	generateRetry := 3
	for i := 0; i < generateRetry; i++ {
		serveTo, err = th.generateServeTorrent(ctx, digest, torrentFile)
		if err != nil {
			return "", errors.Wrapf(err, "generate serve torrent failed")
		}
		v, _ := th.CheckTorrentLocalExist(ctx, digest)
		if v != nil {
			logctx.Infof(ctx, "check torrent exist in db")
			break
		}
		if i == generateRetry-1 {
			return "", errors.Errorf("generate torrent failed because no torrent in db")
		}
		logctx.Warnf(ctx, "generate torrent succes but no torrent in db, should generate again(i=%d)", i)
	}
	if serveTo == nil {
		return "", errors.Errorf("generate torrent is nil")
	}

	serveMi := serveTo.Metainfo()
	var buffer bytes.Buffer
	if err = serveMi.Write(&buffer); err != nil {
		return "", errors.Wrapf(err, "get torrent bytes failed")
	}

	th.torrentCache.Store(digest, serveTo)
	torrentBase64 = base64.StdEncoding.EncodeToString(buffer.Bytes())
	logctx.Infof(ctx, "generate serve torrent success: (too long)")
	th.storeTorrent(ctx, digest, torrentBase64)

	return torrentBase64, nil
}

func (th *TorrentHandler) generateServeTorrent(ctx context.Context, digest, layerFile string) (*torrent.Torrent, error) {
	fi, err := os.Stat(layerFile)
	if err != nil {
		return nil, err
	}
	pieceLength := metainfo.ChoosePieceLength(fi.Size())
	info := metainfo.Info{
		PieceLength: pieceLength,
	}
	if err = info.BuildFromFilePath(layerFile); err != nil {
		return nil, errors.Wrapf(err, "build torrent metainfo from file '%s' failed", layerFile)
	}
	mi := metainfo.MetaInfo{
		InfoBytes: bencode.MustMarshal(info),
	}
	ih := mi.HashInfoBytes()
	to, _ := th.client.AddTorrentOpt(torrent.AddTorrentOpts{
		InfoHash: ih,
		Storage:  storage.NewMMapWithCompletion(th.op.TorrentPath, th.pc),
		//Storage: storage.NewFileOpts(storage.NewFileClientOpts{
		//	ClientBaseDir: layerFile,
		//	FilePathMaker: func(opts storage.FilePathMakerOpts) string {
		//		return filepath.Join(opts.File.BestPath()...)
		//	},
		//	TorrentDirMaker: nil,
		//	PieceCompletion: th.pc,
		//}),
		ChunkSize: 131072,
	})
	if err = to.MergeSpec(&torrent.TorrentSpec{
		DisplayName: digest,
		InfoBytes:   mi.InfoBytes,
		Trackers:    [][]string{{th.op.TorrentAnnounce}},
	}); err != nil {
		return nil, errors.Wrapf(err, "setting trackers failed")
	}
	logctx.Infof(ctx, "generate torrent success")
	return to, nil
}

// CheckTorrentLocalExist check torrent local exist
func (th *TorrentHandler) CheckTorrentLocalExist(ctx context.Context, digest string) (*torrent.Torrent, string) {
	torrentObjs, torrentStrings := th.returnLocalTorrents(ctx)
	return torrentObjs[digest], torrentStrings[digest]
}

func (th *TorrentHandler) returnLocalTorrents(ctx context.Context) (map[string]*torrent.Torrent, map[string]string) {
	ts := th.client.Torrents()
	torrentObjs := make(map[string]*torrent.Torrent)
	torrentStrings := make(map[string]string)
	for _, t := range ts {
		if t == nil {
			continue
		}
		ti := t.Info()
		if ti == nil {
			continue
		}
		mi := t.Metainfo()
		var buffer bytes.Buffer
		if err := mi.Write(&buffer); err != nil {
			logctx.Errorf(ctx, "torrent get bytes failed: %s", err.Error())
			continue
		}
		digest := strings.TrimSuffix(ti.Name, ".tar.gzip")
		torrentObjs[digest] = t
		torrentStrings[digest] = base64.StdEncoding.EncodeToString(buffer.Bytes())
	}
	return torrentObjs, torrentStrings
}

func (th *TorrentHandler) gotTorrentInfo(t *torrent.Torrent) error {
	n := (t.Length() / 10000000) + 1
	gotInfoTimeout := time.After(time.Duration(n*60) * time.Second)
	select {
	case <-gotInfoTimeout:
		return errors.Errorf("got torrent info timeout")
	case <-t.GotInfo():
		return nil
	}
}

func waitForPieces(ctx context.Context, t *torrent.Torrent, beginIndex, endIndex int) {
	sub := t.SubscribePieceStateChanges()
	defer sub.Close()
	expected := storage.Completion{
		Complete: true,
		Ok:       true,
	}
	pending := make(map[int]struct{})
	for i := beginIndex; i < endIndex; i++ {
		if t.Piece(i).State().Completion != expected {
			pending[i] = struct{}{}
		}
	}
	for {
		if len(pending) == 0 {
			return
		}
		select {
		case ev := <-sub.Values:
			if ev.Completion == expected {
				delete(pending, ev.Index)
			}
		case <-ctx.Done():
			return
		}
	}
}

// DownloadTorrent download the file by torrent
func (th *TorrentHandler) DownloadTorrent(ctx context.Context, rw http.ResponseWriter, located, digest,
	torrentBase64 string) (bool, int64, error) {
	remedy, transmitted, err := th.downloadTorrent(ctx, rw, located, digest, torrentBase64)
	if err != nil {
		return remedy, transmitted, err
	}
	torrentFile := path.Join(th.op.TorrentPath, utils.LayerFileName(digest))
	defer func() {
		if removeErr := os.RemoveAll(torrentFile); removeErr != nil {
			logctx.Warnf(ctx, "remove torrent file '%s' failed: %s", torrentFile, removeErr.Error())
		} else {
			logctx.Infof(ctx, "remove torrent file '%s' success", torrentFile)
		}
	}()
	logical, physical, isSparse, err := utils.IsSparseFile(torrentFile)
	if err != nil {
		return true, 0, errors.Wrapf(err, "check sparse file failed")
	}
	if isSparse {
		return true, 0, errors.Errorf("file '%s' is sparse file, logical: %d, physical: %d",
			torrentFile, logical, physical)
	}
	logctx.Infof(ctx, "torrent file '%s' is normal, logical: %d, physical: %d", torrentFile, logical, physical)

	tf, err := os.Open(torrentFile)
	if err != nil {
		return true, 0, errors.Wrapf(err, "open torrent file '%s' failed", torrentFile)
	}
	defer tf.Close()
	//buf := make([]byte, 1024*1024)
	//if written, err := io.CopyBuffer(rw, tf, buf); err != nil {
	//	return true, written, errors.Wrapf(err, "rewrite torrent file '%s' failed", torrentFile)
	//}
	if written, err := io.Copy(rw, tf); err != nil {
		return true, written, errors.Wrapf(err, "rewrite torrent file '%s' failed", torrentFile)
	}
	return true, 0, nil
}

func (th *TorrentHandler) downloadTorrent(ctx context.Context, rw http.ResponseWriter, located, digest,
	torrentBase64 string) (bool, int64, error) {
	torrentBytes, err := base64.StdEncoding.DecodeString(torrentBase64)
	if err != nil {
		return true, 0, errors.Wrapf(err, "base64 decode '%s' failed", torrentBase64)
	}
	mi, err := metainfo.Load(bytes.NewBuffer(torrentBytes))
	if err != nil {
		return true, 0, errors.Wrapf(err, "load metainfo '%s' failed", torrentBase64)
	}
	t, err := th.client.AddTorrent(mi)
	if err != nil {
		return true, 0, errors.Wrapf(err, "add torrent '%s' failed", torrentBase64)
	}
	if err = th.gotTorrentInfo(t); err != nil {
		return true, 0, err
	}
	defer func() {
		if _, ok := th.torrentCache.Load(digest); ok {
			return
		}
		// drop torrent after download, if dest-file not exist in local
		t.Drop()
	}()
	// ignore chunk error
	t.SetOnWriteChunkError(func(err error) {})
	th.semaphore <- struct{}{}
	defer func() { <-th.semaphore }()
	t.DownloadAll()
	logctx.Infof(ctx, "torrent start downloading")
	start := time.Now()
	done := make(chan struct{})
	//go func() {
	//	defer close(done)
	//	waitForPieces(ctx, t, 0, t.NumPieces())
	//}()

	interval := 5 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	recorderTicker := time.NewTicker(30 * time.Second)
	defer recorderTicker.Stop()
	var currentBytes, prevBytes int64
	completedSlice := make([]int64, 0)
	allSize := humanize.Bytes(uint64(t.Length()))
	for {
		currentBytes = t.BytesCompleted()
		byteRate := (currentBytes - prevBytes) * int64(time.Second) / int64(interval)
		var progress float64 = 0
		if t.Length() != 0 {
			progress = float64(currentBytes) / float64(t.Length()) * 100
		}
		select {
		case <-ticker.C:
			logctx.Infof(ctx, "torrent downloading(%v): %s/%s: %v/s, completed(bytes): %.2f%%",
				time.Since(start),
				humanize.Bytes(uint64(currentBytes)),
				allSize,
				humanize.Bytes(uint64(byteRate)),
				float64(t.BytesCompleted())/float64(t.Length())*100,
			)
			if currentBytes == t.Length() {
				close(done)
				break
			}
			completedSlice = append(completedSlice, currentBytes)
			prevBytes = currentBytes

			if currentBytes == 0 {
				noDownloadPoints := 36
				// 找寻 36 个点以前(180s 前)的数据，确认当前是否 3min 仍然未能开始下载
				if len(completedSlice) > noDownloadPoints {
					oldPieces := completedSlice[len(completedSlice)-noDownloadPoints]
					if currentBytes == oldPieces {
						return true, 0, errors.Errorf("torrent start download failed for a long time " +
							"with can reverse")
					}
				}
			} else {
				noSpeedPoints := 12
				// 找寻 12 个点以前(60s 前)的数据，确认当前是否 1min 没有速度
				if len(completedSlice) > noSpeedPoints {
					oldPieces := completedSlice[len(completedSlice)-noSpeedPoints]
					if currentBytes == oldPieces {
						return true, 0, errors.Errorf("download torrent no speed")
					}
				}
			}
		case <-recorderTicker.C:
			recorder.GlobalRecorder().SendObjEvent(ctx, recorder.Normal,
				fmt.Sprintf("Download-by-torrent ‘%s’ process: %.2f%% (%s/%s)", digest, progress,
					humanize.Bytes(uint64(currentBytes)), allSize))
		case <-ctx.Done():
			return true, 0, errors.Errorf("download torrent context exceeded")
		case <-done:
			logctx.Infof(ctx, "torrent download completed")
			return true, 0, nil
		}
	}
}
