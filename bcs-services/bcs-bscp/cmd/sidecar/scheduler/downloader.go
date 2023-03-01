/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	sfs "bscp.io/pkg/sf-share"

	"golang.org/x/sync/semaphore"
)

const (
	// TODO: consider config these options.
	defaultSwapBufferSize              = 2 * 1024 * 1024
	defaultRangeDownloadByteSize       = 5 * defaultSwapBufferSize
	requestAwaitResponseTimeoutSeconds = 10
	defaultDownloadSemaphoreWight      = 5
)

// Downloader implements all the supported operations which used to download
// files from repository.
type Downloader interface {
	// Download the configuration items from repository.
	Download(vas *kit.Vas, downloadUri string, fileSize uint64, toFile string) error
}

// InitDownloader init the downloader instance.
func InitDownloader(auth cc.SidecarAuthentication, tlsBytes *sfs.TLSBytes) (map[cc.StorageMode]Downloader, error) {

	tlsC, err := tlsConfigFromTLSBytes(tlsBytes)
	if err != nil {
		return nil, fmt.Errorf("build tls config failed, err: %v", err)
	}

	weight, err := setupDownloadSemWeight()
	if err != nil {
		return nil, fmt.Errorf("get download sem weight failed, err: %v", err)
	}

	downloaderMap := make(map[cc.StorageMode]Downloader, 2)
	downloaderMap[cc.BK_REPO] = &downloader{
		tls:                     tlsC,
		basicAuth:               auth,
		sem:                     semaphore.NewWeighted(weight),
		balanceDownloadByteSize: defaultRangeDownloadByteSize,
	}
	downloaderMap[cc.S3] = &downloaderS3{}
	return downloaderMap, err
}

// setupDownloadSemWeight maximum combined weight for concurrent download access.
func setupDownloadSemWeight() (int64, error) {
	weightEnv := os.Getenv(constant.EnvMaxDownloadFileGoroutines)
	if len(weightEnv) == 0 {
		return defaultDownloadSemaphoreWight, nil
	}

	weight, err := strconv.ParseInt(weightEnv, 10, 64)
	if err != nil {
		return 0, err
	}

	if weight < 1 {
		return 0, errors.New("invalid download sem weight, should >= 1")
	}

	if weight > 15 {
		return 0, errors.New("invalid download sem weight, should <= 15")
	}

	return weight, nil
}

// downloader is used to download the configuration items from repository.
type downloaderS3 struct {
	AccessKeyID     string
	SecretAccessKey string
	Url             string
	Name            string
}

