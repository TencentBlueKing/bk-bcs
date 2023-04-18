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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"bscp.io/cmd/sidecar/stream"
	"bscp.io/pkg/cc"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"

	"github.com/gofrs/flock"
)

// SchOptions defines scheduler related options
type SchOptions struct {
	Settings      cc.SidecarSetting
	RepositoryTLS *sfs.TLSBytes
	AppReloads    map[uint32]*sfs.Reload
	Stream        stream.Interface
}

var (
	// errorFLockFailed is error of file lock failed.
	errorFLockFailed = errors.New("can't get flock, try again later")

	// tryLockTimeout is the timeout time to get the file lock.
	tryLockTimeout = time.Minute
)

// LockFile locks target file.
func LockFile(file string, needBlock bool) (*flock.Flock, error) {
	fl := flock.New(file)

	isLocked, err := fl.TryLock()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		// lock file is not exit, create it for not.
		fd, err := os.Create(file)
		if err != nil {
			return nil, fmt.Errorf("create lock file: %s failed, err: %v", file, err)
		}
		fd.Close()

	}

	if isLocked {
		return fl, nil
	}

	if !needBlock {
		return nil, errorFLockFailed
	}

	start := time.Now()
	for {
		if time.Since(start) > tryLockTimeout {
			return nil, fmt.Errorf("get file lock timeout after %s", time.Since(start).String())
		}

		fl := flock.New(file)
		isLocked, err := fl.TryLock()
		if err != nil {
			return nil, err
		}

		if isLocked {
			return fl, nil
		}

		time.Sleep(time.Second)
	}

}

// UnlockFile unlocks target file lock.
func UnlockFile(fl *flock.Flock) error {
	if fl == nil {
		return errors.New("flock is nil")
	}

	if err := fl.Unlock(); err != nil {
		return err
	}

	return nil
}

func tlsConfigFromTLSBytes(tlsBytes *sfs.TLSBytes) (*tls.Config, error) {
	if tlsBytes == nil {
		return new(tls.Config), nil
	}

	var caPool *x509.CertPool
	if len(tlsBytes.CaFileBytes) != 0 {
		caPool = x509.NewCertPool()
		if ok := caPool.AppendCertsFromPEM([]byte(tlsBytes.CaFileBytes)); ok != true {
			return nil, fmt.Errorf("append ca cert failed")
		}
	}

	var certificate tls.Certificate
	if len(tlsBytes.CertFileBytes) == 0 && len(tlsBytes.CertFileBytes) == 0 {
		return &tls.Config{
			InsecureSkipVerify: tlsBytes.InsecureSkipVerify,
			ClientCAs:          caPool,
			Certificates:       []tls.Certificate{certificate},
			ClientAuth:         tls.RequireAndVerifyClientCert,
		}, nil
	}

	tlsCert, err := tls.X509KeyPair([]byte(tlsBytes.CertFileBytes), []byte(tlsBytes.KeyFileBytes))
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: tlsBytes.InsecureSkipVerify,
		ClientCAs:          caPool,
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}, nil
}

type currentRelease struct {
	// lock the currentRelease.
	lock sync.Mutex
	// currentRelease describe the latest release's id which is handled successfully by
	// this Jobs instance. if the incoming event's release id is same with it, then the
	// incoming event can be dropped.
	release *sfs.ReleaseEventMetaV1
	// CursorID is the event's cursor which is bound to the upper release.
	cursorID uint32
}

// ReleaseID return the current release's id, if not exist, return 0.
func (cr *currentRelease) ReleaseID() uint32 {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	if cr.release == nil {
		return 0
	}

	return cr.release.ReleaseID
}

// Release return the current release metadata if it exists.
func (cr *currentRelease) Release() (releaseID uint32, cursorID uint32, exist bool) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	if cr.release == nil {
		return 0, 0, false
	}

	return cr.release.ReleaseID, cr.cursorID, true
}

// Set NOTES
func (cr *currentRelease) Set(to *sfs.ReleaseEventMetaV1) {
	cr.lock.Lock()
	defer cr.lock.Unlock()

	if to == nil {
		cr.release = nil
		return
	}

	// deep copy sfs.ReleaseEventMetaV1.
	cr.release = &sfs.ReleaseEventMetaV1{
		AppID:      to.AppID,
		ReleaseID:  to.ReleaseID,
		CIMetas:    make([]*sfs.ConfigItemMetaV1, 0),
		Repository: new(sfs.RepositoryV1),
	}

	for _, one := range to.CIMetas {
		meta := &sfs.ConfigItemMetaV1{
			ID:             one.ID,
			ContentSpec:    nil,
			ConfigItemSpec: nil,
			RepositoryPath: one.RepositoryPath,
		}

		if one.ContentSpec != nil {
			meta.ContentSpec = &pbcontent.ContentSpec{
				Signature: one.ContentSpec.Signature,
				ByteSize:  one.ContentSpec.ByteSize,
			}
		}

		if one.ConfigItemSpec != nil {
			meta.ConfigItemSpec = &pbci.ConfigItemSpec{
				Name:       one.ConfigItemSpec.Name,
				Path:       one.ConfigItemSpec.Path,
				FileType:   one.ConfigItemSpec.FileType,
				FileMode:   one.ConfigItemSpec.FileMode,
				Memo:       one.ConfigItemSpec.Memo,
				Permission: nil,
			}

			if one.ConfigItemSpec.Permission != nil {
				meta.ConfigItemSpec.Permission = &pbci.FilePermission{
					User:      one.ConfigItemSpec.Permission.User,
					UserGroup: one.ConfigItemSpec.Permission.UserGroup,
					Privilege: one.ConfigItemSpec.Permission.Privilege,
				}
			}
		}

		cr.release.CIMetas = append(cr.release.CIMetas, meta)
	}

	if to.Repository != nil {
		cr.release.Repository = &sfs.RepositoryV1{
			Root: to.Repository.Root,
			TLS:  nil,
		}

		if to.Repository.TLS != nil {
			cr.release.Repository.TLS = &sfs.TLSBytes{
				InsecureSkipVerify: to.Repository.TLS.InsecureSkipVerify,
				CaFileBytes:        to.Repository.TLS.CaFileBytes,
				CertFileBytes:      to.Repository.TLS.CertFileBytes,
				KeyFileBytes:       to.Repository.TLS.KeyFileBytes,
			}
		}
	}

	return
}

type repositoryHeader struct {
	kv map[string]string
}

// Clone the http header
func (hd repositoryHeader) Clone() http.Header {
	cloned := http.Header{}
	for k, v := range hd.kv {
		cloned.Set(k, v)
	}

	return cloned
}

type appReleaseMetadata struct {
	DownloadedAt string                  `json:"downloadedAt"`
	CostTime     string                  `json:"costTime"`
	Release      *sfs.ReleaseEventMetaV1 `json:"release"`
}

// saveAppReleaseMetadata save the app's release metadata to the local file.
func saveAppReleaseMetadata(meta *appReleaseMetadata, metaFileName string) error {
	// reset the TLSBytes, it's sensitive data.
	meta.Release.Repository.TLS = nil

	content, err := jsoni.MarshalIndent(meta, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal app release metadata failed, err: %v", err)
	}

	metaFile, err := os.OpenFile(metaFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open the app metadata file failed, err: %v", err)
	}
	defer metaFile.Close()

	if _, err := metaFile.Write(content); err != nil {
		return fmt.Errorf("write app release metadata content to file failed, err: %v", err)
	}

	return nil
}