// Download the configuration items from repository.
func (dl *downloaderS3) Download(vas *kit.Vas, downloadUri string, fileSize uint64, toFile string) error {

	file, err := os.OpenFile(toFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open the target file failed, err: %v", err)
	}
	defer file.Close()
	s3Client, err := minio.New(dl.Url, &minio.Options{
		Creds:  credentials.NewStaticV4(dl.AccessKeyID, dl.SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return err
	}
	reader, err := s3Client.GetObject(context.Background(), dl.Name, downloadUri, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	readerSize, _ := reader.Stat()
	if _, err := io.CopyN(file, reader, readerSize.Size); err != nil {
		return err
	}
	return nil
}

// downloader is used to download the configuration items from repository.
type downloader struct {
	tls       *tls.Config
	basicAuth cc.SidecarAuthentication
	sem       *semaphore.Weighted
	// balanceDownloadByteSize determines when to download the file with range policy
	// if the configuration item's content size is larger than this, then it
	// will be downloaded with range policy, otherwise, it will be downloaded directly
	// without range policy.
	balanceDownloadByteSize uint64
}

// Download the configuration items from repository.
func (dl *downloader) Download(vas *kit.Vas, downloadUri string, fileSize uint64, toFile string) error {
	file, err := os.OpenFile(toFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open the target file failed, err: %v", err)
	}
	defer file.Close()

	header := &repositoryHeader{
		kv: map[string]string{
			constant.RidKey: vas.Rid,
		},
	}

	exec := &execDownload{
		dl:          dl,
		file:        file,
		client:      dl.initClient(),
		header:      header,
		vas:         vas,
		downloadUri: downloadUri,
		fileSize:    fileSize,
		toFile:      toFile,
	}

	return exec.do()
}

func (dl *downloader) initClient() *http.Client {
	// TODO: find a way to manage these configuration options.
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     dl.tls,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost: 10,
		// TODO: confirm this
		ResponseHeaderTimeout: 15 * time.Minute,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   0,
	}
}

type execDownload struct {
	dl          *downloader
	file        *os.File
	client      *http.Client
	header      *repositoryHeader
	vas         *kit.Vas
	downloadUri string
	fileSize    uint64
	toFile      string
}

func (exec *execDownload) do() error {
	if exec.fileSize <= exec.dl.balanceDownloadByteSize {
		// the file size is not big enough, download directly
		if err := exec.downloadDirectly(requestAwaitResponseTimeoutSeconds); err != nil {
			return fmt.Errorf("download directly failed, err: %v", err)
		}

		return nil
	}

	size, yes, err := exec.isRepositorySupportRangeDownload()
	if err != nil {
		return fmt.Errorf("check if repository support range download failed because of %v", err)
	}

	if yes {
		if size != exec.fileSize {
			return fmt.Errorf("the to be download file size: %d is not as what we expected %d", size, exec.fileSize)
		}

		if err := exec.downloadWithRange(); err != nil {
			return fmt.Errorf("download with range failed, err: %v", err)
		}

		return nil
	}

	logs.Warnf("repository do not support download with range policy, download directly now. rid: %s", exec.vas.Rid)

	if err := exec.downloadDirectly(requestAwaitResponseTimeoutSeconds); err != nil {
		return fmt.Errorf("download directly failed, err: %v", err)
	}

	return nil
}

// isRepositorySupportRangeDownload return the configuration item's content size if
// it supports range download.
func (exec *execDownload) isRepositorySupportRangeDownload() (uint64, bool, error) {
	req, err := http.NewRequest(http.MethodHead, exec.downloadUri, nil)
	if err != nil {
		return 0, false, fmt.Errorf("new request failed, err: %v", err)
	}

	req.WithContext(exec.vas.Ctx)
	req.Header = exec.header.Clone()
	req.Header.Set("Request-Timeout", strconv.Itoa(15))

	if exec.dl.basicAuth.IsEnabled() {
		req.SetBasicAuth(exec.dl.basicAuth.User, exec.dl.basicAuth.Token)
	}

	resp, err := exec.client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("do request failed, err: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("request to repository failed http code: %d", resp.StatusCode)
	}

	mode, exist := resp.Header["Accept-Ranges"]
	if !exist {
		return 0, false, nil
	}

	if len(mode) == 0 {
		return 0, false, nil
	}

	if mode[0] != "bytes" {
		return 0, false, nil
	}

	length, exist := resp.Header["Content-Length"]
	if !exist {
		return 0, false, errors.New("can not get the content length form header")
	}

	if len(length) == 0 {
		return 0, false, errors.New("can not get the content length form header")
	}

	size, err := strconv.ParseUint(length[0], 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parse content length failed, err: %v", err)
	}

	return size, true, nil

}

// downloadDirectly download file without range.
func (exec *execDownload) downloadDirectly(timeoutSeconds int) error {
	if err := exec.dl.sem.Acquire(exec.vas.Ctx, 1); err != nil {
		return fmt.Errorf("acquire semaphore failed, err: %v", err)
	}

	defer exec.dl.sem.Release(1)

	start := time.Now()
	header := exec.header.Clone()
	body, err := exec.doRequest(http.MethodGet, header, timeoutSeconds)
	if err != nil {
		return err
	}
	defer body.Close()

	if err := exec.writeToFile(body, exec.fileSize, 0); err != nil {
		return err
	}

	logs.V(0).Infof("file[%s], download directly success, cost: %s, rid: %s", exec.downloadUri,
		time.Since(start).String(), exec.vas.Rid)

	return nil
}

func (exec *execDownload) downloadWithRange() error {

	logs.Infof("start download file[%s] with range, rid: %s", exec.downloadUri, exec.vas.Rid)

	start, end := uint64(0), uint64(0)
	batchSize := 2 * exec.dl.balanceDownloadByteSize
	// calculate the total parts to be downloaded
	totalParts := int(exec.fileSize / batchSize)
	if (exec.fileSize % batchSize) > 0 {
		totalParts += 1
	}

	var hitError error
	wg := sync.WaitGroup{}

	for part := 0; part < totalParts; part++ {
		if err := exec.dl.sem.Acquire(exec.vas.Ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore failed, err: %v", err)
		}

		start = uint64(part) * batchSize

		if part == totalParts-1 {
			end = exec.fileSize
		} else {
			end = start + batchSize
		}

		end -= 1

		wg.Add(1)

		go func(pos int, from uint64, to uint64) {
			defer func() {
				wg.Done()
				exec.dl.sem.Release(1)
			}()

			start := time.Now()
			if err := exec.downloadOneRangedPart(from, to); err != nil {
				hitError = err
				logs.Errorf("download file[%s] part %d failed, start: %d, err: %v", exec.downloadUri, pos, from, err)
				return
			}

			logs.V(0).Infof("download file range part %d success, range [%d, %d], cost: %s, rid: %s", pos, from, to,
				time.Since(start).String(), exec.vas.Rid)

		}(part, start, end)

	}

	wg.Wait()

	if hitError != nil {
		return hitError
	}

	logs.V(1).Infof("download full file[%s] success, rid: %s", exec.downloadUri, exec.vas.Rid)

	return nil
}

func (exec *execDownload) downloadOneRangedPart(start uint64, end uint64) error {
	if start > end {
		return errors.New("invalid start or end to do range download")
	}

	header := exec.header.Clone()
	// set ranged part.
	if start == end {
		header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	} else {
		header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	}

	body, err := exec.doRequest(http.MethodGet, header, 6*requestAwaitResponseTimeoutSeconds)
	if err != nil {
		return err
	}

	defer body.Close()

	if err := exec.writeToFile(body, end-start+1, start); err != nil {
		return err
	}

	return nil
}

func (exec *execDownload) doRequest(method string, header http.Header, timeoutSeconds int) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, exec.downloadUri, nil)
	if err != nil {
		return nil, fmt.Errorf("new request failed, err: %v", err)
	}

	req.Header = header
	// Note: do not use request context to control timeout, the context is
	// managed by the upper scheduler.
	if timeoutSeconds > 0 {
		req.Header.Set("Request-Timeout", strconv.Itoa(timeoutSeconds))
	}

	req.WithContext(exec.vas.Ctx)

	if exec.dl.basicAuth.IsEnabled() {
		req.SetBasicAuth(exec.dl.basicAuth.User, exec.dl.basicAuth.Token)
	}

	resp, err := exec.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed, err: %v", err)
	}

	// download with range's http status code is 206:StatusPartialContent
	// reference:
	// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Range_requests
	// https://datatracker.ietf.org/doc/html/rfc7233#page-8
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("request to repository, but returned with http code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (exec *execDownload) writeToFile(body io.ReadCloser, expectSize uint64, start uint64) error {
	totalSize := uint64(0)
	swap := make([]byte, defaultSwapBufferSize)
	for {
		select {
		case <-exec.vas.Ctx.Done():
			return fmt.Errorf("download context done, %v", exec.vas.Ctx.Err())

		default:
		}

		picked, err := body.Read(swap)
		// we should always process the n > 0 bytes returned before
		// considering the error err. Doing so correctly handles I/O errors
		// that happen after reading some bytes and also both of the
		// allowed EOF behaviors.
		if picked > 0 {
			var cnt int
			cnt, err = exec.file.WriteAt(swap[0:picked], int64(start+totalSize))
			if err != nil {
				return fmt.Errorf("write data to file failed, err: %v", err)
			}

			if cnt != picked {
				return fmt.Errorf("writed to file's size: %d is not as what we expected: %d", cnt, picked)
			}

			totalSize += uint64(picked)
		}

		if err == nil {
			continue
		}

		if err != io.EOF {
			return fmt.Errorf("read data from response body failed, err: %v", err)
		}

		break
	}

	// the file has already been downloaded to the local, check the file size now.
	if totalSize != expectSize {
		return fmt.Errorf("the downloaded file's total size %d is not what we expected %d", totalSize, expectSize)
	}

	return nil
}
